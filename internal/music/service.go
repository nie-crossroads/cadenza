package music

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxQueryRunes = 80
	audioCacheTTL = 8 * time.Minute
	addBudget     = 90 * time.Second
	probeBudget   = 25 * time.Second
)

// Service 是三扇门共用的核心。
type Service struct {
	Source        Source
	Store         Store
	Blog          Blog
	AllowPrivate  bool
	MaxAudioBytes int64
	MaxCoverBytes int64
	SearchTimeout time.Duration
	FetchTimeout  time.Duration

	mediaOnce sync.Once
	media     *http.Client

	cacheMu   sync.Mutex
	cache     map[string]cachedAudio
	playable  map[string]time.Time
	qualities map[string]string
	tracks    map[string]Track

	relaysMu sync.Mutex
	relays   map[string]*audioRelay
}

type cachedAudio struct {
	url   string
	until time.Time
}

func (s *Service) Search(ctx context.Context, query string) ([]Track, error) {
	return s.search(ctx, query, "", 0, 0, false)
}

// SearchQuery 按来源和分页取结果。limit 小于 1 时表示这一页的默认条数由调用方决定，这里按 30 收口。
func (s *Service) SearchQuery(ctx context.Context, query, source string, offset, limit int) ([]Track, error) {
	return s.search(ctx, query, source, offset, limit, true)
}

func (s *Service) search(ctx context.Context, query, source string, offset, limit int, paged bool) ([]Track, error) {
	query = trimQuery(query)
	if query == "" {
		return []Track{}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, s.searchTimeout())
	defer cancel()
	var (
		items []Track
		err   error
	)
	if paged {
		if finder, ok := s.Source.(filteredSearcher); ok {
			items, err = finder.SearchFiltered(ctx, query, source, offset, limit)
		} else {
			items, err = s.Source.Search(ctx, query)
			if err == nil {
				items = sliceTracks(items, source, offset, limit)
			}
		}
	} else {
		items, err = s.Source.Search(ctx, query)
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		log.Printf("搜索失败")
		return nil, fail(CodeSearchFailed, statusOf(CodeSearchFailed))
	}
	if items == nil {
		items = []Track{}
	}
	for i := range items {
		items[i].Quality = normalizeQuality(items[i].Quality)
		items[i].Name = strings.TrimSpace(items[i].Name)
		items[i].Author = strings.TrimSpace(items[i].Author)
	}
	s.noteQuality(items)
	return items, nil
}

type filteredSearcher interface {
	SearchFiltered(ctx context.Context, query, source string, offset, limit int) ([]Track, error)
}

// Preview 从已经开始的整段下载里切出 Range，避免浏览器每次拖动都重新连接源站。
func (s *Service) Preview(ctx context.Context, sourceID, rangeHeader string) (int, http.Header, io.ReadCloser, error) {
	if strings.TrimSpace(sourceID) == "" {
		return 0, nil, nil, ErrNotFound
	}
	relay := s.relayFor(sourceID)
	total, ctype, err := relay.waitMeta(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, ErrNotFound) {
			return 0, nil, nil, err
		}
		var f *Fail
		if errors.As(err, &f) {
			return 0, nil, nil, f
		}
		log.Printf("试听失败")
		return 0, nil, nil, fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	h := make(http.Header)
	h.Set("Content-Type", ctype)
	h.Set("Accept-Ranges", "bytes")
	h.Set("Cache-Control", "private, no-store")
	if total == 0 {
		h.Set("Content-Length", "0")
		return http.StatusOK, h, io.NopCloser(strings.NewReader("")), nil
	}
	start, end, partial, bad := sliceRange(cleanRange(rangeHeader), total)
	if bad {
		h.Set("Content-Range", "bytes */"+strconv.FormatInt(total, 10))
		return http.StatusRequestedRangeNotSatisfiable, h, io.NopCloser(strings.NewReader("")), nil
	}
	h.Set("Content-Length", strconv.FormatInt(end-start, 10))
	status := http.StatusOK
	if partial {
		status = http.StatusPartialContent
		h.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end-1, total))
	}
	return status, h, &rangeBody{relay: relay, ctx: ctx, pos: start, end: end}, nil
}

func (s *Service) Add(ctx context.Context, sourceID, quality string) (id int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, addBudget)
	defer cancel()
	resolved, err := s.resolve(ctx, sourceID)
	if err != nil {
		return 0, err
	}
	track := resolved.Track
	track.Name = strings.TrimSpace(track.Name)
	track.Author = strings.TrimSpace(track.Author)
	if track.Name == "" || UTF16Len(track.Name) > maxNameRunes || UTF16Len(track.Author) > maxNameRunes {
		return 0, fail(CodeBlogValidation, statusOf(CodeBlogValidation))
	}
	if s.Store == nil || s.Blog == nil {
		log.Printf("添加失败：转存或博客未配置")
		if s.Store == nil {
			return 0, fail(CodeCOSFailed, statusOf(CodeCOSFailed))
		}
		return 0, fail(CodeUnknown, statusOf(CodeUnknown))
	}

	quality, ok := CanonicalQuality(quality)
	if !ok {
		return 0, fail(CodeBlogValidation, statusOf(CodeBlogValidation))
	}
	if loc, isLocator := s.Source.(AudioLocator); isLocator {
		raw, audioErr := loc.AudioURL(ctx, sourceID, quality)
		if audioErr != nil {
			if errors.Is(audioErr, ErrNotFound) || errors.Is(audioErr, context.Canceled) {
				return 0, audioErr
			}
			if errors.Is(audioErr, context.DeadlineExceeded) {
				return 0, fail(CodeTimeout, statusOf(CodeTimeout))
			}
			log.Printf("取曲目失败")
			return 0, fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
		}
		if _, err = parseRemoteURL(raw); err != nil {
			return 0, fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
		}
		resolved.Audio.URL = raw
	}
	audio, audioType, err := s.download(ctx, resolved.Audio.URL, s.maxAudio())
	if err != nil {
		return 0, err
	}
	ext, contentType := sniffAudio(audio, audioType)
	var cover []byte
	var coverType string
	if resolved.Cover != "" {
		cover, coverType, err = s.download(ctx, resolved.Cover, s.maxCover())
		if err != nil {
			var f *Fail
			if errors.As(err, &f) && f.Code == CodeTimeout {
				return 0, err
			}
			if errors.As(err, &f) && f.Code != CodeTooLarge {
				log.Printf("封面下载失败，继续写入歌曲")
			}
			cover = nil
			err = nil
		}
	}

	audioURL, audioKey, err := s.Store.Put(ctx, "m", ext, bytesReader(audio), int64(len(audio)), contentType)
	if err != nil {
		return 0, asFail(err, CodeCOSFailed)
	}
	defer func() {
		if err != nil && audioKey != "" {
			s.deleteQuiet(audioKey)
		}
	}()

	coverURL := ""
	coverKey := ""
	if len(cover) > 0 {
		cext, ctype := sniffCover(cover, coverType)
		coverURL, coverKey, err = s.Store.Put(ctx, "c", cext, bytesReader(cover), int64(len(cover)), ctype)
		if err != nil {
			return 0, asFail(err, CodeCOSFailed)
		}
		defer func() {
			if err != nil && coverKey != "" {
				s.deleteQuiet(coverKey)
			}
		}()
	}
	if len(audioURL) > 80 || len(coverURL) > 100 {
		return 0, fail(CodeBlogValidation, statusOf(CodeBlogValidation))
	}

	id, err = s.Blog.Save(ctx, track, audioURL, coverURL)
	if err != nil {
		return 0, asFail(err, CodeUnknown)
	}
	return id, nil
}

// Playable 只确认所选音质能否取出一小段音频，不下载整首，也不写入试听缓存。
func (s *Service) Playable(ctx context.Context, sourceID, quality string) (bool, error) {
	quality, ok := CanonicalQuality(quality)
	if !ok {
		return false, fail(CodeBlogValidation, statusOf(CodeBlogValidation))
	}
	sourceID = strings.TrimSpace(sourceID)
	if s.playableFresh(sourceID, quality) {
		return true, nil
	}
	// 试听只拉标准音质。这一首已经在播，标准档不必再问文件服务器。
	if quality == "std" && s.stdPreviewReady(sourceID) {
		s.markPlayable(sourceID, quality)
		return true, nil
	}
	ctx, cancel := context.WithTimeout(ctx, probeBudget)
	defer cancel()
	raw, err := s.audioForQuality(ctx, sourceID, quality)
	if err != nil {
		if errors.Is(err, context.Canceled) && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return false, err
		}
		return false, nil
	}
	ok, err = s.probePlayable(ctx, raw)
	if err != nil {
		return false, err
	}
	if ok {
		s.markPlayable(sourceID, quality)
	}
	return ok, nil
}

func (s *Service) audioForQuality(ctx context.Context, sourceID, quality string) (string, error) {
	if loc, ok := s.Source.(AudioLocator); ok {
		raw, err := loc.AudioURL(ctx, sourceID, quality)
		if err != nil {
			return "", err
		}
		if _, err = parseRemoteURL(raw); err != nil {
			return "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
		}
		return raw, nil
	}
	resolved, err := s.resolve(ctx, sourceID)
	if err != nil {
		return "", err
	}
	if _, err = parseRemoteURL(resolved.Audio.URL); err != nil {
		return "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	return resolved.Audio.URL, nil
}

func (s *Service) probePlayable(ctx context.Context, raw string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return false, nil
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Range", "bytes=0-31")
	if ref := mediaReferer(raw); ref != "" {
		req.Header.Set("Referer", ref)
	}
	resp, err := s.mediaClient().Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return false, context.Canceled
		}
		return false, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return false, nil
	}
	buf := make([]byte, 32)
	n, readErr := io.ReadAtLeast(io.LimitReader(resp.Body, 32), buf, 1)
	if n < 1 {
		if errors.Is(ctx.Err(), context.Canceled) && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return false, context.Canceled
		}
		if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, io.ErrUnexpectedEOF) {
			return false, nil
		}
		return false, nil
	}
	return playablePrefix(buf[:n], resp.Header.Get("Content-Type")), nil
}

func (s *Service) audioURL(ctx context.Context, sourceID string) (string, error) {
	if raw, ok := s.cached(sourceID); ok {
		return raw, nil
	}
	var raw string
	var err error
	if loc, ok := s.Source.(AudioLocator); ok {
		raw, err = loc.AudioURL(ctx, sourceID, "std")
	} else {
		var resolved Resolved
		resolved, err = s.Source.Resolve(ctx, sourceID)
		if err == nil {
			raw = resolved.Audio.URL
		}
	}
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, context.Canceled) {
			return "", err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return "", fail(CodeTimeout, statusOf(CodeTimeout))
		}
		log.Printf("取曲目失败")
		return "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	if _, err = parseRemoteURL(raw); err != nil {
		return "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	s.remember(sourceID, raw)
	return raw, nil
}

func (s *Service) resolve(ctx context.Context, sourceID string) (Resolved, error) {
	resolved, err := s.Source.Resolve(ctx, sourceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Resolved{}, err
		}
		if errors.Is(err, context.Canceled) {
			return Resolved{}, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return Resolved{}, fail(CodeTimeout, statusOf(CodeTimeout))
		}
		log.Printf("取曲目失败")
		return Resolved{}, fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	if _, err := parseRemoteURL(resolved.Audio.URL); err != nil {
		return Resolved{}, fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	resolved.Track = s.fillTrack(resolved.Track)
	if strings.TrimSpace(resolved.Track.Name) == "" {
		return Resolved{}, ErrNotFound
	}
	return resolved, nil
}

func (s *Service) download(ctx context.Context, raw string, max int64) ([]byte, string, error) {
	if _, err := parseRemoteURL(raw); err != nil {
		return nil, "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	ctx, cancel := context.WithTimeout(ctx, s.fetchTimeout())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	req.Header.Set("User-Agent", userAgent)
	if ref := mediaReferer(raw); ref != "" {
		req.Header.Set("Referer", ref)
	}
	resp, err := s.mediaClient().Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, "", fail(CodeTimeout, statusOf(CodeTimeout))
		}
		log.Printf("下载失败")
		return nil, "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("下载失败")
		return nil, "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	if resp.ContentLength > max {
		return nil, "", fail(CodeTooLarge, statusOf(CodeTooLarge))
	}
	buf, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, "", fail(CodeTimeout, statusOf(CodeTimeout))
		}
		log.Printf("下载失败")
		return nil, "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	if int64(len(buf)) > max {
		return nil, "", fail(CodeTooLarge, statusOf(CodeTooLarge))
	}
	if len(buf) == 0 {
		return nil, "", fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
	}
	return buf, resp.Header.Get("Content-Type"), nil
}

func (s *Service) mediaClient() *http.Client {
	s.mediaOnce.Do(func() {
		s.media = newMediaClient(s.AllowPrivate)
	})
	return s.media
}

func (s *Service) noteQuality(items []Track) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if s.qualities == nil {
		s.qualities = map[string]string{}
	}
	if s.tracks == nil {
		s.tracks = map[string]Track{}
	}
	if len(s.qualities) > 400 {
		s.qualities = map[string]string{}
		s.tracks = map[string]Track{}
	}
	for _, item := range items {
		if item.SourceID == "" {
			continue
		}
		s.qualities[item.SourceID] = normalizeQuality(item.Quality)
		if item.Name != "" {
			s.tracks[item.SourceID] = item
		}
	}
}

func (s *Service) qualityOf(sourceID string) string {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if q := s.qualities[sourceID]; q != "" {
		return q
	}
	return "hi"
}

func (s *Service) fillTrack(track Track) Track {
	s.cacheMu.Lock()
	remembered := s.tracks[track.SourceID]
	s.cacheMu.Unlock()
	if track.Name == "" {
		track.Name = remembered.Name
	}
	if track.Author == "" {
		track.Author = remembered.Author
	}
	if track.Cover == "" {
		track.Cover = remembered.Cover
	}
	if track.Duration == 0 {
		track.Duration = remembered.Duration
	}
	if track.Source == "" {
		track.Source = remembered.Source
	}
	if track.Quality == "" {
		track.Quality = remembered.Quality
	}
	if track.Quality == "" {
		track.Quality = "std"
	}
	return track
}

func (s *Service) stdPreviewReady(sourceID string) bool {
	s.relaysMu.Lock()
	relay := s.relays[sourceID]
	s.relaysMu.Unlock()
	if relay == nil {
		return false
	}
	return relay.hasPlayableAudio()
}

func playableKey(sourceID, quality string) string {
	return sourceID + "\n" + quality
}

func (s *Service) playableFresh(sourceID, quality string) bool {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	until, ok := s.playable[playableKey(sourceID, quality)]
	return ok && time.Now().Before(until)
}

func (s *Service) markPlayable(sourceID, quality string) {
	if sourceID == "" || quality == "" {
		return
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if s.playable == nil {
		s.playable = map[string]time.Time{}
	}
	if len(s.playable) > 400 {
		s.playable = map[string]time.Time{}
	}
	s.playable[playableKey(sourceID, quality)] = time.Now().Add(audioCacheTTL)
}

func (s *Service) cached(id string) (string, bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	item, ok := s.cache[id]
	if !ok || time.Now().After(item.until) || item.url == "" {
		return "", false
	}
	return item.url, true
}

func (s *Service) remember(id, raw string) {
	if id == "" || raw == "" {
		return
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if s.cache == nil {
		s.cache = map[string]cachedAudio{}
	}
	s.cache[id] = cachedAudio{url: raw, until: time.Now().Add(audioCacheTTL)}
}

func (s *Service) dropCache(id string) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	delete(s.cache, id)
}

func (s *Service) deleteQuiet(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Store.Delete(ctx, key); err != nil {
		log.Printf("清理转存文件失败")
	}
}

func (s *Service) searchTimeout() time.Duration {
	if s.SearchTimeout <= 0 {
		return 12 * time.Second
	}
	return s.SearchTimeout
}

func (s *Service) fetchTimeout() time.Duration {
	if s.FetchTimeout <= 0 {
		return 45 * time.Second
	}
	return s.FetchTimeout
}

func (s *Service) maxAudio() int64 {
	if s.MaxAudioBytes <= 0 {
		return 40 << 20
	}
	return s.MaxAudioBytes
}

func (s *Service) maxCover() int64 {
	if s.MaxCoverBytes <= 0 {
		return 2 << 20
	}
	return s.MaxCoverBytes
}

func trimQuery(q string) string {
	q = strings.TrimSpace(q)
	runes := []rune(q)
	if len(runes) > maxQueryRunes {
		q = string(runes[:maxQueryRunes])
	}
	return strings.TrimSpace(q)
}

func cleanRange(h string) string {
	h = strings.TrimSpace(h)
	if len(h) == 0 || len(h) > 64 || strings.ContainsAny(h, "\r\n") {
		return ""
	}
	if !strings.HasPrefix(strings.ToLower(h), "bytes=") {
		return ""
	}
	return h
}

func mediaReferer(raw string) string {
	u, err := parseRemoteURL(raw)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if host == "music.163.com" || strings.HasSuffix(host, ".music.163.com") || host == "126.net" || strings.HasSuffix(host, ".126.net") {
		return "https://music.163.com/"
	}
	switch {
	case strings.Contains(host, "kuwo"):
		return "https://www.kuwo.cn/"
	case strings.Contains(host, "kugou"):
		return "https://www.kugou.com/"
	case strings.Contains(host, "qqmusic") || strings.HasSuffix(host, ".qq.com"):
		return "https://y.qq.com/"
	case strings.Contains(host, "migu"):
		return "https://m.music.migu.cn/"
	case strings.Contains(host, "1ting"):
		return "https://www.1ting.com/"
	case strings.Contains(host, "qianqian") || strings.Contains(host, "taihe") || strings.Contains(host, "91q"):
		return "https://music.91q.com/"
	}
	return ""
}

func asFail(err error, fallback string) error {
	var f *Fail
	if errors.As(err, &f) {
		return f
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fail(CodeTimeout, statusOf(CodeTimeout))
	}
	return fail(fallback, statusOf(fallback))
}

func bytesReader(b []byte) io.Reader {
	return bytes.NewReader(b)
}

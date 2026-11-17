package music

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	neteaseBase   = "https://music.163.com"
	neteasePrefix = "wy:"
	userAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// ErrNotFound 表示 sourceId 不存在或格式不对。
var ErrNotFound = errors.New("not found")

// Netease 直接请求网易接口。临时播放地址只留在 Resolve 的结果里。
type Netease struct {
	base   string
	cookie string
	client *http.Client
}

func NewNetease(cookie string) *Netease {
	return newNetease(neteaseBase, cookie, nil)
}

func newNetease(base, cookie string, client *http.Client) *Netease {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	return &Netease{base: base, cookie: strings.TrimSpace(cookie), client: client}
}

func (n *Netease) Label() string { return sourceNetease }

func (n *Netease) Search(ctx context.Context, query string) ([]Track, error) {
	return n.SearchPage(ctx, query, 0, 30)
}

func (n *Netease) SearchPage(ctx context.Context, query string, offset, limit int) ([]Track, error) {
	offset, limit = clampPage(offset, limit)
	form := url.Values{}
	form.Set("s", query)
	form.Set("type", "1")
	form.Set("limit", strconv.Itoa(limit))
	form.Set("offset", strconv.Itoa(offset))
	body, err := n.post(ctx, "/api/cloudsearch/pc", form)
	if err != nil {
		return nil, err
	}
	songs, err := decodeSongs(body)
	if err != nil {
		return nil, err
	}
	out := make([]Track, 0, len(songs))
	seen := map[string]struct{}{}
	for _, song := range songs {
		track, ok := song.track(n.cookie != "")
		if !ok {
			continue
		}
		if _, dup := seen[track.SourceID]; dup {
			continue
		}
		seen[track.SourceID] = struct{}{}
		out = append(out, track)
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func (n *Netease) AudioURL(ctx context.Context, sourceID, quality string) (string, error) {
	id, err := parseNeteaseID(sourceID)
	if err != nil {
		return "", ErrNotFound
	}
	if quality == "" {
		quality = "hi"
	}
	raw, _, _, err := n.audioURL(ctx, id, quality)
	return raw, err
}

func (n *Netease) Resolve(ctx context.Context, sourceID string) (Resolved, error) {
	id, err := parseNeteaseID(sourceID)
	if err != nil {
		return Resolved{}, ErrNotFound
	}
	detail, err := n.detail(ctx, id)
	if err != nil {
		return Resolved{}, err
	}
	track, ok := detail.track(n.cookie != "")
	if !ok {
		return Resolved{}, ErrNotFound
	}
	audio, level, br, err := n.audioURL(ctx, id, track.Quality)
	if err != nil {
		return Resolved{}, err
	}
	track.Quality = qualityFromFile(level, br, track.Quality)
	return Resolved{Track: track, Audio: Remote{URL: audio}, Cover: track.Cover}, nil
}

func (n *Netease) detail(ctx context.Context, id int64) (neSong, error) {
	q := url.Values{}
	q.Set("ids", fmt.Sprintf("[%d]", id))
	body, err := n.get(ctx, "/api/song/detail/?"+q.Encode())
	if err != nil {
		return neSong{}, err
	}
	var wrap struct {
		Code  int      `json:"code"`
		Songs []neSong `json:"songs"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return neSong{}, err
	}
	if wrap.Code != 200 || len(wrap.Songs) == 0 || wrap.Songs[0].ID != id {
		return neSong{}, ErrNotFound
	}
	return wrap.Songs[0], nil
}

func (n *Netease) audioURL(ctx context.Context, id int64, quality string) (string, string, int, error) {
	for _, br := range bitrateSteps(quality) {
		q := url.Values{}
		q.Set("ids", fmt.Sprintf("[%d]", id))
		q.Set("br", strconv.Itoa(br))
		body, err := n.get(ctx, "/api/song/enhance/player/url?"+q.Encode())
		if err != nil {
			return "", "", 0, err
		}
		var wrap struct {
			Code int `json:"code"`
			Data []struct {
				URL   string `json:"url"`
				BR    int    `json:"br"`
				Level string `json:"level"`
				Code  int    `json:"code"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &wrap); err != nil {
			return "", "", 0, err
		}
		if wrap.Code != 200 || len(wrap.Data) == 0 {
			continue
		}
		item := wrap.Data[0]
		if item.URL == "" || item.Code != 200 {
			continue
		}
		if _, err := parseRemoteURL(item.URL); err != nil {
			continue
		}
		return item.URL, item.Level, item.BR, nil
	}
	return "", "", 0, ErrNotFound
}

func bitrateSteps(quality string) []int {
	switch normalizeQuality(quality) {
	case "hr":
		return []int{1_999_000}
	case "hi":
		return []int{320_000}
	default:
		return []int{128_000}
	}
}

func qualityFromFile(level string, br int, fallback string) string {
	got := QualityOf(level, level, br, true)
	if got == "std" && br == 0 {
		return normalizeQuality(fallback)
	}
	return got
}

func (n *Netease) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, n.base+path, nil)
	if err != nil {
		return nil, err
	}
	n.decorate(req)
	return n.do(req)
}

func (n *Netease) post(ctx context.Context, path string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.base+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	n.decorate(req)
	return n.do(req)
}

func (n *Netease) decorate(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", n.base+"/")
	if n.cookie != "" {
		req.Header.Set("Cookie", n.cookie)
		return
	}
	req.Header.Set("Cookie", "os=pc")
}

func (n *Netease) do(req *http.Request) ([]byte, error) {
	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("source status")
	}
	return body, nil
}

func parseNeteaseID(sourceID string) (int64, error) {
	if !strings.HasPrefix(sourceID, neteasePrefix) {
		return 0, ErrNotFound
	}
	raw := sourceID[len(neteasePrefix):]
	if raw == "" || len(raw) > 20 || strings.Trim(raw, "0123456789") != "" {
		return 0, ErrNotFound
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrNotFound
	}
	return id, nil
}

func SourceID(id int64) string {
	return neteasePrefix + strconv.FormatInt(id, 10)
}

type neSong struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	DT        int64    `json:"dt"`
	Duration  int64    `json:"duration"`
	AR        []neName `json:"ar"`
	Artists   []neName `json:"artists"`
	AL        neAlbum  `json:"al"`
	Album     neAlbum  `json:"album"`
	Privilege struct {
		Maxbr      int    `json:"maxbr"`
		MaxBrLevel string `json:"maxBrLevel"`
		PlLevel    string `json:"plLevel"`
	} `json:"privilege"`
}

type neName struct {
	Name string `json:"name"`
}

type neAlbum struct {
	PicURL string `json:"picUrl"`
}

func (s neSong) track(preferMax bool) (Track, bool) {
	name := strings.TrimSpace(s.Name)
	if s.ID <= 0 || name == "" {
		return Track{}, false
	}
	people := s.AR
	if len(people) == 0 {
		people = s.Artists
	}
	names := make([]string, 0, len(people))
	for _, p := range people {
		names = append(names, p.Name)
	}
	ms := s.DT
	if ms <= 0 {
		ms = s.Duration
	}
	cover := s.AL.PicURL
	if cover == "" {
		cover = s.Album.PicURL
	}
	return Track{
		Name:     name,
		Author:   ChooseAuthor(names),
		Cover:    cover,
		Duration: Seconds(ms),
		Source:   sourceNetease,
		SourceID: SourceID(s.ID),
		Quality:  QualityOf(s.Privilege.PlLevel, s.Privilege.MaxBrLevel, s.Privilege.Maxbr, preferMax),
	}, true
}

func decodeSongs(body []byte) ([]neSong, error) {
	var wrap struct {
		Code   int `json:"code"`
		Result struct {
			Songs []neSong `json:"songs"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, err
	}
	if wrap.Code != 200 {
		return nil, errors.New("source status")
	}
	return wrap.Result.Songs, nil
}

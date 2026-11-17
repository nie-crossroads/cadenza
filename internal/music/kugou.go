package music

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const kugouPrefix = "kg:"

type Kugou struct {
	client *http.Client
}

func newKugou(client *http.Client) *Kugou {
	if client == nil {
		client = sourceClient()
	}
	return &Kugou{client: client}
}

func (k *Kugou) Label() string { return sourceKugou }

func (k *Kugou) Search(ctx context.Context, query string) ([]Track, error) {
	return k.SearchPage(ctx, query, 0, 30)
}

func (k *Kugou) SearchPage(ctx context.Context, query string, offset, limit int) ([]Track, error) {
	page, size, skip := upstreamWindow(offset, limit)
	u := "http://mobilecdn.kugou.com/api/v3/search/song?" + url.Values{
		"format":   {"json"},
		"keyword":  {query},
		"page":     {strconv.Itoa(page)},
		"pagesize": {strconv.Itoa(size)},
		"showtype": {"1"},
	}.Encode()
	body, resp, err := fetch(ctx, k.client, "GET", u, "https://www.kugou.com/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return nil, errSource
	}
	var wrap struct {
		Status int `json:"status"`
		Data   struct {
			Info []kgSong `json:"info"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil || wrap.Status != 1 {
		return nil, errSource
	}
	out := make([]Track, 0, len(wrap.Data.Info))
	seen := map[string]struct{}{}
	for _, song := range wrap.Data.Info {
		track, ok := song.track()
		if !ok {
			continue
		}
		if _, dup := seen[track.SourceID]; dup {
			continue
		}
		seen[track.SourceID] = struct{}{}
		out = append(out, track)
	}
	return takeWindow(out, skip, limit), nil
}

func (k *Kugou) AudioURL(ctx context.Context, sourceID, _ string) (string, error) {
	hash, albumAudio, err := parseKugouID(sourceID)
	if err != nil {
		return "", err
	}
	if raw := k.tracker(ctx, hash, albumAudio); raw != "" {
		return raw, nil
	}
	if raw := k.playInfo(ctx, hash); raw != "" {
		return raw, nil
	}
	return "", ErrNotFound
}

func (k *Kugou) Resolve(ctx context.Context, sourceID string) (Resolved, error) {
	hash, albumAudio, err := parseKugouID(sourceID)
	if err != nil {
		return Resolved{}, ErrNotFound
	}
	audio, err := k.AudioURL(ctx, sourceID, "std")
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{
		Track: Track{Source: sourceKugou, SourceID: kugouID(hash, albumAudio), Quality: "std"},
		Audio: Remote{URL: audio},
	}, nil
}

func (k *Kugou) tracker(ctx context.Context, hash, albumAudio string) string {
	mid := "123456789012345678901234567890"
	key := md5hex(strings.ToUpper(hash) + "57ae12eb6890223e355ccfcb74edf70d" + "1005" + mid + "0")
	u := "https://trackercdn.kugou.com/v5/url?" + url.Values{
		"appid": {"1005"}, "hash": {strings.ToUpper(hash)}, "key": {key},
		"userid": {"0"}, "pid": {"2"}, "behavior": {"play"}, "mid": {mid},
		"album_audio_id": {albumAudio},
	}.Encode()
	body, resp, err := fetch(ctx, k.client, "GET", u, "https://www.kugou.com/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	return firstHTTP(body)
}

func (k *Kugou) playInfo(ctx context.Context, hash string) string {
	u := "http://m.kugou.com/app/i/getSongInfo.php?" + url.Values{
		"cmd": {"playInfo"}, "hash": {strings.ToLower(hash)},
	}.Encode()
	body, resp, err := fetch(ctx, k.client, "GET", u, "https://m.kugou.com/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	var wrap struct {
		URL string `json:"url"`
	}
	if json.Unmarshal(body, &wrap) != nil {
		return ""
	}
	if _, err := parseRemoteURL(wrap.URL); err != nil {
		return ""
	}
	return wrap.URL
}

type kgSong struct {
	Songname     string `json:"songname"`
	Singername   string `json:"singername"`
	Duration     int    `json:"duration"`
	Hash         string `json:"hash"`
	AlbumAudioID int64  `json:"album_audio_id"`
	SQFilesize   int    `json:"sqfilesize"`
	HQFilesize   int    `json:"320filesize"`
}

func (s kgSong) track() (Track, bool) {
	name := strings.TrimSpace(s.Songname)
	hash := strings.ToLower(strings.TrimSpace(s.Hash))
	if name == "" || hash == "" || s.AlbumAudioID <= 0 {
		return Track{}, false
	}
	quality := "std"
	switch {
	case s.SQFilesize > 20_000_000:
		quality = "hr"
	case s.HQFilesize > 0:
		quality = "hi"
	}
	return Track{
		Name: name, Author: ChooseAuthor(strings.Split(s.Singername, "、")),
		Duration: s.Duration, Source: sourceKugou,
		SourceID: kugouID(hash, strconv.FormatInt(s.AlbumAudioID, 10)), Quality: quality,
	}, true
}

func kugouID(hash, albumAudio string) string {
	return kugouPrefix + strings.ToLower(hash) + ":" + albumAudio
}

func parseKugouID(sourceID string) (string, string, error) {
	id, err := plainID(sourceID, kugouPrefix)
	if err != nil {
		return "", "", err
	}
	hash, album, ok := strings.Cut(id, ":")
	if !ok || len(hash) != 32 || album == "" {
		return "", "", ErrNotFound
	}
	return hash, album, nil
}

func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func firstHTTP(body []byte) string {
	var wrap map[string]any
	if json.Unmarshal(body, &wrap) != nil {
		return ""
	}
	if raw := digAudio(wrap); raw != "" {
		if _, err := parseRemoteURL(raw); err == nil {
			return raw
		}
	}
	return ""
}

func digAudio(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for _, key := range []string{"url", "play_url", "backup_url"} {
			if s := digAudio(t[key]); s != "" {
				return s
			}
		}
	case []any:
		for _, item := range t {
			if s := digAudio(item); s != "" {
				return s
			}
		}
	case string:
		if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
			return t
		}
	}
	return ""
}

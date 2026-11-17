package music

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	qianqianPrefix = "bd:"
	qianqianApp    = "16073360"
	qianqianSecret = "0b50b02fd0d73a9c4c8c3a781c30845f"
)

type Qianqian struct {
	client *http.Client
}

func newQianqian(client *http.Client) *Qianqian {
	if client == nil {
		client = sourceClient()
	}
	return &Qianqian{client: client}
}

func (q *Qianqian) Label() string { return sourceQianqian }

func (q *Qianqian) Search(ctx context.Context, query string) ([]Track, error) {
	return q.SearchPage(ctx, query, 0, 30)
}

func (q *Qianqian) SearchPage(ctx context.Context, query string, offset, limit int) ([]Track, error) {
	page, size, skip := upstreamWindow(offset, limit)
	body, err := q.signed(ctx, "/v1/search", map[string]string{
		"word": query, "type": "1", "pageNo": strconv.Itoa(page), "pageSize": strconv.Itoa(size),
	})
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Data struct {
			TypeTrack []qqianSong `json:"typeTrack"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &wrap) != nil {
		return nil, errSource
	}
	out := make([]Track, 0, len(wrap.Data.TypeTrack))
	seen := map[string]struct{}{}
	for _, song := range wrap.Data.TypeTrack {
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

func (q *Qianqian) AudioURL(ctx context.Context, sourceID, _ string) (string, error) {
	id, err := plainID(sourceID, qianqianPrefix)
	if err != nil {
		return "", err
	}
	body, err := q.signed(ctx, "/v1/song/tracklink", map[string]string{"TSID": id})
	if err != nil {
		return "", err
	}
	var wrap map[string]any
	if json.Unmarshal(body, &wrap) != nil {
		return "", errSource
	}
	raw := digFile(wrap["data"])
	if _, err := parseRemoteURL(raw); err != nil {
		return "", ErrNotFound
	}
	return raw, nil
}

func (q *Qianqian) Resolve(ctx context.Context, sourceID string) (Resolved, error) {
	audio, err := q.AudioURL(ctx, sourceID, "std")
	if err != nil {
		return Resolved{}, err
	}
	id, _ := plainID(sourceID, qianqianPrefix)
	return Resolved{
		Track: Track{Source: sourceQianqian, SourceID: qianqianPrefix + id, Quality: "std"},
		Audio: Remote{URL: audio},
	}, nil
}

func (q *Qianqian) signed(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	params["appid"] = qianqianApp
	params["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	params["sign"] = qianqianSign(params)
	u := "https://music.91q.com" + path + "?" + encodeValues(params)
	body, resp, err := fetch(ctx, q.client, "GET", u, "https://music.91q.com/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return nil, errSource
	}
	return body, nil
}

func qianqianSign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "sign" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, key := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(params[key])
	}
	b.WriteString(qianqianSecret)
	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func encodeValues(params map[string]string) string {
	v := url.Values{}
	for key, val := range params {
		v.Set(key, val)
	}
	return v.Encode()
}

type qqianSong struct {
	Title    string `json:"title"`
	TSID     string `json:"TSID"`
	Duration int    `json:"duration"`
	Pic      string `json:"pic"`
	Artist   []struct {
		Name string `json:"name"`
	} `json:"artist"`
	Rate128 struct {
		Path string `json:"path"`
	} `json:"eq_rate_128"`
	Rate320 struct {
		Path string `json:"path"`
	} `json:"eq_rate_320"`
	RateFlac struct {
		Path string `json:"path"`
	} `json:"eq_rate_flac"`
}

func (s qqianSong) track() (Track, bool) {
	name := strings.TrimSpace(s.Title)
	id := strings.TrimSpace(s.TSID)
	if name == "" || id == "" {
		return Track{}, false
	}
	names := make([]string, 0, len(s.Artist))
	for _, artist := range s.Artist {
		names = append(names, artist.Name)
	}
	quality := "std"
	switch {
	case s.RateFlac.Path != "":
		quality = "hr"
	case s.Rate320.Path != "":
		quality = "hi"
	}
	return Track{
		Name: name, Author: ChooseAuthor(names), Cover: cleanCover(s.Pic),
		Duration: secondsOf(s.Duration), Source: sourceQianqian,
		SourceID: qianqianPrefix + id, Quality: quality,
	}, true
}

func digFile(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for key, child := range t {
			low := strings.ToLower(key)
			if strings.Contains(low, "lyric") || strings.Contains(low, "pic") || strings.Contains(low, "img") {
				continue
			}
			if s, ok := child.(string); ok {
				if raw := fileURL(s); raw != "" {
					return raw
				}
			}
			if raw := digFile(child); raw != "" {
				return raw
			}
		}
	case []any:
		for _, child := range t {
			if raw := digFile(child); raw != "" {
				return raw
			}
		}
	}
	return ""
}

func fileURL(s string) string {
	s = strings.TrimSpace(s)
	low := strings.ToLower(s)
	if !(strings.Contains(low, ".mp3") || strings.Contains(low, ".m4a") || strings.Contains(low, ".flac") || strings.Contains(low, ".aac")) {
		return ""
	}
	if strings.HasPrefix(s, "//") {
		s = "https:" + s
	}
	if strings.HasPrefix(s, "/") {
		s = "https://music.91q.com" + s
	}
	if _, err := parseRemoteURL(s); err != nil {
		return ""
	}
	return s
}

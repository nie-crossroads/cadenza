package music

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const kuwoPrefix = "kw:"

type Kuwo struct {
	client *http.Client
}

func newKuwo(client *http.Client) *Kuwo {
	if client == nil {
		client = sourceClient()
	}
	return &Kuwo{client: client}
}

func (k *Kuwo) Label() string { return sourceKuwo }

func (k *Kuwo) Search(ctx context.Context, query string) ([]Track, error) {
	return k.SearchPage(ctx, query, 0, 30)
}

func (k *Kuwo) SearchPage(ctx context.Context, query string, offset, limit int) ([]Track, error) {
	page, size, skip := upstreamWindow(offset, limit)
	u := "https://search.kuwo.cn/r.s?" + url.Values{
		"all": {query}, "ft": {"music"}, "itemset": {"web_2013"}, "client": {"kt"},
		"pn": {strconv.Itoa(page - 1)}, "rn": {strconv.Itoa(size)},
		"rformat": {"json"}, "encoding": {"utf8"},
	}.Encode()
	body, resp, err := fetch(ctx, k.client, "GET", u, "https://www.kuwo.cn/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return nil, errSource
	}
	parsed, err := looseJSON(body)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		AbsList []kwSong `json:"abslist"`
	}
	if err := json.Unmarshal(parsed, &wrap); err != nil {
		return nil, errSource
	}
	out := make([]Track, 0, len(wrap.AbsList))
	seen := map[string]struct{}{}
	for _, song := range wrap.AbsList {
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

func (k *Kuwo) AudioURL(ctx context.Context, sourceID, _ string) (string, error) {
	id, err := plainID(sourceID, kuwoPrefix)
	if err != nil || strings.Trim(id, "0123456789") != "" {
		return "", ErrNotFound
	}
	u := "http://antiserver.kuwo.cn/anti.s?" + url.Values{
		"type": {"convert_url"}, "rid": {"MUSIC_" + id}, "format": {"mp3"}, "response": {"url"},
	}.Encode()
	body, resp, err := fetch(ctx, k.client, "GET", u, "https://www.kuwo.cn/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return "", errSource
	}
	raw := strings.TrimSpace(string(body))
	if _, err := parseRemoteURL(raw); err != nil {
		return "", ErrNotFound
	}
	return raw, nil
}

func (k *Kuwo) Resolve(ctx context.Context, sourceID string) (Resolved, error) {
	id, err := plainID(sourceID, kuwoPrefix)
	if err != nil {
		return Resolved{}, ErrNotFound
	}
	audio, err := k.AudioURL(ctx, sourceID, "std")
	if err != nil {
		return Resolved{}, err
	}
	track := k.detail(ctx, id)
	track.Source = sourceKuwo
	track.SourceID = kuwoPrefix + id
	track.Quality = normalizeQuality(track.Quality)
	if track.Name == "" {
		return Resolved{}, ErrNotFound
	}
	return Resolved{Track: track, Audio: Remote{URL: audio}, Cover: track.Cover}, nil
}

func (k *Kuwo) detail(ctx context.Context, id string) Track {
	u := "http://m.kuwo.cn/newh5/singles/songinfoandlrc?musicId=" + url.QueryEscape(id)
	body, resp, err := fetch(ctx, k.client, "GET", u, "http://m.kuwo.cn/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return Track{}
	}
	var wrap struct {
		Data struct {
			SongName string `json:"songName"`
			Artist   string `json:"artist"`
			Pic      string `json:"pic"`
			Duration int    `json:"duration"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &wrap) != nil {
		return Track{}
	}
	return Track{
		Name: strings.TrimSpace(wrap.Data.SongName), Author: ChooseAuthor([]string{wrap.Data.Artist}),
		Cover: cleanCover(wrap.Data.Pic), Duration: secondsOf(wrap.Data.Duration), Quality: "std",
	}
}

type kwSong struct {
	Name     string `json:"SONGNAME"`
	NameAlt  string `json:"NAME"`
	Artist   string `json:"ARTIST"`
	Duration string `json:"DURATION"`
	MusicRID string `json:"MUSICRID"`
	Pic      string `json:"PICPATH"`
}

func (s kwSong) track() (Track, bool) {
	rid := strings.TrimPrefix(strings.TrimSpace(s.MusicRID), "MUSIC_")
	if rid == "" || strings.Trim(rid, "0123456789") != "" {
		return Track{}, false
	}
	name := cleanText(s.Name)
	if name == "" {
		name = cleanText(s.NameAlt)
	}
	if name == "" {
		return Track{}, false
	}
	sec, _ := strconv.Atoi(strings.TrimSpace(s.Duration))
	return Track{
		Name: name, Author: ChooseAuthor([]string{cleanText(s.Artist)}),
		Cover: kuwoCover(s.Pic), Duration: secondsOf(sec),
		Source: sourceKuwo, SourceID: kuwoPrefix + rid, Quality: "std",
	}, true
}

func looseJSON(raw []byte) ([]byte, error) {
	text := strings.TrimSpace(string(raw))
	if json.Valid([]byte(text)) {
		return []byte(text), nil
	}
	var b strings.Builder
	in := false
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '\\' && in && i+1 < len(text) {
			b.WriteByte(c)
			i++
			b.WriteByte(text[i])
			continue
		}
		if c == '\'' {
			b.WriteByte('"')
			in = !in
			continue
		}
		if c == '"' && in {
			b.WriteString(`\"`)
			continue
		}
		b.WriteByte(c)
	}
	out := []byte(b.String())
	if !json.Valid(out) {
		return nil, errSource
	}
	return out, nil
}

func cleanText(s string) string {
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return strings.TrimSpace(s)
}

func kuwoCover(path string) string {
	path = cleanText(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return "https://img2.kuwo.cn/star/albumcover/" + strings.TrimPrefix(path, "/")
}

func secondsOf(n int) int {
	if n <= 0 {
		return 0
	}
	if n > 10000 {
		return Seconds(int64(n))
	}
	return n
}

func cleanCover(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	if _, err := parseRemoteURL(raw); err != nil {
		return ""
	}
	return raw
}

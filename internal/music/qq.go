package music

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const qqPrefix = "qq:"

type QQ struct {
	client *http.Client
}

func newQQ(client *http.Client) *QQ {
	if client == nil {
		client = sourceClient()
	}
	return &QQ{client: client}
}

func (q *QQ) Label() string { return sourceQQ }

func (q *QQ) Search(ctx context.Context, query string) ([]Track, error) {
	return q.SearchPage(ctx, query, 0, 30)
}

func (q *QQ) SearchPage(ctx context.Context, query string, offset, limit int) ([]Track, error) {
	page, size, skip := upstreamWindow(offset, limit)
	u := "https://c.y.qq.com/soso/fcgi-bin/client_search_cp?" + url.Values{
		"w":      {query},
		"format": {"json"},
		"p":      {strconv.Itoa(page)},
		"n":      {strconv.Itoa(size)},
		"cr":     {"1"},
	}.Encode()
	body, resp, err := fetch(ctx, q.client, "GET", u, "https://y.qq.com/", nil, "")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errSource
	}
	var wrap struct {
		Code int `json:"code"`
		Data struct {
			Song struct {
				List []qqSong `json:"list"`
			} `json:"song"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil || wrap.Code != 0 {
		return nil, errSource
	}
	out := make([]Track, 0, len(wrap.Data.Song.List))
	seen := map[string]struct{}{}
	for _, song := range wrap.Data.Song.List {
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

func (q *QQ) AudioURL(ctx context.Context, sourceID, quality string) (string, error) {
	songmid, media, err := parseQQID(sourceID)
	if err != nil {
		return "", err
	}
	for _, name := range qqFilenames(media, quality) {
		raw, err := q.vkey(ctx, songmid, name)
		if err == nil && raw != "" {
			return raw, nil
		}
	}
	return "", ErrNotFound
}

func (q *QQ) Resolve(ctx context.Context, sourceID string) (Resolved, error) {
	songmid, media, err := parseQQID(sourceID)
	if err != nil {
		return Resolved{}, ErrNotFound
	}
	audio, err := q.AudioURL(ctx, sourceID, "hi")
	if err != nil {
		return Resolved{}, err
	}
	track := Track{Source: sourceQQ, SourceID: qqID(songmid, media), Quality: "std"}
	return Resolved{Track: track, Audio: Remote{URL: audio}}, nil
}

func (q *QQ) vkey(ctx context.Context, songmid, filename string) (string, error) {
	payload := map[string]any{
		"comm": map[string]any{
			"cv": 4747474, "ct": 24, "format": "json",
			"inCharset": "utf-8", "outCharset": "utf-8",
			"platform": "yqq.json", "needNewCode": 1, "uin": 0, "g_tk": 5381,
		},
		"req_1": map[string]any{
			"module": "vkey.GetVkeyServer",
			"method": "CgiGetVkey",
			"param": map[string]any{
				"guid": "10000", "songmid": []string{songmid}, "filename": []string{filename},
				"songtype": []int{0}, "uin": "0", "loginflag": 1, "platform": "20",
			},
		},
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	u := "https://u6.y.qq.com/cgi-bin/musicu.fcg?" + url.Values{
		"format": {"json"}, "platform": {"yqq.json"}, "needNewCode": {"0"}, "data": {string(buf)},
	}.Encode()
	body, resp, err := fetch(ctx, q.client, "GET", u, "https://y.qq.com/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return "", errSource
	}
	var wrap struct {
		Code int `json:"code"`
		Req  struct {
			Data struct {
				Sip        []string `json:"sip"`
				MidURLInfo []struct {
					Purl string `json:"purl"`
				} `json:"midurlinfo"`
			} `json:"data"`
		} `json:"req_1"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil || wrap.Code != 0 {
		return "", errSource
	}
	if len(wrap.Req.Data.MidURLInfo) == 0 || wrap.Req.Data.MidURLInfo[0].Purl == "" {
		return "", ErrNotFound
	}
	host := "https://ws.stream.qqmusic.qq.com/"
	for _, sip := range wrap.Req.Data.Sip {
		if strings.HasPrefix(sip, "https://") {
			host = sip
			break
		}
		if strings.HasPrefix(sip, "http://") && host == "https://ws.stream.qqmusic.qq.com/" {
			host = sip
		}
	}
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	raw := host + strings.TrimPrefix(wrap.Req.Data.MidURLInfo[0].Purl, "/")
	if _, err := parseRemoteURL(raw); err != nil {
		return "", ErrNotFound
	}
	return raw, nil
}

type qqSong struct {
	Songname string `json:"songname"`
	Songmid  string `json:"songmid"`
	MediaMid string `json:"media_mid"`
	Interval int    `json:"interval"`
	Size320  int    `json:"size320"`
	Sizeflac int    `json:"sizeflac"`
	Albummid string `json:"albummid"`
	Singer   []struct {
		Name string `json:"name"`
	} `json:"singer"`
}

func (s qqSong) track() (Track, bool) {
	name := strings.TrimSpace(s.Songname)
	media := strings.TrimSpace(s.MediaMid)
	if media == "" {
		media = strings.TrimSpace(s.Songmid)
	}
	if name == "" || s.Songmid == "" || media == "" {
		return Track{}, false
	}
	names := make([]string, 0, len(s.Singer))
	for _, singer := range s.Singer {
		names = append(names, singer.Name)
	}
	quality := "std"
	switch {
	case s.Sizeflac > 0:
		quality = "hr"
	case s.Size320 > 0:
		quality = "hi"
	}
	cover := ""
	if s.Albummid != "" {
		cover = "https://y.gtimg.cn/music/photo_new/T002R300x300M000" + s.Albummid + ".jpg"
	}
	return Track{
		Name: name, Author: ChooseAuthor(names), Cover: cover,
		Duration: s.Interval, Source: sourceQQ, SourceID: qqID(s.Songmid, media), Quality: quality,
	}, true
}

func qqID(songmid, media string) string {
	return qqPrefix + songmid + ":" + media
}

func parseQQID(sourceID string) (string, string, error) {
	id, err := plainID(sourceID, qqPrefix)
	if err != nil {
		return "", "", err
	}
	songmid, media, ok := strings.Cut(id, ":")
	if !ok || songmid == "" || media == "" {
		return "", "", ErrNotFound
	}
	return songmid, media, nil
}

func qqFilenames(media, quality string) []string {
	switch normalizeQuality(quality) {
	case "hr":
		return []string{"F000" + media + ".flac"}
	case "hi":
		return []string{"M800" + media + ".mp3"}
	default:
		return []string{"M500" + media + ".mp3"}
	}
}

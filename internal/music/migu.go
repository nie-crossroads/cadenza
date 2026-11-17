package music

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const miguPrefix = "mg:"

type Migu struct {
	client   *http.Client
	redirect *http.Client
}

func newMigu(client *http.Client) *Migu {
	if client == nil {
		client = sourceClient()
	}
	return &Migu{client: client, redirect: &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}}
}

func (m *Migu) Label() string { return sourceMigu }

func (m *Migu) Search(ctx context.Context, query string) ([]Track, error) {
	return m.SearchPage(ctx, query, 0, 30)
}

func (m *Migu) SearchPage(ctx context.Context, query string, offset, limit int) ([]Track, error) {
	page, size, skip := upstreamWindow(offset, limit)
	u := "https://app.c.nf.migu.cn/MIGUM3.0/v1.0/content/search_all.do?" + url.Values{
		"text": {query}, "pageNo": {strconv.Itoa(page)}, "pageSize": {strconv.Itoa(size)},
		"searchSwitch": {`{"song":1}`},
	}.Encode()
	body, resp, err := fetch(ctx, m.client, "GET", u, "https://m.music.migu.cn/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return nil, errSource
	}
	var wrap struct {
		Code string `json:"code"`
		Song struct {
			Result []miguSong `json:"result"`
		} `json:"songResultData"`
	}
	if json.Unmarshal(body, &wrap) != nil || wrap.Code != "000000" {
		return nil, errSource
	}
	out := make([]Track, 0, len(wrap.Song.Result))
	seen := map[string]struct{}{}
	for _, song := range wrap.Song.Result {
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

func (m *Migu) AudioURL(ctx context.Context, sourceID, quality string) (string, error) {
	contentID, copyrightID, err := parseMiguID(sourceID)
	if err != nil {
		return "", err
	}
	flag := "PQ"
	switch normalizeQuality(quality) {
	case "hr":
		flag = "SQ"
	case "hi":
		flag = "HQ"
	}
	if raw := m.listen(ctx, contentID, copyrightID, flag); raw != "" {
		return raw, nil
	}
	return "", ErrNotFound
}

func (m *Migu) Resolve(ctx context.Context, sourceID string) (Resolved, error) {
	audio, err := m.AudioURL(ctx, sourceID, "std")
	if err != nil {
		return Resolved{}, err
	}
	track := m.detail(ctx, sourceID)
	track.Source = sourceMigu
	track.SourceID = sourceID
	if track.Quality == "" {
		track.Quality = "std"
	}
	if track.Name == "" {
		return Resolved{}, ErrNotFound
	}
	return Resolved{Track: track, Audio: Remote{URL: audio}, Cover: track.Cover}, nil
}

func (m *Migu) listen(ctx context.Context, contentID, copyrightID, flag string) string {
	u := "https://app.pd.nf.migu.cn/MIGUM2.0/v1.0/content/sub/listenSong.do?" + url.Values{
		"contentId": {contentID}, "copyrightId": {copyrightID}, "resourceType": {"2"},
		"toneFlag": {flag}, "netType": {"00"}, "channel": {"0146951"},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://h5.nf.migu.cn/")
	resp, err := m.redirect.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc := resp.Header.Get("Location")
		if strings.HasPrefix(loc, "//") {
			loc = "https:" + loc
		}
		if _, err := parseRemoteURL(loc); err == nil {
			return loc
		}
	}
	return ""
}

func (m *Migu) detail(ctx context.Context, sourceID string) Track {
	_, copyrightID, err := parseMiguID(sourceID)
	if err != nil {
		return Track{}
	}
	u := "https://app.c.nf.migu.cn/MIGUM2.0/v1.0/content/resourceinfo.do?" + url.Values{
		"copyrightId": {copyrightID}, "resourceType": {"2"},
	}.Encode()
	body, resp, err := fetch(ctx, m.client, "GET", u, "https://m.music.migu.cn/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return Track{SourceID: sourceID}
	}
	var wrap struct {
		Resource []miguSong `json:"resource"`
	}
	if json.Unmarshal(body, &wrap) != nil || len(wrap.Resource) == 0 {
		return Track{SourceID: sourceID}
	}
	track, ok := wrap.Resource[0].track()
	if !ok {
		return Track{SourceID: sourceID}
	}
	track.SourceID = sourceID
	return track
}

type miguSong struct {
	Name        string `json:"name"`
	SongName    string `json:"songName"`
	ContentID   string `json:"contentId"`
	CopyrightID string `json:"copyrightId"`
	Singer      []struct {
		Name string `json:"name"`
	} `json:"singer"`
	Singers []struct {
		Name string `json:"name"`
	} `json:"singers"`
	ImgItems []struct {
		Img string `json:"img"`
	} `json:"imgItems"`
}

func (s miguSong) track() (Track, bool) {
	name := strings.TrimSpace(s.Name)
	if name == "" {
		name = strings.TrimSpace(s.SongName)
	}
	if name == "" || (s.ContentID == "" && s.CopyrightID == "") {
		return Track{}, false
	}
	people := s.Singer
	if len(people) == 0 {
		people = s.Singers
	}
	names := make([]string, 0, len(people))
	for _, singer := range people {
		names = append(names, singer.Name)
	}
	cover := ""
	if len(s.ImgItems) > 0 {
		cover = cleanCover(s.ImgItems[0].Img)
	}
	id := s.ContentID
	if s.CopyrightID != "" {
		id = miguPrefix + s.ContentID + ":" + s.CopyrightID
	}
	return Track{
		Name: name, Author: ChooseAuthor(names), Cover: cover,
		Source: sourceMigu, SourceID: id, Quality: "std",
	}, true
}

func parseMiguID(sourceID string) (string, string, error) {
	id, err := plainID(sourceID, miguPrefix)
	if err != nil {
		return "", "", err
	}
	contentID, copyrightID, ok := strings.Cut(id, ":")
	if !ok || contentID == "" || copyrightID == "" {
		return "", "", ErrNotFound
	}
	return contentID, copyrightID, nil
}

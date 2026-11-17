package music

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const yitingPrefix = "yt:"

type Yiting struct {
	client *http.Client
}

func newYiting(client *http.Client) *Yiting {
	if client == nil {
		client = sourceClient()
	}
	return &Yiting{client: client}
}

func (y *Yiting) Label() string { return sourceYiting }

func (y *Yiting) Search(ctx context.Context, query string) ([]Track, error) {
	return y.SearchPage(ctx, query, 0, 30)
}

func (y *Yiting) SearchPage(ctx context.Context, query string, offset, limit int) ([]Track, error) {
	page, size, skip := upstreamWindow(offset, limit)
	u := "http://so.1ting.com/song/json?" + url.Values{
		"q": {query}, "page": {strconv.Itoa(page)}, "size": {strconv.Itoa(size)},
	}.Encode()
	body, resp, err := fetch(ctx, y.client, "GET", u, "https://www.1ting.com/", nil, "")
	if err != nil || resp.StatusCode != 200 {
		return nil, errSource
	}
	var wrap struct {
		Results []ytSong `json:"results"`
	}
	if json.Unmarshal(body, &wrap) != nil {
		return nil, errSource
	}
	out := make([]Track, 0, len(wrap.Results))
	seen := map[string]struct{}{}
	for _, song := range wrap.Results {
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

func (y *Yiting) AudioURL(ctx context.Context, sourceID, _ string) (string, error) {
	_, path, err := parseYitingID(sourceID)
	if err != nil || path == "" {
		return "", ErrNotFound
	}
	for _, raw := range yitingCandidates(path) {
		if y.reachable(ctx, raw) {
			return raw, nil
		}
	}
	return "", ErrNotFound
}

func (y *Yiting) Resolve(ctx context.Context, sourceID string) (Resolved, error) {
	audio, err := y.AudioURL(ctx, sourceID, "std")
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{
		Track: Track{Source: sourceYiting, SourceID: sourceID, Quality: "std"},
		Audio: Remote{URL: audio},
	}, nil
}

func (y *Yiting) reachable(ctx context.Context, raw string) bool {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Range", "bytes=0-3")
	req.Header.Set("Referer", "https://www.1ting.com/")
	client := y.client
	if client == nil {
		client = sourceClient()
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return false
	}
	ctype := strings.ToLower(resp.Header.Get("Content-Type"))
	return strings.Contains(ctype, "audio") || strings.Contains(ctype, "octet-stream") || strings.Contains(ctype, "video")
}

type ytSong struct {
	Name   string `json:"song_name"`
	ID     string `json:"song_id"`
	Singer string `json:"singer_name"`
	Cover  string `json:"album_cover"`
	Path   string `json:"song_filepath"`
}

func (s ytSong) track() (Track, bool) {
	name := strings.TrimSpace(s.Name)
	id := strings.TrimSpace(s.ID)
	if name == "" || id == "" || strings.Trim(id, "0123456789") != "" {
		return Track{}, false
	}
	return Track{
		Name: name, Author: ChooseAuthor([]string{s.Singer}), Cover: cleanCover(s.Cover),
		Source: sourceYiting, SourceID: yitingID(id, s.Path), Quality: "std",
	}, true
}

func yitingID(id, path string) string {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/") {
		return yitingPrefix + id
	}
	return yitingPrefix + id + ":" + base64.RawURLEncoding.EncodeToString([]byte(path))
}

func parseYitingID(sourceID string) (string, string, error) {
	raw, err := plainID(sourceID, yitingPrefix)
	if err != nil {
		return "", "", err
	}
	id, enc, ok := strings.Cut(raw, ":")
	if !ok || id == "" {
		return raw, "", nil
	}
	buf, err := base64.RawURLEncoding.DecodeString(enc)
	if err != nil || len(buf) == 0 || buf[0] != '/' {
		return "", "", ErrNotFound
	}
	return id, string(buf), nil
}

func yitingCandidates(path string) []string {
	path = "/" + strings.TrimPrefix(path, "/")
	return []string{
		"https://www.1ting.com" + path,
		"http://www.1ting.com" + path,
		"http://h5.1ting.com" + path,
	}
}

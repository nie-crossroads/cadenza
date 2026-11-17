package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nie-crossroads/cadenza/internal/config"
	"github.com/nie-crossroads/cadenza/internal/music"
)

func TestHTTPLoginSearchAddAndPreview(t *testing.T) {
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("ID3xxxx"))
	}))
	defer media.Close()

	svc := &music.Service{
		Source:       gateSource{media: media.URL},
		Store:        &httpMemStore{},
		Blog:         &httpMemBlog{},
		AllowPrivate: true,
	}
	h := newHandler(config.Config{
		WebPassword:  "right-password",
		AccessToken:  "rest-token",
		SessionTTL:   time.Hour,
		CookieSecure: false,
	}, svc)
	srv := httptest.NewServer(h)
	defer srv.Close()

	res := do(t, srv, http.MethodPost, "/api/login", `{"password":"nope"}`, "")
	if res.StatusCode != http.StatusUnauthorized || codeOf(t, res) != "LOGIN_INVALID" {
		t.Fatalf("bad login %d %s", res.StatusCode, codeOf(t, res))
	}

	login := do(t, srv, http.MethodPost, "/api/login", `{"password":"right-password"}`, "")
	if login.StatusCode != http.StatusNoContent {
		t.Fatalf("login %d", login.StatusCode)
	}
	login.Body.Close()
	cookie := login.Header.Get("Set-Cookie")
	if cookie == "" || !bytes.Contains([]byte(cookie), []byte("HttpOnly")) {
		t.Fatalf("cookie %s", cookie)
	}

	session := do(t, srv, http.MethodGet, "/api/session", "", cookie)
	session.Body.Close()
	if session.StatusCode != http.StatusNoContent {
		t.Fatalf("session %d", session.StatusCode)
	}

	search := do(t, srv, http.MethodGet, "/api/search?q=aqua", "", "")
	if search.StatusCode != http.StatusOK {
		t.Fatalf("search %d", search.StatusCode)
	}
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(search.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	_ = search.Body.Close()
	if len(payload.Items) != 1 || payload.Items[0]["sourceId"] != "wy:7" {
		t.Fatalf("%v", payload.Items)
	}
	if _, ok := payload.Items[0]["url"]; ok {
		t.Fatal("search returned a play url")
	}

	anon := do(t, srv, http.MethodPost, "/api/music", `{"sourceId":"wy:7"}`, "")
	if anon.StatusCode != http.StatusUnauthorized || codeOf(t, anon) != "UNAUTHENTICATED" {
		t.Fatalf("anon add %d %s", anon.StatusCode, codeOf(t, anon))
	}

	added := do(t, srv, http.MethodPost, "/api/music", `{"sourceId":"wy:7"}`, cookie)
	if added.StatusCode != http.StatusOK {
		t.Fatalf("cookie add %d %s", added.StatusCode, codeOf(t, added))
	}
	added.Body.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/music", bytes.NewBufferString(`{"sourceId":"wy:7"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer rest-token")
	rest, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rest.Body.Close()
	if rest.StatusCode != http.StatusOK {
		t.Fatalf("bearer add %d", rest.StatusCode)
	}

	preview, err := http.NewRequest(http.MethodGet, srv.URL+"/api/preview?sourceId=wy:7", nil)
	if err != nil {
		t.Fatal(err)
	}
	preview.Header.Set("Range", "bytes=0-2")
	pres, err := http.DefaultClient.Do(preview)
	if err != nil {
		t.Fatal(err)
	}
	defer pres.Body.Close()
	if pres.StatusCode != http.StatusPartialContent || pres.Header.Get("Accept-Ranges") != "bytes" {
		t.Fatalf("preview %d %s", pres.StatusCode, pres.Header.Get("Accept-Ranges"))
	}
	buf, _ := io.ReadAll(pres.Body)
	if string(buf) != "ID3" {
		t.Fatalf("preview body %q", buf)
	}
	if pres.Header.Get("Location") != "" {
		t.Fatal("preview exposed a redirect")
	}

	anonProbe := do(t, srv, http.MethodPost, "/api/music/playable", `{"sourceId":"wy:7","quality":"std"}`, "")
	if anonProbe.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anon playable %d", anonProbe.StatusCode)
	}
	probed := do(t, srv, http.MethodPost, "/api/music/playable", `{"sourceId":"wy:7","quality":"std"}`, cookie)
	if probed.StatusCode != http.StatusOK {
		t.Fatalf("playable %d %s", probed.StatusCode, codeOf(t, probed))
	}
	var playable struct {
		Playable bool `json:"playable"`
	}
	if err := json.NewDecoder(probed.Body).Decode(&playable); err != nil {
		t.Fatal(err)
	}
	_ = probed.Body.Close()
	if !playable.Playable {
		t.Fatal("expected playable")
	}

	page, err := http.Get(srv.URL + "/results")
	if err != nil {
		t.Fatal(err)
	}
	defer page.Body.Close()
	pageBody, _ := io.ReadAll(page.Body)
	if page.StatusCode != http.StatusOK || !bytes.Contains(pageBody, []byte("Cadenza")) {
		t.Fatalf("spa %d", page.StatusCode)
	}
}

func do(t *testing.T, srv *httptest.Server, method, path, body, cookie string) *http.Response {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(method, srv.URL+path, rdr)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookieHeader(cookie))
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func cookieHeader(setCookie string) string {
	part := setCookie
	if i := bytes.IndexByte([]byte(setCookie), ';'); i >= 0 {
		part = setCookie[:i]
	}
	return part
}

func codeOf(t *testing.T, res *http.Response) string {
	t.Helper()
	defer res.Body.Close()
	var body struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(res.Body).Decode(&body)
	return body.Code
}

type gateSource struct {
	media string
}

func (g gateSource) Search(context.Context, string) ([]music.Track, error) {
	return []music.Track{{
		Name: "aqua", Author: "坂本龍一", Cover: "", Duration: 270,
		Source: "网易", SourceID: "wy:7", Quality: "hi",
	}}, nil
}

func (g gateSource) Resolve(context.Context, string) (music.Resolved, error) {
	return music.Resolved{
		Track: music.Track{Name: "aqua", Author: "坂本龍一", Source: "网易", SourceID: "wy:7", Quality: "hi"},
		Audio: music.Remote{URL: g.media + "/a.mp3"},
	}, nil
}

type httpMemStore struct{}

func (httpMemStore) Put(context.Context, string, string, io.Reader, int64, string) (string, string, error) {
	return "https://cdn.example.com/m/abcdefghijkl.mp3", "m/abcdefghijkl.mp3", nil
}

func (httpMemStore) Delete(context.Context, string) error { return nil }

type httpMemBlog struct{}

func (httpMemBlog) Save(context.Context, music.Track, string, string) (int64, error) {
	return 9, nil
}

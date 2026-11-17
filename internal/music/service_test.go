package music

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestAddUploadsShortAddress(t *testing.T) {
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/a.mp3" {
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write(append([]byte("ID3"), []byte("audio")...))
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
	}))
	defer media.Close()

	store := &memStore{domain: "https://cdn.example.com"}
	blog := &memBlog{status: http.StatusOK, id: 42}
	svc := &Service{
		Source: fakeSource{resolved: Resolved{
			Track: Track{Name: "aqua", Author: "坂本龍一", Cover: media.URL + "/c.jpg", Source: "网易", SourceID: "wy:1", Quality: "hi"},
			Audio: Remote{URL: media.URL + "/a.mp3"},
			Cover: media.URL + "/c.jpg",
		}},
		Store:        store,
		Blog:         blog,
		AllowPrivate: true,
		FetchTimeout: 5_000_000_000,
	}
	id, err := svc.Add(context.Background(), "wy:1", "std")
	if err != nil || id != 42 {
		t.Fatalf("id %d err %v", id, err)
	}
	if blog.audio == "" || blog.audio == media.URL+"/a.mp3" {
		t.Fatalf("stored address %s", blog.audio)
	}
	if len(blog.audio) > 80 || len(blog.cover) > 100 {
		t.Fatalf("too long audio %d cover %d", len(blog.audio), len(blog.cover))
	}
	if blog.name != "aqua" || blog.author != "坂本龍一" {
		t.Fatalf("metadata %s %s", blog.name, blog.author)
	}
	if len(store.objects) != 2 {
		t.Fatalf("objects %d", len(store.objects))
	}
}

func TestPlayableReusesStandardPreviewAndRecentCheck(t *testing.T) {
	var hits int
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("ID3xxxx"))
	}))
	defer media.Close()
	svc := &Service{
		Source: qualSource{urls: map[string]string{
			"hi": media.URL + "/hi.mp3",
		}},
		AllowPrivate: true,
		relays: map[string]*audioRelay{
			"wy:1": {data: []byte("ID3played"), ctype: "audio/mpeg", total: 8},
		},
	}
	ok, err := svc.Playable(context.Background(), "wy:1", "std")
	if err != nil || !ok || hits != 0 {
		t.Fatalf("std from preview playable %v err %v hits %d", ok, err, hits)
	}
	ok, err = svc.Playable(context.Background(), "wy:1", "hi")
	if err != nil || !ok || hits != 1 {
		t.Fatalf("hi probed playable %v err %v hits %d", ok, err, hits)
	}
	ok, err = svc.Playable(context.Background(), "wy:1", "hi")
	if err != nil || !ok || hits != 1 {
		t.Fatalf("hi cached playable %v err %v hits %d", ok, err, hits)
	}
}

func TestPlayableUsesChosenQuality(t *testing.T) {
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/std.mp3":
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write([]byte("ID3xxxx"))
		case "/hi.html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>no</html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer media.Close()
	svc := &Service{
		Source: qualSource{urls: map[string]string{
			"std": media.URL + "/std.mp3",
			"hi":  media.URL + "/hi.html",
		}},
		AllowPrivate: true,
	}
	ok, err := svc.Playable(context.Background(), "wy:1", "std")
	if err != nil || !ok {
		t.Fatalf("std playable %v %v", ok, err)
	}
	ok, err = svc.Playable(context.Background(), "wy:1", "hi")
	if err != nil || ok {
		t.Fatalf("hi playable %v %v", ok, err)
	}
	ok, err = svc.Playable(context.Background(), "wy:1", "hr")
	if err != nil || ok {
		t.Fatalf("missing playable %v %v", ok, err)
	}
	if _, err = svc.Playable(context.Background(), "wy:1", "lossless"); err == nil {
		t.Fatal("invalid quality accepted")
	}
}

func TestAddDuplicateDeletesUpload(t *testing.T) {
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ID3xxxx"))
	}))
	defer media.Close()
	store := &memStore{domain: "https://cdn.example.com"}
	svc := &Service{
		Source: fakeSource{resolved: Resolved{
			Track: Track{Name: "aqua", Author: "坂本龍一", SourceID: "wy:1"},
			Audio: Remote{URL: media.URL + "/a.mp3"},
		}},
		Store:        store,
		Blog:         &memBlog{status: http.StatusConflict},
		AllowPrivate: true,
	}
	_, err := svc.Add(context.Background(), "wy:1", "std")
	var f *Fail
	if !errors.As(err, &f) || f.Code != CodeDuplicate || f.Status != http.StatusConflict {
		t.Fatalf("%v", err)
	}
	if len(store.objects) != 0 {
		t.Fatalf("orphan objects %d", len(store.objects))
	}
}

func TestAddRejectsOversizedAudio(t *testing.T) {
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ID3-too-big"))
	}))
	defer media.Close()
	store := &memStore{domain: "https://cdn.example.com"}
	svc := &Service{
		Source: fakeSource{resolved: Resolved{
			Track: Track{Name: "aqua", Author: "坂本龍一", SourceID: "wy:1"},
			Audio: Remote{URL: media.URL + "/a.mp3"},
		}},
		Store:         store,
		Blog:          &memBlog{status: http.StatusOK, id: 1},
		AllowPrivate:  true,
		MaxAudioBytes: 4,
	}
	_, err := svc.Add(context.Background(), "wy:1", "std")
	var f *Fail
	if !errors.As(err, &f) || f.Code != CodeTooLarge {
		t.Fatalf("%v", err)
	}
	if len(store.objects) != 0 {
		t.Fatal("uploaded despite size limit")
	}
}

func TestPreviewRange(t *testing.T) {
	var hits int
	body := []byte("ID3xxxxx")
	media := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.Header.Get("Range") != "" {
			t.Errorf("upstream range %s", r.Header.Get("Range"))
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write(body)
	}))
	defer media.Close()
	svc := &Service{
		Source: fakeSource{resolved: Resolved{
			Track: Track{Name: "aqua", SourceID: "wy:1"},
			Audio: Remote{URL: media.URL + "/a.mp3"},
		}},
		AllowPrivate: true,
	}
	status, hdr, stream, err := svc.Preview(context.Background(), "wy:1", "bytes=0-3")
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if status != http.StatusPartialContent {
		t.Fatalf("status %d", status)
	}
	if hdr.Get("Accept-Ranges") != "bytes" || hdr.Get("Content-Range") != "bytes 0-3/8" {
		t.Fatalf("headers %v", hdr)
	}
	buf, _ := io.ReadAll(stream)
	if string(buf) != "ID3x" {
		t.Fatalf("body %q", buf)
	}
	status, hdr, stream2, err := svc.Preview(context.Background(), "wy:1", "bytes=4-7")
	if err != nil {
		t.Fatal(err)
	}
	defer stream2.Close()
	buf, _ = io.ReadAll(stream2)
	if status != http.StatusPartialContent || string(buf) != "xxxx" {
		t.Fatalf("second status %d body %q", status, buf)
	}
	if hits != 1 {
		t.Fatalf("upstream hits %d", hits)
	}
	for k, values := range hdr {
		for _, v := range values {
			if stringsContains(v, media.URL) {
				t.Fatalf("header %s leaked upstream", k)
			}
		}
	}
}

func TestPreviewBlocksPrivateAddress(t *testing.T) {
	svc := &Service{
		Source: fakeSource{resolved: Resolved{
			Track: Track{Name: "aqua", SourceID: "wy:1"},
			Audio: Remote{URL: "http://127.0.0.1:9/secret.mp3"},
		}},
	}
	_, _, _, err := svc.Preview(context.Background(), "wy:1", "")
	if err == nil {
		t.Fatal("private address was fetched")
	}
}

type fakeSource struct {
	resolved Resolved
}

func (f fakeSource) Search(context.Context, string) ([]Track, error) {
	return []Track{f.resolved.Track}, nil
}

func (f fakeSource) Resolve(context.Context, string) (Resolved, error) {
	return f.resolved, nil
}

type qualSource struct {
	urls map[string]string
}

func (qualSource) Search(context.Context, string) ([]Track, error) {
	return nil, nil
}

func (q qualSource) Resolve(context.Context, string) (Resolved, error) {
	return Resolved{Audio: Remote{URL: q.urls["std"]}}, nil
}

func (q qualSource) AudioURL(_ context.Context, _ string, quality string) (string, error) {
	raw, ok := q.urls[quality]
	if !ok || raw == "" {
		return "", ErrNotFound
	}
	return raw, nil
}

type memStore struct {
	mu      sync.Mutex
	domain  string
	objects map[string][]byte
}

func (m *memStore) Put(_ context.Context, kind, ext string, body io.Reader, _ int64, _ string) (string, string, error) {
	buf, err := io.ReadAll(body)
	if err != nil {
		return "", "", err
	}
	key, err := objectKey(kind, ext)
	if err != nil {
		return "", "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.objects == nil {
		m.objects = map[string][]byte{}
	}
	m.objects[key] = buf
	return m.domain + "/" + key, key, nil
}

func (m *memStore) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
	return nil
}

type memBlog struct {
	status int
	id     int64
	name   string
	author string
	audio  string
	cover  string
}

func (b *memBlog) Save(_ context.Context, track Track, audioURL, coverURL string) (int64, error) {
	b.name = track.Name
	b.author = track.Author
	b.audio = audioURL
	b.cover = coverURL
	if b.status != http.StatusOK {
		return 0, mapBlogStatus(b.status)
	}
	return b.id, nil
}

func stringsContains(s, part string) bool {
	return len(part) > 0 && len(s) >= len(part) && (s == part || len(s) > 0 && (func() bool {
		return len(s) >= len(part) && (s == part || contains(s, part))
	})())
}

func contains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}

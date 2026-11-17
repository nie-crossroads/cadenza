package music

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNeteaseSearch(t *testing.T) {
	const body = `{"code":200,"result":{"songs":[{"name":"aqua","id":500071,"dt":270281,"ar":[{"name":"坂本龍一"}],"al":{"picUrl":"https://img.example/a.jpg"},"privilege":{"maxbr":999000,"maxBrLevel":"lossless","plLevel":"exhigh"}}]}}`
	var sawCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawCookie = r.Header.Get("Cookie")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	src := newNetease(srv.URL, "MUSIC_U=secret", srv.Client())
	items, err := src.Search(context.Background(), "aqua")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items %d", len(items))
	}
	got := items[0]
	if got.Name != "aqua" || got.Author != "坂本龍一" || got.Source != "网易" || got.SourceID != "wy:500071" {
		t.Fatalf("%+v", got)
	}
	if got.Duration != 270 || got.Quality != "hi" || got.Cover == "" {
		t.Fatalf("%+v", got)
	}
	if sawCookie != "MUSIC_U=secret" {
		t.Fatal("cookie was not kept on the source request")
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 7 {
		t.Fatalf("fields %v", fields)
	}
	if _, ok := fields["url"]; ok {
		t.Fatal("play url leaked")
	}
}

func TestNeteaseResolve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "song/detail"):
			_, _ = io.WriteString(w, `{"code":200,"songs":[{"name":"aqua","id":500071,"dt":270000,"artists":[{"name":"坂本龍一"}],"album":{"picUrl":"https://img.example/a.jpg"}}]}`)
		default:
			_, _ = io.WriteString(w, `{"code":200,"data":[{"url":"https://m.example/a.mp3","br":320000,"level":"exhigh","code":200}]}`)
		}
	}))
	defer srv.Close()
	src := newNetease(srv.URL, "", srv.Client())
	got, err := src.Resolve(context.Background(), "wy:500071")
	if err != nil {
		t.Fatal(err)
	}
	if got.Audio.URL != "https://m.example/a.mp3" || got.Track.SourceID != "wy:500071" || got.Track.Quality != "hi" {
		t.Fatalf("%+v", got)
	}
	if _, err := src.Resolve(context.Background(), "https://evil.example/a.mp3"); !errorsIsNotFound(err) {
		t.Fatal("non source id should be rejected")
	}
}

func TestChooseAuthorAndQuality(t *testing.T) {
	if got := ChooseAuthor([]string{"坂本龍一", "久石让"}); got != "坂本龍一、久石让" {
		t.Fatal(got)
	}
	long := strings.Repeat("龙", 21)
	if ChooseAuthor([]string{long, "坂本龍一"}) != long {
		t.Fatal("single long name stays intact")
	}
	if QualityOf("exhigh", "hires", 999000, false) != "hi" {
		t.Fatal("playable level")
	}
	if QualityOf("exhigh", "hires", 999000, true) != "hr" {
		t.Fatal("catalog level")
	}
	if UTF16Len("歌") != 1 {
		t.Fatal("utf16")
	}
}

func errorsIsNotFound(err error) bool {
	return err == ErrNotFound
}

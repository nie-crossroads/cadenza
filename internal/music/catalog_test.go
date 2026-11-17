package music

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestUpstreamWindow(t *testing.T) {
	page, size, skip := upstreamWindow(0, 5)
	if page != 1 || size != 5 || skip != 0 {
		t.Fatalf("first page %d %d %d", page, size, skip)
	}
	page, size, skip = upstreamWindow(10, 20)
	if page != 1 || size != 30 || skip != 10 {
		t.Fatalf("window %d %d %d", page, size, skip)
	}
	items := make([]Track, 30)
	got := takeWindow(items, skip, 20)
	if len(got) != 20 {
		t.Fatalf("take %d", len(got))
	}
}

func TestLooseJSON(t *testing.T) {
	raw := []byte(`{'abslist':[{'SONGNAME':'晴天','ARTIST':'周杰伦','DURATION':'269','MUSICRID':'MUSIC_42','PICPATH':'https://img.example/a.jpg'}]}`)
	parsed, err := looseJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	var wrap struct {
		AbsList []kwSong `json:"abslist"`
	}
	if err := json.Unmarshal(parsed, &wrap); err != nil {
		t.Fatal(err)
	}
	track, ok := wrap.AbsList[0].track()
	if !ok || track.Name != "晴天" || track.Author != "周杰伦" || track.SourceID != "kw:42" || track.Duration != 269 {
		t.Fatalf("%+v", track)
	}
}

func TestCatalogKeepsOtherSources(t *testing.T) {
	catalog := &Catalog{sources: []catalogSource{
		stubSource{label: sourceNetease, items: numbered(sourceNetease, "wy:", 8)},
		stubSource{label: sourceKuwo, err: errors.New("down")},
		stubSource{label: sourceQQ, items: numbered(sourceQQ, "qq:", 2)},
	}}
	items, err := catalog.SearchFiltered(context.Background(), "晴天", "", 0, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 7 {
		t.Fatalf("len %d", len(items))
	}
	if items[0].Source != sourceNetease || items[5].Source != sourceQQ {
		t.Fatalf("%s %s", items[0].Source, items[5].Source)
	}
	only, err := catalog.SearchFiltered(context.Background(), "晴天", sourceQQ, 0, 5)
	if err != nil || len(only) != 2 || only[0].Source != sourceQQ {
		t.Fatalf("%v %+v", err, only)
	}
}

type stubSource struct {
	label string
	items []Track
	err   error
}

func (s stubSource) Label() string { return s.label }
func (s stubSource) Search(context.Context, string) ([]Track, error) {
	return s.items, s.err
}
func (s stubSource) SearchPage(_ context.Context, _ string, offset, limit int) ([]Track, error) {
	if s.err != nil {
		return nil, s.err
	}
	return takeWindow(s.items, offset, limit), nil
}
func (s stubSource) Resolve(context.Context, string) (Resolved, error) {
	return Resolved{}, ErrNotFound
}

func numbered(source, prefix string, n int) []Track {
	items := make([]Track, n)
	for i := range items {
		items[i] = Track{Name: "晴天", Author: "周杰伦", Source: source, SourceID: prefix + "n", Quality: "std"}
	}
	return items
}

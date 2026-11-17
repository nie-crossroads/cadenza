package config

import "testing"

func TestNormalizeBlogURL(t *testing.T) {
	got := normalizeBlogURL("https://example.com")
	want := "https://example.com/agent-blog/server/api/cadenza/music"
	if got != want {
		t.Fatalf("origin: got %s", got)
	}
	full := "https://example.com/agent-blog/server/api/cadenza/music"
	if normalizeBlogURL(full+"/") != full {
		t.Fatal("full url should stay")
	}
	if normalizeBlogURL("  ") != "" {
		t.Fatal("blank")
	}
}

func TestNormalizeDomain(t *testing.T) {
	if normalizeDomain("cdn.example.com/") != "https://cdn.example.com" {
		t.Fatal("scheme")
	}
	if normalizeDomain("https://cdn.example.com") != "https://cdn.example.com" {
		t.Fatal("keep")
	}
}

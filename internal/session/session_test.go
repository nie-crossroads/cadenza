package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRoundTrip(t *testing.T) {
	m := New("secret", time.Hour, false)
	rec := httptest.NewRecorder()
	m.Set(rec)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
		if !c.HttpOnly {
			t.Fatal("cookie must be HttpOnly")
		}
	}
	if !m.Valid(req) {
		t.Fatal("expected valid session")
	}
}

func TestRejectsTamperAndEmptyPassword(t *testing.T) {
	m := New("secret", time.Hour, false)
	rec := httptest.NewRecorder()
	m.Set(rec)
	raw := rec.Result().Cookies()[0].Value
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "cadenza_session", Value: raw + "x"})
	if m.Valid(req) {
		t.Fatal("tampered cookie accepted")
	}
	blank := New("", time.Hour, false)
	if blank.Enabled() || blank.Valid(req) {
		t.Fatal("empty password must not issue a session")
	}
}

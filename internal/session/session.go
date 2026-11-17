// Package session 签发网页登录用的 HttpOnly Cookie。
package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const cookieName = "cadenza_session"

type Manager struct {
	key    []byte
	ttl    time.Duration
	secure bool
}

func New(password string, ttl time.Duration, secure bool) *Manager {
	m := &Manager{ttl: ttl, secure: secure}
	if strings.TrimSpace(password) == "" || ttl <= 0 {
		return m
	}
	sum := sha256.Sum256([]byte("cadenza-session:" + password))
	m.key = sum[:]
	return m
}

func (m *Manager) Enabled() bool {
	return len(m.key) > 0
}

func (m *Manager) Set(w http.ResponseWriter) {
	if !m.Enabled() {
		return
	}
	exp := time.Now().Add(m.ttl).Unix()
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return
	}
	payload := strconv.FormatInt(exp, 10) + "." + hex.EncodeToString(buf[:])
	mac := m.sign(payload)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    payload + "." + mac,
		Path:     "/",
		MaxAge:   int(m.ttl.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.secure,
	})
}

func (m *Manager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.secure,
	})
}

func (m *Manager) Valid(r *http.Request) bool {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return false
	}
	return m.verify(c.Value)
}

func (m *Manager) Present(r *http.Request) bool {
	c, err := r.Cookie(cookieName)
	return err == nil && c.Value != ""
}

func (m *Manager) sign(payload string) string {
	h := hmac.New(sha256.New, m.key)
	_, _ = h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func (m *Manager) verify(token string) bool {
	if !m.Enabled() {
		return false
	}
	expText, rest, ok := strings.Cut(token, ".")
	if !ok {
		return false
	}
	nonce, mac, ok := strings.Cut(rest, ".")
	if !ok || nonce == "" || mac == "" {
		return false
	}
	payload := expText + "." + nonce
	if !hmac.Equal([]byte(mac), []byte(m.sign(payload))) {
		return false
	}
	exp, err := strconv.ParseInt(expText, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	return true
}

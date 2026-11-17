package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nie-crossroads/cadenza/internal/music"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeCode(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"code": code})
}

func writeFail(w http.ResponseWriter, err error) {
	var f *music.Fail
	if errors.As(err, &f) && f.Code != "" {
		status := f.Status
		if status == 0 {
			status = http.StatusBadGateway
		}
		writeCode(w, status, f.Code)
		return
	}
	writeCode(w, http.StatusBadGateway, music.CodeUnknown)
}

func newDecoder(r interface{ Read([]byte) (int, error) }) *json.Decoder {
	return json.NewDecoder(r)
}

func sameSecret(got, want string) bool {
	if got == "" || want == "" {
		return false
	}
	a := sha256.Sum256([]byte(got))
	b := sha256.Sum256([]byte(want))
	return hmac.Equal(a[:], b[:])
}

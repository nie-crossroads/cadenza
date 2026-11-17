package music

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

var errSource = errors.New("source status")

func sourceClient() *http.Client {
	return &http.Client{Timeout: 8 * time.Second}
}

func noRedirectClient() *http.Client {
	return &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func fetch(ctx context.Context, client *http.Client, method, raw, referer string, body io.Reader, contentType string) ([]byte, *http.Response, error) {
	if client == nil {
		client = sourceClient()
	}
	req, err := http.NewRequestWithContext(ctx, method, raw, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, resp, err
	}
	return buf, resp, nil
}

func plainID(sourceID, prefix string) (string, error) {
	if len(sourceID) <= len(prefix) || sourceID[:len(prefix)] != prefix {
		return "", ErrNotFound
	}
	id := sourceID[len(prefix):]
	if id == "" || len(id) > 180 {
		return "", ErrNotFound
	}
	for _, r := range id {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r == ':' || r == '_' || r == '-':
		default:
			return "", ErrNotFound
		}
	}
	return id, nil
}

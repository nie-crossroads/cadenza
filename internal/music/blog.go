package music

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// Blog 调用博客已有的添加接口，不连接博客数据库。
type Blog interface {
	Save(ctx context.Context, track Track, audioURL, coverURL string) (int64, error)
}

type blogClient struct {
	endpoint string
	token    string
	client   *http.Client
}

func NewBlog(endpoint, token string, timeout time.Duration) Blog {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &blogClient{
		endpoint: strings.TrimSpace(endpoint),
		token:    strings.TrimSpace(token),
		client:   &http.Client{Timeout: timeout},
	}
}

type blogBody struct {
	Name   string `json:"name"`
	Author string `json:"author,omitempty"`
	URL    string `json:"url"`
	Cover  string `json:"cover,omitempty"`
	Status int    `json:"status"`
}

func (b *blogClient) Save(ctx context.Context, track Track, audioURL, coverURL string) (int64, error) {
	payload := blogBody{
		Name:   track.Name,
		Author: track.Author,
		URL:    audioURL,
		Cover:  coverURL,
		Status: 1,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, fail(CodeUnknown, statusOf(CodeUnknown))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.endpoint, bytes.NewReader(raw))
	if err != nil {
		return 0, fail(CodeUnknown, statusOf(CodeUnknown))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.token)
	resp, err := b.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return 0, fail(CodeTimeout, statusOf(CodeTimeout))
		}
		return 0, fail(CodeUnknown, statusOf(CodeUnknown))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == http.StatusOK {
		id, ok := parseID(body)
		if !ok {
			return 0, fail(CodeUnknown, statusOf(CodeUnknown))
		}
		return id, nil
	}
	return 0, mapBlogStatus(resp.StatusCode)
}

func mapBlogStatus(status int) *Fail {
	switch status {
	case http.StatusUnauthorized, 419:
		return fail(CodeBlogToken, statusOf(CodeBlogToken))
	case http.StatusConflict:
		return fail(CodeDuplicate, statusOf(CodeDuplicate))
	case http.StatusBadRequest:
		return fail(CodeBlogValidation, statusOf(CodeBlogValidation))
	default:
		return fail(CodeUnknown, statusOf(CodeUnknown))
	}
}

func parseID(body []byte) (int64, bool) {
	dec := json.NewDecoder(bytes.NewReader(bytes.TrimSpace(body)))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return 0, false
	}
	switch n := v.(type) {
	case json.Number:
		id, err := n.Int64()
		return id, err == nil && id > 0
	case string:
		return parseID([]byte(n))
	default:
		return 0, false
	}
}

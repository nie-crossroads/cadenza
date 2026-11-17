package music

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// Store 把音频和封面放到对象存储，返回能放进博客字段的短地址。
type Store interface {
	Put(ctx context.Context, kind string, ext string, body io.Reader, size int64, contentType string) (publicURL, key string, err error)
	Delete(ctx context.Context, key string) error
}

type cosStore struct {
	client *cos.Client
	domain string
}

func NewCOS(secretID, secretKey, bucket, region, domain string) (Store, error) {
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region))
	if err != nil {
		return nil, err
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Timeout: 60 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
	store := &cosStore{client: client, domain: strings.TrimRight(domain, "/")}
	if !store.fits() {
		return nil, fmt.Errorf("cos domain is too long for blog columns")
	}
	return store, nil
}

func (s *cosStore) fits() bool {
	audio := s.domain + "/m/aaaaaaaaaaaa.flac"
	cover := s.domain + "/c/aaaaaaaaaaaa.webp"
	return len(audio) <= 80 && len(cover) <= 100
}

func (s *cosStore) Put(ctx context.Context, kind, ext string, body io.Reader, size int64, contentType string) (string, string, error) {
	key, err := objectKey(kind, ext)
	if err != nil {
		return "", "", err
	}
	public := s.domain + "/" + key
	limit := 80
	if kind == "c" {
		limit = 100
	}
	if len(public) > limit {
		return "", "", fail(CodeBlogValidation, statusOf(CodeBlogValidation))
	}
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType:   contentType,
			ContentLength: size,
		},
	}
	_, err = s.client.Object.Put(ctx, key, body, opt)
	if err != nil {
		if ctx.Err() != nil {
			return "", "", fail(CodeTimeout, statusOf(CodeTimeout))
		}
		return "", "", fail(CodeCOSFailed, statusOf(CodeCOSFailed))
	}
	return public, key, nil
}

func (s *cosStore) Delete(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	_, err := s.client.Object.Delete(ctx, key)
	return err
}

func objectKey(kind, ext string) (string, error) {
	if kind != "m" && kind != "c" {
		return "", fail(CodeUnknown, statusOf(CodeUnknown))
	}
	switch ext {
	case ".mp3", ".flac", ".jpg", ".jpeg", ".png", ".webp", ".gif":
	default:
		return "", fail(CodeUnknown, statusOf(CodeUnknown))
	}
	var buf [9]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return kind + "/" + base62(buf[:]) + ext, nil
}

func base62(b []byte) string {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	out := make([]byte, len(b))
	for i, v := range b {
		out[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(out)
}

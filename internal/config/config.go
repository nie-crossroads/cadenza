package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAddr        = ":8080"
	defaultMaxAudio    = 40 << 20
	defaultMaxCover    = 2 << 20
	blogMusicPath      = "/agent-blog/server/api/cadenza/music"
	defaultSessionTTL  = 7 * 24 * time.Hour
	defaultSearchLimit = 12 * time.Second
	defaultFetchLimit  = 45 * time.Second
	defaultBlogLimit   = 15 * time.Second
)

// Config 只从环境变量读取。页面、日志和仓库里不放密钥。
type Config struct {
	Addr          string
	WebPassword   string
	AccessToken   string
	CookieSecure  bool
	SessionTTL    time.Duration
	COSSecretID   string
	COSSecretKey  string
	COSBucket     string
	COSRegion     string
	COSDomain     string
	BlogURL       string
	BlogToken     string
	SourceKey     string
	MaxAudioBytes int64
	MaxCoverBytes int64
	SearchTimeout time.Duration
	FetchTimeout  time.Duration
	BlogTimeout   time.Duration
}

func Load() Config {
	loadDotEnv(".env")
	return FromEnv()
}

func FromEnv() Config {
	cfg := Config{
		Addr:          env("CADENZA_ADDR", defaultAddr),
		WebPassword:   strings.TrimSpace(os.Getenv("CADENZA_WEB_PASSWORD")),
		AccessToken:   strings.TrimSpace(os.Getenv("CADENZA_ACCESS_TOKEN")),
		CookieSecure:  strings.EqualFold(strings.TrimSpace(os.Getenv("CADENZA_COOKIE_SECURE")), "true"),
		SessionTTL:    defaultSessionTTL,
		COSSecretID:   strings.TrimSpace(os.Getenv("CADENZA_COS_SECRET_ID")),
		COSSecretKey:  strings.TrimSpace(os.Getenv("CADENZA_COS_SECRET_KEY")),
		COSBucket:     strings.TrimSpace(os.Getenv("CADENZA_COS_BUCKET")),
		COSRegion:     strings.TrimSpace(os.Getenv("CADENZA_COS_REGION")),
		COSDomain:     normalizeDomain(os.Getenv("CADENZA_COS_DOMAIN")),
		BlogURL:       normalizeBlogURL(os.Getenv("CADENZA_BLOG_URL")),
		BlogToken:     strings.TrimSpace(os.Getenv("CADENZA_BLOG_TOKEN")),
		SourceKey:     strings.TrimSpace(os.Getenv("CADENZA_SOURCE_KEY")),
		MaxAudioBytes: envInt64("CADENZA_MAX_AUDIO_BYTES", defaultMaxAudio),
		MaxCoverBytes: envInt64("CADENZA_MAX_COVER_BYTES", defaultMaxCover),
		SearchTimeout: defaultSearchLimit,
		FetchTimeout:  defaultFetchLimit,
		BlogTimeout:   defaultBlogLimit,
	}
	if cfg.MaxAudioBytes < 1 {
		cfg.MaxAudioBytes = defaultMaxAudio
	}
	if cfg.MaxCoverBytes < 1 {
		cfg.MaxCoverBytes = defaultMaxCover
	}
	return cfg
}

func (c Config) COSReady() bool {
	return c.COSSecretID != "" && c.COSSecretKey != "" && c.COSBucket != "" && c.COSRegion != "" && c.COSDomain != ""
}

func (c Config) BlogReady() bool {
	return c.BlogURL != "" && c.BlogToken != ""
}

func env(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func envInt64(key string, fallback int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func normalizeDomain(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "/")
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	return raw
}

func normalizeBlogURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "/")
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	if strings.HasSuffix(raw, blogMusicPath) {
		return raw
	}
	// 只给了站点源时补上博客已有的添加路径。
	slash := strings.Index(raw, "://")
	rest := raw[slash+3:]
	if !strings.Contains(rest, "/") {
		return raw + blogMusicPath
	}
	return raw
}

// loadDotEnv 读取本地 .env，不覆盖已经存在的环境变量。
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		_ = os.Setenv(key, val)
	}
}

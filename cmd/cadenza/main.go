package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nie-crossroads/cadenza/internal/config"
	"github.com/nie-crossroads/cadenza/internal/music"
	"github.com/nie-crossroads/cadenza/internal/session"
	"github.com/nie-crossroads/cadenza/web"
)

func main() {
	cfg := config.Load()
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		if err := serveMCP(cfg, newService(cfg)); err != nil {
			log.Fatal(err)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] != "serve" {
		log.Fatal("用法: cadenza [serve|mcp]")
	}
	if err := serveHTTP(cfg); err != nil {
		log.Fatal(err)
	}
}

func newService(cfg config.Config) *music.Service {
	svc := &music.Service{
		Source:        music.NewCatalog(cfg.SourceKey),
		MaxAudioBytes: cfg.MaxAudioBytes,
		MaxCoverBytes: cfg.MaxCoverBytes,
		SearchTimeout: cfg.SearchTimeout,
		FetchTimeout:  cfg.FetchTimeout,
	}
	if cfg.COSReady() {
		store, err := music.NewCOS(cfg.COSSecretID, cfg.COSSecretKey, cfg.COSBucket, cfg.COSRegion, cfg.COSDomain)
		if err != nil {
			log.Printf("对象存储不可用：公开域名过长，地址放不进博客字段")
		} else {
			svc.Store = store
		}
	} else {
		log.Printf("对象存储未配置，添加不可用")
	}
	if cfg.BlogReady() {
		svc.Blog = music.NewBlog(cfg.BlogURL, cfg.BlogToken, cfg.BlogTimeout)
	} else {
		log.Printf("博客接口未配置，添加不可用")
	}
	return svc
}

func serveHTTP(cfg config.Config) error {
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           newHandler(cfg, newService(cfg)),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shut)
	}()
	log.Printf("cadenza 正在监听 %s", cfg.Addr)
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

type app struct {
	svc      *music.Service
	sessions *session.Manager
	password string
	token    string
	pages    http.Handler
}

func newHandler(cfg config.Config, svc *music.Service) http.Handler {
	pages := http.NotFoundHandler()
	if sub, err := fs.Sub(web.Dist, "dist"); err == nil {
		pages = spaHandler(sub)
	}
	a := &app{
		svc:      svc,
		sessions: session.New(cfg.WebPassword, cfg.SessionTTL, cfg.CookieSecure),
		password: cfg.WebPassword,
		token:    cfg.AccessToken,
		pages:    pages,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/session", a.session)
	mux.HandleFunc("POST /api/login", a.login)
	mux.HandleFunc("POST /api/logout", a.logout)
	mux.HandleFunc("GET /api/search", a.search)
	mux.HandleFunc("GET /api/preview", a.preview)
	mux.HandleFunc("HEAD /api/preview", a.preview)
	mux.HandleFunc("POST /api/music/playable", a.playable)
	mux.HandleFunc("POST /api/music", a.add)
	mux.Handle("/", a.pages)
	return mux
}

func (a *app) session(w http.ResponseWriter, r *http.Request) {
	if a.sessions.Valid(r) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeCode(w, http.StatusUnauthorized, "UNAUTHENTICATED")
}

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	if a.password == "" || !a.sessions.Enabled() {
		writeCode(w, http.StatusServiceUnavailable, "LOGIN_UNAVAILABLE")
		return
	}
	if !jsonContent(r) {
		writeCode(w, http.StatusBadRequest, "LOGIN_UNAVAILABLE")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := readJSON(w, r, &body); err != nil {
		writeCode(w, http.StatusBadRequest, "LOGIN_UNAVAILABLE")
		return
	}
	if !sameSecret(body.Password, a.password) {
		writeCode(w, http.StatusUnauthorized, "LOGIN_INVALID")
		return
	}
	a.sessions.Set(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) logout(w http.ResponseWriter, r *http.Request) {
	a.sessions.Clear(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	source := strings.TrimSpace(q.Get("source"))
	offset := queryInt(q.Get("offset"))
	limit := queryInt(q.Get("limit"))
	if limit == 0 {
		if source == "" {
			limit = 5
		} else {
			limit = 30
		}
	}
	items, err := a.svc.SearchQuery(r.Context(), q.Get("q"), source, offset, limit)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		writeFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func queryInt(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func (a *app) preview(w http.ResponseWriter, r *http.Request) {
	status, hdr, body, err := a.svc.Preview(r.Context(), r.URL.Query().Get("sourceId"), r.Header.Get("Range"))
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		code := http.StatusBadGateway
		if errors.Is(err, music.ErrNotFound) {
			code = http.StatusNotFound
		}
		var f *music.Fail
		if errors.As(err, &f) && f.Status != 0 {
			code = f.Status
		}
		w.WriteHeader(code)
		return
	}
	defer body.Close()
	for k, values := range hdr {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(status)
	if r.Method == http.MethodHead {
		return
	}
	buf := make([]byte, 32*1024)
	flusher, _ := w.(http.Flusher)
	for {
		n, readErr := body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if readErr != nil {
			return
		}
	}
}

func (a *app) playable(w http.ResponseWriter, r *http.Request) {
	if !a.authorize(w, r) {
		return
	}
	if !jsonContent(r) {
		writeCode(w, http.StatusBadRequest, music.CodeBlogValidation)
		return
	}
	var body struct {
		SourceID string `json:"sourceId"`
		Quality  string `json:"quality"`
	}
	if err := readJSON(w, r, &body); err != nil || strings.TrimSpace(body.SourceID) == "" {
		writeCode(w, http.StatusBadRequest, music.CodeBlogValidation)
		return
	}
	ok, err := a.svc.Playable(r.Context(), strings.TrimSpace(body.SourceID), body.Quality)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		writeFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"playable": ok})
}

func (a *app) add(w http.ResponseWriter, r *http.Request) {
	if !a.authorize(w, r) {
		return
	}
	if !jsonContent(r) {
		writeCode(w, http.StatusBadRequest, music.CodeBlogValidation)
		return
	}
	var body struct {
		SourceID string `json:"sourceId"`
		Quality  string `json:"quality"`
	}
	if err := readJSON(w, r, &body); err != nil || strings.TrimSpace(body.SourceID) == "" {
		writeCode(w, http.StatusBadRequest, music.CodeBlogValidation)
		return
	}
	quality, ok := music.CanonicalQuality(body.Quality)
	if !ok {
		writeCode(w, http.StatusBadRequest, music.CodeBlogValidation)
		return
	}
	id, err := a.svc.Add(r.Context(), strings.TrimSpace(body.SourceID), quality)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, music.ErrNotFound) {
			writeCode(w, http.StatusBadGateway, music.CodeDownloadFailed)
			return
		}
		writeFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

func (a *app) authorize(w http.ResponseWriter, r *http.Request) bool {
	if got, ok := bearer(r); ok && a.token != "" && sameSecret(got, a.token) {
		return true
	}
	if a.sessions.Valid(r) {
		return true
	}
	if a.sessions.Present(r) {
		writeCode(w, http.StatusUnauthorized, "SESSION_INVALID")
		return false
	}
	writeCode(w, http.StatusUnauthorized, "UNAUTHENTICATED")
	return false
}

func spaHandler(root fs.FS) http.Handler {
	files := http.FS(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}
		f, err := root.Open(name)
		if err != nil {
			if strings.Contains(name, ".") {
				http.NotFound(w, r)
				return
			}
			serveIndex(w, r, files)
			return
		}
		st, err := f.Stat()
		_ = f.Close()
		if err != nil || st.IsDir() {
			if strings.Contains(name, ".") {
				http.NotFound(w, r)
				return
			}
			serveIndex(w, r, files)
			return
		}
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		http.FileServer(files).ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request, files http.FileSystem) {
	w.Header().Set("Cache-Control", "no-cache")
	r2 := r.Clone(r.Context())
	r2.URL.Path = "/"
	http.FileServer(files).ServeHTTP(w, r2)
}

func jsonContent(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(strings.ToLower(ct), "application/json")
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	defer r.Body.Close()
	dec := newDecoder(r.Body)
	return dec.Decode(dst)
}

func bearer(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) < len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(h[len(prefix):])
	if token == "" || strings.ContainsAny(token, "\r\n") {
		return "", false
	}
	return token, true
}

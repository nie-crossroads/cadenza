package music

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// audioRelay 把一首歌完整拉下来，试听的多次 Range 都从这份缓冲里读。
type audioRelay struct {
	mu    sync.Mutex
	data  []byte
	total int64
	ctype string
	err   error
	done  bool
}

func newAudioRelay() *audioRelay {
	return &audioRelay{total: -1}
}

func (r *audioRelay) hasPlayableAudio() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done && r.err != nil && len(r.data) == 0 {
		return false
	}
	n := len(r.data)
	if n == 0 {
		return false
	}
	if n > 32 {
		n = 32
	}
	return playablePrefix(r.data[:n], r.ctype)
}

func (r *audioRelay) dead() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.done && r.err != nil && len(r.data) == 0
}

func (r *audioRelay) fail(err error) {
	r.mu.Lock()
	r.err = err
	r.done = true
	r.mu.Unlock()
}

func (r *audioRelay) setMeta(total int64, ctype string) {
	r.mu.Lock()
	r.total = total
	if ctype != "" {
		r.ctype = ctype
	}
	r.mu.Unlock()
}

func (r *audioRelay) append(p []byte) {
	r.mu.Lock()
	r.data = append(r.data, p...)
	r.mu.Unlock()
}

func (r *audioRelay) finish(err error) {
	r.mu.Lock()
	if err != nil && !errors.Is(err, io.EOF) {
		r.err = err
	}
	if r.total < 0 {
		r.total = int64(len(r.data))
	}
	r.done = true
	r.mu.Unlock()
}

func (r *audioRelay) waitMeta(ctx context.Context) (int64, string, error) {
	err := r.poll(ctx, func() bool {
		return r.total >= 0 || (r.done && r.err != nil && len(r.data) == 0)
	})
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.total < 0 {
		if err == nil {
			err = r.err
		}
		if err == nil {
			err = fail(CodeDownloadFailed, statusOf(CodeDownloadFailed))
		}
		return 0, "", err
	}
	ctype := r.ctype
	if ctype == "" || !strings.HasPrefix(ctype, "audio/") {
		ctype = "audio/mpeg"
	}
	return r.total, ctype, nil
}

func (r *audioRelay) readAt(ctx context.Context, off int64, p []byte, end int64) (int, error) {
	if off >= end {
		return 0, io.EOF
	}
	if int64(len(p)) > end-off {
		p = p[:end-off]
	}
	err := r.poll(ctx, func() bool {
		return int64(len(r.data)) > off || r.done
	})
	if err != nil {
		return 0, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if int64(len(r.data)) <= off {
		if r.err != nil {
			return 0, r.err
		}
		return 0, io.EOF
	}
	n := copy(p, r.data[off:min(int64(len(r.data)), end)])
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (r *audioRelay) poll(ctx context.Context, ready func() bool) error {
	t := time.NewTicker(15 * time.Millisecond)
	defer t.Stop()
	for {
		r.mu.Lock()
		ok := ready()
		r.mu.Unlock()
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

type rangeBody struct {
	relay *audioRelay
	ctx   context.Context
	pos   int64
	end   int64
}

func (b *rangeBody) Read(p []byte) (int, error) {
	n, err := b.relay.readAt(b.ctx, b.pos, p, b.end)
	b.pos += int64(n)
	return n, err
}

func (b *rangeBody) Close() error { return nil }

func (s *Service) relayFor(sourceID string) *audioRelay {
	s.relaysMu.Lock()
	defer s.relaysMu.Unlock()
	if s.relays == nil {
		s.relays = map[string]*audioRelay{}
	}
	if r, ok := s.relays[sourceID]; ok && !r.dead() {
		return r
	}
	r := newAudioRelay()
	s.relays[sourceID] = r
	if len(s.relays) > 2 {
		for id := range s.relays {
			if id != sourceID {
				delete(s.relays, id)
				break
			}
		}
	}
	go s.fillRelay(r, sourceID)
	return r
}

func (s *Service) fillRelay(r *audioRelay, sourceID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	raw, err := s.audioURL(ctx, sourceID)
	if err != nil {
		r.fail(err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		r.fail(fail(CodeDownloadFailed, statusOf(CodeDownloadFailed)))
		return
	}
	req.Header.Set("User-Agent", userAgent)
	if ref := mediaReferer(raw); ref != "" {
		req.Header.Set("Referer", ref)
	}
	resp, err := s.mediaClient().Do(req)
	if err != nil {
		s.dropCache(sourceID)
		log.Printf("试听失败")
		r.fail(fail(CodeDownloadFailed, statusOf(CodeDownloadFailed)))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		s.dropCache(sourceID)
		log.Printf("试听失败")
		r.fail(fail(CodeDownloadFailed, statusOf(CodeDownloadFailed)))
		return
	}
	total := resp.ContentLength
	if total > s.maxAudio() {
		r.fail(fail(CodeTooLarge, statusOf(CodeTooLarge)))
		return
	}
	ctype := resp.Header.Get("Content-Type")
	if idx := strings.Index(ctype, ";"); idx >= 0 {
		ctype = strings.TrimSpace(ctype[:idx])
	}
	if total >= 0 {
		r.setMeta(total, ctype)
	} else {
		r.setMeta(-1, ctype)
	}
	buf := make([]byte, 32*1024)
	var got int64
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			got += int64(n)
			if got > s.maxAudio() {
				r.fail(fail(CodeTooLarge, statusOf(CodeTooLarge)))
				return
			}
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			r.append(chunk)
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				r.finish(nil)
				return
			}
			r.finish(readErr)
			return
		}
	}
}

func sliceRange(header string, total int64) (start, end int64, partial, bad bool) {
	if total < 0 {
		return 0, 0, false, true
	}
	if header == "" {
		return 0, total, false, false
	}
	spec := strings.TrimPrefix(strings.ToLower(header), "bytes=")
	if spec == "" || strings.Contains(spec, ",") {
		return 0, 0, false, true
	}
	left, right, ok := strings.Cut(spec, "-")
	if !ok {
		return 0, 0, false, true
	}
	if left == "" {
		n, err := strconv.ParseInt(right, 10, 64)
		if err != nil || n <= 0 {
			return 0, 0, false, true
		}
		if n > total {
			n = total
		}
		return total - n, total, true, false
	}
	start, err := strconv.ParseInt(left, 10, 64)
	if err != nil || start < 0 || (total > 0 && start >= total) {
		return 0, 0, false, true
	}
	end = total
	if right != "" {
		last, err := strconv.ParseInt(right, 10, 64)
		if err != nil || last < start {
			return 0, 0, false, true
		}
		end = last + 1
		if end > total {
			end = total
		}
	}
	if end < start {
		return 0, 0, false, true
	}
	return start, end, true, false
}

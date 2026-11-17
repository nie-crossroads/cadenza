package music

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"
)

func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() && !ip.IsMulticast() && !ip.IsUnspecified()
}

type guardedDialer struct {
	allowPrivate bool
	dialer       net.Dialer
}

func (g *guardedDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	var last error
	for _, ip := range ips {
		if !g.allowPrivate && !isPublicIP(ip.IP) {
			last = errors.New("blocked address")
			continue
		}
		conn, dialErr := g.dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		last = dialErr
	}
	if last == nil {
		last = errors.New("no address")
	}
	return nil, last
}

func newMediaClient(allowPrivate bool) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&guardedDialer{
			allowPrivate: allowPrivate,
			dialer:       net.Dialer{Timeout: 10 * time.Second},
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          32,
		IdleConnTimeout:       90 * time.Second,
		ReadBufferSize:        256 << 10,
		WriteBufferSize:       256 << 10,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	// Timeout 为 0：试听要跟着整段音频走，由请求的 context 结束，不在这里掐断。
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errors.New("bad scheme")
			}
			return nil
		},
	}
}

func parseRemoteURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
		return nil, errors.New("bad url")
	}
	return u, nil
}

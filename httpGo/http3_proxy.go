/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/quic-go/quic-go/http3"

	"github.com/ruslanBik4/logs"
)

// NewHTTP3Proxy creates an HTTP/3 UDP server that proxies requests to your
// existing fasthttp TCP server.
//
// publicUDPAddr:    ":443"
// cfg.UpstreamAddr: the private fasthttp listener requests are forwarded to,
// "host:port" or ":port" (= 127.0.0.1:port); empty means 127.0.0.1:8080.
// HttpGo itself always passes the address of the loopback listener it opened
// (see listenH3Upstream), so it never depends on a fixed port.
//
// tlsCfg is the SAME config used to create your TCP tlsListener.
// Its GetCertificate callback continues to select the correct certificate by SNI.
func NewHTTP3Proxy(
	h3TLS *tls.Config,
	publicUDPAddr string,
	cfg *HTTP3ProxyConfig,
) (*http3.Server, error) {
	if h3TLS == nil {
		return nil, fmt.Errorf("tls config is required")
	}

	if cfg == nil {
		return nil, fmt.Errorf("http proxy config is required")
	}
	// UDP: HTTP/3 proxy.
	// Point this at a private fasthttp listener.
	// cfg is not modified: it is the user's config (shown by /httpgo/cfg/)
	upstream := cfg.UpstreamAddr
	switch {
	case upstream == "":
		upstream = "127.0.0.1:8080"
	case strings.HasPrefix(upstream, ":"):
		upstream = "127.0.0.1" + upstream
	}

	target, err := url.Parse("http://" + upstream)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL: %w", err)
	}

	logs.StatusLog(target)
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)

			// Preserve the original browser host for multi-site routing.
			pr.Out.Host = pr.In.Host

			pr.SetXForwarded()
			pr.Out.Header.Set("X-Forwarded-Proto", "https")
			pr.Out.Header.Set("X-Forwarded-HTTP-Version", "3")
		},
		Transport: newHTTP3Transport(cfg),
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("X-Httpgo-Transport", "h3-proxy")
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, context.Canceled) ||
			errors.Is(r.Context().Err(), context.Canceled) {
			// Client has already gone away; no response can usefully be sent.
			logs.DebugLog("HTTP/3 client cancelled request: %s %s",
				r.Method, r.URL.String())
			return
		}
		logs.ErrorLog(err)
		http.Error(w, "fasthttp upstream unavailable:"+err.Error(), http.StatusBadGateway)
	}

	h3TLS.MinVersion = tls.VersionTLS13 // HTTP/3 requirement

	return &http3.Server{
		Addr:      publicUDPAddr,
		TLSConfig: http3.ConfigureTLSConfig(h3TLS),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Forwarded-Proto", "https")
			r.Header.Set("X-Forwarded-HTTP-Version", "3")
			proxy.ServeHTTP(w, r)
		}),
	}, nil
}

// listenH3Upstream opens the private TCP listener the HTTP/3 proxy forwards to.
//
// It is always loopback-only: this listener speaks plain HTTP and trusts the
// X-Forwarded-* headers set by the proxy, so it must never be reachable from
// outside. Empty addr picks a free port (unique per process). A plain
// net.Listen (no SO_REUSEPORT) is used deliberately: with reuseport two httpgo
// processes on one host silently shared a fixed port and the kernel spread
// HTTP/3 requests of one app across both.
func listenH3Upstream(addr string) (net.Listener, error) {
	switch {
	case addr == "":
		addr = "127.0.0.1:0"
	case strings.HasPrefix(addr, ":"):
		addr = "127.0.0.1" + addr
	}

	return net.Listen("tcp4", addr)
}

// h3AltSvc builds the Alt-Svc header value that announces HTTP/3 on the port
// of addr. Only the port is advertised (the browser keeps the same host), so
// ":443", "0.0.0.0:443" and "[::]:443" all give `h3=":443"`.
func h3AltSvc(addr string) (string, error) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	if port == "" || port == "0" {
		return "", fmt.Errorf("no port to advertise in %q", addr)
	}

	return fmt.Sprintf(`h3=":%s"; ma=86400`, port), nil
}

type HTTP3ProxyConfig struct {
	UpstreamAddr          string        `yaml:"UpstreamAddr" json:"UpstreamAddr,omitempty"`
	IdleConnTimeout       time.Duration `yaml:"IdleConnTimeout" json:"IdleConnTimeout,omitempty"`
	ResponseHeaderTimeout time.Duration `yaml:"ResponseHeaderTimeout" json:"ResponseHeaderTimeout,omitempty"`
	MaxIdleConns          int           `yaml:"MaxIdleConns" json:"MaxIdleConns,omitempty"`
	MaxIdleConnsPerHost   int           `yaml:"MaxIdleConnsPerHost" json:"MaxIdleConnsPerHost,omitempty"`
	MaxConnsPerHost       int           `yaml:"MaxConnsPerHost" json:"MaxConnsPerHost,omitempty"`
}

func newHTTP3Transport(c *HTTP3ProxyConfig) *http.Transport {
	idleTimeout := c.IdleConnTimeout
	if idleTimeout == 0 {
		idleTimeout = 90 * time.Second
	}

	return &http.Transport{
		// Keep the local fasthttp upstream on HTTP/1.1.
		// It is private and avoids requiring net/http HTTP/2 support here.
		ForceAttemptHTTP2: false,

		// A non-nil map disables Go's automatic HTTP/2 support.
		TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},

		MaxIdleConns:          max(c.MaxIdleConns, 100),
		MaxIdleConnsPerHost:   max(c.MaxIdleConnsPerHost, 100),
		MaxConnsPerHost:       c.MaxConnsPerHost, // 0 = unlimited
		IdleConnTimeout:       idleTimeout,
		ResponseHeaderTimeout: c.ResponseHeaderTimeout, // 0 = disabled
	}
}

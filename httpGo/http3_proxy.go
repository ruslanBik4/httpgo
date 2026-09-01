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
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/quic-go/quic-go/http3"

	"github.com/ruslanBik4/logs"
)

// NewHTTP3Proxy creates an HTTP/3 UDP server that proxies requests to your
// existing fasthttp TCP server.
//
// publicUDPAddr: ":443"
// upstream:       "http://127.0.0.1:8080"  (recommended private fasthttp port)
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
	upstream := cfg.UpstreamAddr
	if upstream == "" {
		upstream = "http://127.0.0.1:8080"
		cfg.UpstreamAddr = ":8080"
	} else {
		upstream = "http://127.0.0.1" + upstream
	}

	target, err := url.Parse(upstream)
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

/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"crypto/tls"
	"fmt"
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
// publicUDPAddr: ":443"
// upstream:       "http://127.0.0.1:8080"  (recommended private fasthttp port)
//
// tlsCfg is the SAME config used to create your TCP tlsListener.
// Its GetCertificate callback continues to select the correct certificate by SNI.
func NewHTTP3Proxy(
	h3TLS *tls.Config,
	publicUDPAddr string,
	upstream string,
) (*http3.Server, error) {
	if h3TLS == nil {
		return nil, fmt.Errorf("tls config is required")
	}

	if strings.HasPrefix(upstream, ":") {
		upstream = "http://127.0.0.1" + upstream
	}

	target, err := url.Parse(upstream)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		// Keep the local fasthttp upstream on HTTP/1.1.
		// It is private and avoids requiring net/http HTTP/2 support here.
		ForceAttemptHTTP2: false,
		// A non-nil map disables Go's automatic HTTP/2 support.
		TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},

		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
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

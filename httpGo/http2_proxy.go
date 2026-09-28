/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/httpGo/http2"
	"github.com/ruslanBik4/logs"
)

// H2TLSProto is the ALPN id; fasthttp hands the conn to our HTTP/2 server
// via NextProto when "h2" is negotiated during the TLS handshake.
const H2TLSProto = "h2"

// NewHTTP2Server builds the HTTP/2 server from cfg and registers it on
// cfg.Server for ALPN "h2". Returns nil when HTTP/2 is not configured.
//
// It may be called before NewHttpgo finishes wrapping cfg.Server.Handler
// (domains, Alt-Svc, IP blocking): the handler is resolved per request, so
// HTTP/2 requests always go through the same final chain as HTTP/1.1 ones.
// Call it after cfg.Server.Logger is set so HTTP/2 logs go to the same place.
func NewHTTP2Server(cfg *CfgHttp) *http2.Server {
	if cfg == nil || cfg.HTTP2 == nil || cfg.Server == nil {
		return nil
	}

	srv := cfg.Server
	h2 := http2.NewServer(func(ctx *fasthttp.RequestCtx) { srv.Handler(ctx) }, HTTP2Config(cfg))
	srv.NextProto(H2TLSProto, h2.ServeConn)

	logs.StatusLog("HTTP/2 enabled (httpgo/http2)")
	return h2
}

// HTTP2Config converts CfgHttp into http2.Config. Values set explicitly in
// cfg.HTTP2 win; zero values are inherited from cfg.Server, so HTTP/2 enforces
// the same limits as HTTP/1.1. cfg.HTTP2 itself is not modified.
func HTTP2Config(cfg *CfgHttp) *http2.Config {
	c := *cfg.HTTP2
	s := cfg.Server

	if c.MaxRequestBodySize == 0 {
		c.MaxRequestBodySize = s.MaxRequestBodySize
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = s.IdleTimeout
		if c.IdleTimeout == 0 {
			c.IdleTimeout = s.ReadTimeout
		}
	}
	if c.Logger == nil {
		c.Logger = s.Logger
	}
	if c.ServerName == "" && !s.NoDefaultServerHeader {
		c.ServerName = s.Name
	}
	c.ReduceMemoryUsage = c.ReduceMemoryUsage || s.ReduceMemoryUsage

	return &c
}

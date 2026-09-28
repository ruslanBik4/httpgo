/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"time"

	"github.com/valyala/fasthttp"
)

// Config tunes the HTTP/2 server side. Zero values fall back to defaults.
type Config struct {
	// PingInterval: send a keep-alive PING after this much read silence; if it
	// isn't answered within another interval the connection is closed.
	// 0 = default (10s, same as domsolutions/http2), negative = disabled.
	PingInterval time.Duration
	// Debug logs connection/stream errors through Logger.
	Debug bool
	// ServerName is sent as the "server" response header unless the handler sets one.
	ServerName string
	// MaxConcurrentStreams advertised to the client (SETTINGS_MAX_CONCURRENT_STREAMS).
	MaxConcurrentStreams uint32
	// InitialWindowSize is the per-stream receive window (SETTINGS_INITIAL_WINDOW_SIZE).
	InitialWindowSize uint32
	// ConnWindowSize is the connection-level receive window.
	ConnWindowSize uint32
	// MaxFrameSize we accept (SETTINGS_MAX_FRAME_SIZE), 16KiB..16MiB.
	MaxFrameSize uint32
	// MaxHeaderListSize bounds decoded request headers (SETTINGS_MAX_HEADER_LIST_SIZE).
	MaxHeaderListSize uint32
	// MaxRequestBodySize; larger bodies get 413. Matches fasthttp.Server semantics.
	MaxRequestBodySize int
	// IdleTimeout closes the connection when no frames arrive for this long.
	IdleTimeout time.Duration
	// Logger is passed to fasthttp.RequestCtx (ctx.Logger()).
	Logger fasthttp.Logger `json:"-" yaml:"-"`
	// DisablePush turns server push off even if the client allows it.
	DisablePush bool
	// ReduceMemoryUsage mirrors fasthttp.Server.ReduceMemoryUsage for ctx buffers.
	ReduceMemoryUsage bool
}

const (
	defaultMaxConcurrentStreams = 250
	defaultInitialWindowSize    = 1 << 20 // 1 MiB per stream
	defaultConnWindowSize       = 1 << 24 // 16 MiB per connection
	defaultMaxHeaderListSize    = 1 << 20
	defaultMaxRequestBodySize   = 4 << 20
	defaultIdleTimeout          = 3 * time.Minute
	defaultPingInterval         = 10 * time.Second
)

func (c *Config) withDefaults() Config {
	var cfg Config
	if c != nil {
		cfg = *c
	}
	if cfg.MaxConcurrentStreams == 0 {
		cfg.MaxConcurrentStreams = defaultMaxConcurrentStreams
	}
	if cfg.InitialWindowSize == 0 {
		cfg.InitialWindowSize = defaultInitialWindowSize
	}
	cfg.InitialWindowSize = min(cfg.InitialWindowSize, maxWindowSize)
	if cfg.ConnWindowSize == 0 {
		cfg.ConnWindowSize = defaultConnWindowSize
	}
	cfg.ConnWindowSize = min(max(cfg.ConnWindowSize, defaultWindowSize), maxWindowSize)
	if cfg.MaxFrameSize == 0 {
		cfg.MaxFrameSize = minMaxFrameSize
	}
	cfg.MaxFrameSize = min(max(cfg.MaxFrameSize, minMaxFrameSize), maxMaxFrameSize)
	if cfg.MaxHeaderListSize == 0 {
		cfg.MaxHeaderListSize = defaultMaxHeaderListSize
	}
	if cfg.MaxRequestBodySize <= 0 {
		cfg.MaxRequestBodySize = defaultMaxRequestBodySize
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = defaultIdleTimeout
	}
	if cfg.PingInterval == 0 {
		cfg.PingInterval = defaultPingInterval
	}
	return cfg
}

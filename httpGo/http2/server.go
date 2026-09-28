/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// Server tracks live HTTP/2 connections so they can be shut down gracefully.
//
// fasthttp.Server.Shutdown waits for every open connection to close, but it
// cannot close connections handed over via NextProto — call Server.Shutdown
// first (or concurrently), otherwise idle h2 connections keep it waiting.
type Server struct {
	Handler fasthttp.RequestHandler
	Config  *Config

	mu       sync.Mutex
	conns    map[*Conn]struct{}
	stopping bool
}

// NewServer creates a Server; cfg may be nil.
func NewServer(handler fasthttp.RequestHandler, cfg *Config) *Server {
	return &Server{Handler: handler, Config: cfg, conns: make(map[*Conn]struct{})}
}

// ServeConn serves one negotiated "h2" connection. Plug it into fasthttp:
//
//	srv.NextProto("h2", h2srv.ServeConn)
func (s *Server) ServeConn(c net.Conn) error {
	conn := NewConnConfig(c, s.Handler, s.Config)
	s.mu.Lock()
	if s.stopping {
		s.mu.Unlock()
		conn.Shutdown()
	} else {
		s.conns[conn] = struct{}{}
		s.mu.Unlock()
	}
	defer func() {
		s.mu.Lock()
		delete(s.conns, conn)
		s.mu.Unlock()
	}()
	return conn.Serve()
}

// Shutdown sends GOAWAY on all connections and waits until they finish
// in-flight requests or ctx expires.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.stopping = true
	for c := range s.conns {
		c.Shutdown()
	}
	s.mu.Unlock()

	t := time.NewTicker(20 * time.Millisecond)
	defer t.Stop()
	for {
		s.mu.Lock()
		n := len(s.conns)
		s.mu.Unlock()
		if n == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

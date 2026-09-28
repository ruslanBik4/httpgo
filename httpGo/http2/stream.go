/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"sync"

	"github.com/valyala/fasthttp"
)

// stream is one HTTP/2 request/response exchange.
//
// Ownership rules (this is what keeps the package race-free):
//   - While in Conn.pending, the stream and its ctx are owned by the read loop.
//   - After END_STREAM it leaves pending and is owned by the handler goroutine;
//     the read loop then only touches the fields guarded by Conn.mu.
//   - The handler goroutine returns the stream to the pool when it's done.
type stream struct {
	ctx fasthttp.RequestCtx

	id uint32
	c  *Conn

	// --- handler goroutine owned ---
	pushed     bool // server-initiated (promised) stream
	responding bool // response is being written: no more PUSH_PROMISE

	// --- read-loop owned (while pending) ---
	declaredCL  int64
	recvWindow  int64
	recvUnacked uint32

	// --- guarded by Conn.mu ---
	sendWindow int64
	reset      bool // client sent RST_STREAM or conn is gone
}

var streamPool = sync.Pool{New: func() any { return new(stream) }}

func (c *Conn) acquireStream(id uint32) *stream {
	s := streamPool.Get().(*stream)
	s.id = id
	s.c = c
	s.recvWindow = int64(c.cfg.InitialWindowSize)
	s.ctx.Init2(c.conn, c.cfg.Logger, c.cfg.ReduceMemoryUsage)
	return s
}

func releaseStream(s *stream) {
	s.ctx.Request.Reset()
	s.ctx.Response.Reset()
	s.ctx.ResetUserValues()
	s.id = 0
	s.c = nil
	s.pushed, s.responding = false, false
	s.declaredCL = -1
	s.recvWindow, s.recvUnacked = 0, 0
	s.sendWindow, s.reset = 0, false
	streamPool.Put(s)
}

// decodeState collects one header block into a request (read loop only).
type decodeState struct {
	req            *fasthttp.Request
	size           uint32
	contentLength  int64
	scheme         string
	trailers       bool
	discard        bool
	tooLarge       bool
	malformed      bool
	regularSeen    bool
	hasMethod      bool
	hasPath        bool
	hasScheme      bool
	hasAuthority   bool
	isConnect      bool
	expectContinue bool
}

func (d *decodeState) reset(req *fasthttp.Request, trailers, discard bool) {
	*d = decodeState{req: req, trailers: trailers, discard: discard, contentLength: -1}
}

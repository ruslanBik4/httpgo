/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"bytes"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

// runHandler executes the fasthttp handler for a fully received request and
// writes the response. It runs in its own goroutine and owns s.
func (c *Conn) runHandler(s *stream) {
	ctx := &s.ctx
	defer func() {
		if r := recover(); r != nil {
			c.logf("http2: panic in handler for %s %s: %v\n%s", ctx.Method(), ctx.RequestURI(), r, debug.Stack())
			_ = c.writeFrames(func(fr *Framer) error { return fr.WriteRSTStream(s.id, ErrCodeInternal) })
		}
		c.mu.Lock()
		delete(c.streams, s.id)
		if s.pushed {
			c.pushActive--
		}
		c.mu.Unlock()
		releaseStream(s)
	}()

	c.handler(ctx)
	s.responding = true // Push is no longer allowed on this stream

	if err := c.writeResponse(s); err != nil && err != errStreamClosed && err != errConnClosed {
		_ = c.writeFrames(func(fr *Framer) error { return fr.WriteRSTStream(s.id, ErrCodeInternal) })
	}
}

// writeResponse sends HEADERS + DATA frames back to the client.
func (c *Conn) writeResponse(s *stream) error {
	ctx := &s.ctx
	resp := &ctx.Response

	status := resp.StatusCode()
	noBody := ctx.IsHead() || status < 200 || status == fasthttp.StatusNoContent || status == fasthttp.StatusNotModified
	isStream := resp.IsBodyStream()

	var body []byte
	if noBody {
		if isStream {
			_ = resp.CloseBodyStream()
			isStream = false
		}
	} else if !isStream {
		body = resp.Body()
	}
	headersOnly := !isStream && len(body) == 0

	// HEADERS (and, if the window allows, the whole body) in one flush.
	c.writeMu.Lock()
	if err := c.checkWritable(s); err != nil {
		c.writeMu.Unlock()
		return err
	}
	c.encodeResponseHeaders(ctx, status, isStream, len(body))
	fr := c.framer
	maxFrame := c.peerMaxFrameSize.Load()
	if err := fr.WriteHeaders(s.id, headersOnly, c.henc.Bytes(), maxFrame); err != nil {
		c.writeMu.Unlock()
		return err
	}
	// Fast path: take whatever window is available without blocking.
	for len(body) > 0 {
		n := c.tryTakeWindow(s, len(body))
		if n == 0 {
			break
		}
		end := n == len(body)
		if err := fr.WriteData(s.id, end, body[:n]); err != nil {
			c.writeMu.Unlock()
			return err
		}
		body = body[n:]
	}
	err := fr.Flush()
	c.writeMu.Unlock()
	if err != nil || headersOnly {
		return err
	}

	if !isStream {
		if len(body) == 0 {
			return nil // already sent with END_STREAM in the fast path
		}
		return c.writeData(s, body, true)
	}

	// Streaming body (SetBodyStream / SetBodyStreamWriter): flush every chunk,
	// so SSE / long-poll style handlers see their data delivered promptly.
	if err := resp.BodyWriteTo(&streamWriter{c: c, s: s}); err != nil {
		return err
	}
	return c.writeData(s, nil, true)
}

func (c *Conn) checkWritable(s *stream) error {
	if c.closed.Load() {
		return errConnClosed
	}
	c.mu.Lock()
	reset := s.reset
	c.mu.Unlock()
	if reset {
		return errStreamClosed
	}
	return nil
}

// encodeResponseHeaders fills c.henc; caller holds writeMu.
func (c *Conn) encodeResponseHeaders(ctx *fasthttp.RequestCtx, status int, isStream bool, bodyLen int) {
	e := c.henc
	e.Reset()
	e.WriteField(":status", statusText(status))
	c.sawDate, c.sawServer = false, false
	ctx.Response.Header.VisitAll(c.visitFn)
	if !c.sawDate {
		e.WriteField("date", httpDate())
	}
	if !c.sawServer && c.cfg.ServerName != "" {
		e.WriteField("server", c.cfg.ServerName)
	}
	switch {
	case status < 200 || status == fasthttp.StatusNoContent || status == fasthttp.StatusNotModified:
	case ctx.IsHead():
		if cl := ctx.Response.Header.ContentLength(); cl >= 0 {
			e.WriteField("content-length", strconv.Itoa(cl))
		}
	case !isStream:
		e.WriteField("content-length", strconv.Itoa(bodyLen))
	}
}

// visitRespHeader encodes one fasthttp response header (caller holds writeMu).
func (c *Conn) visitRespHeader(k, v []byte) {
	switch len(k) {
	case 4:
		if bytes.EqualFold(k, strDate) {
			c.sawDate = true
		}
	case 6:
		if bytes.EqualFold(k, strServer) {
			c.sawServer = true
		}
	case 7:
		if bytes.EqualFold(k, strUpgrade) {
			return
		}
	case 10:
		if bytes.EqualFold(k, strConnection) || bytes.EqualFold(k, strKeepAlive) {
			return
		}
	case 14:
		if bytes.EqualFold(k, strContentLength) {
			return // computed from the actual body
		}
	case 16:
		if bytes.EqualFold(k, strProxyConn) {
			return
		}
	case 17:
		if bytes.EqualFold(k, strTransferEnc) {
			return
		}
	}
	c.henc.WriteFieldBytes(k, v)
}

var (
	strDate          = []byte("Date")
	strServer        = []byte("Server")
	strUpgrade       = []byte("Upgrade")
	strConnection    = []byte("Connection")
	strKeepAlive     = []byte("Keep-Alive")
	strContentLength = []byte("Content-Length")
	strProxyConn     = []byte("Proxy-Connection")
	strTransferEnc   = []byte("Transfer-Encoding")
)

// writeData sends p as DATA frames, blocking on flow control as needed.
func (c *Conn) writeData(s *stream, p []byte, endStream bool) error {
	for {
		n := 0
		if len(p) > 0 {
			var err error
			if n, err = c.takeWindow(s, len(p)); err != nil {
				return err
			}
		}
		last := n == len(p)
		c.writeMu.Lock()
		if c.closed.Load() {
			c.writeMu.Unlock()
			return errConnClosed
		}
		err := c.framer.WriteData(s.id, endStream && last, p[:n])
		if err == nil {
			err = c.framer.Flush()
		}
		c.writeMu.Unlock()
		if err != nil {
			return err
		}
		p = p[n:]
		if last {
			return nil
		}
	}
}

// tryTakeWindow reserves up to want bytes of send window without blocking.
func (c *Conn) tryTakeWindow(s *stream, want int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s.reset {
		return 0
	}
	return c.takeLocked(s, want)
}

// takeWindow reserves 1..want bytes of send window, waiting for WINDOW_UPDATE.
func (c *Conn) takeWindow(s *stream, want int) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for {
		if c.closed.Load() {
			return 0, errConnClosed
		}
		if s.reset {
			return 0, errStreamClosed
		}
		if n := c.takeLocked(s, want); n > 0 {
			return n, nil
		}
		c.cond.Wait()
	}
}

func (c *Conn) takeLocked(s *stream, want int) int {
	n := min(int64(want), c.sendWindow, s.sendWindow, int64(c.peerMaxFrameSize.Load()))
	if n <= 0 {
		return 0
	}
	c.sendWindow -= n
	s.sendWindow -= n
	return int(n)
}

// streamWriter adapts a stream to io.Writer for Response.BodyWriteTo.
type streamWriter struct {
	c *Conn
	s *stream
}

func (w *streamWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if err := w.c.writeData(w.s, p, false); err != nil {
		return 0, err
	}
	return len(p), nil
}

// ---------------------------------------------------------------------------
// small caches
// ---------------------------------------------------------------------------

var statusStrings = func() (a [600]string) {
	for i := 100; i < len(a); i++ {
		a[i] = strconv.Itoa(i)
	}
	return
}()

func statusText(code int) string {
	if code == 0 {
		code = fasthttp.StatusOK
	}
	if code >= 100 && code < len(statusStrings) {
		return statusStrings[code]
	}
	return strconv.Itoa(code)
}

type dateEntry struct {
	sec int64
	s   string
}

var dateCache atomic.Pointer[dateEntry]

// httpDate returns the current time in IMF-fixdate, recomputed once per second.
func httpDate() string {
	now := time.Now()
	sec := now.Unix()
	if d := dateCache.Load(); d != nil && d.sec == sec {
		return d.s
	}
	d := &dateEntry{sec: sec, s: string(fasthttp.AppendHTTPDate(nil, now))}
	dateCache.Store(d)
	return d.s
}

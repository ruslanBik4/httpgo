/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"errors"
	"strings"

	"github.com/valyala/fasthttp"
)

// Server push (RFC 9113 §8.4).
//
// Usage inside any fasthttp handler — safe to call for HTTP/1.x requests too:
//
//	func index(ctx *fasthttp.RequestCtx) {
//		_ = http2.Push(ctx, "/static/app.css", nil)
//		_ = http2.Push(ctx, "/static/app.js", nil)
//		ctx.SetContentType("text/html")
//		ctx.SetBody(page)
//	}
//
// Push must be called from the handler goroutine before it returns (i.e.
// before the response is written). The pushed resource is produced by running
// the same fasthttp handler for a synthetic GET request, in its own goroutine.

var (
	// ErrPushNotSupported: the request is not HTTP/2 (or the ctx user values were reset).
	ErrPushNotSupported = errors.New("http2: push not supported for this request")
	// ErrPushDisabled: the client sent SETTINGS_ENABLE_PUSH=0, push is disabled in Config,
	// or the connection is going away.
	ErrPushDisabled = errors.New("http2: push disabled by client")
	// ErrPushLimit: the client's SETTINGS_MAX_CONCURRENT_STREAMS is reached.
	ErrPushLimit = errors.New("http2: too many pushed streams")
	// ErrPushNotAllowed: pushing from a pushed stream, or after the response started.
	ErrPushNotAllowed = errors.New("http2: push not allowed on this stream")
	// ErrPushBadTarget: target must be a path ("/x") or an absolute URL of the same origin.
	ErrPushBadTarget = errors.New("http2: invalid push target")
	// ErrPushBadMethod: only safe, cacheable methods without body (GET, HEAD) may be pushed.
	ErrPushBadMethod = errors.New("http2: push method must be GET or HEAD")
)

type pushKey struct{}

// PushOptions describes a pushed request.
type PushOptions struct {
	// Method is GET (default) or HEAD.
	Method string
	// Header holds extra request headers for the promised request (keys are lower-cased).
	Header map[string]string
}

// pushInheritHeaders are copied from the parent request so the pushed response
// matches what the client would have asked for (content negotiation, auth cookies).
var pushInheritHeaders = [...]string{"accept-encoding", "accept-language", "user-agent", "cookie", "authorization"}

// IsPushSupported reports whether Push can currently succeed for ctx.
func IsPushSupported(ctx *fasthttp.RequestCtx) bool {
	s, ok := ctx.UserValue(pushKey{}).(*stream)
	return ok && s.c != nil && s.c.canPush() == nil
}

// Push sends PUSH_PROMISE for target on ctx's stream and schedules the pushed response.
func Push(ctx *fasthttp.RequestCtx, target string, opts *PushOptions) error {
	s, ok := ctx.UserValue(pushKey{}).(*stream)
	if !ok || s.c == nil || &s.ctx != ctx {
		return ErrPushNotSupported
	}
	return s.c.push(s, target, opts)
}

func (c *Conn) canPush() error {
	if !c.peerEnablePush.Load() || c.closed.Load() || c.draining.Load() {
		return ErrPushDisabled
	}
	return nil
}

func (c *Conn) push(parent *stream, target string, opts *PushOptions) error {
	if err := c.canPush(); err != nil {
		return err
	}
	if parent.pushed || parent.responding {
		return ErrPushNotAllowed
	}

	method := fasthttp.MethodGet
	if opts != nil && opts.Method != "" {
		method = strings.ToUpper(opts.Method)
	}
	if method != fasthttp.MethodGet && method != fasthttp.MethodHead {
		return ErrPushBadMethod
	}

	pctx := &parent.ctx
	scheme := string(pctx.URI().Scheme())
	authority := string(pctx.Host())
	path, err := pushPath(target, scheme, authority)
	if err != nil {
		return err
	}

	// Build the promised request (outside any lock).
	child := c.acquireStream(0)
	child.pushed = true
	req := &child.ctx.Request
	req.Header.SetMethod(method)
	req.Header.SetRequestURI(path)
	req.Header.SetHost(authority)
	req.Header.SetProtocol("HTTP/2.0")
	req.URI().SetScheme(scheme)
	for _, h := range pushInheritHeaders {
		if v := pctx.Request.Header.Peek(h); len(v) > 0 {
			req.Header.SetBytesV(h, v)
		}
	}
	if opts != nil {
		for k, v := range opts.Header {
			req.Header.Set(strings.ToLower(k), v)
		}
	}

	// Allocate the promised id, register the stream and write PUSH_PROMISE in
	// one writeMu section: ids hit the wire in increasing order and the HPACK
	// encoder state stays consistent.
	c.writeMu.Lock()
	if c.closed.Load() {
		c.writeMu.Unlock()
		releaseStream(child)
		return errConnClosed
	}
	c.mu.Lock()
	switch {
	case parent.reset:
		err = errStreamClosed
	case c.pushActive >= c.peerMaxConcurrent.Load():
		err = ErrPushLimit
	case c.lastPushID.Load() >= maxWindowSize-1: // stream ids exhausted
		err = ErrPushDisabled
	}
	if err != nil {
		c.mu.Unlock()
		c.writeMu.Unlock()
		releaseStream(child)
		return err
	}
	id := c.lastPushID.Load() + 2
	child.id = id
	child.sendWindow = c.peerInitialWindow
	c.streams[id] = child
	c.pushActive++
	c.lastPushID.Store(id)
	c.mu.Unlock()

	e := c.henc
	e.Reset()
	e.WriteField(":method", method)
	e.WriteField(":scheme", scheme)
	e.WriteField(":authority", authority)
	e.WriteField(":path", path)
	req.Header.VisitAll(func(k, v []byte) {
		if len(k) == 4 && strings.EqualFold(string(k), "host") {
			return // carried by :authority
		}
		e.WriteFieldBytes(k, v)
	})
	err = c.framer.WritePushPromise(parent.id, id, e.Bytes(), c.peerMaxFrameSize.Load())
	if err == nil {
		err = c.framer.Flush()
	}
	c.writeMu.Unlock()

	if err != nil {
		c.mu.Lock()
		delete(c.streams, id)
		c.pushActive--
		c.mu.Unlock()
		releaseStream(child)
		return err
	}

	go c.runHandler(child) // stream is now "reserved (local)"; the handler's HEADERS open it
	return nil
}

// pushPath validates target and returns the :path to promise.
func pushPath(target, scheme, authority string) (string, error) {
	if strings.HasPrefix(target, "/") && !strings.HasPrefix(target, "//") {
		return target, nil
	}
	// absolute URL: only same origin can be pushed on this connection
	var u fasthttp.URI
	if err := u.Parse(nil, []byte(target)); err != nil {
		return "", ErrPushBadTarget
	}
	if string(u.Scheme()) != scheme || !strings.EqualFold(string(u.Host()), authority) {
		return "", ErrPushBadTarget
	}
	return string(u.RequestURI()), nil
}

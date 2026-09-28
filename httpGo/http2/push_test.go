/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"crypto/tls"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	xhttp2 "golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

// rawClient is a minimal HTTP/2 client that accepts pushes
// (Go's own client always sends ENABLE_PUSH=0).
type rawClient struct {
	t   *testing.T
	c   *tls.Conn
	fr  *xhttp2.Framer
	dec *hpack.Decoder
	enc *hpack.Encoder
	buf []byte
}

func dialRaw(t *testing.T, addr string, settings ...xhttp2.Setting) *rawClient {
	c, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"h2"}})
	if err != nil {
		t.Fatal(err)
	}
	_ = c.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := c.Write([]byte(clientPreface)); err != nil {
		t.Fatal(err)
	}
	rc := &rawClient{t: t, c: c, fr: xhttp2.NewFramer(c, c), dec: hpack.NewDecoder(4096, nil)}
	rc.fr.MaxHeaderListSize = 1 << 20
	if err := rc.fr.WriteSettings(settings...); err != nil {
		t.Fatal(err)
	}
	return rc
}

func (rc *rawClient) get(streamID uint32, path, host string) {
	var b = new(bytesBuffer)
	enc := hpack.NewEncoder(b)
	for _, f := range [][2]string{{":method", "GET"}, {":scheme", "https"}, {":authority", host}, {":path", path}, {"accept-encoding", "gzip"}} {
		_ = enc.WriteField(hpack.HeaderField{Name: f[0], Value: f[1]})
	}
	if err := rc.fr.WriteHeaders(xhttp2.HeadersFrameParam{StreamID: streamID, BlockFragment: b.b, EndStream: true, EndHeaders: true}); err != nil {
		rc.t.Fatal(err)
	}
}

type bytesBuffer struct{ b []byte }

func (b *bytesBuffer) Write(p []byte) (int, error) { b.b = append(b.b, p...); return len(p), nil }

type exchange struct {
	status  string
	headers map[string]string
	body    string
	done    bool
	reset   bool
}

type pushResult struct {
	promises   map[uint32]map[string]string // promisedID -> request headers
	promiseOrd []uint32
	streams    map[uint32]*exchange
	events     []string
}

func (rc *rawClient) decode(block []byte) map[string]string {
	fields, err := rc.dec.DecodeFull(block)
	if err != nil {
		rc.t.Fatal(err)
	}
	m := map[string]string{}
	for _, f := range fields {
		m[f.Name] = f.Value
	}
	return m
}

// readUntil reads frames until every stream in want is finished (END_STREAM or RST).
func (rc *rawClient) readUntil(res *pushResult, done func(*pushResult) bool, onPromise func(id uint32)) {
	if res.promises == nil {
		res.promises, res.streams = map[uint32]map[string]string{}, map[uint32]*exchange{}
	}
	ex := func(id uint32) *exchange {
		if res.streams[id] == nil {
			res.streams[id] = &exchange{}
		}
		return res.streams[id]
	}
	for !done(res) {
		f, err := rc.fr.ReadFrame()
		if err != nil {
			rc.t.Fatalf("read: %v (events %v)", err, res.events)
		}
		switch f := f.(type) {
		case *xhttp2.SettingsFrame:
			if !f.IsAck() {
				_ = rc.fr.WriteSettingsAck()
			}
		case *xhttp2.PushPromiseFrame:
			h := rc.decode(f.HeaderBlockFragment())
			res.promises[f.PromiseID] = h
			res.promiseOrd = append(res.promiseOrd, f.PromiseID)
			res.events = append(res.events, fmt.Sprintf("promise:%d on %d", f.PromiseID, f.StreamID))
			if onPromise != nil {
				onPromise(f.PromiseID)
			}
		case *xhttp2.HeadersFrame:
			h := rc.decode(f.HeaderBlockFragment())
			e := ex(f.StreamID)
			e.status, e.headers = h[":status"], h
			e.done = e.done || f.StreamEnded()
			res.events = append(res.events, fmt.Sprintf("headers:%d", f.StreamID))
		case *xhttp2.DataFrame:
			e := ex(f.StreamID)
			e.body += string(f.Data())
			e.done = e.done || f.StreamEnded()
			if n := uint32(len(f.Data())); n > 0 {
				_ = rc.fr.WriteWindowUpdate(0, n)
				_ = rc.fr.WriteWindowUpdate(f.StreamID, n)
			}
		case *xhttp2.RSTStreamFrame:
			ex(f.StreamID).reset = true
		case *xhttp2.GoAwayFrame:
			rc.t.Fatalf("unexpected GOAWAY %v", f.ErrCode)
		}
	}
}

func finished(ids ...uint32) func(*pushResult) bool {
	return func(r *pushResult) bool {
		for _, id := range ids {
			if e := r.streams[id]; e == nil || !(e.done || e.reset) {
				return false
			}
		}
		return true
	}
}

func pushHandler(slowAssets bool) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/index":
			e1 := Push(ctx, "/style.css", nil)
			e2 := Push(ctx, "https://"+string(ctx.Host())+"/app.js?v=1", &PushOptions{Header: map[string]string{"X-Pushed": "yes"}})
			e3 := Push(ctx, "/x", &PushOptions{Method: "POST"})
			e4 := Push(ctx, "https://evil.example/a.js", nil)
			ctx.Response.Header.Set("X-Push-Err", fmt.Sprint(e1, "|", e2, "|", e3, "|", e4))
			ctx.SetBodyString("index")
		case "/style.css":
			if slowAssets {
				time.Sleep(300 * time.Millisecond)
			}
			ctx.Response.Header.Set("X-AE", string(ctx.Request.Header.Peek("Accept-Encoding")))
			ctx.Response.Header.Set("X-Nested", fmt.Sprint(Push(ctx, "/nested", nil)))
			ctx.SetContentType("text/css")
			ctx.SetBodyString("body{}")
		case "/app.js":
			ctx.Response.Header.Set("X-Got", string(ctx.Request.Header.Peek("X-Pushed"))+"|"+string(ctx.QueryArgs().Peek("v")))
			ctx.SetBodyString("js")
		case "/big":
			_ = Push(ctx, "/huge", nil)
			ctx.SetBodyString("big")
		case "/huge":
			ctx.SetBody(make([]byte, 20<<20))
		default:
			ctx.SetBodyString("ok")
		}
	}
}

func TestPush(t *testing.T) {
	addr, stop := startServer(t, pushHandler(false), nil)
	defer stop()
	rc := dialRaw(t, addr)
	defer rc.c.Close()
	rc.get(1, "/index", addr)

	var res pushResult
	rc.readUntil(&res, finished(1, 2, 4), nil)

	if got := res.promiseOrd; len(got) != 2 || got[0] != 2 || got[1] != 4 {
		t.Fatalf("promised ids %v, events %v", got, res.events)
	}
	if p := res.promises[2]; p[":path"] != "/style.css" || p[":method"] != "GET" || p[":authority"] != addr || p["accept-encoding"] != "gzip" {
		t.Errorf("promise 2 headers %v", p)
	}
	if p := res.promises[4]; p[":path"] != "/app.js?v=1" || p["x-pushed"] != "yes" {
		t.Errorf("promise 4 headers %v", p)
	}
	// PUSH_PROMISE must precede the parent's response (§8.4: before referencing content).
	if res.events[0] != "promise:2 on 1" || res.events[1] != "promise:4 on 1" {
		t.Errorf("order: %v", res.events)
	}
	if e := res.streams[1]; e.body != "index" ||
		e.headers["x-push-err"] != fmt.Sprint(nil, "|", nil, "|", ErrPushBadMethod, "|", ErrPushBadTarget) {
		t.Errorf("parent: %+v", e)
	}
	if e := res.streams[2]; e.status != "200" || e.body != "body{}" || e.headers["x-ae"] != "gzip" ||
		e.headers["x-nested"] != ErrPushNotSupported.Error() {
		t.Errorf("css: %+v", e)
	}
	if e := res.streams[4]; e.body != "js" || e.headers["x-got"] != "yes|1" {
		t.Errorf("js: %+v", e)
	}
}

func TestPushDisabledByClient(t *testing.T) {
	addr, stop := startServer(t, pushHandler(false), nil)
	defer stop()
	rc := dialRaw(t, addr, xhttp2.Setting{ID: xhttp2.SettingEnablePush, Val: 0})
	defer rc.c.Close()
	rc.get(1, "/index", addr)
	var res pushResult
	rc.readUntil(&res, finished(1), nil)
	if len(res.promises) != 0 || res.streams[1].headers["x-push-err"][:len(ErrPushDisabled.Error())] != ErrPushDisabled.Error() {
		t.Fatalf("promises %v, parent %+v", res.promises, res.streams[1])
	}
}

func TestPushLimit(t *testing.T) {
	addr, stop := startServer(t, pushHandler(true), nil)
	defer stop()
	rc := dialRaw(t, addr, xhttp2.Setting{ID: xhttp2.SettingMaxConcurrentStreams, Val: 1})
	defer rc.c.Close()
	rc.get(1, "/index", addr)
	var res pushResult
	rc.readUntil(&res, finished(1, 2), nil)
	errs := res.streams[1].headers["x-push-err"]
	if len(res.promises) != 1 || errs != fmt.Sprint(nil, "|", ErrPushLimit, "|", ErrPushBadMethod, "|", ErrPushBadTarget) {
		t.Fatalf("promises %v, errs %q", res.promises, errs)
	}
}

func TestPushCancelledByClient(t *testing.T) {
	addr, stop := startServer(t, pushHandler(false), nil)
	defer stop()
	rc := dialRaw(t, addr, xhttp2.Setting{ID: xhttp2.SettingInitialWindowSize, Val: 1 << 16})
	defer rc.c.Close()
	rc.get(1, "/big", addr)
	var res pushResult
	// Refuse the 20 MiB push as soon as it's promised (a client that has it cached).
	rc.readUntil(&res, finished(1), func(id uint32) { _ = rc.fr.WriteRSTStream(id, xhttp2.ErrCodeCancel) })
	if res.streams[1].body != "big" {
		t.Fatalf("parent %+v", res.streams[1])
	}
	// Connection keeps working afterwards.
	rc.get(3, "/other", addr)
	rc.readUntil(&res, finished(3), nil)
	if res.streams[3].body != "ok" {
		t.Fatalf("after cancel: %+v", res.streams[3])
	}
	if e := res.streams[2]; e != nil && len(e.body) >= 20<<20 {
		t.Fatalf("pushed stream was not cancelled")
	}
}

func TestPushNotHTTP2(t *testing.T) {
	var ctx fasthttp.RequestCtx
	if err := Push(&ctx, "/a", nil); !errors.Is(err, ErrPushNotSupported) || IsPushSupported(&ctx) {
		t.Fatal(err)
	}
}

func TestKeepAlivePing(t *testing.T) {
	addr, stop := startServer(t, pushHandler(false), &Config{PingInterval: 150 * time.Millisecond, IdleTimeout: time.Hour})
	defer stop()

	// A client that never answers PING: server must drop the half-open conn.
	rc := dialRaw(t, addr)
	defer rc.c.Close()
	start, pings := time.Now(), 0
	for {
		f, err := rc.fr.ReadFrame()
		if err != nil { // EOF: server closed us
			break
		}
		if p, ok := f.(*xhttp2.PingFrame); ok && !p.IsAck() {
			pings++
		}
	}
	if d := time.Since(start); pings != 1 || d < 250*time.Millisecond || d > 2*time.Second {
		t.Fatalf("pings=%d closed after %v", pings, d)
	}

	// A client that answers stays connected well past several intervals.
	rc2 := dialRaw(t, addr)
	defer rc2.c.Close()
	deadline := time.Now().Add(700 * time.Millisecond)
	for time.Now().Before(deadline) {
		f, err := rc2.fr.ReadFrame()
		if err != nil {
			t.Fatalf("answering client dropped: %v", err)
		}
		if p, ok := f.(*xhttp2.PingFrame); ok && !p.IsAck() {
			_ = rc2.fr.WritePing(true, p.Data)
		}
	}
}

func TestUserValuesIsolatedAndServerName(t *testing.T) {
	h := func(ctx *fasthttp.RequestCtx) {
		prev := ctx.UserValue("leak")
		ctx.SetUserValue("leak", string(ctx.Path()))
		ctx.SetBodyString(fmt.Sprint(prev))
	}
	addr, stop := startServer(t, h, &Config{ServerName: "httpgo-test"})
	defer stop()
	cl := h2Client()
	for i := 0; i < 50; i++ { // same connection, pooled ctx reused
		resp, err := cl.Get(fmt.Sprintf("https://%s/r%d", addr, i))
		if err != nil {
			t.Fatal(err)
		}
		b := make([]byte, 64)
		n, _ := resp.Body.Read(b)
		resp.Body.Close()
		if string(b[:n]) != "<nil>" || resp.Header.Get("Server") != "httpgo-test" {
			t.Fatalf("req %d: user value leaked %q / server %q", i, b[:n], resp.Header.Get("Server"))
		}
	}
}

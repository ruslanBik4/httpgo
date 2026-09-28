/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

// Conn is a server-side HTTP/2 connection that dispatches requests to a
// fasthttp.RequestHandler.
//
// Concurrency model:
//   - one read loop (Serve) owns the Framer read side, the HPACK decoder and
//     all not-yet-dispatched streams;
//   - every request runs its handler in its own goroutine (true multiplexing);
//   - all frame writes go through writeMu (Framer write side + HPACK encoder);
//   - flow-control windows and the stream map are guarded by mu, with cond
//     used to park writers waiting for WINDOW_UPDATE.
//
// Lock order: writeMu -> mu. Never acquire writeMu while holding mu.
type Conn struct {
	conn    net.Conn
	handler fasthttp.RequestHandler
	cfg     Config

	framer *Framer
	hdec   *HPACKDecoder

	writeMu sync.Mutex
	henc    *HPACKEncoder

	mu         sync.Mutex
	cond       sync.Cond
	streams    map[uint32]*stream
	sendWindow int64 // connection-level send window (peer's receive window)
	// peerInitialWindow is the peer's SETTINGS_INITIAL_WINDOW_SIZE (new streams start with it).
	peerInitialWindow int64
	pushActive        uint32 // open server-initiated (pushed) streams

	// --- server push, written under writeMu, read atomically by the read loop ---
	lastPushID        atomic.Uint32
	peerEnablePush    atomic.Bool
	peerMaxConcurrent atomic.Uint32

	closed           atomic.Bool
	draining         atomic.Bool
	peerMaxFrameSize atomic.Uint32

	// --- read-loop owned ---
	// pending holds streams still receiving headers/body. Only the read loop
	// touches them, so it never reads fields of a stream a handler owns.
	pending         map[uint32]*stream
	lastStreamID    uint32
	recvWindow      int64
	recvUnacked     uint32
	goAwayRecv      bool
	goAwaySent      bool
	pingOutstanding bool
	closedRing      closedRing
	settingsSeen    bool
	contStream      uint32 // stream awaiting CONTINUATION, 0 if none
	contEndStream   bool
	contSelfDep     bool // HEADERS depends on itself: stream PROTOCOL_ERROR after decoding
	hdrBlock        []byte
	dec             decodeState
	emitFn          func(name, value string)

	// --- write side scratch (under writeMu) ---
	visitFn   func(k, v []byte)
	sawDate   bool
	sawServer bool
	ping      [8]byte
}

// NewConn creates a new HTTP/2 connection with default settings.
func NewConn(c net.Conn, handler fasthttp.RequestHandler) *Conn {
	return NewConnConfig(c, handler, nil)
}

// NewConnConfig creates a new HTTP/2 connection with the given settings.
func NewConnConfig(c net.Conn, handler fasthttp.RequestHandler, cfg *Config) *Conn {
	sc := &Conn{
		conn:              c,
		handler:           handler,
		cfg:               cfg.withDefaults(),
		streams:           make(map[uint32]*stream),
		pending:           make(map[uint32]*stream),
		sendWindow:        defaultWindowSize,
		peerInitialWindow: defaultWindowSize,
		henc:              NewHPACKEncoder(),
	}
	sc.cond.L = &sc.mu
	sc.peerMaxFrameSize.Store(minMaxFrameSize)
	sc.peerEnablePush.Store(!sc.cfg.DisablePush) // client default is ENABLE_PUSH=1
	sc.peerMaxConcurrent.Store(^uint32(0))       // unlimited until the client says otherwise
	sc.framer = NewFramer(c, c)
	sc.framer.MaxReadFrameSize = sc.cfg.MaxFrameSize
	sc.hdec = NewHPACKDecoder(int(sc.cfg.MaxHeaderListSize))
	sc.recvWindow = int64(sc.cfg.ConnWindowSize)
	sc.emitFn = sc.emitField
	sc.visitFn = sc.visitRespHeader
	return sc
}

// Serve runs the HTTP/2 connection until it is closed or an error occurs.
// It returns nil on a clean close (EOF, idle timeout, client GOAWAY).
func (c *Conn) Serve() (err error) {
	defer func() { err = c.shutdown(err) }()

	_ = c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err := c.readClientPreface(); err != nil {
		return err
	}

	// Server connection preface: our SETTINGS + enlarge connection window.
	if err := c.writeFrames(func(fr *Framer) error {
		_ = fr.WriteSettings(
			Setting{SettingMaxConcurrentStreams, c.cfg.MaxConcurrentStreams},
			Setting{SettingInitialWindowSize, c.cfg.InitialWindowSize},
			Setting{SettingMaxFrameSize, c.cfg.MaxFrameSize},
			Setting{SettingMaxHeaderListSize, c.cfg.MaxHeaderListSize},
			Setting{SettingEnablePush, 0},
		)
		if d := c.cfg.ConnWindowSize - defaultWindowSize; d > 0 {
			return fr.WriteWindowUpdate(0, d)
		}
		return nil
	}); err != nil {
		return err
	}

	// Wake up at least every PingInterval to check liveness.
	wait := c.cfg.IdleTimeout
	pingOn := c.cfg.PingInterval > 0
	if pingOn {
		wait = min(wait, c.cfg.PingInterval)
	}
	lastDeadline := time.Time{}
	lastRead := time.Now()   // any frame: peer is alive
	lastActivity := lastRead // any frame but PING: peer is using the connection
	var drainStart time.Time
	for {
		if now := time.Now(); now.Sub(lastDeadline) > time.Second {
			_ = c.conn.SetReadDeadline(now.Add(wait))
			lastDeadline = now
		}
		// Checked after SetReadDeadline so a concurrent Shutdown can't be missed.
		if c.draining.Load() {
			if !c.goAwaySent {
				c.goAwaySent, drainStart = true, time.Now()
				last := c.lastStreamID
				_ = c.writeFrames(func(fr *Framer) error { return fr.WriteGoAway(last, ErrCodeNo, nil) })
			}
			if c.activeStreams() == 0 || time.Since(drainStart) > c.cfg.IdleTimeout {
				return nil
			}
			_ = c.conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond)) // poll for handlers to finish
			lastDeadline = time.Time{}
		}

		f, err := c.framer.ReadFrame()
		if err != nil {
			if err != errIdle {
				return err
			}
			lastDeadline = time.Time{}
			now := time.Now()
			if pingOn && now.Sub(lastRead) >= c.cfg.PingInterval {
				if c.pingOutstanding {
					if c.cfg.Debug {
						c.logf("http2: %s: keep-alive PING not answered, closing", c.conn.RemoteAddr())
					}
					return nil // peer is dead (half-open TCP): nothing to tell it
				}
				c.pingOutstanding = true
				if err := c.writeFrames(func(fr *Framer) error { return fr.WritePing(false, keepAlivePing) }); err != nil {
					return err
				}
				lastRead = now // measure the next interval from the PING
			}
			// Idle close only when no request is in flight: long-running
			// handlers (SSE, long-poll) keep the connection alive.
			if !c.draining.Load() && c.activeStreams() == 0 && now.Sub(lastActivity) >= c.cfg.IdleTimeout {
				return nil
			}
			continue
		}
		lastRead = time.Now()
		if f.Type != FramePing {
			lastActivity = lastRead
		}

		if err := c.processFrame(f); err != nil {
			var se StreamError
			if errors.As(err, &se) {
				if c.cfg.Debug {
					c.logf("%v (%s)", se, c.conn.RemoteAddr())
				}
				c.resetStream(se.StreamID, se.Code)
				continue
			}
			if c.cfg.Debug {
				c.logf("%v (%s)", err, c.conn.RemoteAddr())
			}
			return err
		}
	}
}

var keepAlivePing = [8]byte{'h', 't', 't', 'p', 'g', 'o', 'k', 'a'}

// Shutdown starts a graceful close: GOAWAY is sent, new streams are refused,
// in-flight requests are completed, then Serve returns. Safe to call from any goroutine.
func (c *Conn) Shutdown() {
	if c.draining.CompareAndSwap(false, true) {
		_ = c.conn.SetReadDeadline(time.Now()) // wake the read loop
	}
}

// shutdown converts the loop error into GOAWAY and wakes all waiting writers.
func (c *Conn) shutdown(err error) error {
	// A handler stuck writing to a peer that stopped reading holds writeMu;
	// bound it so the final GOAWAY (and Serve) can't hang forever.
	_ = c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	code, send := ErrCodeNo, true
	var ce ConnError
	switch {
	case errors.As(err, &ce):
		code = ce.Code
	case err == nil:
	case errors.Is(err, io.EOF), errors.Is(err, net.ErrClosed), isTimeout(err):
		send, err = false, nil // peer is gone, nobody to tell
	default:
		send = false
	}
	if send && !(c.goAwaySent && code == ErrCodeNo) {
		var debug []byte
		if ce.Reason != "" {
			debug = []byte(ce.Reason)
		}
		_ = c.writeFrames(func(fr *Framer) error { return fr.WriteGoAway(c.lastStreamID, code, debug) })
	}

	c.closed.Store(true)
	c.mu.Lock()
	for _, s := range c.streams {
		s.reset = true
	}
	for id := range c.pending {
		delete(c.streams, id)
	}
	c.cond.Broadcast()
	c.mu.Unlock()
	for id, s := range c.pending { // read-loop owned: safe to release here
		delete(c.pending, id)
		releaseStream(s)
	}
	return err
}

func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// readClientPreface reads and validates the HTTP/2 client connection preface.
func (c *Conn) readClientPreface() error {
	var buf [len(clientPreface)]byte
	if _, err := io.ReadFull(c.framer.r, buf[:]); err != nil {
		return err
	}
	if string(buf[:]) != clientPreface {
		return ErrProtocol
	}
	return nil
}

// ---------------------------------------------------------------------------
// frame dispatch (read loop)
// ---------------------------------------------------------------------------

func (c *Conn) processFrame(f *Frame) error {
	if !c.settingsSeen {
		if f.Type != FrameSettings || f.Flags.Has(FlagAck) {
			return connErr(ErrCodeProtocol, "first frame must be SETTINGS")
		}
		c.settingsSeen = true
	}
	if c.contStream != 0 && (f.Type != FrameContinuation || f.StreamID != c.contStream) {
		return connErr(ErrCodeProtocol, "expected CONTINUATION for stream %d", c.contStream)
	}

	switch f.Type {
	case FrameData:
		return c.handleData(f)
	case FrameHeaders:
		return c.handleHeaders(f)
	case FrameContinuation:
		return c.handleContinuation(f)
	case FrameSettings:
		return c.handleSettings(f)
	case FrameWindowUpdate:
		return c.handleWindowUpdate(f)
	case FramePing:
		return c.handlePing(f)
	case FrameRSTStream:
		return c.handleRSTStream(f)
	case FrameGoAway:
		if f.StreamID != 0 {
			return connErr(ErrCodeProtocol, "GOAWAY on stream %d", f.StreamID)
		}
		c.goAwayRecv = true // refuse new streams, finish the in-flight ones
		return nil
	case FramePriority:
		if f.StreamID == 0 {
			return connErr(ErrCodeProtocol, "PRIORITY on stream 0")
		}
		if f.Length != 5 {
			return streamErr(f.StreamID, ErrCodeFrameSize, "PRIORITY length")
		}
		if binary.BigEndian.Uint32(f.Payload)&(1<<31-1) == f.StreamID {
			return streamErr(f.StreamID, ErrCodeProtocol, "stream depends on itself")
		}
		return nil // RFC 9113 deprecates the priority scheme
	case FramePushPromise:
		return connErr(ErrCodeProtocol, "client sent PUSH_PROMISE")
	default:
		return nil // unknown frame types MUST be ignored
	}
}

func (c *Conn) handleSettings(f *Frame) error {
	if f.StreamID != 0 {
		return connErr(ErrCodeProtocol, "SETTINGS on stream %d", f.StreamID)
	}
	if f.Flags.Has(FlagAck) {
		if f.Length != 0 {
			return connErr(ErrCodeFrameSize, "SETTINGS ACK with payload")
		}
		return nil
	}
	if f.Length%6 != 0 {
		return connErr(ErrCodeFrameSize, "SETTINGS length %d", f.Length)
	}
	p := f.Payload
	tableSize, tableSizeSet := uint32(0), false
	for ; len(p) > 0; p = p[6:] {
		id := SettingID(binary.BigEndian.Uint16(p))
		v := binary.BigEndian.Uint32(p[2:])
		switch id {
		case SettingHeaderTableSize:
			tableSize, tableSizeSet = v, true
		case SettingEnablePush:
			if v > 1 {
				return connErr(ErrCodeProtocol, "ENABLE_PUSH=%d", v)
			}
			c.peerEnablePush.Store(v == 1 && !c.cfg.DisablePush)
		case SettingMaxConcurrentStreams:
			c.peerMaxConcurrent.Store(v) // limits streams WE open, i.e. pushes
		case SettingInitialWindowSize:
			if v > maxWindowSize {
				return connErr(ErrCodeFlowControl, "INITIAL_WINDOW_SIZE=%d", v)
			}
			c.mu.Lock()
			delta := int64(v) - c.peerInitialWindow
			c.peerInitialWindow = int64(v)
			for _, s := range c.streams {
				s.sendWindow += delta
				if s.sendWindow > maxWindowSize {
					c.mu.Unlock()
					return connErr(ErrCodeFlowControl, "stream window overflow")
				}
			}
			c.cond.Broadcast()
			c.mu.Unlock()
		case SettingMaxFrameSize:
			if v < minMaxFrameSize || v > maxMaxFrameSize {
				return connErr(ErrCodeProtocol, "MAX_FRAME_SIZE=%d", v)
			}
			c.peerMaxFrameSize.Store(v)
		}
	}
	return c.writeFrames(func(fr *Framer) error {
		if tableSizeSet {
			c.henc.SetMaxDynamicTableSizeLimit(tableSize)
		}
		return fr.WriteSettingsAck()
	})
}

func (c *Conn) handlePing(f *Frame) error {
	if f.StreamID != 0 {
		return connErr(ErrCodeProtocol, "PING on stream %d", f.StreamID)
	}
	if f.Length != 8 {
		return connErr(ErrCodeFrameSize, "PING length %d", f.Length)
	}
	if f.Flags.Has(FlagAck) {
		c.pingOutstanding = false
		return nil
	}
	copy(c.ping[:], f.Payload) // f.Payload is reused by the next ReadFrame
	return c.writeFrames(func(fr *Framer) error { return fr.WritePing(true, c.ping) })
}

func (c *Conn) handleWindowUpdate(f *Frame) error {
	if f.Length != 4 {
		return connErr(ErrCodeFrameSize, "WINDOW_UPDATE length %d", f.Length)
	}
	incr := int64(binary.BigEndian.Uint32(f.Payload) & (1<<31 - 1))
	if f.StreamID == 0 {
		if incr == 0 {
			return connErr(ErrCodeProtocol, "WINDOW_UPDATE increment 0")
		}
		c.mu.Lock()
		c.sendWindow += incr
		overflow := c.sendWindow > maxWindowSize
		c.cond.Broadcast()
		c.mu.Unlock()
		if overflow {
			return connErr(ErrCodeFlowControl, "connection window overflow")
		}
		return nil
	}
	if c.isIdle(f.StreamID) {
		return connErr(ErrCodeProtocol, "WINDOW_UPDATE on idle stream %d", f.StreamID)
	}
	if incr == 0 {
		return streamErr(f.StreamID, ErrCodeProtocol, "WINDOW_UPDATE increment 0")
	}
	c.mu.Lock()
	s := c.streams[f.StreamID]
	overflow := false
	if s != nil {
		s.sendWindow += incr
		overflow = s.sendWindow > maxWindowSize
		c.cond.Broadcast()
	}
	c.mu.Unlock()
	if overflow {
		return streamErr(f.StreamID, ErrCodeFlowControl, "stream window overflow")
	}
	return nil
}

func (c *Conn) handleRSTStream(f *Frame) error {
	if f.StreamID == 0 {
		return connErr(ErrCodeProtocol, "RST_STREAM on stream 0")
	}
	if f.Length != 4 {
		return connErr(ErrCodeFrameSize, "RST_STREAM length %d", f.Length)
	}
	if c.isIdle(f.StreamID) {
		return connErr(ErrCodeProtocol, "RST_STREAM on idle stream %d", f.StreamID)
	}
	c.closeStreamFromReader(f.StreamID)
	c.closedRing.add(f.StreamID, false)
	return nil
}

// closeStreamFromReader marks a stream reset. A dispatched stream is released
// by its handler goroutine; an undispatched one is released right here.
func (c *Conn) closeStreamFromReader(id uint32) {
	ps := c.pending[id]
	c.mu.Lock()
	if s := c.streams[id]; s != nil {
		s.reset = true
		if ps != nil {
			delete(c.streams, id)
		}
		c.cond.Broadcast()
	}
	c.mu.Unlock()
	if ps != nil {
		delete(c.pending, id)
		releaseStream(ps)
	}
}

// resetStream sends RST_STREAM and forgets the stream.
func (c *Conn) resetStream(id uint32, code ErrCode) {
	c.closeStreamFromReader(id)
	c.closedRing.add(id, true)
	_ = c.writeFrames(func(fr *Framer) error { return fr.WriteRSTStream(id, code) })
}

func (c *Conn) activeStreams() int {
	c.mu.Lock()
	n := len(c.streams)
	c.mu.Unlock()
	return n
}

// isIdle reports whether id was never opened: odd ids are client-initiated,
// even ids are ours (server push).
func (c *Conn) isIdle(id uint32) bool {
	if id%2 == 0 {
		return id > c.lastPushID.Load()
	}
	return id > c.lastStreamID
}

// isDispatched reports whether id is open on the handler side (half-closed remote).
func (c *Conn) isDispatched(id uint32) bool {
	c.mu.Lock()
	_, ok := c.streams[id]
	c.mu.Unlock()
	return ok && c.pending[id] == nil
}

// closedStreamFrame classifies a DATA/HEADERS frame for a non-pending stream
// id <= lastStreamID (RFC 9113 §5.1).
func (c *Conn) closedStreamFrame(id uint32, what string) error {
	if c.isDispatched(id) { // half-closed (remote)
		return streamErr(id, ErrCodeStreamClosed, what+" after END_STREAM")
	}
	known, byUs := c.closedRing.find(id)
	switch {
	case byUs:
		return nil // we sent RST_STREAM: frames already in flight MUST be ignored (§5.4.2)
	case known || what == "DATA":
		return connErr(ErrCodeStreamClosed, "%s on closed stream %d", what, id)
	default: // never-opened id lower than lastStreamID
		return connErr(ErrCodeProtocol, "%s on stream %d lower than last %d", what, id, c.lastStreamID)
	}
}

// closedRing remembers recently closed stream ids (read loop only).
type closedRing struct {
	ids  [64]uint32
	byUs [64]bool
	n    int
}

func (r *closedRing) add(id uint32, byUs bool) {
	if i := r.index(id); i >= 0 { // update in place (e.g. dispatched, then we reset it)
		r.byUs[i] = r.byUs[i] || byUs
		return
	}
	i := r.n % len(r.ids)
	r.ids[i], r.byUs[i] = id, byUs
	r.n++
}

func (r *closedRing) index(id uint32) int {
	for i := range min(r.n, len(r.ids)) {
		if r.ids[i] == id {
			return i
		}
	}
	return -1
}

func (r *closedRing) find(id uint32) (known, byUs bool) {
	if i := r.index(id); i >= 0 {
		return true, r.byUs[i]
	}
	return false, false
}

// ---------------------------------------------------------------------------
// HEADERS / CONTINUATION
// ---------------------------------------------------------------------------

func (c *Conn) handleHeaders(f *Frame) error {
	if f.StreamID == 0 || f.StreamID%2 == 0 {
		return connErr(ErrCodeProtocol, "HEADERS on invalid stream %d", f.StreamID)
	}
	frag, dependsOn, err := headerBlockFragment(f)
	if err != nil {
		return err
	}
	endStream := f.Flags.Has(FlagEndStream)
	c.contSelfDep = dependsOn == f.StreamID
	if f.Flags.Has(FlagEndHeaders) {
		return c.onHeaderBlock(f.StreamID, frag, endStream) // zero-copy fast path
	}
	c.contStream, c.contEndStream = f.StreamID, endStream
	c.hdrBlock = append(c.hdrBlock[:0], frag...)
	return nil
}

func (c *Conn) handleContinuation(f *Frame) error {
	if c.contStream == 0 {
		return connErr(ErrCodeProtocol, "unexpected CONTINUATION")
	}
	c.hdrBlock = append(c.hdrBlock, f.Payload...)
	// CONTINUATION flood protection (CVE-2024-27316 class): bound buffered block.
	if uint32(len(c.hdrBlock)) > c.cfg.MaxHeaderListSize+c.cfg.MaxFrameSize {
		return connErr(ErrCodeEnhanceYourCalm, "header block too large")
	}
	if !f.Flags.Has(FlagEndHeaders) {
		return nil
	}
	id, end := c.contStream, c.contEndStream
	c.contStream = 0
	err := c.onHeaderBlock(id, c.hdrBlock, end)
	if cap(c.hdrBlock) > 64<<10 {
		c.hdrBlock = nil
	}
	return err
}

// onHeaderBlock handles a complete header block. It ALWAYS decodes the block,
// even for streams it will refuse, because HPACK state is connection-wide.
func (c *Conn) onHeaderBlock(id uint32, block []byte, endStream bool) error {
	if s := c.pending[id]; s != nil { // trailers
		c.dec.reset(&s.ctx.Request, true, false)
		if err := c.hdec.Decode(block, c.emitFn); err != nil {
			return err
		}
		if !endStream || c.dec.malformed {
			return streamErr(id, ErrCodeProtocol, "malformed trailers")
		}
		return c.endRequest(s)
	}

	if id <= c.lastStreamID {
		c.dec.reset(nil, false, true)
		if err := c.hdec.Decode(block, c.emitFn); err != nil {
			return err
		}
		return c.closedStreamFrame(id, "HEADERS")
	}
	c.lastStreamID = id

	if c.contSelfDep {
		c.dec.reset(nil, false, true)
		if err := c.hdec.Decode(block, c.emitFn); err != nil {
			return err
		}
		return streamErr(id, ErrCodeProtocol, "stream depends on itself")
	}

	if c.goAwayRecv || c.goAwaySent || c.activeStreams() >= int(c.cfg.MaxConcurrentStreams) {
		c.dec.reset(nil, false, true)
		if err := c.hdec.Decode(block, c.emitFn); err != nil {
			return err
		}
		return streamErr(id, ErrCodeRefusedStream, "too many streams")
	}

	s := c.acquireStream(id)
	c.dec.reset(&s.ctx.Request, false, false)
	if err := c.hdec.Decode(block, c.emitFn); err != nil {
		releaseStream(s)
		return err
	}
	d := &c.dec
	if d.tooLarge {
		releaseStream(s)
		c.closedRing.add(id, true)
		return c.rejectStream(id, fasthttp.StatusRequestHeaderFieldsTooLarge)
	}
	if d.malformed || !d.hasMethod ||
		(d.isConnect && (!d.hasAuthority || d.hasPath || d.hasScheme)) ||
		(!d.isConnect && (!d.hasScheme || !d.hasPath)) {
		releaseStream(s)
		return streamErr(id, ErrCodeProtocol, "malformed request headers")
	}
	req := &s.ctx.Request
	req.Header.SetProtocol("HTTP/2.0")
	if d.scheme != "" {
		req.URI().SetScheme(d.scheme)
	}
	if d.contentLength > int64(c.cfg.MaxRequestBodySize) {
		releaseStream(s)
		c.closedRing.add(id, true)
		return c.rejectStream(id, fasthttp.StatusRequestEntityTooLarge)
	}
	s.declaredCL = d.contentLength

	c.mu.Lock()
	s.sendWindow = c.peerInitialWindow
	c.streams[id] = s
	c.mu.Unlock()
	c.pending[id] = s

	if endStream {
		return c.endRequest(s)
	}
	if d.expectContinue {
		return c.writeFrames(func(fr *Framer) error {
			c.henc.Reset()
			c.henc.WriteField(":status", "100")
			return fr.WriteHeaders(id, false, c.henc.Bytes(), c.peerMaxFrameSize.Load())
		})
	}
	return nil
}

// endRequest is called on END_STREAM: hands the stream over to a handler goroutine.
// After this the read loop must not touch s except via c.streams under c.mu.
func (c *Conn) endRequest(s *stream) error {
	n := len(s.ctx.Request.Body())
	if s.declaredCL >= 0 && s.declaredCL != int64(n) {
		return streamErr(s.id, ErrCodeProtocol, "content-length mismatch")
	}
	if n > 0 || s.declaredCL >= 0 {
		s.ctx.Request.Header.SetContentLength(n)
	}
	delete(c.pending, s.id)
	c.closedRing.add(s.id, false)
	// Lets Push(ctx, ...) find this stream; no allocation once the pooled ctx
	// has user-value capacity.
	s.ctx.SetUserValue(pushKey{}, s)
	go c.runHandler(s)
	return nil
}

// emitField is called by the HPACK decoder for every field of the current block.
func (c *Conn) emitField(name, value string) {
	d := &c.dec
	d.size += uint32(len(name) + len(value) + 32)
	if d.size > c.cfg.MaxHeaderListSize {
		d.tooLarge = true
	}
	if d.discard || d.tooLarge || d.malformed {
		return
	}
	req := d.req

	if len(name) > 0 && name[0] == ':' {
		if d.regularSeen || d.trailers {
			d.malformed = true
			return
		}
		switch name {
		case ":method":
			if d.hasMethod {
				d.malformed = true
				return
			}
			d.hasMethod, d.isConnect = true, value == fasthttp.MethodConnect
			req.Header.SetMethod(value)
		case ":path":
			if d.hasPath || value == "" {
				d.malformed = true
				return
			}
			d.hasPath = true
			req.Header.SetRequestURI(value)
		case ":scheme":
			if d.hasScheme {
				d.malformed = true
				return
			}
			d.hasScheme, d.scheme = true, value
		case ":authority":
			if d.hasAuthority {
				d.malformed = true
				return
			}
			d.hasAuthority = true
			req.Header.SetHost(value)
		default: // :status, :protocol (extended CONNECT unsupported), unknown
			d.malformed = true
		}
		return
	}

	d.regularSeen = true
	for i := 0; i < len(name); i++ {
		if b := name[i]; b >= 'A' && b <= 'Z' {
			d.malformed = true // RFC 9113 §8.2.1: field names MUST be lowercase
			return
		}
	}
	switch name {
	case "connection", "keep-alive", "proxy-connection", "transfer-encoding", "upgrade":
		d.malformed = true // §8.2.2 connection-specific fields
		return
	case "te":
		if value != "trailers" {
			d.malformed = true
		}
		return
	case "host":
		if !d.hasAuthority && !d.trailers {
			req.Header.SetHost(value)
		}
		return
	case "content-length":
		if d.trailers {
			return
		}
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil || n < 0 || (d.contentLength >= 0 && d.contentLength != n) {
			d.malformed = true
			return
		}
		d.contentLength = n
		return
	case "expect":
		if strings.EqualFold(value, "100-continue") {
			d.expectContinue = true
			return
		}
	}
	// cookie: fasthttp merges multiple "cookie" fields (§8.2.3) on Add.
	req.Header.Add(name, value)
}

// ---------------------------------------------------------------------------
// DATA (request body) + receive flow control
// ---------------------------------------------------------------------------

func (c *Conn) handleData(f *Frame) error {
	if f.StreamID == 0 {
		return connErr(ErrCodeProtocol, "DATA on stream 0")
	}
	if c.isIdle(f.StreamID) {
		return connErr(ErrCodeProtocol, "DATA on idle stream %d", f.StreamID)
	}
	// Connection-level accounting covers every DATA frame, padding included.
	l := f.Length
	c.recvWindow -= int64(l)
	if c.recvWindow < 0 {
		return connErr(ErrCodeFlowControl, "connection receive window exceeded")
	}
	c.recvUnacked += l
	if c.recvUnacked >= c.cfg.ConnWindowSize/2 {
		incr := c.recvUnacked
		c.recvWindow += int64(incr)
		c.recvUnacked = 0
		if err := c.writeFrames(func(fr *Framer) error { return fr.WriteWindowUpdate(0, incr) }); err != nil {
			return err
		}
	}

	s := c.pending[f.StreamID]
	if s == nil {
		return c.closedStreamFrame(f.StreamID, "DATA")
	}
	s.recvWindow -= int64(l)
	if s.recvWindow < 0 {
		return streamErr(f.StreamID, ErrCodeFlowControl, "stream receive window exceeded")
	}
	data, err := stripPadding(f)
	if err != nil {
		return err
	}
	req := &s.ctx.Request
	if len(req.Body())+len(data) > c.cfg.MaxRequestBodySize {
		c.closeStreamFromReader(f.StreamID)
		c.closedRing.add(f.StreamID, true)
		return c.rejectStream(f.StreamID, fasthttp.StatusRequestEntityTooLarge)
	}
	req.AppendBody(data)

	if f.Flags.Has(FlagEndStream) {
		return c.endRequest(s)
	}
	s.recvUnacked += l
	if s.recvUnacked >= c.cfg.InitialWindowSize/2 {
		incr := s.recvUnacked
		s.recvWindow += int64(incr)
		s.recvUnacked = 0
		return c.writeFrames(func(fr *Framer) error { return fr.WriteWindowUpdate(s.id, incr) })
	}
	return nil
}

// rejectStream answers with a bodiless status and tells the client to stop
// sending (RST_STREAM NO_ERROR after a complete response, RFC 9113 §8.1).
func (c *Conn) rejectStream(id uint32, status int) error {
	return c.writeFrames(func(fr *Framer) error {
		c.henc.Reset()
		c.henc.WriteField(":status", statusText(status))
		if err := fr.WriteHeaders(id, true, c.henc.Bytes(), c.peerMaxFrameSize.Load()); err != nil {
			return err
		}
		return fr.WriteRSTStream(id, ErrCodeNo)
	})
}

// ---------------------------------------------------------------------------
// writing
// ---------------------------------------------------------------------------

// writeFrames runs fn under writeMu and flushes.
func (c *Conn) writeFrames(fn func(fr *Framer) error) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.closed.Load() {
		return errConnClosed
	}
	if err := fn(c.framer); err != nil {
		return err
	}
	return c.framer.Flush()
}

func (c *Conn) logf(format string, args ...any) {
	if c.cfg.Logger != nil {
		c.cfg.Logger.Printf(format, args...)
		return
	}
	log.Printf(format, args...)
}

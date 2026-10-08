/*
 * Copyright (c) 2025-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// implementation support SSE on backend
package httpGo

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/apis/crud"
	"github.com/ruslanBik4/logs"
)

const (
	logName     = "sse"
	pingEvery   = 5 * time.Second
	maxLogLines = 5000 // history kept per log; older lines are dropped
)

var (
	pingData   = []byte("1")
	closedData = []byte("Stream finished successfully")
)

// HandleNoticeSSE streams a LogWriter kept in the app Store (see
// Store.StartSSEJob) to the browser as server-sent events:
//
//	GET /httpgo/store/sse?id=<key>
//
// Every line is an event with a growing id, so a client that reconnects
// (Last-Event-ID) continues where it stopped instead of replaying the log; a
// client that connects late still gets everything from the start. The stream
// ends with a proper `closed` event, and a `ping` every few seconds is what
// readEvents() in forms.js uses to tell a live connection from a dead one.
//
// The route must have IsServerEvents set: the event-stream headers are written
// by ApiRoute.CheckAndRun.
func HandleNoticeSSE(ctx *fasthttp.RequestCtx) (any, error) {
	store, ok := ctx.UserValue(apis.AppStore).(*Store)
	if !ok {
		return nil, &apis.ErrMethodNotAllowed{}
	}
	id, _ := ctx.UserValue(crud.ParamsIDReq.Name).(int32)
	l, ok := store.Get(uint64(id), logName).(*LogWriter)
	if !ok {
		return nil, &apis.ErrMethodNotAllowed{}
	}

	// read request data NOW: the stream function runs after the handler returns
	from := 0
	if v, err := strconv.Atoi(gotools.BytesToString(ctx.Request.Header.Peek("Last-Event-ID"))); err == nil && v >= 0 {
		from = v + 1
	}

	ctx.Response.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			if e := recover(); e != nil {
				logs.StatusLog(e)
			}
		}()
		streamLog(w, l, from)
	})

	return nil, nil
}

var ansiToCSS = map[string]string{
	"30": "#000000", // Black
	"31": "#FF0000", // Red
	"32": "#00FF00", // Green
	"33": "#FFFF00", // Yellow
	"34": "#0000FF", // Blue
	"35": "#FF00FF", // Magenta
	"36": "#00FFFF", // Cyan
	"37": "#FFFFFF", // White
	"90": "#808080", // Bright Black (Gray)
	"91": "#FF5555", // Bright Red
	"92": "#55FF55", // Bright Green
	"93": "#FFFF55", // Bright Yellow
	"94": "#5555FF", // Bright Blue
	"95": "#FF55FF", // Bright Magenta
	"96": "#55FFFF", // Bright Cyan
	"97": "#FFFFFF", // Bright White
}

// \033[31m text \033[0m  (text may contain spaces; the old `[\S]+` did not match them)
var ansiRegex = regexp.MustCompile(`\033\[(\d+)(?:;1)?m([^\033]*)\033\[0m`)

// ConvertANSICodeToCSS turns ANSI colour sequences into <span style="color:..">
// in ONE pass over input (no second regexp match per hit, no string<->bytes
// round trips). Unknown codes are left as they were (the old version deleted
// the text). Returns input itself when there is nothing to convert.
func ConvertANSICodeToCSS(input []byte) []byte {
	matches := ansiRegex.FindAllSubmatchIndex(input, -1)
	if matches == nil {
		return input
	}

	out := make([]byte, 0, len(input)+len(matches)*40)
	last := 0
	for _, m := range matches { // m: [whole.. , code.. , text..]
		color, exists := ansiToCSS[gotools.BytesToString(input[m[2]:m[3]])]
		if !exists {
			continue // stays in input[last:] and is copied later
		}
		out = append(out, input[last:m[0]]...)
		out = append(out, `<span style="color:`...)
		out = append(out, color...)
		out = append(out, `;">`...)
		out = append(out, input[m[4]:m[5]]...)
		out = append(out, `</span>`...)
		last = m[1]
	}

	return append(out, input[last:]...)
}

// escapeHTML returns a NEW slice with & < > " ' replaced by entities (same
// replacements as html.EscapeString, but bytes in, bytes out).
func escapeHTML(src []byte) []byte {
	dst := make([]byte, 0, len(src)+len(src)/8)
	for _, c := range src {
		switch c {
		case '&':
			dst = append(dst, "&amp;"...)
		case '<':
			dst = append(dst, "&lt;"...)
		case '>':
			dst = append(dst, "&gt;"...)
		case '"':
			dst = append(dst, "&#34;"...)
		case '\'':
			dst = append(dst, "&#39;"...)
		default:
			dst = append(dst, c)
		}
	}
	return dst
}

var (
	lf = []byte("\n")
	cr = []byte("\r")
)

// writeSSE writes ONE event and flushes it.
//
// A browser dispatches an event only if it has at least one data: line AND is
// ended by a blank line - an `event: x` without data is silently dropped (that
// is why "closed" and "ping" never reached the client before). Multi-line data
// gets one data: line per row, otherwise the extra rows break the framing.
// id < 0 means "no id" (ping/closed must not move Last-Event-ID).
//
// bufio.Writer keeps the first write error and returns it from every later
// call, so only Flush's result needs checking.
func writeSSE(w *bufio.Writer, id int, event string, data []byte) error {
	if id >= 0 {
		_, _ = fmt.Fprintf(w, "id: %d\n", id)
	}
	if event != "" {
		_, _ = w.WriteString("event: ")
		_, _ = w.WriteString(event)
		_ = w.WriteByte('\n')
	}
	for {
		line, rest, more := bytes.Cut(data, lf)
		_, _ = w.WriteString("data: ")
		_, _ = w.Write(bytes.TrimSuffix(line, cr))
		_ = w.WriteByte('\n')
		if !more {
			break
		}
		data = rest
	}
	_ = w.WriteByte('\n')

	return w.Flush()
}

// LogWriter collects the output of a long job (docker build, ...) and lets any
// number of SSE clients read it - from the beginning, or from where a
// reconnecting client stopped (Last-Event-ID). It never blocks the producer
// and never loses lines: the old unbuffered channel + `default:` dropped every
// line written before the browser had connected.
type LogWriter struct {
	mu     sync.Mutex
	lines  [][]byte      // retained history
	base   int           // sequence number of lines[0]
	closed bool          // producer finished
	wake   chan struct{} // closed (and replaced) on every change
}

func NewLogWriter() *LogWriter {
	return &LogWriter{wake: make(chan struct{})}
}

// Write is the io.Writer for RAW process output: it is HTML-escaped first
// (container output is user-controlled and the client inserts it as HTML),
// then ANSI colours become spans. p is not kept (io.Writer contract): the
// escaped copy is.
func (l *LogWriter) Write(p []byte) (int, error) {
	l.add(ConvertANSICodeToCSS(escapeHTML(bytes.TrimRight(p, "\r\n"))))

	return len(p), nil
}

// Send adds one message that is ALREADY html (e.g. fabric.ShrinkOut output).
// msg is used without copying - strings are immutable and never modified here.
func (l *LogWriter) Send(msg string) {
	l.add(gotools.StringToBytes(msg))
}

// add appends one finished message and takes ownership of b.
func (l *LogWriter) add(b []byte) {
	if len(b) == 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	l.lines = append(l.lines, b)
	if n := len(l.lines) - maxLogLines; n > 0 {
		l.lines = append(l.lines[:0], l.lines[n:]...)
		l.base += n
	}
	l.notify()
}

// Close marks the log as finished; readers send "closed" after the last line.
func (l *LogWriter) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.closed {
		l.closed = true
		l.notify()
	}
}

func (l *LogWriter) notify() { // l.mu held
	close(l.wake)
	l.wake = make(chan struct{})
}

// next returns the lines with sequence >= from, the sequence of the first one,
// a channel that is closed on the next change, and whether the log is finished.
func (l *LogWriter) next(from int) (lines [][]byte, first int, wake <-chan struct{}, closed bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if from < l.base {
		from = l.base
	}
	if i := from - l.base; i < len(l.lines) {
		lines = append([][]byte(nil), l.lines[i:]...)
	}
	return lines, from, l.wake, l.closed
}

// streamLog sends the log to one client, starting at sequence `from`, until the
// log is finished or a write fails (client gone). The ping both keeps proxies
// from timing the connection out and lets us notice a vanished client.
func streamLog(w *bufio.Writer, l *LogWriter, from int) {
	ping := time.NewTicker(pingEvery)
	defer ping.Stop()

	for {
		lines, first, wake, closed := l.next(from)
		for i, line := range lines {
			if err := writeSSE(w, first+i, "", line); err != nil {
				return
			}
		}
		from = first + len(lines)

		if closed { // everything delivered
			_ = writeSSE(w, -1, "closed", closedData)
			return
		}

		select {
		case <-wake:
		case <-ping.C:
			if err := writeSSE(w, -1, "ping", pingData); err != nil {
				return
			}
		}
	}
}

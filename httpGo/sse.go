/*
 * Copyright (c) 2025. Author: Ruslan Bikchentaev. All rights reserved.
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

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/apis/crud"
	"github.com/ruslanBik4/logs"
)

const logName = "sse"

// HandleNoticeSSE  for demo version
func HandleNoticeSSE(ctx *fasthttp.RequestCtx) (any, error) {
	l, ok := ctx.UserValue(apis.AppStore).(*Store).Get(uint64(ctx.UserValue(crud.ParamsIDReq.Name).(int32)), logName).(*LogWriter)
	if !ok {
		return nil, &apis.ErrMethodNotAllowed{}
	}

	ctx.Response.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			if e := recover(); e != nil {
				logs.StatusLog(e)
			}
		}()
		for b := range l.ch {
			_, err := fmt.Fprintf(w, "data: %s\n\n", b)
			if err != nil {
				logs.ErrorLog(err)
				return
			}

			if err := w.Flush(); err != nil {
				logs.ErrorLog(err)
				return
			}
		}
		_, _ = fmt.Fprintf(w, "event: closed\ndata: Stream finished successfully\n\n")
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

var ansiRegex = regexp.MustCompile(`\033\[(\d+)(;1)?m([\S]+)\033\[0m`)

func ConvertANSICodeToCSS(input []byte) []byte {
	// Regex to match ANSI escape codes (like \033[31m for red)

	return ansiRegex.ReplaceAllFunc(input, func(match []byte) []byte {
		groups := ansiRegex.FindSubmatch(match)
		if len(groups) > 1 {
			if color, exists := ansiToCSS[gotools.BytesToString(groups[1])]; exists {
				return gotools.StringToBytes(fmt.Sprintf(`<span style="color:%s;">%s</span>`, color, groups[3]))
			}
		}
		return nil
	})
}

type LogWriter struct {
	buf *bytes.Buffer
	ch  chan []byte
}

func (l *LogWriter) Write(p []byte) (n int, err error) {

	f := len(p)
	select {
	case l.ch <- ConvertANSICodeToCSS(p):
	default:
	}

	return f, nil
}

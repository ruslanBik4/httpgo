/*
 * Copyright (c) 2024-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"bytes"
	"net"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/logs"
)

func RunRedirectNoSecure(cfg *CfgHttp) *fasthttp.Server {
	ln, err := net.Listen("tcp", cfg.PortRedirect)
	if err != nil {
		// port is occupied - work without redirection
		logs.ErrorLog(err, "port is occupied - work without redirection")
		return nil
	}
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			uri := ctx.Request.URI()
			uri.SetScheme("https")
			if h := bytes.Split(uri.Host(), []byte(":")); len(h) > 1 {
				uri.SetHostBytes(h[0])
			}

			ctx.RedirectBytes(uri.FullURI(), fasthttp.StatusMovedPermanently)
			logs.DebugLog("redirect %s", uri.FullURI())
		},
		Logger:          &fastHTTPLogger{},
		ErrorHandler:    renderError,
		CloseOnShutdown: true,
	}
	go func() {
		logs.StatusLog("Redirect service starting %s on port %s", time.Now(), cfg.PortRedirect)
		err = s.Serve(ln)
		if err != nil {
			logs.ErrorLog(err, "fasthttpServe")
		}
	}()
	return s
}

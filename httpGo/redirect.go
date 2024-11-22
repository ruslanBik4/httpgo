/*
 * Copyright (c) 2024. Author: Ruslan Bikchentaev. All rights reserved.
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

func RunRedirectNoSecure(port string) {
	if ln, err := net.Listen("tcp", port); err != nil {
		// port is occupied - work without redirection
		logs.ErrorLog(err, "port is occupied - work without redirection")
	} else {
		go func() {
			logs.StatusLog("Redirect service starting %s on port %s", time.Now(), port)
			err = fasthttp.Serve(ln, func(ctx *fasthttp.RequestCtx) {
				uri := ctx.Request.URI()
				uri.SetScheme("https")
				if h := bytes.Split(uri.Host(), []byte(":")); len(h) > 1 {
					uri.SetHostBytes(h[0])
				}

				ctx.RedirectBytes(uri.FullURI(), fasthttp.StatusMovedPermanently)
				logs.DebugLog("redirect %s", string(uri.FullURI()))
			})
			if err != nil {
				logs.ErrorLog(err, "fasthttpServe")
			}
		}()
	}
}

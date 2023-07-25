/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package auth

import "github.com/valyala/fasthttp"

type FncAuth interface {
	Auth(ctx *fasthttp.RequestCtx) bool
	AdminAuth(ctx *fasthttp.RequestCtx) bool
	String() string
}

func SetAuthManager(ctx *fasthttp.RequestCtx, a FncAuth) {
	if a != nil {
		ctx.SetUserValue(authManager, a)
	}
}

func GetAuthManager(ctx *fasthttp.RequestCtx) (a FncAuth, ok bool) {
	a, ok = ctx.UserValue(authManager).(FncAuth)
	return
}

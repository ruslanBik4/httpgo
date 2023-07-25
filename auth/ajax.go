/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package auth

import (
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/views"
)

// AjaxOnly wrap auth manager for ajax endpoint on case full refresh web-page
type AjaxOnly struct {
	auth FncAuth
}

func NewAjaxOnly(auth FncAuth) *AjaxOnly {
	return &AjaxOnly{auth: auth}
}

func (r *AjaxOnly) Auth(ctx *fasthttp.RequestCtx) bool {
	return !views.IsAJAXRequest(&ctx.Request) || (r.auth != nil && r.auth.Auth(ctx)) ||
		r.GetAuthManager(ctx, false)
}

func (r *AjaxOnly) AdminAuth(ctx *fasthttp.RequestCtx) bool {
	return !views.IsAJAXRequest(&ctx.Request) || (r.auth != nil && r.auth.AdminAuth(ctx)) ||
		r.GetAuthManager(ctx, true)
}

func (r *AjaxOnly) GetAuthManager(ctx *fasthttp.RequestCtx, isAdmin bool) bool {
	a, ok := GetAuthManager(ctx)
	// without manager allowed request
	return !ok || isAdmin && a.AdminAuth(ctx) || !isAdmin && a.Auth(ctx)
}

func (r *AjaxOnly) String() string {
	if r.auth != nil {
		return r.auth.String()
	}
	return `ajax wrap for auth`
}

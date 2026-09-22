/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package auth

import (
	"fmt"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/gotools"
)

type AuthBearer struct {
	tokens Tokens
}

func NewAuthBearer(tokens Tokens) *AuthBearer {
	if tokens == nil {
		tokens = NewMapTokens(tokenExpires)
	}

	return &AuthBearer{tokens}
}

// TokenIssuer is the minimal capability a route needs from whatever
// auth.FncAuth is currently wired in as the app's auth manager: mint a
// Bearer token for a freshly-authenticated TokenData. Neither FncAuth nor
// FncPasskey (passkey.go) declares this - both only guarantee
// Auth/AdminAuth/String - so a route that issues its own token (e.g.
// /user/signin/'s password login) asserts to this interface instead of a
// concrete type. *AuthBearer satisfies it directly (NewToken below);
// *WebAuthnPasskey satisfies it too, by promotion, since it embeds
// *AuthBearer - so a route written against TokenIssuer works unchanged
// whichever of the two the app hands to apis.NewApis(ctx, routes, fncAuth)
// as its app-wide auth manager.
type TokenIssuer interface {
	NewToken(userData TokenData) (string, error)
}

func (a *AuthBearer) NewToken(userData TokenData) (string, error) {
	return a.tokens.NewToken(userData)
}

func (a *AuthBearer) RemoveToken(s string) error {
	return a.tokens.RemoveToken(s)
}

func (a *AuthBearer) GetToken(ctx *fasthttp.RequestCtx) TokenData {
	bearer := a.getBearer(ctx)
	if bearer == "" {
		return nil
	}

	return a.tokens.GetToken(bearer)
}

var regBearer = regexp.MustCompile(`Bearer\s+(\S+)`)

func (a *AuthBearer) getBearer(ctx *fasthttp.RequestCtx) string {
	b := regBearer.FindSubmatch(ctx.Request.Header.Peek("Authorization"))
	if len(b) == 0 {
		return ""
	}

	return gotools.BytesToString(b[1])
}

func (a *AuthBearer) String() string {

	return `implement auth for Bearer standart: 
	 user:` + getStringOfFnc(reflect.ValueOf(a.Auth).Pointer()) + `
	 admin: ` + getStringOfFnc(reflect.ValueOf(a.AdminAuth).Pointer())
}

func (a *AuthBearer) Auth(ctx *fasthttp.RequestCtx) bool {

	token := a.GetToken(ctx)
	if token == nil {
		return false
	}

	ctx.SetUserValue(UserValueToken, token)

	return true
}

func (a *AuthBearer) AdminAuth(ctx *fasthttp.RequestCtx) bool {

	return a.Auth(ctx) && (ctx.UserValue(UserValueToken).(TokenData).IsAdmin())
}

func getStringOfFnc(pc uintptr) string {
	fnc := runtime.FuncForPC(pc)

	fName, line := fnc.FileLine(0)
	fncName := fnc.Name()
	shortName := strings.TrimSuffix(path.Base(fncName), "-fm")

	return fmt.Sprintf(`<a href='https://pkg.go.dev/%s/%s' target='_blank'>%s(), %s:%d</a>`,
		path.Dir(fncName),
		strings.ReplaceAll(strings.ReplaceAll(shortName, ".(*", "#"), ")", ""),
		shortName,
		path.Base(fName),
		line,
	)
}

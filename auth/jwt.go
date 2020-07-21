// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"

	"github.com/valyala/fasthttp"
)

type AuthBearer struct {
	tokens Tokens
}

func NewAuthBearer(tokens Tokens) AuthBearer {
	if tokens == nil {
		tokens = &mapTokens{
			expiresIn: tokenExpires,
			tokens:    make(map[int64]*mapToken, 0),
		}
	}

	return AuthBearer{tokens}
}

func (a AuthBearer) AddToken(hash int64, id int, ctx map[string]interface{}) int64 {
	return a.tokens.addToken(hash, id, ctx)
}

func (a AuthBearer) GetToken(ctx *fasthttp.RequestCtx) int64 {
	bearer := a.getBearer(ctx)
	if bearer == "" {
		return -1
	}

	return a.tokens.getToken(bearer)
}

func (a AuthBearer) getBearer(ctx *fasthttp.RequestCtx) string {
	b := ctx.Request.Header.Peek("Authorization")
	if len(b) == 0 || !bytes.HasPrefix(b, []byte("Bearer ")) {
		return ""
	}

	return string(bytes.TrimPrefix(b, []byte("Bearer ")))
}

func (a AuthBearer) String() string {

	return `implement auth for Bearer standart: 
	 user: ` + getStringOfFnc(reflect.ValueOf(a.Auth).Pointer()) + `
	 admin: ` + getStringOfFnc(reflect.ValueOf(a.AdminAuth).Pointer())
}

func (a AuthBearer) Auth(ctx *fasthttp.RequestCtx) bool {

	token := a.GetToken(ctx)
	if token < 0 {
		return false
	}

	ctx.SetUserValue(UserValueToken, token)

	return true
}

func (a AuthBearer) AdminAuth(ctx *fasthttp.RequestCtx) bool {

	return a.Auth(ctx) && (ctx.UserValue(UserValueToken).(*mapToken).isAdmin)
}

func getStringOfFnc(pc uintptr) string {
	fnc := runtime.FuncForPC(pc)
	fName, line := fnc.FileLine(0)

	return fmt.Sprintf("%s:%d %s()", fName, line, fnc.Name())
}

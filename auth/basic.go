// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"bytes"
	"encoding/base64"
	"regexp"

	"github.com/pkg/errors"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"
)

type fncNewTokenData func(ctx *fasthttp.RequestCtx, user, pass []byte) TokenData

type AuthBasic struct {
	tokens Tokens
	// maps basic string to tokens
	mapBasicToTokens map[string]string
	NewTokenData     fncNewTokenData
}

func NewAuthBasic(tokens Tokens, fnc fncNewTokenData) *AuthBasic {
	return &AuthBasic{tokens, make(map[string]string), fnc}
}

func (a *AuthBasic) Auth(ctx *fasthttp.RequestCtx) bool {

	b := a.getBasic(ctx)
	if len(b) == 0 {
		return false
	}
	token, ok := a.mapBasicToTokens[string(b)]
	if ok {
		data := a.tokens.GetToken(token)
		if data != nil {
			ctx.Response.Header.Set("token", token)
			return true
		}
	}

	u, p, ok := a.getUserPass(b)
	if ok {
		data := a.NewTokenData(ctx, u, p)
		if data == nil {
			return false
		}

		token, err := a.tokens.NewToken(data)
		if err != nil {
			logs.ErrorLog(errors.Wrap(err, ""))
			return false
		}

		a.mapBasicToTokens[string(b)] = token
		ctx.Response.Header.Set("token", token)

		return true
	}

	ctx.Response.Header.Set("WWW-Authenticate", `Basic realm="User Visible Realm", charset="UTF-8"`)
	logs.DebugLog("%s:'%s'", string(p), string(u))

	return false
}

func (a *AuthBasic) AdminAuth(ctx *fasthttp.RequestCtx) bool {
	return false
}

func (a *AuthBasic) String() string {
	return "Basic access authentication"
}

var regBasic = regexp.MustCompile(`Basic\s+(\S+)`)

func (a *AuthBasic) getBasic(ctx *fasthttp.RequestCtx) []byte {
	b := regBasic.FindSubmatch(ctx.Request.Header.Peek("Authorization"))
	if len(b) == 0 {
		return nil
	}

	return b[1]
}

func (a *AuthBasic) getUserPass(b []byte) (user []byte, pass []byte, ok bool) {
	enc := base64.StdEncoding
	dst := make([]byte, enc.DecodedLen(len(b)))
	n, err := enc.Decode(dst, b)
	if err != nil {
		logs.ErrorLog(err, "Decode "+string(b))
		return nil, nil, false
	}

	pair := bytes.SplitN(dst[:n], []byte(":"), -1)

	if len(pair) != 2 {
		logs.DebugLog("'%s' <- '%s'", string(dst[:n]), string(b))
		return nil, nil, false
	}

	return pair[0], pair[1], true
}

/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package auth

import (
	"reflect"

	"github.com/valyala/fasthttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

type OAuth2 struct {
	*oauth2.Config
	*AuthBearer
}

func NewOAuth2(tokens Tokens, clientID, redirectURL string) *OAuth2 {

	return &OAuth2{
		&oauth2.Config{
			ClientID:     clientID,
			ClientSecret: "x",
			Endpoint:     endpoints.GitHub,
			RedirectURL:  redirectURL,
			Scopes:       []string{"read"},
		},
		NewAuthBearer(tokens),
	}
}

func (a *OAuth2) Auth(ctx *fasthttp.RequestCtx) bool {
	token := a.GetToken(ctx)
	if token == nil {
		return false
	}

	ctx.SetUserValue(UserValueToken, token)

	return true
}

func (a *OAuth2) AdminAuth(ctx *fasthttp.RequestCtx) bool {
	return a.Auth(ctx) && (ctx.UserValue(UserValueToken).(TokenData).IsAdmin())
}

func (a *OAuth2) String() string {
	return `implement auth for Bearer standart: 
	 user:` + getStringOfFnc(reflect.ValueOf(a.Auth).Pointer()) + `
	 admin: ` + getStringOfFnc(reflect.ValueOf(a.AdminAuth).Pointer())
}

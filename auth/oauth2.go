/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package auth

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"

	"github.com/ruslanBik4/logs"
)

type AuthServer uint8

const (
	Amazon = iota
	Bitbucket
	GitHub
	GitLab
	Facebook
	Instagram
	LinkedIn
	Microsoft
	PayPal
)

type OAuth2 struct {
	*oauth2.Config
	*AuthBearer
}

func NewOAuth2(clientID, clientSecret, redirectURL string, scopes ...string) *OAuth2 {

	return &OAuth2{
		&oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			//Endpoint:     nil,
			RedirectURL: redirectURL,
			Scopes:      scopes,
		},
		NewAuthBearer(nil),
	}
}

func NewOAuth2WithCustomTokens(tokens Tokens, clientID, clientSecret, redirectURL string, scopes ...string) *OAuth2 {

	return &OAuth2{
		&oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			//Endpoint:     nil,
			RedirectURL: redirectURL,
			Scopes:      scopes,
		},
		NewAuthBearer(tokens),
	}
}

func (a *OAuth2) DoAuth(ctx *fasthttp.RequestCtx, s AuthServer) error {
	switch s {
	case Amazon:
		a.Endpoint = endpoints.Amazon
	case Bitbucket:
		a.Endpoint = endpoints.Bitbucket
	case GitHub:
		a.Endpoint = endpoints.GitHub
	case GitLab:
		a.Endpoint = endpoints.GitLab
	case Facebook:
		a.Endpoint = endpoints.Facebook
	case Instagram:
		a.Endpoint = endpoints.Instagram
	case LinkedIn:
		a.Endpoint = endpoints.LinkedIn
	case Microsoft:
		a.Endpoint = endpoints.Microsoft
	case PayPal:
		a.Endpoint = endpoints.PayPal
	default:
		return errors.New("unknown server")
	}
	url := a.AuthCodeURL("read")
	logs.StatusLog(url)
	ctx.Redirect(url, 200)
	return nil
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

/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"go/types"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/auth"
)

var (
	ParamsID = apis.InParam{
		Name: "id",
		Req:  false,
		Type: apis.NewTypeInParam(types.Int32),
	}
	ParamsIDReq = apis.InParam{
		Name: "id",
		Req:  true,
		Type: apis.NewTypeInParam(types.Int32),
	}
	ParamsLang = apis.InParam{
		Name:     "lang",
		Desc:     "language of response (also may use header 'Accept-Language')",
		DefValue: apis.NewDefValueHeader("Accept-Language", "en"),
		Req:      true,
		Type:     apis.NewTypeInParam(types.String),
	}
	ParamsHTML = apis.InParam{
		Name:     "html",
		Desc:     "need for get result in html instead JSON",
		DefValue: apis.NewDefValueHeader("HX-Request", "false"),
		Req:      false,
		Type:     apis.NewTypeInParam(types.Bool),
	}
	ParamsGetFormActions = apis.InParam{
		Name: "is_get_form_actions",
		Desc: "need to get form actions in response",
		Req:  false,
		Type: apis.NewTypeInParam(types.Bool),
	}
	ParamsLimit = apis.InParam{
		Name:     "limit",
		Desc:     "max count of queries results",
		DefValue: 100,
		Req:      true,
		Type:     apis.NewTypeInParam(types.Int),
	}
	ParamsOffset = apis.InParam{
		Name: "offset",
		Desc: "offset of queries results",
		Req:  false,
		Type: apis.NewTypeInParam(types.Int),
	}
	ParamsEmail = apis.InParam{
		Name: "email",
		Desc: "email for login",
		Req:  true,
		Type: apis.NewTypeInParam(types.String),
	}
	ParamsPassword = apis.InParam{
		Name: "key",
		Desc: "password or other key word (on future)",
		Req:  true,
		Type: apis.NewTypeInParam(types.String),
	}
	// ParamsLogin identifies the account for a passkey (WebAuthn) ceremony
	// - register/begin (auth.WebAuthnPasskey.BeginRegistration) reads this
	// via ctx.UserValue(auth.PasskeyLoginParam). Its Name is set FROM
	// auth.PasskeyLoginParam (not the other way around: auth can't import
	// this package - see PasskeyLoginParam's own doc comment for the
	// import-cycle reason) so the two stay in sync automatically.
	//
	// NOTE: register/begin's request today is a fetch() POST from
	// passkey-auth.js, most likely with a JSON body (BeginLogin, the
	// sibling ceremony, parses its own "login" field straight out of the
	// raw JSON body rather than through this InParam mechanism at all).
	// This wasn't verified against apis's own param-extraction code this
	// session - confirm ctx.UserValue(auth.PasskeyLoginParam) actually gets
	// populated for that request's real Content-Type before relying on
	// this; if apis only extracts from form/multipart bodies or a query
	// string, either adjust the client call or have BeginRegistration fall
	// back to a manual json.Unmarshal(ctx.PostBody(), ...) the same way
	// BeginLogin already does.
	ParamsLogin = apis.InParam{
		Name: auth.PasskeyLoginParam,
		Desc: "login (email or username) identifying the account for a passkey ceremony",
		Req:  true,
		Type: apis.NewTypeInParam(types.String),
	}
	ParamsWhere = apis.InParam{
		Name: "where",
		Desc: "conditions for query ('where' clause)",
		Type: apis.NewTypeInParam(types.String),
	}
	ParamsOrderBy = apis.InParam{
		Name: "order_by",
		Desc: "conditions for sort queries data ('order by' clause)",
		Type: apis.NewTypeInParam(types.String),
	}
	ParamsSelect = apis.InParam{
		Name: "select[]",
		Desc: "list of columns for query ('select' clause)",
		Type: apis.NewSliceTypeInParam(types.String),
	}
	ParamsName = apis.InParam{
		Name:     "name",
		Desc:     "name of parameters. operation, etc.",
		DefValue: apis.ApisValues(apis.ChildRoutePath),
		Req:      true,
		Type:     apis.NewTypeInParam(types.String),
	}
	BasicParams = []apis.InParam{
		ParamsHTML,
		ParamsLang,
	}
	APIQueriesParams = []apis.InParam{
		ParamsGetFormActions,
		ParamsLimit,
		ParamsOffset,
		ParamsWhere,
		ParamsOrderBy,
		ParamsSelect,
		ParamsLang,
		ParamsHTML,
	}
)

const PathVersion = "/api/v1"

type DbRouteType int

const (
	DbRouteType_Insert DbRouteType = iota
	DbRouteType_Update
	DbRouteType_Select
)

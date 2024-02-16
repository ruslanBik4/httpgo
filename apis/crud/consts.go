/*
 * Copyright (c) 2022-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"go/types"

	"github.com/ruslanBik4/httpgo/apis"
)

var (
	ParamsID = apis.InParam{
		Name: "id",
		Req:  false,
		Type: apis.NewTypeInParam(types.Int32),
	}
	ParamsLang = apis.InParam{
		Name:     "lang",
		Desc:     "language of response (also may use header 'Accept-Language')",
		DefValue: apis.NewDefValueHeader("Accept-Language", "en"),
		Req:      true,
		Type:     apis.NewTypeInParam(types.String),
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
		DefValue: 1000,
		Req:      true,
		Type:     apis.NewTypeInParam(types.Int),
	}
	ParamsOffset = apis.InParam{
		Name: "offset",
		Desc: "offset of queries results",
		Req:  false,
		Type: apis.NewTypeInParam(types.Int),
	}
	ParamsHTML = apis.InParam{
		Name: "html",
		Desc: "need for get result in html instead JSON",
		Req:  false,
		Type: apis.NewTypeInParam(types.Bool),
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

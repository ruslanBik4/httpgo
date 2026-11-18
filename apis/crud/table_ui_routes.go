/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"fmt"
	"go/types"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/logs"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/auth"
	"github.com/ruslanBik4/httpgo/views"
)

// ParamsSetBool are the params of the route that sets a boolean column of the
// whole table (TableSetBool).
var ParamsSetBool = []apis.InParam{
	{
		Name: "column",
		Desc: "name of the boolean column",
		Req:  true,
		Type: apis.NewTypeInParam(types.String),
	},
	{
		Name: "value",
		Desc: "true or false",
		Req:  true,
		Type: apis.NewTypeInParam(types.String),
	},
}

// IsAdminRequest reports that the user of the request is an administrator.
func IsAdminRequest(ctx *fasthttp.RequestCtx) bool {
	token, ok := ctx.UserValue(auth.UserValueToken).(auth.TokenData)

	return ok && token.IsAdmin()
}

// TableSetBool sets a boolean column of EVERY row of the table ("set_bool_all",
// the buttons 'true' and 'false' of the filter of the table - the route is for
// administrators). The visible rows are changed by the "put/batch" route.
func TableSetBool(db *dbEngine.DB, table dbEngine.Table) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (any, error) {
		column, _ := ctx.UserValue("column").(string)
		value, _ := ctx.UserValue("value").(string)

		query, args, err := BuildSetBoolSQL(table, column, value)
		if err != nil {
			return map[string]string{"column": err.Error()}, apis.ErrWrongParamsList
		}

		var changed int64
		if err = db.Conn.SelectOneAndScan(ctx, &changed, query, args...); err != nil {
			return nil, errors.Wrap(err, table.Name())
		}

		logs.StatusLog("%s.%s set %s: %d rows", table.Name(), column, value, changed)
		ctx.SetStatusCode(fasthttp.StatusAccepted)

		return map[string]any{
			"updated": changed,
			"message": fmt.Sprintf("%s: %d rows changed", column, changed),
		}, nil
	}
}

// TableCard writes the short card of a row of the table as HTML: what is under
// the mouse over a foreign key of a table view. Columns for administrators only
// ({"admin_only": true} in the comment of the column) are not shown to others.
func TableCard(table dbEngine.Table, pk string) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (any, error) {
		id := ctx.UserValue(pk)
		if id == nil {
			return map[string]string{pk: "required params"}, apis.ErrWrongParamsList
		}

		title, fields, found, err := ReadCard(ctx, table, pk, id, ctx.UserValue(ParamsLang.Name), IsAdminRequest(ctx))
		if err != nil {
			return nil, errors.Wrap(err, table.Name())
		}
		if !found {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return nil, nil
		}

		views.WriteHeadersHTML(ctx)
		_, err = ctx.WriteString(RenderCard(table.Name(), id, title, fields))

		return nil, err
	}
}

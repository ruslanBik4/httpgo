// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crud

import (
	"github.com/pkg/errors"
	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/valyala/fasthttp"
)

func TableSelect(table dbEngine.Table, params []string) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (interface{}, error) {
		args := make([]interface{}, 0, len(params))
		colNames := make([]string, 0, len(params))
		for _, key := range params {
			if v := ctx.UserValue(key); v != nil {
				args = append(args, v)
				colNames = append(colNames, key)
			}
		}

		res := make([]map[string]interface{}, 0)
		err := table.SelectAndRunEach(ctx,
			func(values []interface{}, columns []dbEngine.Column) error {
				r := make(map[string]interface{}, len(columns))
				for key, col := range columns {
					r[col.Name()] = values[key]
				}

				res = append(res, r)

				return nil
			},
			dbEngine.WhereForSelect(colNames...),
			dbEngine.ArgsForSelect(args...),
		)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}

		if len(res) == 0 {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return nil, nil
		}

		return res, nil
	}
}

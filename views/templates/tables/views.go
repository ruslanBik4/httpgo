// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tables

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/system/routeTable"
)

func ViewRoute(preRoute string, table dbEngine.Table, DB *dbEngine.DB) *apis.ApiRoute {
	return &apis.ApiRoute{
		Desc:     "show data of table " + table.Name(),
		NeedAuth: true,
		Fnc:      TableView(preRoute, table, DB),
	}
}

func TableView(preRoute string, table dbEngine.Table, DB *dbEngine.DB) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (interface{}, error) {
		rows := make([][]interface{}, 0)
		row := make([]interface{}, len(table.Columns()))

		href := preRoute + table.Name() + "/form"
		if table.Name() == "forms" {
			href = preRoute + table.Name() + "/form_editor"
		}

		onClick := "$('#content').load(this.href); return false;"
		row[0] = fmt.Sprintf(`<a href="%s?html" onclick="%s">New</a>`,
			href, onClick)

		rows = append(rows, row)

		err := table.SelectAndRunEach(ctx,
			func(row []interface{}, columns []dbEngine.Column) error {

				row[0] = fmt.Sprintf(`<a href="%s?id=%v&html" onclick="%s">Edit</a>`,
					href, row[0], onClick)
				for i, col := range columns[1:] {
					if v := getForeigthValue(DB, col, row[i+1], ctx.UserValue("lang")); v != nil {
						row[i+1] = v
					} else {
						row[i+1] = convertStrings(row[i+1])
					}
				}
				rows = append(rows, row)

				return nil
			},
		)
		if err != nil {
			return nil, errors.Wrap(err, "select")
		}

		views.RenderHTMLPage(ctx, layouts.WritePutHeadForm)

		colDecors := make([]*forms.ColumnDecor, len(table.Columns()))
		for i, col := range table.Columns() {
			colDecors[i] = forms.NewColumnDecor(col, nil)
		}
		routeTable.WriteTableRow(ctx, colDecors, rows)

		return nil, nil
	}
}

func getForeigthValue(DB *dbEngine.DB, col dbEngine.Column, val, lang interface{}) interface{} {
	if strings.HasPrefix(col.Name(), "id_") {
		table, ok := DB.Tables[strings.TrimPrefix(col.Name(), "id_")]
		if ok {
			colName := "name"
			if l, ok := lang.(string); ok && (table.FindColumn(colName+"_"+l) != nil) {
				colName += "_" + l
			}
			err := table.SelectAndRunEach(context.Background(),
				func(values []interface{}, columns []dbEngine.Column) error {
					val = values[0]
					return nil
				},
				dbEngine.ColumnsForSelect(colName),
				dbEngine.WhereForSelect("id"),
				dbEngine.ArgsForSelect(val),
			)
			if err != nil {
				logs.ErrorLog(err, "")
			}
			return val
		}
	}

	return nil
}

func getForeigthVal(DB *dbEngine.DB, colDec *forms.ColumnDecor, id interface{}) {
	if strings.HasPrefix(colDec.Name(), "id_") {
		table, ok := DB.Tables[strings.TrimPrefix(colDec.Name(), "id_")]
		if ok {
			colDec.SelectOptions = make(map[string]string)
			// selectParams := []dbEngine.BuildSqlOptions{
			//
			// }
			// if id != nil {
			// 	selectParams = append(selectParams, dbEngine.WhereForSelect("id"))
			// 	selectParams = append(selectParams, dbEngine.)
			// }
			err := table.SelectAndRunEach(context.Background(),
				func(values []interface{}, columns []dbEngine.Column) error {
					colDec.SelectOptions[values[1].(string)] = strconv.Itoa(int(values[0].(int32)))
					if id != nil {
						colDec.Value = id
					}
					return nil
				},
				dbEngine.ColumnsForSelect("id", "name"),
			)
			if err != nil {
				logs.ErrorLog(err, "")
			}
		}
	}

}

func convertStrings(values interface{}) interface{} {
	// todo- move to dbEngine
	switch v := values.(type) {
	case pgtype.VarcharArray:
		str := make([]string, len(v.Elements))
		for i, val := range v.Elements {
			str[i] = val.String
		}
		return str
	case *pgtype.VarcharArray:
		str := make([]string, len(v.Elements))
		for i, val := range v.Elements {
			str[i] = val.String
		}
		return str
	default:
		logs.DebugLog("%T", values)
		return values
	}
}

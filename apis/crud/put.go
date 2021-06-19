// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crud

import (
	"fmt"
	"go/types"
	"io/ioutil"
	"mime/multipart"
	"strings"

	"github.com/pkg/errors"
	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"
)

// TableInsert insert data of params into table
func TableInsert(preRoute string, DB *dbEngine.DB, table dbEngine.Table, params []string) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (interface{}, error) {

		args := make([]interface{}, 0, len(params))
		colSel := make([]string, 0, len(params))
		msg := ""
		for _, name := range params {
			arg := ctx.UserValue(name)
			if arg == nil {
				continue
			}

			colName := strings.TrimSuffix(name, "[]")
			col := table.FindColumn(colName)
			if col == nil {
				logs.ErrorLog(dbEngine.ErrNotFoundColumn{Table: table.Name(), Column: colName})
				continue
			}

			if col.BasicType() == types.UnsafePointer {

				switch val := arg.(type) {
				case nil, string:
				case []*multipart.FileHeader:

					names, bytea, err := readByteA(val)
					if err != nil {
						logs.DebugLog(names)
						return map[string]string{colName: err.Error()},
							apis.ErrWrongParamsList
					}

					switch len(bytea) {
					case 0:
					case 1:
						args = append(args, bytea[0])
						colSel = append(colSel, colName)
						msg += "[file]"
					default:
						args = append(args, bytea)
						colSel = append(colSel, colName)
						msg += "[files]"
					}
				default:
					return map[string]string{colName: fmt.Sprintf("%v", val)},
						apis.ErrWrongParamsList
				}

				continue
			}

			args = append(args, arg)
			colSel = append(colSel, colName)
			msg += fmt.Sprintf(" %v", arg)

		}

		id, err := table.Insert(ctx,
			dbEngine.ColumnsForSelect(colSel...),
			dbEngine.ArgsForSelect(args...),
		)
		if err != nil {
			return CreateErrResult(err)
		}

		return createResult(ctx, id, msg, colSel, preRoute+table.Name())
	}
}

// TableUpdate
func TableUpdate(preRoute string, table dbEngine.Table, columns, priColumns []string) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (interface{}, error) {

		badParams := make(map[string]string, 0)
		for _, key := range priColumns {
			if ctx.UserValue(key) == nil {
				badParams[key] = "required params"
			}
		}

		if len(badParams) > 0 {
			return badParams, apis.ErrWrongParamsList

		}

		args := make([]interface{}, 0, len(columns))
		colSel := make([]string, 0, len(columns))
		msg := ""
		for _, name := range columns {
			isPrimary := false
			for _, priName := range priColumns {
				if name == priName {
					arg := ctx.UserValue("new." + priName)
					if arg != nil {
						colSel = append(colSel, priName)
						args = append(args, arg)
						msg += fmt.Sprintf(" %v", arg)
					}
					isPrimary = true
					break
				}
			}

			if isPrimary {
				continue
			}

			arg := ctx.UserValue(name)
			if arg == nil {
				continue
			}

			colName := strings.TrimSuffix(name, "[]")
			col := table.FindColumn(colName)
			if col.BasicType() == types.UnsafePointer {
				switch val := arg.(type) {
				case nil, string:
				case []*multipart.FileHeader:
					names, bytea, err := readByteA(val)
					if err != nil {
						logs.DebugLog(names)
						return map[string]string{colName: err.Error()}, apis.ErrWrongParamsList
					}

					switch len(bytea) {
					case 0:
					case 1:
						args = append(args, bytea[0])
						colSel = append(colSel, colName)
						msg += "[file]"
					default:
						args = append(args, bytea)
						colSel = append(colSel, colName)
						msg += "[files]"
					}
				default:
					return map[string]string{colName: fmt.Sprintf("%v", val)}, apis.ErrWrongParamsList
				}

				continue
			}

			args = append(args, arg)
			colSel = append(colSel, colName)
			msg += fmt.Sprintf(" %v", arg)

		}

		for _, name := range priColumns {
			args = append(args, ctx.UserValue(name))
		}

		i, err := table.Update(ctx,
			dbEngine.ColumnsForSelect(colSel...),
			dbEngine.WhereForSelect(priColumns...),
			dbEngine.ArgsForSelect(args...),
		)
		if err != nil {
			return nil, err
		}
		if i <= 0 {
			logs.DebugLog(colSel, priColumns)
			logs.DebugLog(args)
			return map[string]string{"update": fmt.Sprintf("%d", i)}, apis.ErrWrongParamsList
		}

		msg = "Success update: " + strings.Join(colSel, ", ") + " values:\n" + msg

		ctx.SetStatusCode(fasthttp.StatusAccepted)
		g, ok := ctx.UserValue(ParamsGetFormActions.Name).(bool)
		if ok && g {
			urlSuffix := "/browse"
			lang := ctx.UserValue("lang")
			if l, ok := lang.(string); ok {
				urlSuffix += "?lang=" + l
			}

			return insertResult{
				FormActions: []FormActions{
					{
						Typ: "redirect",
						Url: preRoute + table.Name() + urlSuffix,
					},
				},
				Msg: msg,
			}, nil
		}

		return msg, nil
	}
}

type insertResult struct {
	FormActions []FormActions `json:"formActions"`
	Id          int64         `json:"id,omitempty"`
	Msg         string        `json:"message"`
}

func readByteA(fHeaders []*multipart.FileHeader) ([]string, [][]byte, error) {
	bytea := make([][]byte, len(fHeaders))
	names := make([]string, len(fHeaders))
	for i, fHeader := range fHeaders {

		f, err := fHeader.Open()
		if err != nil {
			logs.DebugLog(err, fHeader)
			return nil, nil, errors.Wrap(err, fHeader.Filename)
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			logs.DebugLog(err, fHeader)
			return nil, nil, errors.Wrap(err, "read "+fHeader.Filename)
		}

		bytea[i] = b
		names[i] = fHeader.Filename
	}

	return names, bytea, nil
}

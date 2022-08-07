/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package crud

import (
	"bytes"
	"fmt"
	"go/types"
	"io/ioutil"
	"mime/multipart"
	"strings"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/logs"
)

// TableInsert insert data of params into table
func TableInsert(preRoute string, DB *dbEngine.DB, table dbEngine.Table, params []string) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (interface{}, error) {

		args := make([]interface{}, 0, len(params))
		colSel := make([]string, 0, len(params))
		badParams := make(map[string]string, 0)
		buf := bytes.NewBufferString("")
		for _, name := range params {
			arg := ctx.UserValue(name)
			if arg == nil {
				continue
			}

			AddColumnAndValue(name, table, arg, args, colSel, buf, badParams)
		}

		if len(badParams) > 0 {
			return badParams, apis.ErrWrongParamsList
		}

		id, err := table.Insert(ctx,
			dbEngine.ColumnsForSelect(colSel...),
			dbEngine.ArgsForSelect(args...),
		)
		if err != nil {
			return CreateErrResult(err)
		}

		return RenderCreatedResult(ctx, id, buf, colSel, preRoute+table.Name())
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

		args := make([]interface{}, 0, len(columns))
		colSel := make([]string, 0, len(columns))
		buf := bytes.NewBufferString("")
		for _, name := range columns {
			isPrimary := false
			for _, priName := range priColumns {
				if name == priName {
					arg := ctx.UserValue("new." + priName)
					if arg != nil {
						colSel = append(colSel, priName)
						args = append(args, arg)
						_, err := fmt.Fprintf(buf, " %v", arg)
						if err != nil {
							return nil, err
						}
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

			AddColumnAndValue(name, table, arg, args, colSel, buf, badParams)
		}

		if len(badParams) > 0 {
			return badParams, apis.ErrWrongParamsList
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

		return RenderAcceptedResult(ctx, colSel, buf, preRoute+table.Name())
	}
}

func AddColumnAndValue(name string, table dbEngine.Table, arg interface{}, args []interface{}, colSel []string,
	buf *bytes.Buffer, badParams map[string]string) {

	colName := strings.TrimSuffix(name, "[]")
	col := table.FindColumn(colName)
	if col == nil {
		badParams[colName] = dbEngine.ErrNotFoundColumn{Table: table.Name(), Column: colName}.Error()
		return
	}

	switch col.BasicType() {
	case types.UnsafePointer:
		switch val := arg.(type) {
		case nil, string:
			badParams[colName] = "wrong type of file"
		case []*multipart.FileHeader:
			names, bytea, err := readByteA(val)
			if err != nil {
				logs.DebugLog(names)
				badParams[colName] = err.Error()
			}

			switch len(bytea) {
			case 0:
				badParams[colName] = "empty file"
			case 1:
				args = append(args, bytea[0])
				colSel = append(colSel, colName)
				buf.WriteString("[file]")
			default:
				args = append(args, bytea)
				colSel = append(colSel, colName)
			}
			buf.WriteString("[file]")
		default:
			badParams[colName] = fmt.Sprintf("unknown type of value: %T", val)
		}

	default:
		args = append(args, arg)
		colSel = append(colSel, colName)
		_, err := fmt.Fprintf(buf, " %v", arg)
		logs.ErrorLog(err)
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

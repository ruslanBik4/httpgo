/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"mime/multipart"
	"strings"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/logs"
)

// TableInsert insert data of params into table
func TableInsert(preRoute string, table dbEngine.Table, params []string) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (any, error) {

		args := make([]any, 0, len(params))
		colSel := make([]string, 0, len(params))
		badParams := make(map[string]string, 0)
		buf := bytes.NewBufferString("")
		for _, name := range params {
			arg := ctx.UserValue(name)
			if arg == nil {
				continue
			}

			n, a := AddColumnAndValue(name, table, arg, buf, badParams)
			colSel = append(colSel, n)
			args = append(args, a)
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
	return func(ctx *fasthttp.RequestCtx) (any, error) {

		badParams := make(map[string]string, 0)
		for _, key := range priColumns {
			if ctx.UserValue(key) == nil {
				badParams[key] = "required params"
			}
		}

		args := make([]any, 0, len(columns))
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

			n, a := AddColumnAndValue(name, table, arg, buf, badParams)
			colSel = append(colSel, n)
			args = append(args, a)
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

func AddColumnAndValue(name string, table dbEngine.Table, arg any, buf io.Writer, badParams map[string]string) (string, any) {

	colName := strings.TrimSuffix(name, "[]")
	col := table.FindColumn(colName)
	if col == nil {
		badParams[colName] = dbEngine.ErrNotFoundColumn{Table: table.Name(), Column: colName}.Error()
		return "", nil
	}

	switch col.BasicType() {
	case types.UnsafePointer:
		switch val := arg.(type) {
		case nil, string:
			badParams[colName] = "wrong type of file"
		case []*multipart.FileHeader:
			names, bytea, err := ReadByteA(val)
			if err != nil {
				logs.DebugLog(names)
				badParams[colName] = err.Error()
				return "", nil
			}

			switch len(bytea) {
			case 0:
				badParams[colName] = "empty file"
			case 1:
				_, err := buf.Write([]byte("[file]"))
				if err != nil {
					logs.ErrorLog(err)
				}
				return colName, bytea[0]
			default:
				_, err := buf.Write([]byte("[files]"))
				if err != nil {
					logs.ErrorLog(err)
				}
				return colName, bytea
			}
		default:
			badParams[colName] = fmt.Sprintf("unknown type of value: %T", val)
		}

	default:
		_, err := fmt.Fprintf(buf, " %v", arg)
		logs.ErrorLog(err)
		return colName, arg
	}
	return "", nil
}

type insertResult struct {
	FormActions []FormActions `json:"formActions"`
	Id          int64         `json:"id,omitempty"`
	Msg         string        `json:"message"`
}

func ReadByteA(fHeaders []*multipart.FileHeader) ([]string, [][]byte, error) {
	bytea := make([][]byte, len(fHeaders))
	names := make([]string, len(fHeaders))
	for i, fHeader := range fHeaders {

		f, err := fHeader.Open()
		if err != nil {
			logs.DebugLog(err, fHeader)
			return nil, nil, errors.Wrap(err, fHeader.Filename)
		}

		b, err := io.ReadAll(f)
		if err != nil {
			logs.DebugLog(err, fHeader)
			return nil, nil, errors.Wrap(err, "read "+fHeader.Filename)
		}

		bytea[i] = b
		names[i] = fHeader.Filename
	}

	return names, bytea, nil
}

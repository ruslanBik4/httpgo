/*
 * Copyright (c) 2022-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/views"
)

func isNilOrEmpty[T any](v T) bool {
	// fast nil test
	var i interface{} = v
	// check by type
	switch x := i.(type) {
	case nil:
		return true

	case string:
		return x == ""

	case []byte:
		return len(x) == 0

	case []string:
		return len(x) == 0

	case map[string]any:
		return len(x) == 0

	case map[int]any:
		return len(x) == 0

	case map[string]string:
		return len(x) == 0

	case map[int]string:
		return len(x) == 0

	default:
		return false
	}
}

func RenderArrayResult[T any](ctx *fasthttp.RequestCtx, res []T, err error) (any, error) {

	if err != nil {
		return CreateErrResult(err)
	}
	if len(res) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return nil, nil
	}

	views.WriteJSONHeaders(ctx)
	return res, nil
}

func RenderAny[T any](ctx *fasthttp.RequestCtx, res T, err error) (any, error) {

	if err != nil {
		return CreateErrResult(err)
	}

	if isNilOrEmpty(res) {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return nil, nil
	}

	views.WriteJSONHeaders(ctx)
	return res, nil
}

func RenderCreatedResult(ctx *fasthttp.RequestCtx, id int64, buf *bytes.Buffer, colSel []string, url string) (any, error) {
	msg := fmt.Sprintf("Success insert: %s, values:\n%s", strings.Join(colSel, ", "), buf.String())

	ctx.SetStatusCode(fasthttp.StatusCreated)
	if res, ok := createResponse(ctx, msg, url+"/form?html"); ok {
		res.Id = id
		return res, nil
	}

	return id, nil
}

func RenderAcceptedResult(ctx *fasthttp.RequestCtx, colSel []string, buf *bytes.Buffer, url string) (any, error) {
	msg := fmt.Sprintf("Success update: %s, values:\n%s", strings.Join(colSel, ", "), buf.String())

	ctx.SetStatusCode(fasthttp.StatusAccepted)
	if res, ok := createResponse(ctx, msg, url+"/browse"); ok {
		return res, nil
	}

	return msg, nil
}

func createResponse(ctx *fasthttp.RequestCtx, msg, url string) (*insertResult, bool) {

	if g, ok := ctx.UserValue(ParamsGetFormActions.Name).(bool); ok && g {
		if l, ok := ctx.UserValue(ParamsLang.Name).(string); ok {
			url += "?lang=" + l
		}

		return &insertResult{
			FormActions: []FormActions{
				{
					Typ: "redirect",
					Url: url,
				},
			},
			Msg: msg,
		}, true
	}

	return nil, false
}

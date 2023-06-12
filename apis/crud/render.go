/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"bytes"
	"strings"

	"github.com/valyala/fasthttp"
)

func RenderCreatedResult(ctx *fasthttp.RequestCtx, id int64, buf *bytes.Buffer, colSel []string, url string) (any, error) {
	msg := "Success saving: " + strings.Join(colSel, ", ") + " values:\n" + buf.String()

	ctx.SetStatusCode(fasthttp.StatusCreated)

	if g, ok := ctx.UserValue(ParamsGetFormActions.Name).(bool); ok && g {
		url += "/form?html"

		lang := ctx.UserValue("lang")
		if l, ok := lang.(string); ok {
			url += "&lang=" + l
		}

		return insertResult{
			FormActions: []FormActions{
				{
					Typ: "redirect",
					Url: url,
				},
			},
			Id:  id,
			Msg: msg,
		}, nil
	}

	return id, nil
}

func RenderAcceptedResult(ctx *fasthttp.RequestCtx, colSel []string, buf *bytes.Buffer, route string) (any, error) {
	msg := "Success update: " + strings.Join(colSel, ", ") + " values:\n" + buf.String()

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
					Url: route + urlSuffix,
				},
			},
			Msg: msg,
		}, nil
	}

	return msg, nil
}

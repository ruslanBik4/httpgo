/*
 * Copyright (c) 2024-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package views

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"path"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/logs"
)

const (
	// render JSON from any data type
	jsonHEADERSContentType = "application/json; charset=utf-8"
	htmlHEADERSContentType = "text/html; charset=utf-8"
)

const ServerName = "name of server httpgo"
const AgeOfServer = "AGE"

// HEADERS - list standard header for html page - noinspection GoInvalidConstType
var HEADERS = map[string]string{
	"author":           "ruslanBik4",
	"Server":           ServerName,
	"Content-Language": "en,uk",
}

// WriteHeaders выдаем стандартные заголовки страницы
func WriteHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentEncoding("utf-8")
	age, ok := ctx.UserValue(AgeOfServer).(float64)
	if ok {
		ctx.Response.Header.Set("Age", fmt.Sprintf("%f", age))
	}
	ctx.Response.Header.SetLastModified(time.Now().Add(-(time.Second * time.Duration(age))))
	for key, value := range HEADERS {
		if key == "Server" {
			value = ctx.UserValue(ServerName).(string)
		}
		// set header ONLY if it not presents
		if len(ctx.Response.Header.Peek(key)) == 0 {
			ctx.Response.Header.Set(key, value)
		}
	}
}

// WriteJSONHeaders return standart headers for JSON
func WriteJSONHeaders(ctx *fasthttp.RequestCtx) {
	WriteHeaders(ctx)
	ctx.Response.Header.SetContentType(jsonHEADERSContentType)
}

func WriteHeadersHTML(ctx *fasthttp.RequestCtx) {
	WriteHeaders(ctx)
	ctx.Response.Header.SetContentType(htmlHEADERSContentType)
}

func WriteDownloadHeaders(ctx *fasthttp.RequestCtx, lastModify time.Time, fileName string, length int) {
	ctx.Response.Header.Set("Content-Description", "File Transfer")
	ctx.Response.Header.Set("Content-Transfer-Encoding", "binary")
	ctx.Response.Header.Set("Cache-Control", "must-revalidate")
	ctx.Response.Header.SetLastModified(lastModify)
	if length > 0 {
		ctx.Response.Header.SetContentLength(length)
	}

	ct, fileName := GetContentType(ctx, fileName)

	ctx.Response.Header.SetContentType(ct)
	ctx.Response.Header.Set("Content-Disposition", "attachment; filename="+fileName)
	if bytes.Contains(ctx.Request.Header.Peek("Cache-Control"), []byte("max-age=0, no-cache, no-store")) {
		ctx.Response.Header.Set("Cache-Control", "max-age=0, no-cache, no-store")
		ctx.Response.Header.Set("Pragma", "no-cache")
		ctx.Response.Header.Set("Expires", "Wed, 11 Jan 1984 05:00:00 GMT")
	} else {
		ctx.Response.Header.Set("Cache-Control", "must-revalidate")
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func GetContentType(ctx *fasthttp.RequestCtx, fileName string) (string, string) {
	ct := ""
	if ext := path.Ext(fileName); ext > "" {
		ct = mime.TypeByExtension(ext)
	}

	// empty extension or not found MIME type
	if ct == "" {
		ct = http.DetectContentType(ctx.Response.Body())
		if ext, err := mime.ExtensionsByType(ct); err != nil {
			logs.ErrorLog(err)
		} else if len(ext) > 0 {
			fileName += ext[0]
		}

	}
	return ct, fileName
}

func WriteCORSHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, X-Auth-Token, Origin, Authorization, X-Requested-With, X-Requested-By")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	ctx.Response.Header.Set("Access-Control-Max-Age", "86400")
}

func WriteServerEventsHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.DisableNormalizing()
	ctx.Response.Header.SetContentType("text/event-stream")
	ctx.Response.Header.Set("Cache-Control", "no-cache")
	ctx.Response.Header.Set("Connection", "keep-alive")
	ctx.Response.Header.Set("Transfer-Encoding", "chunked")
	ctx.SetStatusCode(fasthttp.StatusOK)
}

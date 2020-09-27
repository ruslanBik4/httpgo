// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fonts сервер отдачи шрифтов (пока реализовано только разделение браузеров на два виде,
// позже планируется учитывать другие параметры пользователя
package fonts

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/logs"
)

var pathWeb string

// GetPath set path with fonts files
func GetPath(path *string) {
	pathWeb = *path
}
func contains(array []string, str string) bool {
	for _, value := range array {
		if strings.Contains(value, str) {
			return true
		}
	}

	return false
}

var fontTypes = map[string]string{
	"ttf":  "font-sfnt",
	"woff": "x-woff",
}

// HandleGetFont push font for some browser
// @/fonts/{font_name}
func HandleGetFont(ctx *fasthttp.RequestCtx) (interface{}, error) {

	ext := ".ttf"
	if browser := string(ctx.Request.Header.Peek("User-Agent")); strings.Contains(browser, "Safari") {
		ext = ".woff"
	} else {
		//http.ServeFile(w, r, pathWeb+r.URL.Path+ext)
		logs.DebugLog("browser=", browser)
	}

	filename := ctx.Path()
	if pos := bytes.Index(filename, []byte(".")); pos > 0 {
		filename = filename[:pos-1]
	}
	data, err := ioutil.ReadFile(path.Join(pathWeb, string(filename)+ext))

	if err != nil {
		if os.IsNotExist(err) && (ext == ".woff") {
			data, err = ioutil.ReadFile(path.Join(pathWeb, string(filename)+".ttf"))
		}
		if err != nil {
			return nil, err
		}
	}

	setHeaderFromFontType(ctx, ext)

	return data, nil
}

// set header on font type
func setHeaderFromFontType(ctx *fasthttp.RequestCtx, ext string) {
	ctx.Response.Header.Set("Content-Type", "mime/type: font/"+fontTypes[ext[:1]])
}

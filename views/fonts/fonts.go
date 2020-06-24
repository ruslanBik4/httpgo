// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fonts сервер отдачи шрифтов (пока реализовано только разделение браузеров на два виде,
// позже планируется учитывать другие параметры пользователя
package fonts

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/views"
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
func HandleGetFont(ctx *fasthttp.RequestCtx) {

	ext := ".ttf"
	if browser := r.Header["User-Agent"]; contains(browser, "Safari") {
		ext = ".woff"
	} else {
		//http.ServeFile(w, r, pathWeb+r.URL.Path+ext)
		logs.DebugLog("browser=", browser)
	}

	filename := r.URL.Path
	if pos := strings.Index(r.URL.Path, "."); pos > 0 {
		filename = filename[:pos-1]
	}
	data, err := ioutil.ReadFile(pathWeb + filename + ext)

	if err != nil {
		if os.IsNotExist(err) && (ext == ".woff") {
			data, err = ioutil.ReadFile(pathWeb + filename + ".ttf")
		}
		if err != nil {
			views.RenderInternalError(w, err)
			return
		}
	}

	setHeaderFromFontType(w, ext)
	w.Write(data)

}

// set header on font type
func setHeaderFromFontType(ctx *fasthttp.RequestCtx, ext string) {
	w.Header().Set("Content-Type", "mime/type: font/"+fontTypes[ext[:1]])
}

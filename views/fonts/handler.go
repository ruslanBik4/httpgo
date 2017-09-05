// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// сервер отдачи шрифтов (пока реализовано только разделение браузеров на два виде,
// позже планируется учитывать другие параметры пользователя
package fonts

import (
	"github.com/ruslanBik4/httpgo/models/logs"
	"io/ioutil"
	"net/http"
	"strings"
	"io"
	"os"
	"github.com/ruslanBik4/httpgo/views"
)

var PathWeb string

func GetPath(path *string) {
	PathWeb = *path
}
func contains(array []string, str string) bool {
	for _, value := range array {
		if strings.Contains(value, str) {
			return true
		}
	}

	return false
}
var fontTypes = map[string] string {
	"ttf" : "font-sfnt",
	"woff" : "x-woff",
}
// push font for some browser
// @/fonts/{font_name}
func HandleGetFont(w http.ResponseWriter, r *http.Request) {

	//w.Header().Set("Content-Type", "mime/type; ttf")

	//PathWeb = "/home/travel/thetravel/web"
	ext := ".ttf"
	if browser := r.Header["User-Agent"]; contains(browser, "Safari") {
		ext = ".woff"
	} else {
		//http.ServeFile(w, r, PathWeb+r.URL.Path+ext)
		logs.DebugLog("browser=", browser)
	}

	filename := r.URL.Path
	if pos := strings.Index(r.URL.Path, "."); pos > 0 {
		filename = filename[:pos-1]
	}
	data, err := ioutil.ReadFile(PathWeb + filename + ext)

	if err != nil {
		if os.IsNotExist(err) && (ext == ".woff") {
			data, err = ioutil.ReadFile(PathWeb + filename + ".ttf")
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
func setHeaderFromFontType(w http.ResponseWriter, ext string) {
	w.Header().Set("Content-Type", "mime/type: font/" + fontTypes[ext[:1]])
}

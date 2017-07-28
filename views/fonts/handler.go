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
func HandleGetFont(w http.ResponseWriter, r *http.Request) {

	//w.Header().Set("Content-Type", "mime/type; ttf")

	//PathWeb = "/home/travel/thetravel/web"
	ext := ".ttf"
	if browser := r.Header["User-Agent"]; contains(browser, "Safari") {
		ext = ".woff"
		w.Header().Set("Content-Type", "mime/type: font/x-woff")
	} else {
		w.Header().Set("Content-Type", "mime/type: font/font-sfnt")
		//http.ServeFile(w, r, PathWeb+r.URL.Path+ext)
		logs.DebugLog("browser=", browser)
	}

	filename := r.URL.Path
	if pos := strings.Index(r.URL.Path, "."); pos > 0 {
		filename = filename[:pos-1]
	}
	if data, err := ioutil.ReadFile(PathWeb + filename + ext); err != nil {
		logs.ErrorLog(err)
	} else {
		w.Write(data)
	}
}

// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/views"
	"io"
)

// @/api/multiroute/?route[]={routes list}
// prepare JSON with fields type from structere DB and + 1 row with data if issue parameter "id"
func HandleMultiRouteJSON(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)

	routes, ok := r.Form["route"]

	if !ok {
		views.RenderNotParamsInPOST(w, "route")
		return
	}

	r.Form.Del("route")
	url := "http://" + r.Host

	comma := `{"`
	for _, val := range routes {
		resp, err := http.PostForm(url+val, r.Form)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(comma + val + `":[`))
		io.Copy(w, resp.Body)
		comma = `],"`
	}
	w.Write([]byte("]}"))

}

func init() {
	http.HandleFunc("/api/v1/multiroute/", system.WrapCatchHandler(HandleMultiRouteJSON))
}

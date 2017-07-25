// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	_ "github.com/ruslanBik4/httpgo/views/templates/json"
	"github.com/ruslanBik4/httpgo/models/services"
	"github.com/ruslanBik4/httpgo/views"
)

// @/api/table/schema/?table={nameTable}
// показ структуры таблицы nameTable
func HandleListAllList(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)
	if list := r.FormValue("list"); list > "" {
		result, err := services.Get("DBlists", "one", list)
		if err != nil {
			views.RenderInternalError(w, err)
		} else {
			views.RenderArrayJSON(w, result.([]map[string] interface{}))
		}
		return

	}
	result, err := services.Get("DBlists", "all-list")
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		//for _, name := range result.([]string) {
			views.RenderStringSliceJSON(w, result.([]string) )
		//}
	}
}

func init()  {
	http.HandleFunc("/api/v1/list/", HandleListAllList )

}

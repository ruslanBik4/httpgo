// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/services"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/json"
	viewsSystem "github.com/ruslanBik4/httpgo/views/templates/system"
	"net/http"
	_ "strings"
)

const _2K = (1 << 10) * 2

// /api/table/form/?table=
// prepare JSON with fields type from structere DB and + 1 row with data if issue parameter "id"
func HandleFieldsJSON(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)
	tableName := r.FormValue("table")

	if tableName == "" {
		views.RenderBadRequest(w)
		return
	}

	defer func() {
		err1 := recover()
		switch err := err1.(type) {
		case schema.ErrNotFoundTable:
			views.RenderInternalError(w, err)
		case nil:
		case error:
			panic(err)
		}
	}()

	qBuilder := qb.Create("id=?", "", "")
	qBuilder.AddTable("", tableName)
	// инши параметры могут быть использованы для суррогатных (вложенных) полей
	qBuilder.PostParams = r.Form

	addJSON := make(map[string]interface{}, 0)
	if id := r.FormValue("id"); id > "" {
		// получаем данные для суррогатных полей
		qBuilder.AddArg(id)
		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			views.RenderInternalError(w, err)
			return
		}

		addJSON["data"] = arrJSON[0]
	}

	views.RenderJSONAnyForm(w, qBuilder.GetFields(), new(json.FormStructure), addJSON)
}

func HandleSchema(w http.ResponseWriter, r *http.Request) {
	tableName := r.FormValue("table")
	if table, err := services.Get("schema", tableName); err != nil {
		views.RenderInternalError(w, err)
	} else {
		w.Write([]byte(viewsSystem.ShowSchema(table.(*schema.FieldsTable))))
	}

}
func HandleRowJSON(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)
	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if (tableName > "") && (id > "") {
		qBuilder := qb.Create("id=?", "", "")
		qBuilder.PostParams = r.Form
		qBuilder.AddTable("a", tableName)
		qBuilder.AddArg(id)
		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			views.RenderInternalError(w, err)
		} else if len(arrJSON) > 0 {
			views.RenderAnyJSON(w, arrJSON[0])
		}
	} else {
		views.RenderBadRequest(w)
	}
}
func HandleAllRowsJSON(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)
	tableName := r.FormValue("table")

	if tableName > "" {
		qBuilder := qb.CreateFromSQL("SELECT * FROM " + tableName)
		qBuilder.PostParams = r.Form
		//qBuilder.AddTable("a", tableName)
		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			views.RenderInternalError(w, err)
		} else if len(arrJSON) == 1 {
			views.RenderAnyJSON(w, arrJSON[0])
		} else if len(arrJSON) > 1 {
			views.RenderArrayJSON(w, arrJSON)
		}
	} else {
		views.RenderBadRequest(w)
	}
}

// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/json"
	_ "strings"
	"github.com/ruslanBik4/httpgo/models/services"
	viewsSystem "github.com/ruslanBik4/httpgo/views/templates/system"
	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/db"
	"log"
)
// prepare JSON with fields type from structere DB and + 1 row with data if issue parameter "id"
func HandleFieldsJSON(w http.ResponseWriter, r *http.Request) {

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
		default:
			panic(err)
		}
	}()

	fields := schema.GetFieldsTable(tableName)
	for idx, field := range fields.Rows {

		if field.SETID || field.NODEID {

			log.Println(field.SQLforFORMList)
			rows, err := db.DoSelect(field.SQLforFORMList)
			if err != nil {
				log.Println(err, field.SQLforFORMList)
			} else {

				defer rows.Close()
				for rows.Next() {
					var key int
					var title string
					if err := rows.Scan(&key, &title); err != nil {
						log.Println(err)
					}

					fields.Rows[idx].SelectValues[key] = title
				}
			}

		}
	}

	addJSON := make(map[string]string, 0)
	if id := r.FormValue("id"); id > "" {
		// получаем данные для суррогатных полей
		qBuilder := qb.Create("id=?", "", "")
		qBuilder.AddTable("a", tableName)
		qBuilder.AddArgs(id)
		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			views.RenderInternalError(w, err)
			return
		}

		// значение приходит в виде строки. Для агрегатных полей нужно формировать вложеность
		//for sKey, sValue := range arrJSON {
		//
		//	for key, value := range sValue {
		//
		//		if strings.HasPrefix(key, "setid_") || strings.HasPrefix(key, "nodeid_") || strings.HasPrefix(key, "tableid_") {
		//
		//			switch vv := value.(type) {
		//			case []map[string]interface{} :
		//				arrJSON[sKey][key] = convertToMultiDimension(vv)
		//			}
		//		}
		//	}
		//}

		addJSON["data"] = json.WriteSliceJSON(arrJSON)
	}

	views.RenderJSONAnyForm(w, fields, new (json.FormStructure), addJSON)
}

func convertToMultiDimension(array [] map[string]interface{}) json.MapMultiDimension {

	var mapToDem = json.MapMultiDimension{}
	for _,val := range array {
		mapToDem = append(mapToDem, val)
	}
	return mapToDem
}
func HandleSchema(w http.ResponseWriter, r *http.Request) {
	tableName := r.FormValue("table")
	if table, err := services.Get("schema", tableName); err != nil {
		views.RenderInternalError(w, err)
	} else {
		w.Write([]byte(viewsSystem.ShowSchema(table.(*schema.FieldsTable) )))
	}

}
func HandleRowJSON(w http.ResponseWriter, r *http.Request) {
	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if (tableName > "") && (id > "") {
		qBuilder := qb.Create("id=?", "", "")
		qBuilder.AddTable("a", tableName)
		qBuilder.AddArgs(id)
		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			views.RenderInternalError(w, err)
		}
		if len(arrJSON) > 0 {
			views.RenderAnyJSON(w, arrJSON[0])
			return
		}
	} else {
		views.RenderBadRequest(w)
	}
}
func HandleAllRowsJSON(w http.ResponseWriter, r *http.Request) {
	tableName := r.FormValue("table")

	if (tableName > "") {
		qBuilder := qb.Create("", "", "")
		qBuilder.AddTable("a", tableName)
		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			views.RenderInternalError(w, err)
			return
		}
		if len(arrJSON) > 0 {
			views.RenderAnyJSON(w, arrJSON[0])
			return
		}
	} else {
		views.RenderBadRequest(w)
	}
}

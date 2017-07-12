// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/views"
	"net/http"
	"strings"
	"github.com/ruslanBik4/httpgo/models/logs"
)

func findField(key string, tables map[string]schema.FieldsTable) *schema.FieldStructure {

	for _, table := range tables {
		if field := table.FindField(key); field != nil {
			return field
		}
	}

	return nil
}
func HandlerSearch(w http.ResponseWriter, r *http.Request) {

	var where string
	var args []interface{}

	r.ParseMultipartForm(_2K)

	tables := make(map[string]schema.FieldsTable, 0)
	tableName := r.FormValue("table")

	if tableName == "" {
		views.RenderBadRequest(w)
		return
	}

	table := schema.GetFieldsTable(tableName)

	r.Form.Del("table")

	joins := make(map[string]*schema.FieldsTable, 0)
	for _, tableName := range r.Form["joins"] {
		joins[tableName] = schema.GetFieldsTable(tableName)
	}
	r.Form.Del("joins")

	comma := ""

	for key, value := range r.Form {

		if (findField(key, tables) == nil) && (table.FindField(key) == nil) {
			logs.StatusLog(key, value)
			continue
		}
		if len(value) > 1 {
			where += comma + key + " in ("
			commaIn := ""
			for _, val := range value {
				args = append(args, val)
				where += commaIn + "?"
				commaIn = ","
			}
			where += ")"
		} else if strings.HasPrefix(key, "id") {
			where += comma + key + "=" + value[0]
		} else {
			where += comma + key + "=?"
			args = append(args, value[0])
		}
		comma = " AND "
	}
	qBuilder := qb.Create(where, "", "")
	qBuilder.AddTable("m", tableName)

	leftTable := tableName
	for name, _ := range tables {
		qBuilder.LeftJoin("", name, "ON " + leftTable + ".id" + "=" + name + ".id_" + leftTable)
		leftTable = name
	}

	qBuilder.AddArgs(args...)
	qBuilder.PostParams = r.Form

	arrJSON, err := qBuilder.SelectToMultidimension()
	if err != nil {
		views.RenderInternalError(w, err)
		return
	}

	views.RenderArrayJSON(w, arrJSON)
}

func init() {
	http.HandleFunc("/api/v1/search/", system.WrapCatchHandler(HandlerSearch))
}

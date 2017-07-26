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
	"github.com/ruslanBik4/httpgo/models/logs"
	"strconv"
	"net/url"
	"strings"
)

const _2K = (1 << 10) * 2

// @/api/table/form/?table={nameTable}
// prepare JSON with fields type from structere DB and + 1 row with data if issue parameter "id"
func HandleFieldsJSON(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)
	tableName := r.FormValue("table")

	if tableName == "" {
		views.RenderNotParamsInPOST(w, "table")
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
	// инши параметры могут быть использованы для суррогатных (вложенных) полей
	qBuilder.PostParams = r.Form
	addFieldsFromPost( qBuilder.AddTable("", tableName), r.Form )

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
// @/api/table/schema/?table={nameTable}
// показ структуры таблицы nameTable
func HandleSchema(w http.ResponseWriter, r *http.Request) {
	tableName := r.FormValue("table")
	if table, err := services.Get("schema", tableName); err != nil {
		views.RenderInternalError(w, err)
	} else {
		views.WriteHeaders(w)
		viewsSystem.WriteShowSchema( w, table.(*schema.FieldsTable) )
	}

}
var wOut http.ResponseWriter
var comma string
// read rows and store in JSON
func PutRowToJSON(fields []*qb.QBField) error {
	// обрамление объекта в JSON
	wOut.Write([]byte(comma + "{"))
	defer func() {
		wOut.Write([]byte("}"))
		comma = ","
	}()

	for idx, field := range fields {
		if idx > 0 {
			wOut.Write( []byte (",") )
		}

		wOut.Write([]byte(`"` + field.Alias + `":`))
		if field.ChildQB != nil {
			wOut.Write([]byte("["))
			if fieldID, ok := field.Table.Fields["id"]; ok {
				// не переводим в int только потому, что в данном случае неважно, отдаем строкой
				field.ChildQB.Args[0] = string(fieldID.Value)
				comma = ""
				err := field.ChildQB.SelectRunFunc(PutRowToJSON)
				if err != nil {
					logs.ErrorLog(err, field.ChildQB)
					return err
				}
			} else if val := string(field.Value); val > "" {
				// проставляем, что в значении есть фильтра
				if i := strings.Index(val, ":"); i > 0 {
					param, suffix := val[i+1:], ""
					// считаем, что окончанием параметра могут быть символы ", )"
					j := strings.IndexAny(param, ", )")
					if j > 0 {
						suffix = param[j:]
						param = param[:j]
					}
					if fieldID, ok := field.Table.Fields[param]; ok {
						field.ChildQB.SetArgs(string(fieldID.Value))
					}

					val = val[:i] + "?" + suffix
					field.ChildQB.SetWhere(val)
					comma = ""
					err := field.ChildQB.SelectRunFunc(PutRowToJSON)
					if err != nil {
						logs.ErrorLog(err, field.ChildQB)
						return err
					}
				}
			} else {
				wOut.Write([]byte(`{"error":"not ID in parent Table"}`))
				logs.DebugLog(field.ChildQB, field.Table)
			}

			wOut.Write([]byte("]"))
		} else if field.Value == nil {
			wOut.Write([]byte("null"))
		} else if field.Schema.SETID || field.Schema.NODEID || field.Schema.IdForeign {
			if field.SelectValues == nil {
				field.GetSelectedValues()
			}
			value, err := strconv.Atoi( string(field.Value) )
			if err != nil {
				logs.ErrorLog(err, "convert RawBytes ", field.Value)
				return err
			}
			json.WriteElement(wOut, field.SelectValues[value] )
		} else {
			json.WriteElement(wOut, field.GetNativeValue(true) )
		}
	}

	return nil
}
// @/api/table/view/?table={nameTable}& other field in this table
// return field with text values for show in site

func HandleTextRowJSON(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)
	tableName := r.FormValue("table")
	if tableName == "" {
		views.RenderNotParamsInPOST(w, "table")
		return
	}
	//TODO: add check ID as integer unsigned
	_, ok := r.Form["id"]

	if !ok && (len(r.Form) < 2) {
		views.RenderNotParamsInPOST(w, "id")
		return
	}
	table := schema.GetFieldsTable(tableName)

	//r.Form.Del("table")

	where, args := PrepareQuery(r.Form, table)
	qBuilder := qb.Create(where, "", "")
	qBuilder.PostParams = r.Form
	qBuilder.AddTable("m", tableName)
	qBuilder.AddArgs(args...)

	wOut = w
	comma = "["
	err := qBuilder.SelectRunFunc(PutRowToJSON)
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		w.Write([]byte("]"))
		views.WriteJSONHeaders(w)
	}
}
func PrepareQuery(rForm url.Values, table *schema.FieldsTable) (where string, args []interface{}) {

	comma := ""
	for key, value := range rForm {

		if key == "table" {
			continue
		}
		if (table.FindField(key) == nil) {
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
		} else {
			where += comma + key + "=?"
			args = append(args, value[0])
		}
		comma = " AND "
	}
	return where, args
}
// @/api/table/row/?table={nameTable}&id={id}
// return row from nameTable from key=id
func HandleRowJSON(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)
	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if (tableName > "") && (id > "") {
		qBuilder := qb.Create("id=?", "", "")
		qBuilder.PostParams = r.Form
		addFieldsFromPost( qBuilder.AddTable("a", tableName), r.Form )
		qBuilder.AddArg(id)
		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			views.RenderInternalError(w, err)
		} else if len(arrJSON) > 0 {
			views.RenderAnyJSON(w, arrJSON[0])
		}
	} else {
		views.RenderNotParamsInPOST(w, "table", "id")
	}
}
// @/api/table/rows/?table={nameTable}
// return all rows from nameTable
func HandleAllRowsJSON(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)
	tableName := r.FormValue("table")

	if tableName > "" {
		qBuilder := qb.Create("", "", "")
		qBuilder.PostParams = r.Form
		addFieldsFromPost( qBuilder.AddTable("a", tableName), r.Form )
		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			views.RenderInternalError(w, err)
		} else if len(arrJSON) == 1 {
			views.RenderAnyJSON(w, arrJSON[0])
		} else if len(arrJSON) > 1 {
			views.RenderArrayJSON(w, arrJSON)
		}
	} else {
		views.RenderNotParamsInPOST(w, "table")
	}
}

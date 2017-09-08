// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// описание вспомогательных функций для роутеров API
package api

import (
	"net/url"
	"net/http"
	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/system"
)
var (
	routes = map[string]http.HandlerFunc{
		"/api/v1/table/form/":   HandleFieldsJSON,
		"/api/v1/table/view/":   HandleTextRowJSON,
		"/api/v1/table/row/":    HandleRowJSON,
		"/api/v1/table/rows/":   HandleAllRowsJSON,
		"/api/v1/table/schema/": HandleSchema,
		"/api/v1/update/":       HandleUpdateServer,
		"/api/v1/restart/":      HandleRestartServer,
		"/api/v1/log/":          HandleLogServer,
		"/api/v1/photos/":       HandlePhotos,
		"/api/v1/video/":        HandleVideos,
		"/api/v1/photos/add/":   HandleAddPhoto,
		// short route
		"/api/table/form/":   HandleFieldsJSON,
		"/api/table/view/":   HandleTextRowJSON,
		"/api/table/row/":    HandleRowJSON,
		"/api/table/rows/":   HandleAllRowsJSON,
		"/api/table/schema/": HandleSchema,
		"/api/update/":       HandleUpdateServer,
		"/api/restart/":      HandleRestartServer,
		"/api/log/":          HandleLogServer,
		"/api/photos/":       HandlePhotos,
		"/api/video/":        HandleVideos,
		"/api/photos/add/":   HandleAddPhoto,
	}

)
// check params "fields" in Post request & add those in qBuilder table
func addFieldsFromPost(table *qb.QBTable, rForm url.Values)  {

	if fields, ok := rForm["fields[]"]; ok {
		for _, val := range fields {
			table.AddField("", val)
		}
	}
}

func findField(key string, tables map[string]schema.FieldsTable) *schema.FieldStructure {

	for _, table := range tables {
		if field := table.FindField(key); field != nil {
			return field
		}
	}

	return nil
}

//func init() {
//	for route, fnc := range routes {
//		http.HandleFunc(route, system.WrapCatchHandler(fnc))
//	}
//
//}

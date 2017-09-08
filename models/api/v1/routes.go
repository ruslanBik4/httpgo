// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
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
		"/api/v1/search/"    :   HandlerSearch,
		"/api/v1/multiroute/":   HandleMultiRouteJSON,
		"/api/v1/list/"      :   HandleListAllList,
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
		"/api/log/errors/": HandleShowErrorsServer,
		"/api/log/status/": HandleShowStatusServer,
		"/api/log/debug/" : HandleShowDebugServer,
		"/api/update/source/":  HandleUpdateSource,
		"/api/update/test/"  :  HandleUpdateTest,
		"/api/update/build/" :  HandleUpdateBuild,
	}

)

func init() {
		for route, fnc := range routes {
			http.HandleFunc(route, system.WrapCatchHandler(fnc))
		}
}

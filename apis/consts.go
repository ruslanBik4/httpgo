/*
 * Copyright (c) 2022-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"regexp"

	"github.com/valyala/fasthttp"
)

// content types
const (
	ContentTypeJSON      = "application/json"
	ContentTypeMultiPart = "multipart/form-data"
)

type tMethod uint8

func (t tMethod) String() string {
	return methodNames[t]
}

// const of tMethod type values
const (
	GET tMethod = iota
	POST
	HEAD
	PUT
	PATCH
	DELETE
	CONNECT
	OPTIONS
	TRACE
	UNKNOWN
)

var methodNames = []string{
	"GET",
	"POST",
	"HEAD",
	"PUT",
	"PATCH",
	"DELETE",
	"CONNECT",
	"OPTIONS",
	"TRACE",
	"UNKNOWN",
}

func methodFromName(nameMethod string) tMethod {
	switch nameMethod {
	case fasthttp.MethodGet:
		return GET
	case fasthttp.MethodPost:
		return POST
	case fasthttp.MethodHead:
		return HEAD
	case fasthttp.MethodPut:
		return PUT
	case fasthttp.MethodPatch:
		return PATCH
	case fasthttp.MethodDelete:
		return DELETE
	case fasthttp.MethodConnect:
		return CONNECT
	case fasthttp.MethodOptions:
		return OPTIONS
	case fasthttp.MethodTrace:
		return TRACE
	default:
		return UNKNOWN
	}
}

// const of values ctx parameters names
const (
	JSONParams      = "JSONparams"
	MultiPartParams = "MultiPartParams"
	ChildRoutePath  = "lastSegment"
	ApiVersion      = "ACC_VERSION"
	IsWrapHandler   = "HAS_HANDLER"

	ServerName    = "name of server httpgo"
	ServerVersion = "version of server httpgo"
	Database      = "DB"
	AppStore      = "store"
)

const testRouteSuffix = "_test"
const PARAM_REQUIRED = "is required parameter"

// vars fr regexp replacer for Docs
var (
	regAbbr  = regexp.MustCompile(`<abbr[^<]*>([^<]+)</abbr>`)
	regTitle = regexp.MustCompile(`(?m)^#\s+([^\n]+)$`)
	regTags  = regexp.MustCompile(`\*([^*]+)\**`)
)

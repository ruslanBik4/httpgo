/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package apis

import "go/types"

// content types
const (
	ContentTypeJSON      = "application/json"
	ContentTypeMultiPart = "multipart/form-data"
)

type tMethod uint8

func (t tMethod) String() string {
	return methodNames[t]
}

//  const of tMethod type values
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
	UNKNOW
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
	for method, name := range methodNames {
		if name == nameMethod {
			return tMethod(method)
		}
	}

	return UNKNOW
}

//  const of values ctx parameters names
const (
	JSONParams      = "JSONparams"
	MultiPartParams = "MultiPartParams"
	ChildRoutePath  = "lastSegment"
	ApiVersion      = "ACC_VERSION"
	AuthManager     = "auth"
	Database        = "DB"
)

const testRouteSuffix = "_test"
const PARAM_REQUIRED = "is required parameter"

var (
	schemas = map[bool]string{
		false: "http",
		true:  "https",
	}
)

var (
	ParamsHTML = InParam{
		Name: "html",
		Desc: "need for get result in html instead JSON",
		Req:  false,
		Type: NewTypeInParam(types.Bool),
	}
	ParamsLang = InParam{
		Name:     "lang",
		Desc:     "need to get result on non-english",
		DefValue: "en",
		Req:      true,
		Type:     NewTypeInParam(types.String),
	}
	ParamsEmail = InParam{
		Name: "email",
		Desc: "email for login",
		Req:  true,
		Type: NewTypeInParam(types.String),
	}
	ParamsPassword = InParam{
		Name: "key",
		Desc: "password or other key word (on future)",
		Req:  true,
		Type: NewTypeInParam(types.String),
	}
	ParamsGetFormActions = InParam{
		Name: "is_get_form_actions",
		Desc: "need to get form actions in response",
		Req:  false,
		Type: NewTypeInParam(types.Bool),
	}
	BasicParams = []InParam{
		ParamsHTML,
		ParamsLang,
	}
)

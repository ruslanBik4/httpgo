// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

const (
	ctJSON      = "application/json"
	ctMultiPart = "multipart/form-data"
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
)

const testRouteSuffix = "_test"
const PARAM_REQUIRED = "is required parameter"

var (
	schemas = map[bool]string{
		false: "http",
		true:  "https",
	}
)

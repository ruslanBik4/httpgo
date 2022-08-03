/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package apis

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

type FncAuth interface {
	Auth(ctx *fasthttp.RequestCtx) bool
	AdminAuth(ctx *fasthttp.RequestCtx) bool
	String() string
}

// RouteDTO must help create some types into routing handling
type RouteDTO interface {
	GetValue() interface{}
	NewValue() interface{}
}

type CheckDTO interface {
	CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool
}

type CompoundDTO interface {
	ReadParams(ctx *fasthttp.RequestCtx)
}

type FncVisit func([]byte, *fastjson.Value)

//todo add description
type Visit interface {
	Each([]byte, *fastjson.Value)
	Result() (interface{}, error)
}

/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

// RouteDTO must help create some types into routing handling
type RouteDTO interface {
	GetValue() any
	NewValue() any
}

type CheckDTO interface {
	CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool
}

type CompoundDTO interface {
	ReadParams(ctx *fasthttp.RequestCtx)
}

type FncVisit func([]byte, *fastjson.Value)

// todo add description
type Visit interface {
	Each([]byte, *fastjson.Value)
	Result() (any, error)
}

type Docs interface {
	Expect() string
	Format() string
	RequestType() string
}

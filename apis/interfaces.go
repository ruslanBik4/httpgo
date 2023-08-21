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

// CheckDTO must implement checking parameters of route
type CheckDTO interface {
	CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool
}

// CompoundDTO fills properties of DTO from ctx.UserValue
type CompoundDTO interface {
	ReadParams(ctx *fasthttp.RequestCtx)
}

type FncVisit func([]byte, *fastjson.Value)

// Visit implement functions for parsing values from JSON body into DTO
type Visit interface {
	Each([]byte, *fastjson.Value)
	Result() (any, error)
}

// Docs implement functions for writing documentation
type Docs interface {
	Expect() string
	Format() string
	RequestType() string
}

// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

func (a *Apis) onboarding(ctx *fasthttp.RequestCtx) (interface{}, error) {

	path := "/" + ctx.UserValue("path").(string)
	params := ctx.UserValue("params").(string)
	port := ctx.UserValue("port").(int32)
	method := strings.ToUpper(ctx.UserValue("method").(string))

	var mRoute tMethod
	for key, val := range methodNames {
		if val == method {
			mRoute = tMethod(key)
			break
		}
	}

	desc, ok := ctx.UserValue("desc").(string)
	if !ok {
		desc = "onboarding routes from local services into APIS"
	}

	newRoute := &ApiRoute{
		Desc: desc,
		Fnc: func(ctx *fasthttp.RequestCtx) (i interface{}, err error) {
			ctx.Request.SetRequestURI(fmt.Sprintf("http://localhost:%d%s", port, path))

			err = (&fasthttp.Client{}).Do(&ctx.Request, &ctx.Response)
			return nil, err
		},
		Method:    mRoute,
		Multipart: ctx.UserValue("multipart").(bool),
		NeedAuth:  ctx.UserValue("auth").(bool),
		OnlyAdmin: ctx.UserValue("admin").(bool),
	}

	err := jsoniter.UnmarshalFromString(params, &(newRoute.Params))
	if err != nil {
		return []string{"params JSON wrong:" + err.Error()}, ErrWrongParamsList
	}

	return newRoute, a.addRoute(path, newRoute)
}

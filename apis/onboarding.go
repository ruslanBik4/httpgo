// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"fmt"
	"strings"

	"go/types"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var 	onboardParams =  []InParam{
			{
				Name: "path",
				Req:  true,
				Type: NewTypeInParam(types.String),
			},
			{
				Name: "desc",
				Req:  false,
				Type: NewTypeInParam(types.String),
			},
			{
				Name: "params",
				Req:  true,
				Type: NewTypeInParam(types.String),
			},
			{
				Name: "port",
				Req:  true,
				Type: NewTypeInParam(types.Int32),
			},
			{
				Name:     "method",
				Req:      true,
				Type:     NewTypeInParam(types.String),
				DefValue: "POST",
			},
			{
				Name:     "multipart",
				Req:      true,
				Type:     NewTypeInParam(types.Bool),
				DefValue: false,
			},
			{
				Name:     "auth",
				Req:      true,
				Type:     NewTypeInParam(types.Bool),
				DefValue: false,
			},
			{
				Name:     "admin",
				Req:      true,
				Type:     NewTypeInParam(types.Bool),
				DefValue: false,
			},
		}
	

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

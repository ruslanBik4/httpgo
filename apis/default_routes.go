/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
)

func (a *Apis) DefaultRoutes() ApiRoutes {
	return ApiRoutes{
		"/apis": {
			Desc:     "full routers list current *APIS*",
			Fnc:      a.renderApis,
			WithCors: true,
			Params: []InParam{
				{
					Name: "json",
				},
			},
		},
		"/swagger.io": {
			Desc: "Scale Your *APIS* Design with Confidence",
			Fnc: func(ctx *fasthttp.RequestCtx) (interface{}, error) {
				views.RenderHTMLPage(ctx, pages.WriteSwaggerPage)
				return nil, nil
			},
			WithCors: true,
		},
		"/onboarding": {
			Desc:      "onboarding routes from local services into *APIS*",
			Fnc:       a.onboarding,
			OnlyLocal: true,
			Params:    onboardParams,
		},
	}

}
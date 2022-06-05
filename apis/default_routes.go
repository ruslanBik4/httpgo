package apis

import (
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/valyala/fasthttp"
)

func (apis Apis) DefaultRoutes() ApiRoutes {
	return ApiRoutes{
		"/apis": {
			Desc:     "full routers list current *APIS*",
			Fnc:      apis.renderApis,
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
			Fnc:       apis.onboarding,
			OnlyLocal: true,
			Params:    onboardParams,
		},
	}

}

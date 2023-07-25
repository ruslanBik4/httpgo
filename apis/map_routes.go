/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"fmt"
	"go/types"
	"path"
	"strings"

	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/typesExt"
	"github.com/ruslanBik4/httpgo/auth"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/logs"
)

type mapRoute struct {
	method tMethod
	path   string
}

type MapRoutes map[mapRoute]*ApiRoute

// NewMapRoutes create APIRotes instance
func NewMapRoutes() MapRoutes {
	m := make(MapRoutes, 0)
	//for method := range methodNames {
	//	m[tMethod(method)] = make(ApiRoutes, 0)
	//}

	return m
}

// AddRoutes add ApiRoute into hair onsafe
func (r MapRoutes) AddRoutes(routes ApiRoutes) (badRouting []string) {
	for url, route := range routes {
		if !strings.HasPrefix(url, "/") {
			url = "/" + url
		}
		m := mapRoute{route.Method, url}
		_, ok := r[m]
		if ok {
			logs.ErrorLog(ErrPathAlreadyExists, url)
			badRouting = append(badRouting, fmt.Sprintf("%s:%s", route.Method, url))
			continue
		}

		r[m] = route
		// may allow OPTIONS request for "preflighted" requests
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
		if route.WithCors && route.Method != OPTIONS && (route.DTO != nil || route.NeedAuth) {
			mOpt := mapRoute{OPTIONS, url}
			_, ok := r[mOpt]
			if !ok {
				// 	add empty OPTIONS route for
				r[mOpt] = &ApiRoute{
					Desc: fmt.Sprintf(`<abbr title="Cross-Origin Resource Sharing">CORS</abbr> [allow for preflighted](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) route '%s %s':  
%s`, route.Method, url, route.Desc),
					DTO:      nil,
					Fnc:      HandlerForPreflightedCORS,
					Method:   OPTIONS,
					WithCors: true,
				}
			}
		}

		if false {
			r.addTestRoute(url, route)
		}

	}

	return
}

func (r MapRoutes) addTestRoute(url string, route *ApiRoute) {
	testRoute := &ApiRoute{
		Desc:      "test handler for " + url,
		DTO:       route.DTO,
		Fnc:       nil,
		FncAuth:   route.TestFncAuth,
		Method:    GET,
		Multipart: false,
		NeedAuth:  false,
		OnlyAdmin: false,
		OnlyLocal: false,
		Params:    make([]InParam, len(route.Params)),
		Resp:      route.Resp,
		WithCors:  true,
	}
	for i, param := range route.Params {
		testRoute.Params[i] = param
		if param.Req {
			testRoute.Params[i].Req = false
			// testRoute.Params[i].DefValue = param.TestValue
		}
	}

	testRoute.Fnc = func(url string, route *ApiRoute) ApiRouteHandler {
		return func(ctx *fasthttp.RequestCtx) (any, error) {
			// if route.Multipart {
			// 	var b bytes.Buffer
			// 	w := multipart.NewWriter(&b)
			// 	for _, param := range route.Params {
			// 		err := w.WriteField(param.Name, param.TestValue)
			// 		if err != nil {
			// 			return nil, err
			// 		}
			// 	}
			// 	if err := w.Close(); err != nil {
			// 		return nil, errors.Wrap(err, "w.Close")
			// 	}
			//
			// 	ctx.Request.Header.SetMethod(fasthttp.MethodPost)
			// 	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())
			// 	ctx.Request.SetBody(b.Bytes())
			// 	return route.Fnc(ctx)
			// }

			views.WriteHeaders(ctx)
			s, err := jsoniter.MarshalToString(testRoute.Resp)
			if err != nil {
				logs.ErrorLog(err, "MarshalToString")
			}
			page := pages.URLTestPage{
				Host:       string(ctx.Host()),
				Method:     methodNames[route.Method],
				Multipart:  route.Multipart,
				Path:       url,
				Language:   "",
				Charset:    "",
				LinkStyles: nil,
				MetaTags:   nil,
				Params:     make([]pages.ParamUrlTestPage, len(testRoute.Params)),
				Resp:       s,
			}
			for i, val := range testRoute.Params {
				arrReq := map[bool]string{
					true:  "(require)",
					false: "",
				}

				page.Params[i] = pages.ParamUrlTestPage{
					Basic:   types.Typ[types.Invalid],
					Name:    val.Name + arrReq[val.Req],
					Req:     val.Req,
					Type:    val.Type.String(),
					Comment: val.Desc,
				}
				if val.TestValue > "" {
					page.Params[i].Value = val.TestValue
				}

				t, ok := val.Type.(TypeInParam)
				if ok {
					page.Params[i].Basic = typesExt.Basic(t.BasicKind)
					page.Params[i].IsSlice = t.isSlice
				}
			}
			page.WriteShowURlTestPage(ctx.Response.BodyWriter())
			return nil, nil
		}
	}(url, route)

	r[mapRoute{GET, path.Join(url, r.GetTestRouteSuffix(route))}] = testRoute
}

func (r MapRoutes) GetTestRouteSuffix(route *ApiRoute) string {

	if route.Method != GET {
		return methodNames[route.Method] + testRouteSuffix
	}

	return testRouteSuffix
}
func (r MapRoutes) GetRoute(ctx *fasthttp.RequestCtx) (*ApiRoute, error) {
	m := mapRoute{methodFromName(string(ctx.Method())), string(ctx.Path())}
	// check exactly
	if route, ok := r[m]; ok {
		return route, nil
	}

	// find parent pathURL with some method
	if route, parent := r.findParentRoute(m); route != nil {
		ctx.SetUserValue(ChildRoutePath, strings.TrimPrefix(m.path, parent))
		return route, nil
	}

	for i := GET; i < UNKNOWN-1; i++ {
		m.method = i
		route, ok := r[m]
		if ok {
			return route, errMethodNotAllowed
		}
	}

	return nil, errNotFoundPage
}

func (r MapRoutes) findParentRoute(m mapRoute) (*ApiRoute, string) {
	for p := getParentPath(m.path); p > ""; p = getParentPath(p) {
		m.path = p
		route, ok := r[m]
		// check method
		if ok {
			return route, p
		}
	}

	return nil, ""
}

func getParentPath(path string) string {
	if strings.HasSuffix(path, "/") {
		return strings.TrimSuffix(path, "/")
	}

	n := strings.LastIndex(path, "/")
	if n < 0 {
		return ""
	}

	return path[:n+1]
}

func HandlerForPreflightedCORS(ctx *fasthttp.RequestCtx) (any, error) {
	ctx.SetStatusCode(fasthttp.StatusNoContent)
	origin := ctx.Request.Header.Peek("Origin")
	ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", origin)

	return nil, nil
}

func NewMapRoutesWithAjaxWrap(endpoints []ApiRoutes, wrapHandler ApiRouteHandler, chgRoute func(*ApiRoute)) MapRoutes {
	mapRoutes := NewMapRoutes()
	for _, r := range endpoints {
		for _, route := range r {
			if !route.IsAJAXRequest {
				continue
			}

			if route.NeedAuth {
				route.FncAuth = auth.NewAjaxOnly(route.FncAuth)
			}

			chgRoute(route)

			route.Fnc = func(handler ApiRouteHandler) ApiRouteHandler {
				return func(ctx *fasthttp.RequestCtx) (any, error) {
					if views.IsAJAXRequest(&ctx.Request) {
						return handler(ctx)
					}
					ctx.SetUserValue(IsWrapHandler, struct{}{})
					return wrapHandler(ctx)
				}
			}(route.Fnc)
		}

		if badRoutings := mapRoutes.AddRoutes(r); len(badRoutings) > 0 {
			logs.ErrorLog(ErrRouteForbidden, badRoutings)
		}
	}

	return mapRoutes
}

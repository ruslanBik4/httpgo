// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"

	. "github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/system/routeTable"
)

type CtxApis map[string]interface{}

func NewCtxApis(cap int) CtxApis {
	ctx := make(CtxApis, cap)
	return ctx
}
func (c CtxApis) AddValue(key string, val interface{}) {
	c[key] = val
}
func (c CtxApis) Deadline() (deadline time.Time, ok bool) {
	return time.Now(), true
}
func (c CtxApis) Done() <-chan struct{} {
	return nil
}
func (c CtxApis) Err() error {
	return nil
}
func (c CtxApis) Value(key interface{}) interface{} {
	if key, ok := key.(string); ok {
		return c[key]
	}
	return nil
}

type FncAuth interface {
	Auth(ctx *fasthttp.RequestCtx) bool
	AdminAuth(ctx *fasthttp.RequestCtx) bool
	String() string
}

// Apis encapsulates REST API configuration and endpoints
// in calling it checks for the presence, validation of parameters and access privileges
type Apis struct {
	Ctx CtxApis
	// authentication method
	fncAuth FncAuth
	// list of endpoints
	routes MapRoutes
	lock   sync.RWMutex
}

// NewApis create new Apis from list of routes, environment values configuration & authentication method
func NewApis(ctx CtxApis, routes MapRoutes, fncAuth FncAuth) *Apis {
	// Apis include all endpoints application
	apis := &Apis{
		Ctx:     ctx,
		routes:  routes,
		fncAuth: fncAuth,
	}

	apisRoutes := ApiRoutes{
		"/apis": {
			Desc: "full routers list",
			Fnc:  apis.renderApis,
			Params: []InParam{
				{
					Name: "json",
				},
			},
		},
		"/onboarding": {
			Desc:      "onboarding routes from local services into APIS",
			Fnc:       apis.onboarding,
			OnlyLocal: true,
			Params:    onboardParams,
		},
	}
	apis.AddRoutes(apisRoutes)

	return apis
}

// Handler find route on request, check & run
func (a *Apis) Handler(ctx *fasthttp.RequestCtx) {

	route, err := a.routes.GetRoute(ctx)
	if err != nil {
		a.renderError(ctx, err, route)
		return
	}

	// add Cfg params to requestCtx
	for name, val := range a.Ctx {
		ctx.SetUserValue(name, val)
	}

	defer func() {
		errRec := recover()
		switch errRec := errRec.(type) {
		case error:
			params := ctx.UserValue(JSONParams)
			if route.Multipart {
				params = ctx.UserValue(MultiPartParams)
			}

			logs.DebugLog("during performs handler %s, params %+v", route.Desc, params)
			a.renderError(ctx, errRec, nil)
		case string:
			a.renderError(ctx, errors.New(errRec), nil)
		case nil:
		default:
			logs.StatusLog(errRec)
		}
	}()

	resp, err := route.CheckAndRun(ctx, a.fncAuth)
	if err != nil {
		logs.DebugLog("'%s' failure - %v, %s, %s, %s",
			string(ctx.Path()),
			resp,
			ctx.Request.Header.ContentType(),
			ctx.Request.Header.Referer(),
			ctx.Request.Header.UserAgent())
		a.renderError(ctx, err, resp)

		return
	}

	// success execution
	switch resp := resp.(type) {
	case nil:
	case []byte:
		ctx.Response.SetBodyString(string(resp))
	case string:
		ctx.Response.SetBodyString(resp)
	default:
		err = WriteJSON(ctx, resp)
		if err != nil {
			a.renderError(ctx, err, resp)
		}
	}

}

// WriteJSON write JSON to response
func WriteJSON(ctx *fasthttp.RequestCtx, r interface{}) (err error) {

	defer func() {
		errR := recover()
		if errR != nil {
			err = errors.Wrap(errR.(error), "marshal json")
		}
	}()

	enc := jsoniter.NewEncoder(ctx)
	err = enc.Encode(r)
	if err != nil {
		return err
	}

	WriteJSONHeaders(ctx)

	return nil
}

// WriteJSONHeaders return standard headers for JSON
func WriteJSONHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(jsonHEADERSContentType)
}

// render JSON from any data type
const jsonHEADERSContentType = "application/json; charset=utf-8"

// renderError send error message into response
func (a *Apis) renderError(ctx *fasthttp.RequestCtx, err error, resp interface{}) {

	statusCode := http.StatusInternalServerError
	errMsg := err.Error()
	switch errDeep := errors.Cause(err); errDeep {
	case errMethodNotAllowed:
		statusCode = http.StatusMethodNotAllowed
		switch r := resp.(type) {
		case string:
			errMsg = fmt.Sprintf(errMsg, string(ctx.Method()), r)
		case *ApiRoute:
			errMsg = fmt.Sprintf(errMsg, string(ctx.Method()), r.Method)
		default:
			errMsg = fmt.Sprintf(errMsg+"%+v", string(ctx.Method()), "", resp)
		}

	case ErrUnAuthorized:
		logs.StatusLog("attempt unauthorized access %s", ctx.Request.Header.Referer())
		statusCode = http.StatusUnauthorized
	case ErrRouteForbidden:
		statusCode = fasthttp.StatusForbidden
	case errRouteOnlyLocal:
		statusCode = fasthttp.StatusForbidden
	case fasthttp.ErrNoMultipartForm:
		statusCode = fasthttp.StatusBadRequest
		_, err := ctx.WriteString("must be multipart-form")
		if err != nil {
			logs.StatusLog(err)
		}

	case ErrWrongParamsList:
		statusCode = http.StatusBadRequest
		errMsg = fmt.Sprintf(errMsg, resp)

		if bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte(ctMultiPart)) {
			logs.DebugLog(ctx.UserValue(MultiPartParams))
		} else if ctx.IsPost() {
			logs.DebugLog(ctx.PostArgs().String())
		} else {
			logs.DebugLog(ctx.QueryArgs().String())
		}

	case errNotFoundPage:
		ctx.NotFound()
		logs.StatusLog("Not Found Page %+v", ctx.Request.String())
		return

	default:
		logs.ErrorStack(errDeep, resp)
	}

	ctx.Error(errMsg, statusCode)
}

// addRoute with safe on concurrent
func (a *Apis) addRoute(path string, route *ApiRoute) error {
	_, ok := a.routes[route.Method][path]
	if ok {
		return ErrPathAlreadyExists
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	a.routes[route.Method][path] = route

	return nil
}

// AddRoutes add routes to Apis for handling service
// return slice with name routes which not adding
func (a *Apis) AddRoutes(routes ApiRoutes) (badRouting []string) {
	return a.routes.AddRoutes(routes)
}

// renderApis show list routers for developers (as JSON)
func (a *Apis) renderApis(ctx *fasthttp.RequestCtx) (interface{}, error) {
	if ctx.UserValue("json") != nil {
		return *a, nil
	}

	columns := SimpleColumns(
		"Path - Method",
		"Descriptor",
		"Auth",
		"Required parameters",
		"Others parameters",
		"DtoFromJSON",
		"Response",
	)

	rows := make([][]interface{}, 0)

	i := 0
	for method, routes := range a.routes {
		sortList := make([]string, 0, len(routes))
		for url := range routes {
			sortList = append(sortList, url)
		}
		sort.Strings(sortList)

		for _, url := range sortList {
			route := routes[url]
			if url == string(ctx.Path()) || strings.HasSuffix(url, testRouteSuffix) {
				continue
			}

			row := make([]interface{}, len(columns))
			testURL := path.Join(url, a.routes.GetTestRouteSuffix(route))

			row[0] = fmt.Sprintf(`<a href="%s" title="see test">%s</a> - %s`, testURL, url, method)
			if route.Multipart {
				row[0] = row[0].(string) + ", MULTIPART"
			}
			row[1] = route.Desc

			s := "use method '"
			if route.FncAuth != nil {
				s += route.FncAuth.String() + "' for checking authorization"
			} else if route.NeedAuth {
				s += a.fncAuth.String() + "' for checking authorization"
			}

			if route.OnlyAdmin {
				s += " only admin request be allowed"
			}

			if route.OnlyLocal {
				s += " only local request be allowed"
			}

			if s > "" {
				row[2] = s
			} else {
				row[2] = false
			}

			r, p := "", ""
			for _, param := range route.Params {
				s := fmt.Sprintf(`<div>"%s" <i>%s</i>, %s `, param.Name, param.Desc, param.Type)
				if param.DefValue != nil {
					s += fmt.Sprintf("Def: '%v'", param.defaultValueOfParams(nil))
				}

				if len(param.PartReq) > 0 {
					s += "one of {" + strings.Join(param.PartReq, ", ") + " and " + param.Name + "} is required"
				}

				if len(param.IncompatibleWiths) > 0 {
					s += "only one of {" + strings.Join(param.IncompatibleWiths, ", ") + " and " + param.Name + "} may use for request"
				}

				if param.Req {
					r += s + "</div>"
				} else {
					p += s + "</div>"
				}
			}

			row[3] = r
			row[4] = p

			if route.DTO != nil {
				row[5] = route.DTO.NewValue()
			}
			row[6] = route.Resp

			rows = append(rows, row)
			i++
		}
	}

	views.RenderHTMLPage(ctx, layouts.WritePutHeadForm)

	colDecors := make([]*forms.ColumnDecor, len(columns))
	for i, col := range columns {
		colDecors[i] = forms.NewColumnDecor(col, nil)
	}
	routeTable.WriteTableRow(ctx, colDecors, rows)

	return nil, nil
}

func getLastSegment(path string) string {
	n := strings.LastIndex(strings.TrimSuffix(path, "/"), "/")
	if n < 0 {
		return ""
	}

	return path[n+1:]
}

func isNotLocalRequest(ctx *fasthttp.RequestCtx) bool {
	host := string(ctx.Request.Header.Host())

	return !strings.Contains(host, "127.0.0.1") && !strings.Contains(host, "localhost")
}

// apiRouteToJSON produces a human-friendly description of Apis.
// Based on real data of the executable application, does not require additional documentation.
func apisToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	apis := *(*Apis)(ptr)
	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	FirstFieldToJSON(stream, "Descriptor", "API Specification, include endpoints description, ect")
	AddObjectToJSON(stream, "ctx", apis.Ctx)
	if apis.fncAuth != nil {
		AddObjectToJSON(stream, "auth", apis.fncAuth.String())
	}

	AddObjectToJSON(stream, "routes", apis.routes)
}

func init() {
	jsoniter.RegisterTypeEncoderFunc("apis.Apis", apisToJSON, func(pointer unsafe.Pointer) bool {
		return false
	})
}

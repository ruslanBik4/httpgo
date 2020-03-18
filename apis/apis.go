// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"bytes"
	"fmt"
	"go/types"
	"net/http"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/views"
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
	routes ApiRoutes
	lock   sync.RWMutex
}

// NewApis create new Apis from list of routes, environment values configuration & authentication method
func NewApis(ctx CtxApis, routes ApiRoutes, fncAuth FncAuth) *Apis {
	// Apis include all endpoints application
	apis := &Apis{
		Ctx:     ctx,
		routes:  routes,
		fncAuth: fncAuth,
	}

	// add system routers, ignore errors
	apisRoute := &ApiRoute{
		Desc: "full routers list",
		Fnc:  apis.renderApis,
		Params: []InParam{
			{
				Name: "json",
			},
		},
	}

	_ = apis.addRoute("/apis", apisRoute)

	onboardingRoute := &ApiRoute{
		Desc:      "onboarding routes from local services into APIS",
		Fnc:       apis.onboarding,
		OnlyLocal: true,
		Params: []InParam{
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
		},
	}
	_ = apis.addRoute("/onboarding", onboardingRoute)

	return apis
}

// Handler find route on request, check & run
func (a *Apis) Handler(ctx *fasthttp.RequestCtx) {

	route, ok := a.isValidPath(ctx)
	if !ok {
		ctx.NotFound()
		logs.StatusLog("Not Found Page %+v", ctx.Request.String())
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

// WriteJSONHeaders return standart headers for JSON
func WriteJSONHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(jsonHEADERSContentType)
}

// render JSON from any data type
const jsonHEADERSContentType = "application/json; charset=utf-8"

// renderError send error message into responce
func (a *Apis) renderError(ctx *fasthttp.RequestCtx, err error, resp interface{}) {

	statusCode := http.StatusInternalServerError
	switch errDeep := errors.Cause(err); errDeep {
	case errMethodNotAllowed:
		statusCode = http.StatusMethodNotAllowed
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

		errMsg := fmt.Sprintf(err.Error(), resp)
		ctx.Error(errMsg, statusCode)

		if bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte(ctMultiPart)) {
			logs.DebugLog(ctx.UserValue(MultiPartParams))
		} else if ctx.IsPost() {
			logs.DebugLog(ctx.PostArgs().String())
		} else {
			logs.DebugLog(ctx.QueryArgs().String())
		}

		return

	default:
		logs.ErrorStack(errDeep, resp)
	}

	ctx.Error(err.Error(), statusCode)
}

// addRoute with safe on concurrent
func (a *Apis) addRoute(path string, route *ApiRoute) error {
	_, ok := a.routes[path]
	if ok {
		return ErrPathAlreadyExists
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	a.routes[path] = route

	return nil
}

// AddRoutes add routes to Apis for handling service
// return slice with name routes which not adding
func (a *Apis) AddRoutes(routes ApiRoutes) (badRouting []string) {
	for path, route := range routes {
		err := a.addRoute(path, route)
		if err != nil {
			logs.ErrorLog(err, path)
			badRouting = append(badRouting, path)
		}
	}

	return
}

// renderApis show list routers for developers (as JSON)
func (a *Apis) renderApis(ctx *fasthttp.RequestCtx) (interface{}, error) {
	if ctx.UserValue("json") != nil {
		return *a, nil
	}

	columns := []string{
		"Path - Method",
		"Descriptor",
		"Auth",
		"Required parameters",
		"Others parameters",
		"DtoFromJSON",
		"Response",
	}

	rows := make([][]interface{}, len(a.routes))

	i := 0
	for url, route := range a.routes {
		row := make([]interface{}, len(columns))
		row[0] = url + " - " + route.Method.String()
		if route.Multipart {
			row[0] = row[0].(string) + ", MULTIPART"
		}
		row[1] = route.Desc

		s := ""
		if route.FncAuth != nil {
			s += "use custom method '" + route.FncAuth.String() + "' for checking authorization"
		} else if route.NeedAuth {
			s += " use standard method for checking authorization"
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
			s := fmt.Sprintf(`<div>"%s": <i>%s</i>, %s `, param.Name, param.Desc, param.Type)
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

		rows[i] = row
		i++
	}
	views.RenderHTMLPage(ctx, layouts.WritePutHeadForm)

	routeTable.WriteTableRow(ctx, columns, rows)

	return nil, nil
}

func (a *Apis) isValidPath(ctx *fasthttp.RequestCtx) (*ApiRoute, bool) {
	path := string(ctx.Path())
	route, ok := a.routes[path]
	// check method
	if ok && route.isValidMethod(ctx) {
		return route, ok
	}

	return a.findParentRoute(ctx, path)
}

func (a *Apis) findParentRoute(ctx *fasthttp.RequestCtx, path string) (route *ApiRoute, ok bool) {
	for p := getParentPath(path); p > ""; p = getParentPath(p) {
		route, ok = a.routes[p]
		// check method
		if ok && route.isValidMethod(ctx) {
			ctx.SetUserValue(ChildRoutePath, strings.TrimPrefix(path, p))
			return route, ok
		}
	}

	return
}

func getParentPath(path string) string {
	n := strings.LastIndex(strings.TrimSuffix(path, "/"), "/")
	if n < 0 {
		return ""
	}

	return path[:n+1]
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
//Based on real data of the executable application, does not require additional documentation.
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

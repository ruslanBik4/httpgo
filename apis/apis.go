// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/valyala/fasthttp"
)

type ApisRender interface {
	WriteJSON(ctx *fasthttp.RequestCtx, r interface{}) bool
}

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

// Apis encapsulates REST API configuration and endpoints
// in calling it checks for the presence, validation of parameters and access privileges
type Apis struct {
	Ctx CtxApis
	// authentication method
	fncAuth func(ctx *fasthttp.RequestCtx) bool
	// list of endpoints
	routes APIRoutes
	lock   sync.RWMutex
}

// NewApis create new Apis from list of routes, environment values configuration & authentication method
func NewApis(ctx CtxApis, routes APIRoutes, fncAuth func(ctx *fasthttp.RequestCtx) bool) *Apis {
	// Apis include all endpoints application
	apis := &Apis{
		Ctx:     ctx,
		routes:  routes,
		fncAuth: fncAuth,
	}

	apisRoute := &APIRoute{Desc: "full routers list", Fnc: apis.renderApis}
	err := apis.addRoute("/apis", apisRoute)
	if err != nil {
		logs.ErrorLog(err)
	}

	return apis
}

// Handler find route on request, check & run
func (a *Apis) Handler(ctx *fasthttp.RequestCtx) {

	path := string(ctx.Path())
	route, ok := a.isValidPath(ctx, path)
	if !ok {
		ctx.NotFound()
		logs.ErrorLog(errNotFoundPage, path, ctx.String(), ctx.Request.String())
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
			errRec = errors.Wrapf(errRec, "route.CheckAndRun")
			a.renderError(ctx, errRec, nil)
		case string:
			a.renderError(ctx, errors.New(errRec), nil)
		case nil:
		default:
			logs.StatusLog(errRec)
		}
	}()

	resp, err := route.CheckAndRun(ctx, a.fncAuth)

	// success execution
	if err != nil {
		logs.DebugLog("route not run successfully - '%s', %v, %s", path, resp, ctx.Request.Header.String())
		a.renderError(ctx, err, resp)
	} else {
		if resp != nil {
			err = a.WriteJSON(ctx, resp)
			if err != nil {
				a.renderError(ctx, err, resp)
			}
		}
	}

}

// WriteJSON write JSON to response
func (a *Apis) WriteJSON(ctx *fasthttp.RequestCtx, r interface{}) (err error) {

	defer func() {
		errR := recover()
		if errR != nil {
			err = errR.(error)
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
	switch errors.Cause(err) {
	case errMethodNotAllowed:
		statusCode = http.StatusMethodNotAllowed
	case ErrUnAuthorized:
		statusCode = http.StatusUnauthorized
	case ErrRouteForbidden:
		statusCode = http.StatusForbidden
	case errRouteOnlyLocal:
		statusCode = http.StatusForbidden
	case fasthttp.ErrNoMultipartForm:
		statusCode = http.StatusBadRequest
		_, err := ctx.WriteString("must be multipart-form")
		if err != nil {
			logs.StatusLog(err)
		}

	case ErrWrongParamsList:
		statusCode = http.StatusBadRequest

		errMsg := fmt.Sprintf(err.Error(), resp)
		if _, err := ctx.WriteString(errMsg); err != nil {
			logs.ErrorLog(err)
		}

		if bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte(ctMultipArt)) {
			logs.DebugLog(ctx.UserValue(MultiPartParams))
		}
		if ctx.IsPost() {
			logs.DebugLog(ctx.PostArgs().String())
		} else {
			logs.DebugLog(ctx.QueryArgs().String())
		}
		ctx.Error(errMsg, statusCode)
		return
	default:
		logs.ErrorStack(err, resp)
	}
	ctx.Error(err.Error(), statusCode)
}

// addRoute with safe on concurents
func (a *Apis) addRoute(path string, route *APIRoute) error {
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
func (a *Apis) AddRoutes(routes APIRoutes) (badRouting []string) {
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
	return a.routes, nil
}

func (a *Apis) isValidPath(ctx *fasthttp.RequestCtx, path string) (*APIRoute, bool) {
	route, ok := a.routes[path]
	// check method
	if ok && route.isValidMethod(ctx) {
		return route, ok
	}

	return a.findRootRoute(ctx, path)
}

func (a *Apis) findRootRoute(ctx *fasthttp.RequestCtx, path string) (route *APIRoute, ok bool) {
	for n := strings.LastIndex(path, "/"); n > -1; n = strings.LastIndex(strings.TrimSuffix(path, "/"), "/") {
		path = path[:n+1]
		route, ok = a.routes[path]
		// check method
		if ok && route.isValidMethod(ctx) {
			return route, ok
		}
	}

	return
}

func isNotLocalRequest(ctx *fasthttp.RequestCtx) bool {
	host := string(ctx.Request.Header.Host())

	return !strings.Contains(host, "127.0.0.1") && !strings.Contains(host, "localhost")
}

/*
 * Copyright (c) 2022-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Package apis consists of interfaces for managements endpoints httpgo
package apis

import (
	"bytes"
	"database/sql"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/httpgo/auth"
	"github.com/ruslanBik4/logs"

	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/system/routeTable"
)

type CtxApis map[string]any

func NewCtxApis(cap int) CtxApis {
	ctx := make(CtxApis, cap)
	return ctx
}
func (c CtxApis) AddValue(key string, val any) {
	c[key] = val
}
func (c CtxApis) Deadline() (deadline time.Time, ok bool) {
	return
}
func (c CtxApis) Done() <-chan struct{} {
	return nil
}
func (c CtxApis) Err() error {
	return nil
}
func (c CtxApis) Value(key any) any {
	if key, ok := key.(string); ok {
		return c[key]
	}
	return nil
}

// Apis encapsulates REST API configuration and endpoints
// in calling it checks for the presence, validation of parameters and access privileges
type Apis struct {
	*sync.RWMutex
	Ctx CtxApis
	// authentication method
	fncAuth auth.FncAuth
	// list of endpoints
	routes    MapRoutes
	Https     bool
	StartTime time.Time
}

// NewApis create new Apis from list of routes, environment values configuration & authentication method
func NewApis(ctx CtxApis, routes MapRoutes, fncAuth auth.FncAuth) *Apis {
	// Apis include all endpoints application
	apis := &Apis{
		Ctx:     ctx,
		RWMutex: &sync.RWMutex{},
		routes:  routes,
		fncAuth: fncAuth,
	}

	apisRoutes := apis.DefaultRoutes()
	apis.AddRoutes(apisRoutes)

	return apis
}

// Handler find route on request, check & run
func (a *Apis) Handler(ctx *fasthttp.RequestCtx) {

	//reset user values for HTTP/2
	ctx.ResetUserValues()
	route, err := a.routes.GetRoute(ctx)
	if err != nil {
		a.renderError(ctx, err, route)
		return
	}

	ctx.SetUserValue(views.AgeOfServer, time.Since(a.StartTime).Seconds())
	// add Cfg params to requestCtx
	for name, val := range a.Ctx {
		ctx.SetUserValue(name, val)
	}

	auth.SetAuthManager(ctx, a.fncAuth)
	defer func() {
		errRec := recover()
		switch errRec := errRec.(type) {
		case nil:
			return

		case error:
			params := ctx.UserValue(JSONParams)
			if params == nil && route.Multipart {
				params = ctx.UserValue(MultiPartParams)
			}

			logs.DebugLog("during performs handler '%s', params %+v", route.Desc, params)
			a.renderError(ctx, errRec, nil)
		case string:
			a.renderError(ctx, errors.New(errRec), nil)
		default:
			logs.StatusLog(errRec)
		}

		if route.WithCors {
			views.WriteCORSHeaders(ctx)
		}

		logs.StatusLog(ctx.Path(), err, errRec)

	}()

	resp, err := route.CheckAndRun(ctx, a.fncAuth)
	if err != nil {
		logs.DebugLog("'%s' failure - %v (%v), %s, %s, %s",
			ctx.Path(),
			err,
			resp,
			ctx.Request.Header.ContentType(),
			ctx.Request.Header.Referer(),
			ctx.Request.Header.UserAgent())
		a.renderError(ctx, err, resp)

		return
	}

	// success execution
	if err := views.WriteResponse(ctx, resp); err != nil {
		a.renderError(ctx, err, resp)
	}
}

// renderError send error message into response
func (a *Apis) renderError(ctx *fasthttp.RequestCtx, err error, resp any) {

	statusCode := fasthttp.StatusInternalServerError
	errMsg := err.Error()

	switch errDeep := errors.Cause(err); e := errDeep.(type) {
	case *ErrMethodNotAllowed:
		statusCode = fasthttp.StatusMethodNotAllowed
	case *ErrorResp:
		a.writeBadRequest(ctx, e.FormErrors)
	default:

		switch errDeep {
		case pgx.ErrNoRows, sql.ErrNoRows:
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return
		// can't send standard error (lost headers & response body)
		case ErrUnAuthorized:
			logs.StatusLog("attempt unauthorized access %s", ctx.Request.Header.Referer())
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(errMsg)
			return

		case ErrRouteForbidden:
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetBodyString(errMsg)
			return

		case errRouteOnlyLocal:
			statusCode = fasthttp.StatusForbidden

		case fasthttp.ErrNoMultipartForm:
			statusCode = fasthttp.StatusBadRequest
			_, err := ctx.WriteString("must be multipart-form")
			if err != nil {
				logs.StatusLog(err)
			}

		case ErrWrongParamsList:

			if f, ok := resp.(map[string]string); ok {
				resp = ErrorResp{
					FormErrors: f,
				}
			}

			errMsg = fmt.Sprintf(errMsg, resp)
			a.writeBadRequest(ctx, resp)

			return

		case errNotFoundPage:
			ctx.NotFound()
			logs.StatusLog("Not Found Page %+v", ctx.Request.String())
			return

		default:
			logs.ErrorStack(errDeep, resp)
		}
	}
	ctx.Error(errMsg, statusCode)
}

func (a *Apis) writeBadRequest(ctx *fasthttp.RequestCtx, resp any) {
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
	if err := views.WriteJSON(ctx, resp); err != nil {
		logs.ErrorLog(err, resp)
	}

	if bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte(ContentTypeMultiPart)) {
		logs.DebugLog(ctx.UserValue(MultiPartParams))
	} else if ctx.IsPost() {
		logs.DebugLog(ctx.PostArgs().String())
	} else {
		logs.DebugLog(ctx.QueryArgs().String())
	}
}

// addRoute with safe on concurrency
func (a *Apis) addRoute(path string, route *ApiRoute) error {

	a.Lock()
	defer a.Unlock()

	m := mapRoute{route.Method, path}
	_, ok := a.routes[m]
	if ok {
		return ErrPathAlreadyExists
	}
	a.routes[m] = route

	return nil
}

// AddRoutes add routes to Apis for handling service
// return slice with name routes which not adding
func (a *Apis) AddRoutes(routes ApiRoutes) (badRouting []string) {
	return a.routes.AddRoutes(routes)
}

// renderApis show list routers for developers (as JSON)
func (a *Apis) renderApis(ctx *fasthttp.RequestCtx) (any, error) {
	if ctx.UserValue("json") != nil {
		return a, nil
	}

	if ctx.UserValue("diagram") != nil {
		return a.getDiagram(ctx)
	}

	columns := dbEngine.SimpleColumns(
		"Path - Method",
		"Descriptor",
		"Auth",
		"Required parameters",
		"Others parameters",
		"Dto for JSON parsing",
		"Response",
	)

	rows := make([][]any, 0)

	i := 0
	sortList := make(map[tMethod][]string, 0)
	for m := range a.routes {
		a := sortList[m.method]
		a = append(a, m.path)
		sortList[m.method] = a
	}
	for method, list := range sortList {
		sort.Strings(list)

		for _, url := range list {
			m := mapRoute{method, url}
			route := a.routes[m]
			if m.path == string(ctx.Path()) || strings.HasSuffix(m.path, testRouteSuffix) {
				continue
			}

			row := make([]any, len(columns))
			testURL := path.Join(m.path, a.routes.GetTestRouteSuffix(route))

			row[0] = fmt.Sprintf(`<a href="%s" title="see test">%s</a> - %s`, testURL, m.path, method)
			if route.Multipart {
				row[0] = row[0].(string) + ", MULTIPART"
			}
			row[1] = route.Desc

			s := ""
			if route.OnlyLocal {
				s += "only local request "
			}

			if route.OnlyAdmin {
				s += "only admin allowed "
			}

			if route.FncAuth != nil {
				s += strings.Replace(route.FncAuth.String(), "\n\r", "</br>", -1)
			} else if route.NeedAuth {
				s += strings.Replace(a.fncAuth.String(), "\n\r", "</br>", -1)
			} else {
				s = ""
			}

			if s > "" {
				row[2] = s
			} else {
				row[2] = false
			}

			r := make(map[string]string)
			p := make(map[string]string)
			for _, param := range route.Params {
				s := fmt.Sprintf(`%s, <i>%s</i>`, param.Type, param.Desc)
				if param.DefValue != nil {
					s += fmt.Sprintf(", Def:%v", param.defaultValueOfParams(nil, nil))
				}

				if len(param.PartReq) > 0 {
					s += ", one of {" + strings.Join(param.PartReq, ", ") + ", " + param.Name + "} is required"
				}

				if len(param.IncompatibleWiths) > 0 {
					s += ", only one of {" + strings.Join(param.IncompatibleWiths, ", ") + " OR " + param.Name + "} may use for request"
				}

				if param.Req {
					r[param.Name] = s
				} else {
					p[param.Name] = s
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

	views.WriteHeadersHTML(ctx)

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

// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"bytes"
	"fmt"
	"go/types"
	"path"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/typesExt"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
)

// BuildRouteOptions implement 'Functional Option' pattern for ApiRoute settings
type BuildRouteOptions func(route *ApiRoute)

// RouteAuth set custom auth method on ApiRoute
func RouteAuth(fncAuth FncAuth) BuildRouteOptions {
	return func(route *ApiRoute) {
		route.FncAuth = fncAuth
	}
}

// OnlyLocal set flag of only local response routing
func OnlyLocal() BuildRouteOptions {
	return func(route *ApiRoute) {
		route.OnlyLocal = true
	}
}

// DTO set custom struct on response params
func DTO(dto RouteDTO) BuildRouteOptions {
	return func(route *ApiRoute) {
		route.DTO = dto
	}
}

// MultiPartForm set flag of multipart checking
func MultiPartForm() BuildRouteOptions {
	return func(route *ApiRoute) {
		route.Multipart = true
	}
}

// RouteDTO must to help create some types into routing handling
type RouteDTO interface {
	GetValue() interface{}
	NewValue() interface{}
}

type (
	ApiRouteHandler  func(ctx *fasthttp.RequestCtx) (interface{}, error)
	ApiRouteFuncAuth func(ctx *fasthttp.RequestCtx) error
)

// ApiRoute implement endpoint info & handler on request
type ApiRoute struct {
	Desc                                      string          `json:"descriptor"`
	DTO                                       RouteDTO        `json:"dto"`
	Fnc                                       ApiRouteHandler `json:"-"`
	FncAuth                                   FncAuth         `json:"-"`
	TestFncAuth                               FncAuth         `json:"-"`
	Method                                    tMethod         `json:"method,string"`
	Multipart, NeedAuth, OnlyAdmin, OnlyLocal bool
	Params                                    []InParam   `json:"parameters,omitempty"`
	Resp                                      interface{} `json:"response,omitempty"`
	lock                                      sync.RWMutex
}

// NewAPIRoute create customizing ApiRoute
func NewAPIRoute(desc string, method tMethod, params []InParam, needAuth bool, fnc ApiRouteHandler,
	resp interface{}, Options ...BuildRouteOptions) *ApiRoute {
	route := &ApiRoute{
		Desc:     desc,
		Fnc:      fnc,
		Method:   method,
		Params:   params,
		NeedAuth: needAuth,
		Resp:     resp,
	}

	for _, setOption := range Options {
		setOption(route)
	}

	return route
}

// CheckAndRun check & run route handler
func (route *ApiRoute) CheckAndRun(ctx *fasthttp.RequestCtx, fncAuth FncAuth) (resp interface{}, err error) {

	// check auth is needed
	// owl func for auth
	if (route.FncAuth != nil) && !route.FncAuth.Auth(ctx) ||
		// only admin access
		(route.FncAuth == nil) && (route.OnlyAdmin && !fncAuth.AdminAuth(ctx) ||
			// access according to FncAuth if need
			!route.OnlyAdmin && route.NeedAuth && !fncAuth.Auth(ctx)) {
		return nil, ErrUnAuthorized
	}

	// compliance check local request is needed
	if route.OnlyLocal && isNotLocalRequest(ctx) {
		return nil, errRouteOnlyLocal
	}

	if bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte(ctJSON)) && (route.DTO != nil) {
		// check JSON parsing

		dto := route.DTO.NewValue()
		err := jsoniter.Unmarshal(ctx.Request.Body(), &dto)
		if err != nil {
			ctx.SetUserValue("bad_params", "json DTO not parse :"+err.Error())
			return nil, ErrWrongParamsList
		}
		ctx.SetUserValue(JSONParams, dto)
		// check multipart params
	} else if route.Multipart {
		if !bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte(ctMultiPart)) {
			return nil, fasthttp.ErrNoMultipartForm
		}

		mf, err := ctx.Request.MultipartForm()
		if err != nil {
			return nil, err
		}

		ctx.SetUserValue(MultiPartParams, mf.Value)

		badParams := make([]string, 0)
		for key, value := range mf.Value {
			val, err := route.checkTypeParam(ctx, key, value)
			if err != nil {
				if val != nil {
					logs.DebugLog(val)
				}

				badParams = append(badParams, key+" wrong type "+strings.Join(value, ",")+err.Error())
			}
			ctx.SetUserValue(key, val)
		}

		for key, files := range mf.File {
			ctx.SetUserValue(key, files)
		}

		if len(badParams) > 0 {
			return badParams, ErrWrongParamsList
		}

	} else {
		var args *fasthttp.Args
		if ctx.IsPost() {
			args = ctx.PostArgs()
		} else {
			args = ctx.QueryArgs()
		}

		badParams := make([]string, 0)

		args.VisitAll(func(k, v []byte) {

			key := string(k)
			val, err := route.checkTypeParam(ctx, key, []string{string(v)})
			if err != nil {
				badParams = append(badParams, fmt.Sprintf("'%s' wrong type %v (%s)", key, val, err))
			} else {
				ctx.SetUserValue(key, val)
			}
		})

		if len(badParams) > 0 {
			return badParams, ErrWrongParamsList
		}
	}

	if badParams := route.CheckParams(ctx); len(badParams) > 0 {
		return badParams, ErrWrongParamsList
	}

	return route.Fnc(ctx)
}

// CheckParams check param of request
func (route *ApiRoute) CheckParams(ctx *fasthttp.RequestCtx) (badParams []string) {
	for _, param := range route.Params {
		value := ctx.UserValue(param.Name)
		if value == nil {
			// param is part of group required params
			if param.presentOtherRegParam(ctx) {
				return
			}

			value = param.defaultValueOfParams(ctx)
			//  not present required param
			if value != nil {
				ctx.SetUserValue(param.Name, value)
			} else if param.Req {
				badParams = append(badParams, param.Name+": is required parameter")
			}
		} else if name, val := param.isHasIncompatibleParams(ctx); name > "" {
			// has present param which not compatible with 'param'
			badParams = append(badParams, fmt.Sprintf("incompatible params: %s=%s & %s=%s", param.Name, value, name, val))
		}
	}

	return
}

func (route *ApiRoute) checkTypeParam(ctx *fasthttp.RequestCtx, name string, values []string) (interface{}, error) {
	// find param in InParams list & convert according to Type
	for _, param := range route.Params {
		if param.Name == name {

			if param.Type == nil {
				return values, nil
			}

			if param.Type.IsSlice() {
				return param.Type.ConvertSlice(ctx, values)
			}

			return param.Type.ConvertValue(ctx, values[0])
		}
	}

	if len(values) == 1 {
		return values[0], nil
	}

	return values, nil
}

func (route *ApiRoute) isValidMethod(ctx *fasthttp.RequestCtx) bool {
	return route.Method == methodFromName(string(ctx.Method()))
}

// apiRouteToJSON produces a human-friendly description of Apis.
//Based on real data of the executable application, does not require additional documentation.
func apiRoutesToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	routes := *(*ApiRoutes)(ptr)

	sortList := make([]string, 0, len(routes))
	for name := range routes {
		sortList = append(sortList, name)
	}
	sort.Strings(sortList)

	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	FirstFieldToJSON(stream, "Descriptor", "routers description, params response format, ect")
	for _, name := range sortList {
		AddObjectToJSON(stream, name, *(routes[name]))
	}
}

// apiRouteToJSON produces a human-friendly description of Apis.
//Based on real data of the executable application, does not require additional documentation.
func apiRouteToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	// todo: add description of the test-based return data
	route := (*ApiRoute)(ptr)

	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	FirstFieldToJSON(stream, "Descriptor", route.Desc)

	AddFieldToJSON(stream, "Method", methodNames[route.Method])

	if route.FncAuth != nil {
		AddFieldToJSON(stream, "AuthCustom", "use custom method '"+route.FncAuth.String()+"' for checking authorization")
	} else if route.NeedAuth {
		AddFieldToJSON(stream, "Auth", "use standard method for checking authorization")
	}

	if route.OnlyAdmin {
		AddFieldToJSON(stream, "AdminOnly", "only admin request be allowed")
	}

	if route.OnlyLocal {
		AddFieldToJSON(stream, "LocalOnly", "only local request be allowed")
	}

	// print parameters
	if len(route.Params) > 0 {
		AddObjectToJSON(stream, "parameters", route.Params)
	}

	if route.DTO != nil {
		AddObjectToJSON(stream, "DtoFromJSON", route.DTO.GetValue())
		AddFieldToJSON(stream, "DtoFromJSONType", fmt.Sprintf("%+#v", route.DTO.GetValue()))
	}

	if route.Resp != nil {
		if resp, ok := route.Resp.(string); ok {
			AddFieldToJSON(stream, "Response", resp)
		} else {
			AddObjectToJSON(stream, "Response", route.Resp)
		}
	}
}

func AddFieldToJSON(stream *jsoniter.Stream, field string, s string) {
	stream.WriteMore()
	FirstFieldToJSON(stream, field, s)
}

func AddObjectToJSON(stream *jsoniter.Stream, field string, s interface{}) {
	stream.WriteMore()
	stream.WriteObjectField(field)
	stream.WriteVal(s)
}

func FirstFieldToJSON(stream *jsoniter.Stream, field string, s string) {
	stream.WriteObjectField(field)
	stream.WriteString(s)
}

func FirstObjectToJSON(stream *jsoniter.Stream, field string, s interface{}) {
	stream.WriteObjectField(field)
	stream.WriteVal(s)
}

// ApiRoutes is hair of APIRoute
type ApiRoutes map[string]*ApiRoute
type MapRoutes map[tMethod]map[string]*ApiRoute

// NewMapRoutes create APIRotes instance
func NewMapRoutes() MapRoutes {
	m := make(MapRoutes, 0)
	for method := range methodNames {
		m[tMethod(method)] = make(ApiRoutes, 0)
	}

	return m
}

// AddRoutes add ApiRoute into hair onsafe
func (r MapRoutes) AddRoutes(routes ApiRoutes) (badRouting []string) {
	for url, route := range routes {
		_, ok := r[route.Method][url]
		if ok {
			logs.ErrorLog(ErrPathAlreadyExists, url)
			badRouting = append(badRouting, url)
			continue
		}

		r[route.Method][url] = route
		testRoute := &ApiRoute{
			Desc:      "test handler for " + url,
			DTO:       nil,
			Fnc:       nil,
			FncAuth:   route.TestFncAuth,
			Method:    GET,
			Multipart: false,
			NeedAuth:  false,
			OnlyAdmin: false,
			OnlyLocal: false,
			Params:    make([]InParam, len(route.Params)),
			Resp:      route.Resp,
		}
		for i, param := range route.Params {
			testRoute.Params[i] = param
			if param.Req {
				testRoute.Params[i].Req = false
				// testRoute.Params[i].DefValue = param.TestValue
			}
		}

		testRoute.Fnc = func(url string, route *ApiRoute) ApiRouteHandler {
			return func(ctx *fasthttp.RequestCtx) (interface{}, error) {
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

		r[GET][path.Join(url, r.GetTestRouteSuffix(route))] = testRoute

	}

	return
}

func (r MapRoutes) GetTestRouteSuffix(route *ApiRoute) string {

	if route.Method != GET {
		return methodNames[route.Method] + testRouteSuffix
	}

	return testRouteSuffix
}
func (r MapRoutes) GetRoute(ctx *fasthttp.RequestCtx) (*ApiRoute, error) {
	path := string(ctx.Path())
	method := methodFromName(string(ctx.Method()))

	// check exactly
	if route, ok := r[method][path]; ok {
		return route, nil
	}

	// find parent path with some method
	if route, parent := r.findParentRoute(method, path); route != nil {
		ctx.SetUserValue(ChildRoutePath, strings.TrimPrefix(path, parent))
		return route, nil
	}

	for m, routes := range r {
		if method == m {
			continue
		}
		route, ok := routes[path]
		if ok {
			return route, errMethodNotAllowed
		}
	}

	return nil, errNotFoundPage
}

func (r MapRoutes) findParentRoute(method tMethod, path string) (*ApiRoute, string) {
	for p := getParentPath(path); p > ""; p = getParentPath(p) {
		route, ok := r[method][p]
		// check method
		if ok {
			return route, p
		}
	}

	return nil, ""
}

func getParentPath(path string) string {
	n := strings.LastIndex(strings.TrimSuffix(path, "/"), "/")
	if n < 0 {
		return ""
	}

	return path[:n+1]
}

func init() {
	jsoniter.RegisterTypeEncoderFunc("apis.ApiRoute", apiRouteToJSON, func(pointer unsafe.Pointer) bool {
		return false
	})
	jsoniter.RegisterTypeEncoderFunc("apis.ApiRoutes", apiRoutesToJSON, func(pointer unsafe.Pointer) bool {
		return false
	})
	jsoniter.RegisterTypeEncoderFunc("apis.MapRoutes", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		routes := *(*MapRoutes)(ptr)
		stream.WriteObjectStart()
		defer stream.WriteObjectEnd()

		isFirst := true
		for key, val := range routes {
			if val == nil || len(val) == 0 {
				continue
			}

			if isFirst {
				isFirst = false
				FirstObjectToJSON(stream, key.String(), &val)
			} else {
				AddObjectToJSON(stream, key.String(), &val)
			}
		}

	}, func(pointer unsafe.Pointer) bool {
		return false
	})
}

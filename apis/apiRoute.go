// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/json-iterator/go"
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/valyala/fasthttp"
)

func init() {
	jsoniter.RegisterTypeEncoderFunc("apis.ApiRoute", routingToJSON, func(pointer unsafe.Pointer) bool {
		return false
	})
	jsoniter.RegisterTypeEncoderFunc("apis.ApiRoutes", apiRoutesToJSON, func(pointer unsafe.Pointer) bool {
		return false
	})
}

// BuildRouteOptions implement 'Functional Option' pattern for ApiRoute settings
type BuildRouteOptions func(route *ApiRoute)

// RouteAuth set custom auth method on ApiRoute
func RouteAuth(fncAuth ApiRouteFuncAuth) BuildRouteOptions {
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
	Desc                           string           `json:"descriptor"`
	DTO                            RouteDTO         `json:"dto"`
	Fnc                            ApiRouteHandler  `json:"-"`
	FncAuth                        ApiRouteFuncAuth `json:"-"`
	Method                         tMethod          `json:"method,string"`
	Multipart, NeedAuth, OnlyLocal bool
	Params                         []InParam   `json:"parameters,omitempty"`
	Resp                           interface{} `json:"response"`
	lock                           sync.RWMutex
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
func (route *ApiRoute) CheckAndRun(ctx *fasthttp.RequestCtx, fncAuth func(ctx *fasthttp.RequestCtx) bool) (resp interface{}, err error) {

	// check auth is needed
	if route.FncAuth != nil {
		// route has his auth method
		if err := route.FncAuth(ctx); err != nil {
			return nil, err
		}
	} else if route.NeedAuth && !fncAuth(ctx) {
		return nil, ErrUnAuthorized
	}

	// compliance check local request is needed
	if route.OnlyLocal && isNotLocalRequest(ctx) {
		return nil, errRouteOnlyLocal
	}

	// check multipart params
	if route.Multipart {
		if !bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte(ctMultipArt)) {
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
					return val, errors.Wrap(ErrWrongParamsList, err.Error())
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

	} else if bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte(ctJSON)) && (route.DTO != nil) {
		// check JSON parsing

		dto := route.DTO.NewValue()
		err := jsoniter.Unmarshal(ctx.Request.Body(), &dto)
		if err != nil {
			ctx.SetUserValue("bad_params", "json DTO not parse :"+err.Error())
			return nil, ErrWrongParamsList
		}
		ctx.SetUserValue(JSONParams, dto)

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
				badParams = append(badParams, key+" wrong type "+val.(string)+err.Error())
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
			if route.isHasPartRegParam(ctx, param) {
				return
			}

			value = route.defaultValueOfParams(ctx, param)
			//  not present required param
			if value != nil {
				ctx.SetUserValue(param.Name, value)
			} else if param.Req {
				badParams = append(badParams, param.Name+": is required parameter")
			}
		} else if name, val := route.isHasIncompatibleParams(ctx, param); name > "" {
			// has present param which not compatible with 'param'
			badParams = append(badParams, fmt.Sprintf("incompatible params: %s=%s & %s=%s", param.Name, value, name, val))
		}
	}

	return
}

type DefValueCalcFnc = func(ctx *fasthttp.RequestCtx) interface{}

// defaultValueOfParams return value as default for param, it is only for single required param
func (route *ApiRoute) defaultValueOfParams(ctx *fasthttp.RequestCtx, param InParam) interface{} {
	switch def := param.DefValue.(type) {
	case DefValueCalcFnc:
		return def(ctx)
	default:
		return param.DefValue
	}
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

// found params incompatible with 'param'
func (route *ApiRoute) isHasIncompatibleParams(ctx *fasthttp.RequestCtx, param InParam) (string, interface{}) {
	for _, name := range param.IncompatibleWiths {
		val := ctx.FormValue(name)
		if len(val) > 0 {
			return name, val
		}
	}

	return "", nil
}

// check 'param' is one part of list required params AND one of other params is present
func (route *ApiRoute) isHasPartRegParam(ctx *fasthttp.RequestCtx, param InParam) bool {
	isPartReq := param.isPartReq()

	if isPartReq {
		// Looking for parameters associated with the original 'param'
		for _, name := range param.PartReq {
			// param 'name' is present
			if ctx.UserValue(name) != nil {
				return true
			}
		}

		return false
	}

	return isPartReq
}

func (route *ApiRoute) isValidMethod(ctx *fasthttp.RequestCtx) bool {
	return methodNames[route.Method] == string(ctx.Method())
}

// routingToJSON produces a human-friendly description of Apis.
//Based on real data of the executable application, does not require additional documentation.
func apiRoutesToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	routes := *(*ApiRoutes)(ptr)

	sortList := make([]string, 0, len(routes))
	for name := range routes {
		sortList = append(sortList, name)
	}
	sort.Strings(sortList)

	stream.WriteObjectStart()
	FirstFieldToJSON(stream, "Descriptor", "routers description, params response format, ect")
	for _, name := range sortList {
		AddObjectToJSON(stream, name, routes[name])
	}

	stream.WriteObjectEnd()

}

// routingToJSON produces a human-friendly description of Apis.
//Based on real data of the executable application, does not require additional documentation.
func routingToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	// todo: add description of the test-based return data
	route := (*ApiRoute)(ptr)

	stream.WriteObjectStart()

	FirstFieldToJSON(stream, "Descriptor", route.Desc)

	AddFieldToJSON(stream, "Method", methodNames[route.Method])

	if route.FncAuth != nil {
		t := reflect.TypeOf(route.FncAuth)
		AddFieldToJSON(stream, "AuthCustom", "use custom method '"+t.Name()+"' for checking autorization")
	} else if route.NeedAuth {
		AddFieldToJSON(stream, "Auth", "use standard method for checking authorization")

	}

	if route.OnlyLocal {
		AddFieldToJSON(stream, "LocalOnly", "only local request be allowed")
	}

	// print parameters
	if len(route.Params) > 0 {
		stream.WriteMore()
		stream.WriteObjectField("parameters")
		stream.WriteObjectStart()
		for i, param := range route.Params {
			if i > 0 {
				stream.WriteMore()
			}

			stream.WriteObjectField(param.Name)
			stream.WriteObjectStart()
			FirstFieldToJSON(stream, "Descriptor", param.Desc)

			if param.Type != nil {
				AddFieldToJSON(stream, "Type", param.Type.String())
			}

			if param.Req {
				if len(param.PartReq) > 0 {
					s := strings.Join(param.PartReq, ",")
					AddFieldToJSON(stream, "Required", "one of "+s+" and "+param.Name+" is required")
				} else {
					AddObjectToJSON(stream, "Required", true)
				}

			}
			stream.WriteObjectEnd()
		}
		stream.WriteObjectEnd()
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

	stream.WriteObjectEnd()
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

// ApiRoutes is hair of APIROute
type ApiRoutes map[string]*ApiRoute

// NewAPIRoutes create APIRotes instance
func NewAPIRoutes() ApiRoutes {
	return make(map[string]*ApiRoute, 0)
}

// AddRoutes add ApiRoute into hair onsafe
func (s ApiRoutes) AddRoutes(routes ApiRoutes) (badRouting []string) {
	for path, route := range routes {
		_, ok := s[path]
		if ok {
			logs.ErrorLog(ErrPathAlreadyExists, path)
			badRouting = append(badRouting, path)
		} else {
			s[path] = route
		}
	}

	return
}

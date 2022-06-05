// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"fmt"
	"github.com/json-iterator/go"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"
	"go/types"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

// Apis has auto generate documentation (created at start http-service)
// We may see it on two special routes:
// 1. /apis - html version, shows routes into table: path, description, parameters, etc.
// 2. /swagger.io - swagger version (it still be tested at now)
// Swagger use data as json getting from /apis?json

func init() {
	jsoniter.RegisterTypeEncoderFunc("apis.Apis", apisToJSON, func(pointer unsafe.Pointer) bool {
		return false
	})
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
		paths := make([]string, 0)
		for method, val := range routes {
			if val == nil || len(val) == 0 {
				continue
			}

		routes:
			for path, route := range val {
				if isFirst {
					isFirst = false
				} else {
					for _, s := range paths {
						if path == s {
							continue routes
						}
					}
					paths = append(paths, path)
					stream.WriteMore()
				}

				stream.WriteObjectField(path)
				stream.WriteObjectStart()
				FirstObjectToJSON(stream, strings.ToLower(method.String()), route)
				stream.WriteObjectEnd()
			}
		}

	}, func(pointer unsafe.Pointer) bool {
		return false
	})
}

// apiRoutesToJSON produces a human-friendly description of Apis.
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
// Based on real data of the executable application, does not require additional documentation.
func apiRouteToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	route := (*ApiRoute)(ptr)

	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	FirstFieldToJSON(stream, "description", route.Desc)
	summary := ""
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

	respErrors := make(map[string]InParam)
	ctx := &fasthttp.RequestCtx{}
	// print parameters
	params := make([]interface{}, 0)
	if len(route.Params) > 0 {
		for _, param := range route.Params {
			if param.DefValue == ChildRoutePath {
				summary += fmt.Sprintf("/{%s}", param.Name)
			}

			params = append(params, param)
			if param.Req {
				param.DefValue = PARAM_REQUIRED
			} else {
				badParams := make(map[string]string)
				param.Check(ctx, badParams)
				if len(badParams) > 0 {
					param.DefValue = badParams
				} else if param.Type == nil {
					param.DefValue = "todo"
				} else {
					param.DefValue = fmt.Sprintf("wrong type, expect: %s", param.Type.String())
				}
			}

			respErrors[param.Name] = param

		}
	}

	if route.DTO != nil {
		value := route.DTO.NewValue()
		v := reflect.ValueOf(value)
		if !v.IsZero() {
			stream.WriteMore()

			params = append(params, writeReflect("JSON", v, stream))
			r, ok := (value).(Visit)
			if ok {
				//fastjson.MustParse(`{}`).GetObject().Visit(r.Each)
				badParams, err := r.Result()
				if err != nil {
					respErrors["body"] = InParam{
						Name:     err.Error(),
						Desc:     "JSON",
						Req:      true,
						DefValue: badParams,
					}
				}
			} else if c, ok := (value).(CheckDTO); ok {
				badParams := make(map[string]string)
				if !c.CheckParams(ctx, badParams) {
					for name, s := range badParams {

						respErrors[name] = InParam{
							Name:     s,
							Desc:     "",
							Req:      true,
							DefValue: s,
						}
					}

				}
			}

		}
	}

	if route.WithCors {
		summary += ", CORS"
	}
	if route.Multipart {
		summary += ", multipart"
	}
	if route.NeedAuth {
		summary += ", only auth access"
	}
	if route.OnlyAdmin {
		summary += ", only admin access"
	}
	if route.OnlyLocal {
		summary += ", only local request"
	}

	AddFieldToJSON(stream, "summary", summary)

	AddObjectToJSON(stream, "parameters", params)

	tags := make([]string, 0)

	if strings.Contains(route.Desc, "test handler") {
		tags = append(tags, "Test handlers (auto generated)")
	} else if strings.Contains(route.Desc, "get form") {
		tags = append(tags, "Forms handlers (auto generated)")
	} else if strings.Contains(route.Desc, "APIS") {
		tags = append(tags, "APIS handlers (auto generated)")
	} else if strings.Contains(route.Desc, "table") {
		tags = append(tags, "CRUD")
	} else if strings.Contains(route.Desc, "httpgo") {
		tags = append(tags, "HttpGo managements")
	} else {
		parts := strings.Split(route.Desc, "#")
		if len(parts) > 2 {
			tags = append(tags, parts[1])
		} else if parts := strings.Split(route.Desc, "*"); len(parts) > 2 {
			tags = append(tags, parts[1])
		}
	}

	AddObjectToJSON(stream, "tags", tags)
	AddObjectToJSON(stream, "consumes", []string{
		"application/json",
	})
	AddObjectToJSON(stream, "produces", []string{
		"application/json",
		"text/plain",
	})
	stream.WriteMore()
	writeResponse(stream, respErrors, route.Resp)

	if route.NeedAuth {
		writeResponseForAuth(stream)
	}

	stream.WriteObjectEnd()
}

func writeReflect(title string, value reflect.Value, stream *jsoniter.Stream) interface{} {
	kind := value.Kind()
	// Handle pointers specially.
	kind, value = indirect(kind, value)
	defer func() {
		e := recover()
		err, ok := e.(error)
		if ok {
			logs.ErrorLog(err, kind.String(), value.String())
		}
	}()

	if kind > reflect.UnsafePointer || kind <= 0 {
		stream.WriteObjectField(title)
		stream.WriteObjectField(value.String())
		stream.WriteString(kind.String())
		desc := ""
		if parts := strings.Split(title, ","); len(parts) > 1 {
			title = parts[0]
			desc = parts[1]
		}

		return InParam{
			Name:              title,
			Desc:              desc,
			Req:               false,
			PartReq:           nil,
			Type:              NewTypeInParam(types.String),
			DefValue:          kind.String(),
			IncompatibleWiths: nil,
			TestValue:         "",
		}
	}

	vType := value.Type()
	sType := vType.String()
	if parts := strings.Split(title, ","); len(parts) > 1 {
		title = parts[0]
		sType += ", " + parts[1]
	}

	stream.WriteObjectField(title)

	return WriteReflectKind(kind, value, stream, sType, title)
}

func WriteReflectKind(kind reflect.Kind, value reflect.Value, stream *jsoniter.Stream, sType, title string) interface{} {
	switch kind {
	case reflect.Struct:
		return WriteStruct(value, stream, title)

	case reflect.Map:
		return WriteMap(value, stream, title)

	case reflect.Array, reflect.Slice:
		return WriteSlice(value, stream, title)

	default:
		stream.WriteString(sType)
		return InParam{
			Name:              title,
			Desc:              "",
			Req:               false,
			PartReq:           nil,
			Type:              &ReflectType{Type: value.Type()},
			DefValue:          kind.String(),
			IncompatibleWiths: nil,
			TestValue:         "",
		}
	}
}

func WriteMap(value reflect.Value, stream *jsoniter.Stream, title string) interface{} {
	// nil maps should be indicated as different than empty maps
	if value.IsNil() {
		stream.WriteEmptyObject()
		return nil
	}

	stream.WriteObjectStart()
	keys := value.MapKeys()
	propers := make([]interface{}, 0)
	for i, v := range keys {
		if i > 0 {
			stream.WriteMore()
		}
		propers = append(propers, writeReflect(fmt.Sprintf("%d: %s %s `%s`", i, v.Kind(), v.Type(), v.String()), v, stream))
	}

	stream.WriteObjectEnd()
	return NewSwaggerParam(propers, title, "object")
}

func WriteSlice(value reflect.Value, stream *jsoniter.Stream, title string) interface{} {
	stream.WriteArrayStart()
	defer stream.WriteArrayEnd()

	vType := value.Type()
	numEntries := value.Len()
	if numEntries == 0 {

		elem := vType.Elem()

		for kind := elem.Kind(); ; kind = elem.Kind() {
			switch kind {
			case reflect.Ptr, reflect.Interface, reflect.UnsafePointer:
				elem = elem.Elem()
				continue

			case reflect.Struct:
				return WriteStruct(reflect.Zero(elem), stream, title)

			default:
				return WriteReflectKind(kind, reflect.New(elem), stream, vType.Elem().String(), title)
			}

		}
	}

	propers := make([]interface{}, 0)

	for i := 0; i < numEntries; i++ {
		if i > 0 {
			stream.WriteMore()
		}
		v := value.Index(i)
		propers = append(propers, writeReflect(fmt.Sprintf("%d: %s %s `%s`", i, v.Kind(), v.Type(), v.String()), v, stream))
	}

	return NewSwaggerArray(title, propers...)
}

func WriteStruct(value reflect.Value, stream *jsoniter.Stream, title string) interface{} {
	propers := make(map[string]interface{}, 0)
	vType := value.Type()
	stream.WriteObjectStart()
	for i, isFirst := 0, true; i < value.NumField(); i++ {
		v := vType.Field(i)
		if !v.IsExported() {
			continue
		}
		if !isFirst {
			stream.WriteMore()
		}

		val := value.Field(i)
		kind := val.Kind()
		kind, val = indirect(kind, val)

		tag := v.Tag.Get("json")
		title := tag // fmt.Sprintf("%s: %s", v.Name, v.Type) + writeTag(v.Tag)
		if tag == "" {
			title = v.Name
		}

		propers[title] = writeReflect(title, val, stream)
		isFirst = false
	}
	stream.WriteObjectEnd()

	return NewSwaggerObject(propers, title)
}

func writeTag(tag reflect.StructTag) string {
	if tag > "" {
		return " `" + string(tag) + "`"
	}
	return ""
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

// apiRouteToJSON produces a human-friendly description of Apis.
// Based on real data of the executable application, does not require additional documentation.
func apisToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	apis := *(*Apis)(ptr)
	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	defer func() {
		e := recover()
		err, ok := e.(error)
		if ok {
			logs.ErrorStack(err)
		}
	}()

	FirstFieldToJSON(stream, "swagger", "2.0")
	stream.WriteMore()

	stream.WriteObjectField("info")
	stream.WriteObjectStart()
	FirstFieldToJSON(stream, "descriptor", "API Specification, include endpoints description, ect")
	version, ok := apis.Ctx.Value(ApiVersion).(string)
	if ok {
		AddFieldToJSON(stream, "version", version)
	}
	AddFieldToJSON(stream, "title", "httpgo")
	stream.WriteMore()

	stream.WriteObjectField("license")
	stream.WriteObjectStart()
	FirstFieldToJSON(stream, "name", "Apache 2.0")
	AddFieldToJSON(stream, "url", "http://www.apache.org/licenses/LICENSE-2.0.html")

	stream.WriteObjectEnd()

	if apis.fncAuth != nil {
		AddObjectToJSON(stream, "auth", apis.fncAuth.String())
	}
	stream.WriteObjectEnd()

	AddObjectToJSON(stream, "schemes", []string{schemas[apis.Https]})
	AddObjectToJSON(stream, "paths", apis.routes)
}

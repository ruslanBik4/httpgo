/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"fmt"
	"go/types"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/logs"
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
	jsoniter.RegisterTypeEncoderFunc("apis.MapRoutes", mapRoutesToJSON, func(pointer unsafe.Pointer) bool {
		return false
	})
}

func mapRoutesToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	mapRoutes := *(*MapRoutes)(ptr)
	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()
	defer func() {
		err, ok := recover().(error)
		if ok {
			logs.ErrorLog(err)
		}
	}()

	paths := make(map[string][]*ApiRoute, 0)
	for m, route := range mapRoutes {
		a := paths[m.path]
		a = append(a, route)
		paths[m.path] = a
	}
	sortList := make([]string, 0, len(paths))
	for name := range paths {
		sortList = append(sortList, name)
	}
	sort.Strings(sortList)
	isFirst := true
	for _, path := range sortList {
		if !isFirst {
			stream.WriteMore()
		} else {
			isFirst = false
		}

		stream.WriteObjectField(path)
		stream.WriteObjectStart()

		for i, route := range paths[path] {
			if i > 0 {
				stream.WriteMore()
			}
			FirstObjectToJSON(stream, strings.ToLower(route.Method.String()), route)
		}
		stream.WriteObjectEnd()
	}
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
	//AddFieldToJSON(stream, "Method", methodNames[route.Method])

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
	params := make([]any, 0)
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
				} else if t, ok := param.Type.(TypeInParam); ok && t.DTO != nil {
					value := t.DTO.NewValue()
					if d, ok := value.(Docs); ok {
						param.DefValue = fmt.Sprintf("wrong type, expect: %s", d.Expect())
					} else {
						param.DefValue = fmt.Sprintf("wrong type, expect: %T", value)
					}
				} else {
					param.DefValue = fmt.Sprintf("wrong type, expect: %s", param.Type.String())
				}
			}

			respErrors[param.Name] = param
		}
	}

	if route.FncAuth != nil {
		summary += route.FncAuth.String()
	}
	if route.DTO != nil {
		value := route.DTO.NewValue()
		v := reflect.ValueOf(value)
		if !v.IsZero() {
			stream.WriteMore()

			p := writeReflect("JSON", v, stream)
			params = append(params, p)
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
						_, ok := respErrors[name]
						if ok {
							continue
						}

						respErrors[name] = InParam{
							Name:     name,
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

	stream.WriteMore()
	in := "query"
	if route.Multipart {
		in = "formdata"
	}
	jParam := NewqInParam(in)
	jParam.WriteSwaggerParams(stream, params)

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
	//AddObjectToJSON(stream, "consumes", []string{
	//	"application/json",
	//})
	//AddObjectToJSON(stream, "produces", []string{
	//	"application/json",
	//	"text/plain",
	//})
	stream.WriteMore()
	writeResponse(stream, respErrors, route.Resp)

	if route.NeedAuth {
		writeResponseForAuth(stream)
	}

	//if route.NeedAuth {
	//	stream.WriteMore()
	//	WriteSecurity(stream, "")
	//}
	stream.WriteObjectEnd()
}

func writeReflect(title string, value reflect.Value, stream *jsoniter.Stream) any {
	i := value.Interface()
	// Handle pointers specially.
	kind, val := indirect(value.Kind(), value)
	defer func() {
		e := recover()
		err, ok := e.(error)
		if ok {
			logs.ErrorLog(err, kind.String(), val.String())
		}
	}()

	if kind > reflect.UnsafePointer || kind <= 0 {
		stream.WriteObjectField(title)
		stream.WriteObjectField(val.String())
		stream.WriteString(kind.String())
		desc := ""
		if parts := strings.Split(title, ","); len(parts) > 1 {
			title = parts[0]
			desc = parts[1]
		}

		logs.StatusLog(title, desc)

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

	vType := val.Type()
	sType := vType.String()
	if parts := strings.Split(title, ","); len(parts) > 1 {
		title = parts[0]
		sType += ", " + parts[1]
	}

	stream.WriteObjectField(title)
	elem := WriteReflectKind(kind, val, stream, sType, title)
	if elem == nil {
		var typ APIRouteParamsType = &ReflectType{Type: value.Type()}
		if d, ok := i.(Docs); ok {
			title = d.Expect()

		} else if d, ok := i.(*Docs); ok {
			title = (*d).Expect()
		}
		if r, ok := i.(RouteDTO); ok {
			typ = NewStructInParam(r)
		}

		stream.WriteString(title)
		elem = InParam{
			Name:              title,
			Desc:              "default",
			Req:               false,
			PartReq:           nil,
			Type:              typ,
			DefValue:          title,
			IncompatibleWiths: nil,
			TestValue:         "",
		}
	}

	return elem
}

func WriteReflectKind(kind reflect.Kind, value reflect.Value, stream *jsoniter.Stream, sType, title string) any {
	switch kind {
	case reflect.Struct:
		return WriteStruct(value, stream, title)

	case reflect.Map:
		return WriteMap(value, stream, title)

	case reflect.Array, reflect.Slice:
		return WriteSlice(value, stream, title)

	default:
		//logs.StatusLog(title, sType, value)
		stream.WriteString(sType) //writeReflect(title, value, stream)
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

func WriteMap(value reflect.Value, stream *jsoniter.Stream, title string) any {
	// nil maps should be indicated as different than empty maps
	if value.IsNil() {
		stream.WriteEmptyObject()
		return nil
	}

	stream.WriteObjectStart()
	keys := value.MapKeys()
	propers := make([]any, 0)
	for i, v := range keys {
		if i > 0 {
			stream.WriteMore()
		}
		propers = append(propers, writeReflect(fmt.Sprintf("%d: %s %s `%s`", i, v.Kind(), v.Type(), v.String()), v, stream))
	}

	stream.WriteObjectEnd()
	return NewSwaggerParam(propers, title, "object")
}

func WriteSlice(value reflect.Value, stream *jsoniter.Stream, title string) any {
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

	propers := make([]any, 0)

	for i := 0; i < numEntries; i++ {
		if i > 0 {
			stream.WriteMore()
		}
		v := value.Index(i)
		propers = append(propers, writeReflect(fmt.Sprintf("%d: %s %s `%s`", i, v.Kind(), v.Type(), v.String()), v, stream))
	}

	return NewSwaggerArray(title, propers...)
}

func WriteStruct(value reflect.Value, stream *jsoniter.Stream, title string) any {
	props := make(map[string]any, 0)
	vType := value.Type()
	writeFields(value, stream, vType, props, map[string]struct{}{})

	if len(props) > 0 {
		return NewSwaggerObject(props, title)
	}

	return nil
}

func writeFields(value reflect.Value, stream *jsoniter.Stream, vType reflect.Type, props map[string]any, overlays map[string]struct{}) {
	list := make(map[string]reflect.StructField)
	titles := make(map[string]struct{})

	for i := 0; i < value.NumField(); i++ {
		tField := vType.Field(i)
		if !tField.IsExported() {
			continue
		}

		name := tField.Name
		if _, ok := overlays[name]; ok {
			continue
		}
		list[name] = tField
		titles[name] = struct{}{}
	}

	if len(list) == 0 {
		return
	}

	stream.WriteObjectStart()
	defer func() {
		err, ok := recover().(error)
		if ok {
			logs.ErrorStack(err)
		}
		stream.WriteObjectEnd()
	}()

	isFirst := true
	for name, tField := range list {

		if !isFirst {
			stream.WriteMore()
		}

		val := value.FieldByName(name)
		tag := tField.Tag.Get("json")
		if tag == "" {
			tag = tField.Name
		}

		if tField.Anonymous {
			kind := val.Kind()
			kind, val = indirect(kind, val)
			writeFields(val, stream, val.Type(), props, titles)
		} else {
			props[tag] = writeReflect(tag, val, stream)
		}
		isFirst = false
	}
}

func indirect(kind reflect.Kind, value reflect.Value) (reflect.Kind, reflect.Value) {
	for kind == reflect.Pointer || kind == reflect.UnsafePointer || kind == reflect.Interface {
		if value.IsZero() {
			value = reflect.New(value.Type().Elem())
		} else {
			value = value.Elem()
		}
		kind = value.Kind()
	}

	return kind, value
}

func AddFieldToJSON(stream *jsoniter.Stream, field string, s string) {
	stream.WriteMore()
	FirstFieldToJSON(stream, field, s)
}

func AddObjectToJSON(stream *jsoniter.Stream, field string, s any) {
	stream.WriteMore()
	stream.WriteObjectField(field)
	stream.WriteVal(s)
}

func FirstFieldToJSON(stream *jsoniter.Stream, field string, s string) {
	stream.WriteObjectField(field)
	stream.WriteString(s)
}

func FirstObjectToJSON(stream *jsoniter.Stream, field string, s any) {
	stream.WriteObjectField(field)
	stream.WriteVal(s)
}

// apisToJSON produces a human-friendly description of Apis.
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

	FirstFieldToJSON(stream, "openapi", "3.0.3")
	stream.WriteMore()

	stream.WriteObjectField("info")
	stream.WriteObjectStart()
	FirstFieldToJSON(stream, "description", "API Specification, include endpoints description, ect")
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
	stream.WriteObjectEnd()

	if apis.fncAuth != nil {
		stream.WriteMore()
		WriteBearer(stream, apis.fncAuth.String())
	}

	AddObjectToJSON(stream, "schemes", []string{schemas[apis.Https]})
	AddObjectToJSON(stream, "paths", apis.routes)
}
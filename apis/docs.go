/*
 * Copyright (c) 2022-2024. Author: Ruslan Bikchentaev. All rights reserved.
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

	"github.com/ruslanBik4/httpgo/auth"
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

const authSecurity = "auth"

func mapRoutesToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	mapRoutes := *(*MapRoutes)(ptr)
	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()
	defer func() {
		switch err := recover().(type) {
		case nil:
		case error:
			logs.ErrorLog(err)
		default:
			logs.StatusLog("recover: %v", err)
		}
	}()

	paths := make(map[string][]*ApiRoute, len(mapRoutes)%2)
	for m, route := range mapRoutes {
		paths[m.path] = append(paths[m.path], route)
	}
	sortList := make([]string, 0, len(paths))
	for name := range paths {
		sortList = append(sortList, name)
	}
	sort.Strings(sortList)
	for i, path := range sortList {
		stream.WriteObjectField(path)
		stream.WriteObjectStart()

		for j, route := range paths[path] {
			FirstObjectToJSON(stream, strings.ToLower(route.Method.String()), route)
			if j+1 < len(paths[path]) {
				stream.WriteMore()
			}
		}
		stream.WriteObjectEnd()
		if i+1 < len(sortList) {
			stream.WriteMore()
		}
	}
}

// apiRoutesToJSON produces a human-friendly description of Apis.
// Based on real data of the executable application, does not require additional documentation.
func apiRoutesToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	routes := *(*ApiRoutes)(ptr)

	sortList := make([]string, 0, len(routes))
	for name := range routes {
		sortList = append(sortList, name)
	}
	sort.Strings(sortList)

	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	FirstFieldToJSON(stream, "description", "routers description, params response format, ect")
	for _, name := range sortList {
		AddObjectToJSON(stream, name, *(routes[name]))
	}
}

// apiRouteToJSON produces a human-friendly description of Apis.
// Based on real data of the executable application, does not require additional documentation.
func apiRouteToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	route := (*ApiRoute)(ptr)
	in := "query"
	if route.Multipart {
		in = "body"
	}

	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	FirstFieldToJSON(stream, "description", route.Desc)
	summary := ""

	if route.FncAuth != nil {
		//todo: create custom security object
		AddFieldToJSON(stream, "AuthCustom", "use custom method '"+route.FncAuth.String()+"' for checking authorization")
	}
	ctx := &fasthttp.RequestCtx{}
	// print parameters
	params := make([]InParam, len(route.Params))
	respErrors := make([]InParam, len(route.Params))
	for i, param := range route.Params {
		params[i] = param
		if param.DefValue == ApisValues(ChildRoutePath) {
			summary += fmt.Sprintf("{%s} ", param.Name)
		}

		respErrors[i] = convertParamForErrResp(ctx, param)
	}
	if route.IsAJAXRequest {
		params = append(params, InParam{
			Name:              "X-Requested-With",
			Desc:              "header - flag to mark AJAX (Asynchronous JavaScript and XML) requests",
			Req:               true,
			PartReq:           nil,
			Type:              HeaderInParam{NewTypeInParam(types.String)},
			DefValue:          "XMLHttpRequest",
			IncompatibleWiths: nil,
			TestValue:         "",
		})
		summary += "(AJAX request)"
	}
	writeSummary(stream, route, summary)

	var valueDTO any
	if dto := route.DTO; dto != nil {
		value := dto.NewValue()
		v := reflect.ValueOf(value)
		if !v.IsZero() {
			valueDTO = writeReflect("properties", v, stream)
		}
		resp := getRespError(ctx, value)
		if resp != nil {
			isDuplicate := false
			for _, r := range respErrors {
				if r.Name == resp.Name {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				respErrors = append(respErrors, *resp)
			}
		}
	}

	if len(params) > 0 || valueDTO != nil {
		stream.WriteMore()
		jParam := NewqInParam(in, valueDTO, route.Multipart, params...)
		jParam.WriteSwaggerParams(stream)
	}

	tags := make([]string, 0)
	if strings.Contains(route.Desc, "test handler") {
		tags = append(tags, "Test handlers (auto generated)")
	} else if strings.Contains(route.Desc, "table") {
		tags = append(tags, "CRUD")
	}

	if parts := regTitle.FindStringSubmatch(route.Desc); len(parts) > 0 {
		tags = append(tags, parts[1])
	}

	if parts := regTags.FindStringSubmatch(route.Desc); len(parts) > 1 {
		tags = append(tags, parts[1:]...)
	}

	if len(tags) > 0 {
		AddObjectToJSON(stream, "tags", tags)
	}
	writeResponses(stream, respErrors, route)
}

func convertParamForErrResp(ctx *fasthttp.RequestCtx, param InParam) InParam {
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
	param.Req = false

	return param
}

func writeSummary(stream *jsoniter.Stream, route *ApiRoute, summary string) {
	desc := regAbbr.ReplaceAllString(route.Desc, "$1")
	replacer := strings.NewReplacer("\n", " ", "*", "", "#", "", "[", " ", "]", " ")
	if len(route.Desc) < 255 {
		summary += replacer.Replace(desc)
	} else {
		summary += replacer.Replace(strings.TrimRightFunc(desc, func(r rune) bool {
			return r == ',' || r == '.' || r == ';' || r == ':' || r == '\n'
		}))
	}
	if route.Multipart {
		summary += ", multipart"
	}
	if route.NeedAuth {
		a := "JWT"
		if route.OnlyAdmin {
			summary += ", only admin access"
			a = "JWT,only admin"
		}
		AddObjectToJSON(stream, "security", []map[string][]string{{authSecurity: []string{a}}})
	}
	if route.OnlyLocal {
		summary += ", only local request"
	}

	AddFieldToJSON(stream, "summary", summary)
}

func getRespError(ctx *fasthttp.RequestCtx, value any) *InParam {
	r, ok := (value).(Visit)
	if ok {
		badParams, err := r.Result()
		if err != nil {
			return &InParam{
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
				return &InParam{
					Name:     name,
					Desc:     "JSON",
					Req:      true,
					DefValue: s,
				}
			}

		}
	}
	return nil
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
		return InParam{
			Name:              title,
			Desc:              sType,
			Req:               false,
			PartReq:           nil,
			Type:              &ReflectType{Type: value.Type()},
			DefValue:          value.String(),
			IncompatibleWiths: nil,
			TestValue:         "",
		}
	}
}

func WriteMap(value reflect.Value, stream *jsoniter.Stream, title string) any {
	// nil maps should be indicated as different than empty maps
	if value.IsNil() {
		return nil
	}

	keys := value.MapKeys()
	props := make([]any, 0)
	for i, v := range keys {
		if i > 0 {
			stream.WriteMore()
		}
		props = append(props, writeReflect(fmt.Sprintf("%d: %s %s `%s`", i, v.Kind(), v.Type(), v.String()), v, stream))
	}

	if len(props) > 0 {
		return NewSwaggerObject(props, title)
	}

	return nil
}

func WriteSlice(value reflect.Value, stream *jsoniter.Stream, title string) any {
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

	props := make([]any, 0)

	for i := 0; i < numEntries; i++ {
		v := value.Index(i)
		props = append(props, writeReflect(fmt.Sprintf("%d: %s %s `%s`", i, v.Kind(), v.Type(), v.String()), v, stream))
	}

	if len(props) > 0 {
		return NewSwaggerArray(title, props...)
	}

	return nil
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

	for name, tField := range list {
		val := value.FieldByName(name)
		tag := tField.Tag.Get("json")
		switch tag {
		case "":
			tag = tField.Name
		case "-":
			continue
		}

		if tField.Anonymous {
			kind := val.Kind()
			kind, val = indirect(kind, val)
			writeFields(val, stream, val.Type(), props, titles)
		} else {
			props[tag] = writeReflect(tag, val, stream)
		}
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
	title := "httpgo"
	if n, ok := apis.Ctx[ServerName].(string); ok && n > "" {
		title = n
	}
	version, hasVersion := apis.Ctx.Value(ServerVersion).(string)
	if !hasVersion {
		version, _ = apis.Ctx.Value(ApiVersion).(string)
	}
	AddFieldToJSON(stream, "version", version)
	AddFieldToJSON(stream, "title", title)
	stream.WriteMore()

	stream.WriteObjectField("license")
	stream.WriteObjectStart()
	FirstFieldToJSON(stream, "name", "Apache 2.0")
	AddFieldToJSON(stream, "url", "https://www.apache.org/licenses/LICENSE-2.0.html")

	stream.WriteObjectEnd()
	stream.WriteObjectEnd()

	if apis.fncAuth != nil {
		stream.WriteMore()
		if a, ok := apis.fncAuth.(*auth.AuthBearer); ok {
			WriteBearer(stream, authSecurity, a.String())
		} else {
			//todo: implements others auth
			WriteBearer(stream, authSecurity, apis.fncAuth.String())
		}
	}

	AddObjectToJSON(stream, "paths", apis.routes)
}

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
	"strings"

	"github.com/go-openapi/spec"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/logs"
)

type SwaggerUnit struct {
	Properties []spec.SchemaProps `json:"properties,omitempty"`
	Items      any                `json:"items,omitempty"`
	Type       string
}

type SwaggerParam map[string]any

func NewSwaggerObjectRoot(props any) SwaggerParam {
	return map[string]any{
		"name": "body",
		"in":   "body",
		"schema": map[string]any{
			"type":       "object",
			"properties": props,
		},
	}
}

func NewSwaggerObject(props any, name string) SwaggerParam {
	return map[string]any{
		//"items": map[string]any{
		"type":       "object",
		"properties": props,
		//},
	}
}

func NewSwaggerArray(desc string, props ...any) SwaggerParam {
	items := make([]spec.Items, len(props))
	for i, prop := range props {
		items[i] = spec.Items{SimpleSchema: spec.SimpleSchema{
			Default: prop,
		},
		}
	}
	logs.StatusLog(items)

	return map[string]any{
		"description": desc,
		"schema": SwaggerUnit{
			Type:  "array",
			Items: props[0],
		},
	}
}

func NewSwaggerParam(name, in, typ, f string) SwaggerParam {
	return map[string]any{
		"name": name,
		"in":   in,
		"schema": map[string]any{
			"type":   typ,
			"format": f,
		},
	}
}

func NewSwaggerArray1(props any, name string) SwaggerParam {
	mapC := NewSwaggerContent(map[string]any{
		"type":  "array",
		"items": props,
	})
	mapC["name"] = name
	mapC["in"] = "body"

	return mapC
}

func NewSwaggerContent(schema map[string]any) map[string]any {
	return map[string]any{
		//"content": map[string]any{
		//	"application/json": map[string]any{
		"schema": schema,
		//	},
		//},
	}
}

type ReflectType struct {
	reflect.Type
	value reflect.Value
	Props any
}

func NewReflectType(value any) *ReflectType {
	val := reflect.ValueOf(value)
	r := &ReflectType{Type: val.Type(), value: val}
	r.Props = r.convertValue(fmt.Sprintf("%v", value), val)

	return r
}

func (r ReflectType) CheckType(ctx *fasthttp.RequestCtx, value string) bool {
	//TODO implement me
	panic("implement me")
}

func (r ReflectType) ConvertValue(ctx *fasthttp.RequestCtx, value string) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (r ReflectType) ConvertSlice(ctx *fasthttp.RequestCtx, values []string) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (r ReflectType) IsSlice() bool {
	//TODO implement me
	panic("implement me")
}

func (r *ReflectType) convertValue(title string, value reflect.Value) any {

	kind := value.Kind()
	// Handle pointers specially.
	kind, value = Indirect(kind, value)
	defer func() {
		e := recover()
		err, ok := e.(error)
		if ok {
			logs.ErrorLog(err, kind.String(), value.String())
		} else if e != nil {
			logs.StatusLog(e)
		}
	}()
	if kind > reflect.UnsafePointer || kind <= 0 {
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

	return r.WriteReflectKind(kind, value, sType, title)
}

func (r *ReflectType) WriteReflectKind(kind reflect.Kind, value reflect.Value, sType, title string) any {
	switch kind {
	case reflect.Struct:
		return r.WriteStruct(value, title)

	case reflect.Map:
		return r.WriteMap(value, title)

	case reflect.Array, reflect.Slice:
		return r.WriteSlice(value, title)

	default:
		//return spec.Schema{SchemaProps: spec.SchemaProps{
		//	Schema:               "",
		//	Description:          "",
		//	Type:                 spec.StringOrArray{sType, value.Type().String()},
		//	Nullable:             true,
		//	FormatDoc:               "",
		//	Title:                title,
		//	Default:              kind.String(),
		//	Maximum:              nil,
		//	ExclusiveMaximum:     false,
		//	Minimum:              nil,
		//	ExclusiveMinimum:     false,
		//	MaxLength:            nil,
		//	MinLength:            nil,
		//	Pattern:              "",
		//	MaxItems:             nil,
		//	MinItems:             nil,
		//	UniqueItems:          false,
		//	MultipleOf:           nil,
		//	Enum:                 nil,
		//	MaxProperties:        nil,
		//	MinProperties:        nil,
		//	Required:             nil,
		//	Items:                nil,
		//	AllOf:                nil,
		//	OneOf:                nil,
		//	AnyOf:                nil,
		//	Not:                  nil,
		//	Properties:           nil,
		//	AdditionalProperties: nil,
		//	PatternProperties:    nil,
		//	Dependencies:         nil,
		//	AdditionalItems:      nil,
		//	Definitions:          nil,
		//},
		//}
		return InParam{
			Name:              title,
			Desc:              sType,
			Req:               false,
			PartReq:           nil,
			Type:              &ReflectType{Type: value.Type()},
			DefValue:          kind.String(),
			IncompatibleWiths: nil,
			TestValue:         "",
		}
	}
}

func (r *ReflectType) WriteMap(value reflect.Value, title string) any {
	// nil maps should be indicated as different than empty maps
	if value.IsNil() {
		logs.StatusLog(title, value)
		return nil
	}

	keys := value.MapKeys()
	propers := make([]any, 0)
	for i, v := range keys {
		propers = append(propers, r.convertValue(fmt.Sprintf("%d: %s %s `%s`", i, v.Kind(), v.Type(), v.String()), v))
	}

	return NewSwaggerObject(propers, title)
}

func (r *ReflectType) WriteSlice(value reflect.Value, title string) any {

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
				return NewSwaggerArray(title, r.WriteStruct(reflect.Zero(elem), title))

			default:
				return NewSwaggerArray(title, r.WriteReflectKind(kind, reflect.New(elem), vType.Elem().String(), title))
			}

		}
	}

	propers := make([]any, numEntries)
	for i := 0; i < numEntries; i++ {
		v := value.Index(i)

		propers[i] = r.convertValue(fmt.Sprintf("%s, %s", v.Kind(), v.Type()), v)
	}

	return NewSwaggerArray(title, propers...)
}

func (r *ReflectType) WriteStruct(value reflect.Value, title string) any {
	propers := make(map[string]any, 0)
	vType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		v := vType.Field(i)
		if !v.IsExported() {
			continue
		}

		val := value.Field(i)
		kind := val.Kind()
		kind, val = Indirect(kind, val)

		tag := v.Tag.Get("json")
		title := tag // fmt.Sprintf("%s: %s", v.Name, v.Type) + writeTag(v.Tag)
		if tag == "" {
			title = v.Name
		}

		propers[title] = r.convertValue(title, val)
	}

	return NewSwaggerObject(propers, title)
}

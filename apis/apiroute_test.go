/*
 * Copyright (c) 2023-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	// "bufio"
	// "go/types"
	// "net"
	// "sync"
	"encoding/json"
	"fmt"
	"go/types"
	"sync"
	"testing"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	// "github.com/json-iterator/go"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/dbEngine/dbEngine/psql"
	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/gotools/typesExt"
	"github.com/ruslanBik4/httpgo/auth"
	"github.com/ruslanBik4/httpgo/views"
)

type commCase string

type PRCommandParams struct {
	Command   commCase `json:"command"`
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	Account   int32    `json:"account"`
	LastQuery commCase `json:"last_query"`
}

// Implementing RouteDTO interface
func (prParams *PRCommandParams) GetValue() any {
	return prParams
}

func (prParams *PRCommandParams) NewValue() any {

	newVal := PRCommandParams{}

	return newVal

}

const jsonText = `{"account":7060246,"command":"adjustments","end_date":"2020-01-25","start_date":"2020-01-01"}`

var (
	route = &ApiRoute{
		Desc:   "test route",
		Method: POST,
		DTO:    &PRCommandParams{},
	}
)

func TestCheckAndRun(t *testing.T) {

	dto := route.DTO.NewValue()
	val := &dto
	// err := jsoniter.UnmarshalFromString(json, &val)

	// assert.Nil(t, err)

	// t.Logf("%+v", DTO)

	err := json.Unmarshal([]byte(jsonText), &val)

	assert.Nil(t, err)

	t.Logf("%+v", dto)
}

type testArgs struct {
	name string
	src  []byte
	col  types.BasicKind
	want string
}

var tests = []testArgs{
	{
		"string",
		[]byte("simple string"),
		types.String,
		`"simple string"`,
	},
	{
		"string",
		[]byte(`<html> \d\s`),
		types.String,
		`"\u003chtml> \\d\\s"`,
	},
	{
		"string",
		[]byte(`{"src": "<html> \d\s", "error": false, "code": 123}`),
		types.String,
		`"{\"src\": \"\u003chtml> \\d\\s\", \"error\": false, \"code\": 123}"`,
	},
	{
		"object",
		[]byte(`{"src": "<html> \d\s", "error": false, "code": 123}`),
		typesExt.TMap,
		`{"src": "<html> \d\s", "error": false, "code": 123}`,
	},
}

func TestWriteElemValue(t *testing.T) {

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := writeTestValue(tt)
			body := ctx.Response.Body()
			assert.Equal(t, tt.want, gotools.BytesToString(body))
		})
	}
}

func writeTestValue(tt testArgs) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	var col dbEngine.Column
	switch tt.col {
	case types.String:
		col = dbEngine.NewStringColumn(tt.name, "comment", false, 0)
	//	todo add other column types
	case typesExt.TMap:
		col = psql.NewColumn(nil, tt.name, "inet", nil, true,
			"", "comment", "inet", 0,
			false, false)
	}
	WriteElemValue(ctx, tt.src, col)
	return ctx
}

func BenchmarkWriteElemValue(b *testing.B) {
	b.ReportAllocs()
	for _, tt := range tests {
		b.ReportAllocs()
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			b.Run(tt.name, func(b *testing.B) {

				ctx := writeTestValue(tt)
				b.Logf("%s", ctx.Response.Body())
			})
		}
		b.ResetTimer()
		b.ReportAllocs()
	}
	b.ReportAllocs()

}

func TestAddFieldToJSON(t *testing.T) {
	type args struct {
		stream *jsoniter.Stream
		field  string
		s      string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddFieldToJSON(tt.args.stream, tt.args.field, tt.args.s)
		})
	}
}

func TestAddObjectToJSON(t *testing.T) {
	type args struct {
		stream *jsoniter.Stream
		field  string
		s      any
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddObjectToJSON(tt.args.stream, tt.args.field, tt.args.s)
		})
	}
}

func TestApiRoute_CheckAndRun(t *testing.T) {
	type fields struct {
		Desc        string
		DTO         RouteDTO
		Fnc         ApiRouteHandler
		FncAuth     auth.FncAuth
		TestFncAuth auth.FncAuth
		Method      tMethod
		Multipart   bool
		NeedAuth    bool
		OnlyAdmin   bool
		OnlyLocal   bool
		WithCors    bool
		Params      []InParam
		Resp        any
	}
	type args struct {
		ctx     *fasthttp.RequestCtx
		fncAuth auth.FncAuth
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp any
		wantErr  assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &ApiRoute{
				Desc:        tt.fields.Desc,
				DTO:         tt.fields.DTO,
				Fnc:         tt.fields.Fnc,
				FncAuth:     tt.fields.FncAuth,
				TestFncAuth: tt.fields.TestFncAuth,
				Method:      tt.fields.Method,
				Multipart:   tt.fields.Multipart,
				NeedAuth:    tt.fields.NeedAuth,
				OnlyAdmin:   tt.fields.OnlyAdmin,
				OnlyLocal:   tt.fields.OnlyLocal,
				WithCors:    tt.fields.WithCors,
				Params:      tt.fields.Params,
				Resp:        tt.fields.Resp,
			}
			gotResp, err := route.CheckAndRun(tt.args.ctx, tt.args.fncAuth)
			if !tt.wantErr(t, err, fmt.Sprintf("CheckAndRun(%v, %v)", tt.args.ctx, tt.args.fncAuth)) {
				return
			}
			assert.Equalf(t, tt.wantResp, gotResp, "CheckAndRun(%v, %v)", tt.args.ctx, tt.args.fncAuth)
		})
	}
}

func TestApiRoute_CheckParams(t *testing.T) {
	type fields struct {
		Desc        string
		DTO         RouteDTO
		Fnc         ApiRouteHandler
		FncAuth     auth.FncAuth
		TestFncAuth auth.FncAuth
		Method      tMethod
		Multipart   bool
		NeedAuth    bool
		OnlyAdmin   bool
		OnlyLocal   bool
		WithCors    bool
		Params      []InParam
		Resp        any
		lock        sync.RWMutex
	}
	type args struct {
		ctx       *fasthttp.RequestCtx
		badParams map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &ApiRoute{
				Desc:        tt.fields.Desc,
				DTO:         tt.fields.DTO,
				Fnc:         tt.fields.Fnc,
				FncAuth:     tt.fields.FncAuth,
				TestFncAuth: tt.fields.TestFncAuth,
				Method:      tt.fields.Method,
				Multipart:   tt.fields.Multipart,
				NeedAuth:    tt.fields.NeedAuth,
				OnlyAdmin:   tt.fields.OnlyAdmin,
				OnlyLocal:   tt.fields.OnlyLocal,
				WithCors:    tt.fields.WithCors,
				Params:      tt.fields.Params,
				Resp:        tt.fields.Resp,
			}
			assert.Equalf(t, tt.want, route.CheckParams(tt.args.ctx, tt.args.badParams), "CheckParams(%v, %v)", tt.args.ctx, tt.args.badParams)
		})
	}
}

func TestApiRoute_checkTypeParam(t *testing.T) {
	type fields struct {
		Desc        string
		DTO         RouteDTO
		Fnc         ApiRouteHandler
		FncAuth     auth.FncAuth
		TestFncAuth auth.FncAuth
		Method      tMethod
		Multipart   bool
		NeedAuth    bool
		OnlyAdmin   bool
		OnlyLocal   bool
		WithCors    bool
		Params      []InParam
		Resp        any
		lock        sync.RWMutex
	}
	type args struct {
		ctx    *fasthttp.RequestCtx
		name   string
		values []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &ApiRoute{
				Desc:        tt.fields.Desc,
				DTO:         tt.fields.DTO,
				Fnc:         tt.fields.Fnc,
				FncAuth:     tt.fields.FncAuth,
				TestFncAuth: tt.fields.TestFncAuth,
				Method:      tt.fields.Method,
				Multipart:   tt.fields.Multipart,
				NeedAuth:    tt.fields.NeedAuth,
				OnlyAdmin:   tt.fields.OnlyAdmin,
				OnlyLocal:   tt.fields.OnlyLocal,
				WithCors:    tt.fields.WithCors,
				Params:      tt.fields.Params,
				Resp:        tt.fields.Resp,
			}
			got, err := route.checkTypeAndConvertParam(tt.args.ctx, tt.args.name, tt.args.values)
			if !tt.wantErr(t, err, fmt.Sprintf("checkTypeAndConvertParam(%v, %v, %v)", tt.args.ctx, tt.args.name, tt.args.values)) {
				return
			}
			assert.Equalf(t, tt.want, got, "checkTypeAndConvertParam(%v, %v, %v)", tt.args.ctx, tt.args.name, tt.args.values)
		})
	}
}

func TestApiRoute_isValidMethod(t *testing.T) {
	type fields struct {
		Desc        string
		DTO         RouteDTO
		Fnc         ApiRouteHandler
		FncAuth     auth.FncAuth
		TestFncAuth auth.FncAuth
		Method      tMethod
		Multipart   bool
		NeedAuth    bool
		OnlyAdmin   bool
		OnlyLocal   bool
		WithCors    bool
		Params      []InParam
		Resp        any
		lock        sync.RWMutex
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &ApiRoute{
				Desc:        tt.fields.Desc,
				DTO:         tt.fields.DTO,
				Fnc:         tt.fields.Fnc,
				FncAuth:     tt.fields.FncAuth,
				TestFncAuth: tt.fields.TestFncAuth,
				Method:      tt.fields.Method,
				Multipart:   tt.fields.Multipart,
				NeedAuth:    tt.fields.NeedAuth,
				OnlyAdmin:   tt.fields.OnlyAdmin,
				OnlyLocal:   tt.fields.OnlyLocal,
				WithCors:    tt.fields.WithCors,
				Params:      tt.fields.Params,
				Resp:        tt.fields.Resp,
			}
			assert.Equalf(t, tt.want, route.isValidMethod(tt.args.ctx), "isValidMethod(%v)", tt.args.ctx)
		})
	}
}

func TestApiRoute_performsJSON(t *testing.T) {
	type fields struct {
		Desc        string
		DTO         RouteDTO
		Fnc         ApiRouteHandler
		FncAuth     auth.FncAuth
		TestFncAuth auth.FncAuth
		Method      tMethod
		Multipart   bool
		NeedAuth    bool
		OnlyAdmin   bool
		OnlyLocal   bool
		WithCors    bool
		Params      []InParam
		Resp        any
		lock        sync.RWMutex
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &ApiRoute{
				Desc:        tt.fields.Desc,
				DTO:         tt.fields.DTO,
				Fnc:         tt.fields.Fnc,
				FncAuth:     tt.fields.FncAuth,
				TestFncAuth: tt.fields.TestFncAuth,
				Method:      tt.fields.Method,
				Multipart:   tt.fields.Multipart,
				NeedAuth:    tt.fields.NeedAuth,
				OnlyAdmin:   tt.fields.OnlyAdmin,
				OnlyLocal:   tt.fields.OnlyLocal,
				WithCors:    tt.fields.WithCors,
				Params:      tt.fields.Params,
				Resp:        tt.fields.Resp,
			}
			got, err := route.performsJSON(tt.args.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("performsJSON(%v)", tt.args.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "performsJSON(%v)", tt.args.ctx)
		})
	}
}

func TestDTO(t *testing.T) {
	type args struct {
		dto RouteDTO
	}
	tests := []struct {
		name string
		args args
		want BuildRouteOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DTO(tt.args.dto), "DTO(%v)", tt.args.dto)
		})
	}
}

func TestFirstFieldToJSON(t *testing.T) {
	type args struct {
		stream *jsoniter.Stream
		field  string
		s      string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FirstFieldToJSON(tt.args.stream, tt.args.field, tt.args.s)
		})
	}
}

func TestFirstObjectToJSON(t *testing.T) {
	type args struct {
		stream *jsoniter.Stream
		field  string
		s      any
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FirstObjectToJSON(tt.args.stream, tt.args.field, tt.args.s)
		})
	}
}

func TestMapRoutes_AddRoutes(t *testing.T) {
	type args struct {
		routes ApiRoutes
	}
	tests := []struct {
		name           string
		r              MapRoutes
		args           args
		wantBadRouting []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantBadRouting, tt.r.AddRoutes(tt.args.routes), "AddRoutes(%v)", tt.args.routes)
		})
	}
}

func TestMapRoutes_GetRoute(t *testing.T) {
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name    string
		r       MapRoutes
		args    args
		want    *ApiRoute
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.GetRoute(tt.args.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("GetRoute(%v)", tt.args.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetRoute(%v)", tt.args.ctx)
		})
	}
}

func TestMapRoutes_GetTestRouteSuffix(t *testing.T) {
	type args struct {
		route *ApiRoute
	}
	tests := []struct {
		name string
		r    MapRoutes
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.r.GetTestRouteSuffix(tt.args.route), "GetTestRouteSuffix(%v)", tt.args.route)
		})
	}
}

func TestMapRoutes_findParentRoute(t *testing.T) {
	type args struct {
		method tMethod
		path   string
	}
	tests := []struct {
		name  string
		r     MapRoutes
		args  args
		want  *ApiRoute
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.r.findParentRoute(mapRoute{tt.args.method, tt.args.path})
			assert.Equalf(t, tt.want, got, "findParentRoute(%v, %v)", tt.args.method, tt.args.path)
			assert.Equalf(t, tt.want1, got1, "findParentRoute(%v, %v)", tt.args.method, tt.args.path)
		})
	}
}

func TestMultiPartForm(t *testing.T) {
	tests := []struct {
		name string
		want BuildRouteOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MultiPartForm(), "MultiPartForm()")
		})
	}
}

func TestNewAPIRoute(t *testing.T) {
	type args struct {
		desc     string
		method   tMethod
		params   []InParam
		needAuth bool
		fnc      ApiRouteHandler
		resp     any
		Options  []BuildRouteOptions
	}
	tests := []struct {
		name string
		args args
		want *ApiRoute
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, funcName, _, _ := getStringOfFnc(tt.args.fnc)
			assert.Equalf(t, tt.want,
				NewAPIRoute(tt.args.desc, tt.args.method, tt.args.params, tt.args.needAuth, tt.args.fnc, tt.args.resp, tt.args.Options...),
				"NewAPIRoute(%v, %v, %v, %v, %v, %v, %+v)",
				tt.args.desc, tt.args.method, tt.args.params, tt.args.needAuth, funcName, tt.args.resp, tt.args.Options)
		})
	}
}

func TestNewAPIRouteWithDBEngine(t *testing.T) {
	type args struct {
		desc      string
		method    tMethod
		needAuth  bool
		params    []InParam
		sqlOrName string
		Options   []BuildRouteOptions
	}
	tests := []struct {
		name string
		args args
		want *ApiRoute
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want,
				NewAPIRouteWithDBEngine(tt.args.desc, tt.args.method, tt.args.needAuth, tt.args.params, tt.args.sqlOrName, tt.args.Options...),
				"NewAPIRouteWithDBEngine(%v, %v, %v, %v, %v, %+v)",
				tt.args.desc, tt.args.method, tt.args.needAuth, tt.args.params, tt.args.sqlOrName, tt.args.Options)
		})
	}
}

func TestNewMapRoutes(t *testing.T) {
	tests := []struct {
		name string
		want MapRoutes
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewMapRoutes(), "NewMapRoutes()")
		})
	}
}

func TestNewSimpleGETRoute(t *testing.T) {
	type args struct {
		desc    string
		params  []InParam
		fnc     ApiSimpleHandler
		Options []BuildRouteOptions
	}
	tests := []struct {
		name string
		args args
		want *ApiRoute
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want,
				NewSimpleGETRoute(tt.args.desc, tt.args.params, tt.args.fnc, tt.args.Options...),
				"NewSimpleGETRoute(%v, %v, %v, %+v)", tt.args.desc, tt.args.params, tt.args.fnc, tt.args.Options)
		})
	}
}

func TestNewSimplePOSTRoute(t *testing.T) {
	type args struct {
		desc    string
		params  []InParam
		fnc     ApiSimpleHandler
		Options []BuildRouteOptions
	}
	tests := []struct {
		name string
		args args
		want *ApiRoute
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want,
				NewSimplePOSTRoute(tt.args.desc, tt.args.params, tt.args.fnc, tt.args.Options...),
				"NewSimplePOSTRoute(%v, %v, %v, %+v)", tt.args.desc, tt.args.params, tt.args.fnc, tt.args.Options)
		})
	}
}

func TestOnlyLocal(t *testing.T) {
	tests := []struct {
		name string
		want BuildRouteOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, OnlyLocal(), "OnlyLocal()")
		})
	}
}

func TestRouteAuth(t *testing.T) {
	type args struct {
		fncAuth auth.FncAuth
	}
	tests := []struct {
		name string
		args args
		want BuildRouteOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, RouteAuth(tt.args.fncAuth), "RouteAuth(%v)", tt.args.fncAuth)
		})
	}
}

func TestRouteNeedAuth(t *testing.T) {
	tests := []struct {
		name string
		want BuildRouteOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, RouteNeedAuth(), "RouteNeedAuth()")
		})
	}
}

func TestRouteOnlyAdmin(t *testing.T) {
	tests := []struct {
		name string
		want BuildRouteOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, RouteOnlyAdmin(), "RouteOnlyAdmin()")
		})
	}
}

func TestWriteElemValue1(t *testing.T) {
	type args struct {
		ctx *fasthttp.RequestCtx
		src []byte
		col dbEngine.Column
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteElemValue(tt.args.ctx, tt.args.src, tt.args.col)
		})
	}
}

func Test_apiRouteToJSON(t *testing.T) {
	type args struct {
		ptr    unsafe.Pointer
		stream *jsoniter.Stream
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiRouteToJSON(tt.args.ptr, tt.args.stream)
		})
	}
}

func Test_apiRoutesToJSON(t *testing.T) {
	type args struct {
		ptr    unsafe.Pointer
		stream *jsoniter.Stream
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiRoutesToJSON(tt.args.ptr, tt.args.stream)
		})
	}
}

func Test_getParentPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getParentPath(tt.args.path), "getParentPath(%v)", tt.args.path)
		})
	}
}

func Test_setCORSHeaders(t *testing.T) {
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			views.WriteCORSHeaders(tt.args.ctx)
		})
	}
}

func Test_writeArray(t *testing.T) {
	type args struct {
		ctx *fasthttp.RequestCtx
		src []byte
		col dbEngine.Column
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, writeArray(tt.args.ctx, tt.args.src, tt.args.col), fmt.Sprintf("writeArray(%v, %v, %v)", tt.args.ctx, tt.args.src, tt.args.col))
		})
	}
}

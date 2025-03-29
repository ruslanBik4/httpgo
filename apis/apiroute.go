/*
 * Copyright (c) 2022-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go/types"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/gotools/typesExt"
	"github.com/ruslanBik4/httpgo/auth"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/json"
	"github.com/ruslanBik4/logs"
)

// BuildRouteOptions implement 'Functional Option' pattern for ApiRoute settings
type BuildRouteOptions func(route *ApiRoute)

// RouteAuth set custom auth method on ApiRoute
func RouteAuth(fncAuth auth.FncAuth) BuildRouteOptions {
	return func(route *ApiRoute) {
		route.FncAuth = fncAuth
	}
}

// RouteNeedAuth set auth on ApiRoute
func RouteNeedAuth() BuildRouteOptions {
	return func(route *ApiRoute) {
		route.NeedAuth = true
	}
}

// RouteOnlyAdmin set admin access for ApiRoute
func RouteOnlyAdmin() BuildRouteOptions {
	return func(route *ApiRoute) {
		route.OnlyAdmin = true
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

type (
	ApiRouteHandler  func(ctx *fasthttp.RequestCtx) (any, error)
	ApiSimpleHandler func() (any, error)
	ApiRouteFuncAuth func(ctx *fasthttp.RequestCtx) error
)

type PreCache interface {
	Equal(ctx *fasthttp.RequestCtx) (any, bool)
	Save(*fasthttp.RequestCtx, any)
}

// ApiRoute implement endpoint info & handler on request
type ApiRoute struct {
	Desc           string                              `json:"descriptor"`
	DTO            RouteDTO                            `json:"DTO"`
	Fnc            ApiRouteHandler                     `json:"-"`
	FncAuth        auth.FncAuth                        `json:"-"`
	FncIsForbidden func(ctx *fasthttp.RequestCtx) bool `json:"-"`
	TestFncAuth    auth.FncAuth                        `json:"-"`
	Method         tMethod                             `json:"method,string"`
	Multipart      bool
	NeedAuth       bool
	OnlyAdmin      bool
	OnlyLocal      bool
	WithCors       bool
	IsAJAXRequest  bool
	IsServerEvents bool
	Params         []InParam `json:"parameters,omitempty"`
	PreCache       PreCache  `json:"-"`
	Resp           any       `json:"response,omitempty"`
}

// NewAPIRoute create customizing ApiRoute
func NewAPIRoute(desc string, method tMethod, params []InParam, needAuth bool, fnc ApiRouteHandler,
	resp any, Options ...BuildRouteOptions) *ApiRoute {
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

// NewSimplePOSTRoute create POST ApiRoute with minimal requirements
func NewSimplePOSTRoute(desc string, params []InParam, fnc ApiSimpleHandler,
	Options ...BuildRouteOptions) *ApiRoute {
	route := &ApiRoute{
		Desc: desc,
		Fnc: func(ctx *fasthttp.RequestCtx) (any, error) {
			return fnc()
		},
		Method: POST,
		Params: params,
	}

	for _, setOption := range Options {
		setOption(route)
	}

	return route
}

// NewSimpleGETRoute create GET ApiRoute with minimal requirements
func NewSimpleGETRoute(desc string, params []InParam, fnc ApiSimpleHandler,
	Options ...BuildRouteOptions) *ApiRoute {
	route := &ApiRoute{
		Desc: desc,
		Fnc: func(ctx *fasthttp.RequestCtx) (any, error) {
			return fnc()
		},
		Method: GET,
		Params: params,
	}

	for _, setOption := range Options {
		setOption(route)
	}

	return route
}

// NewAPIRouteWithDBEngine create customizing ApiRoute
func NewAPIRouteWithDBEngine(desc string, method tMethod, needAuth bool, params []InParam,
	sqlOrName string, Options ...BuildRouteOptions) *ApiRoute {

	route := &ApiRoute{
		Desc: desc,
		Fnc: func(ctx *fasthttp.RequestCtx) (resp any, err error) {
			DB, ok := ctx.UserValue(Database).(*dbEngine.DB)
			if !ok {
				return nil, dbEngine.ErrDBNotFound
			}

			args := make([]any, 0, len(params))
			for _, param := range params {
				p := ctx.UserValue(param.Name)
				if p != nil {
					args = append(args, p)
				}
			}

			views.WriteJSONHeaders(ctx)
			if strings.Index(sqlOrName, " ") < 0 {
				table, ok := DB.Tables[sqlOrName]
				if ok {
					sqlOrName = "select * from " + sqlOrName
					i, comma := 0, "  WHERE "
					for _, param := range params {
						p := ctx.UserValue(param.Name)
						col := table.FindColumn(param.Name)
						if p != nil {
							if col == nil {
								return nil, dbEngine.NewErrNotFoundColumn(table.Name(), param.Name)
							}

							i++
							sqlOrName += fmt.Sprintf("%s %s=$%d", comma, col.Name(), i)
							comma = " AND "
						}
					}
					logs.DebugLog(sqlOrName)
				} else if routine, ok := DB.Routines[sqlOrName]; ok {

					sqlOrName, args, err = routine.BuildSql(dbEngine.ArgsForSelect(args...))
					if err != nil {
						return nil, errors.Wrap(err, "routine.BuildSql")
					}

					if routine.ReturnType() != "record" {
						isFound := false
						err := DB.Conn.SelectAndPerformRaw(ctx,
							func(values [][]byte, columns []dbEngine.Column) error {

								col := columns[0]
								src := values[0]
								if strings.HasPrefix(col.Type(), "_") {
									err := writeArray(ctx, src, col)
									if err != nil {
										return err
									}

								} else {
									WriteElemValue(ctx, src, col)
								}
								isFound = true

								return nil
							},
							sqlOrName, args...)
						if err != nil {
							ctx.ResetBody()
							return nil, errors.Wrap(err, "SelectAndPerformRaw")
						}

						if !isFound {
							//	not found any row
							ctx.SetStatusCode(fasthttp.StatusNoContent)
						}

						return nil, nil
					}
				}
			}

			_, _ = ctx.WriteString("[")
			rowComma := ""
			err = DB.Conn.SelectAndPerformRaw(ctx,
				WriteRecordAsJSON(ctx, &rowComma),
				sqlOrName, args...)

			if err != nil {
				ctx.ResetBody()
				return nil, errors.Wrap(err, "SelectAndPerformRaw")
			}

			if rowComma == "" {
				//	not found any row
				ctx.SetStatusCode(fasthttp.StatusNoContent)
			} else {
				_, _ = ctx.WriteString("]")
			}

			return nil, nil
		},
		Method:   method,
		Params:   params,
		NeedAuth: needAuth,
	}

	for _, setOption := range Options {
		setOption(route)
	}

	return route
}

func WriteRecordAsJSON(ctx *fasthttp.RequestCtx, rowComma *string) func(values [][]byte, columns []dbEngine.Column) error {
	return func(values [][]byte, columns []dbEngine.Column) error {
		_, _ = ctx.WriteString(*rowComma + "{")
		*rowComma = ","
		comma := ""
		for i, col := range columns {
			_, _ = ctx.WriteString(comma + `"` + col.Name() + `":`)
			if strings.HasPrefix(col.Type(), "_") {
				src := values[i]
				err := writeArray(ctx, src, col)
				if err != nil {
					return err
				}

			} else {
				WriteElemValue(ctx, values[i], col)
			}
			comma = ","
		}
		_, _ = ctx.WriteString("}")

		return nil
	}
}

func writeArray(ctx *fasthttp.RequestCtx, src []byte, col dbEngine.Column) error {
	var arrayHeader pgtype.ArrayHeader
	rp, err := arrayHeader.DecodeBinary(nil, src)
	if err != nil {
		return err
	}

	_, _ = ctx.WriteString("[")
	comma := ""
	for i := int32(0); i < arrayHeader.Dimensions[0].Length; i++ {
		elemLen := int(int32(binary.BigEndian.Uint32(src[rp:])))
		rp += 4
		var elemSrc []byte
		if elemLen >= 0 {
			elemSrc = src[rp : rp+elemLen]
			rp += elemLen
		}
		_, _ = ctx.WriteString(comma)
		WriteElemValue(ctx, elemSrc, col)
		comma = ","
	}

	_, _ = ctx.WriteString("]")
	return nil
}

func WriteElemValue(ctx *fasthttp.RequestCtx, src []byte, col dbEngine.Column) {
	basicType := col.BasicType()
	if len(src) == 0 && basicType != types.String {
		_, _ = fmt.Fprint(ctx, "null")
		return
	}

	switch basicType {
	case types.Bool, types.UntypedBool:
		_, _ = fmt.Fprintf(ctx, "%v", src[0] == 't' || src[0] == 'T')

	case types.String, types.UnsafePointer:
		json.WriteByteAsString(ctx, src)
	case types.UntypedFloat:
		decoded := &pgtype.Numeric{}
		err := decoded.DecodeBinary(nil, src)
		if err != nil {
			logs.ErrorLog(err, "decode UntypedFloat")
			return
		}
		_, _ = fmt.Fprintf(ctx, "%sE%d", decoded.Int.String(), decoded.Exp)

	case types.Uint16, types.Byte:
		_, _ = fmt.Fprintf(ctx, "%d", binary.BigEndian.Uint16(src))
	case types.Int8, types.Int16:
		_, _ = fmt.Fprintf(ctx, "%d", int16(binary.BigEndian.Uint16(src)))
	case types.Uint32:
		_, _ = fmt.Fprintf(ctx, "%d", binary.BigEndian.Uint32(src))
	case types.Int32:
		_, _ = fmt.Fprintf(ctx, "%d", int32(binary.BigEndian.Uint32(src)))
	case types.Uint64:
		_, _ = fmt.Fprintf(ctx, "%d", binary.BigEndian.Uint64(src))
	case types.Int64:
		_, _ = fmt.Fprintf(ctx, "%d", int64(binary.BigEndian.Uint64(src)))
	case types.Float32:
		_, _ = fmt.Fprintf(ctx, "%f", math.Float32frombits(binary.BigEndian.Uint32(src)))
	case types.Float64:
		_, _ = fmt.Fprintf(ctx, "%f", math.Float64frombits(binary.BigEndian.Uint64(src)))
	case typesExt.TMap:
		_, _ = fmt.Fprintf(ctx, `%s`, src)
	case typesExt.TStruct:
		switch col.Type() {
		case "date", "timestamp", "timestamptz", "time":
			layout := "2006-01-02"
			if col.Type() != "date" {
				layout += " 15:04:05.999999999"
			}
			t, err := time.Parse(layout, gotools.BytesToString(src))
			if err != nil {
				microsecSinceY2K := int64(binary.BigEndian.Uint64(src))

				const (
					negativeInfinityMicrosecondOffset = -9223372036854775808
					infinityMicrosecondOffset         = 9223372036854775807
					microsecFromUnixEpochToY2K        = 946684800 * 1000000
				)

				switch microsecSinceY2K {
				case infinityMicrosecondOffset:
					_, _ = ctx.WriteString(`"Infinity"`)
				case negativeInfinityMicrosecondOffset:
					_, _ = ctx.WriteString(`"-Infinity"`)
				default:
					microsecSinceUnixEpoch := microsecFromUnixEpochToY2K + microsecSinceY2K
					t := time.Unix(microsecSinceUnixEpoch/1000000, (microsecSinceUnixEpoch%1000000)*1000).UTC()
					_, _ = ctx.WriteString(`"` + t.Format(layout) + `"`)
				}
			} else {
				_, _ = ctx.WriteString(`"` + t.Format(layout) + `"`)
			}
		default:
			_, _ = fmt.Fprintf(ctx, `"%s"`, src)
		}
	default:
		_, _ = fmt.Fprintf(ctx, `"%s"`, src)
	}
}

// CheckAndRun check & run route handler
func (route *ApiRoute) CheckAndRun(ctx *fasthttp.RequestCtx, fncAuth auth.FncAuth) (resp any, err error) {

	if route.WithCors && !route.Multipart {
		views.WriteCORSHeaders(ctx)
	}

	if route.IsServerEvents {
		views.WriteServerEventsHeaders(ctx)
	}
	// check auth is needed
	// owl func for auth
	if (route.FncAuth != nil) && !route.FncAuth.Auth(ctx) ||
		// only admin access
		(route.FncAuth == nil) && (route.OnlyAdmin && !fncAuth.AdminAuth(ctx) ||
			// access according to FncAuth if it needs
			!route.OnlyAdmin && route.NeedAuth && !fncAuth.Auth(ctx)) {
		return nil, ErrUnAuthorized
	}

	// check forbidden
	if route.FncIsForbidden != nil && route.FncIsForbidden(ctx) {
		return nil, ErrRouteForbidden
	}
	// compliance check local request is needed
	if route.OnlyLocal && isNotLocalRequest(ctx) {
		return nil, errRouteOnlyLocal
	}

	contentType := ctx.Request.Header.ContentType()
	if bytes.HasPrefix(contentType, []byte(ContentTypeJSON)) && (route.DTO != nil) {
		return route.performsJSON(ctx)
	}

	badParams := make(map[string]string, 0)

	if route.Multipart {
		// check multipart params¬
		if !bytes.HasPrefix(contentType, []byte(ContentTypeMultiPart)) {
			return nil, fasthttp.ErrNoMultipartForm
		}
		mf, err := ctx.Request.MultipartForm()
		if err != nil {
			if strings.Contains(err.Error(), "form size must be greater than 0") ||
				strings.Contains(err.Error(), "cannot read multipart/form-data body") {
				return ctx.Request.String(), ErrWrongParamsList
			}
			return nil, err
		}

		defer ctx.Request.RemoveMultipartFormFiles()

		ctx.SetUserValue(MultiPartParams, mf.Value)

		for key, value := range mf.Value {
			val, err := route.checkTypeAndConvertParam(ctx, key, value)
			if err != nil {
				if val != nil {
					logs.DebugLog(val)
				}

				badParams[key] = "wrong type " + strings.Join(value, ",") + err.Error()
			} else {
				ctx.SetUserValue(key, val)
			}
		}

		for key, files := range mf.File {
			ctx.SetUserValue(key, files)
		}
	} else {
		fncStoreArgs := func(k, v []byte) {
			key := gotools.BytesToString(k)
			val, err := route.checkTypeAndConvertParam(ctx, key, []string{gotools.BytesToString(v)})
			if err != nil {
				badParams[key] = fmt.Sprintf("has wrong type %v (%s)", val, err)
			} else {
				ctx.SetUserValue(key, val)
			}
		}

		if ctx.IsPost() {
			ctx.PostArgs().VisitAll(fncStoreArgs)
		}
		ctx.QueryArgs().VisitAll(fncStoreArgs)
	}

	if (len(badParams) > 0) || !route.CheckParams(ctx, badParams) {
		return badParams, ErrWrongParamsList
	}

	if route.DTO != nil {
		if dto, ok := route.DTO.NewValue().(CompoundDTO); ok {
			dto.ReadParams(ctx)
			ctx.SetUserValue(JSONParams, dto)
		}
	}

	if p := route.PreCache; p != nil {
		//if gotools.BytesToString(ctx.Request.Header.Peek("Cache-Control")) == "no-cache" {
		//	logs.StatusLog("no cache detected!")
		//	return nil, false
		//}
		// we have result on cache
		if res, ok := p.Equal(ctx); ok {
			return res, nil
		}
		defer func() {
			if err == nil && resp != nil {
				p.Save(ctx, resp)
			}
		}()
	}
	return route.Fnc(ctx)
}

func (route *ApiRoute) performsJSON(ctx *fasthttp.RequestCtx) (any, error) {
	badParams := make(map[string]string, 0)
	// check JSON parsing
	dto := route.DTO.NewValue()

	if r, ok := (dto).(Visit); ok {
		val, err := fastjson.ParseBytes(ctx.Request.Body())
		if err != nil {
			return nil, errors.Wrap(err, "ParseBytes")
		}

		val.GetObject().Visit(r.Each)
		dto, err = r.Result()
		switch err {
		case nil:
		case ErrWrongParamsList:
			return dto, err
		default:
			return nil, errors.Wrap(err, "visit result")
		}
	} else if err := jsoniter.Unmarshal(ctx.Request.Body(), &dto); err != nil {
		errMsg := err.Error()
		parts := strings.Split(errMsg, ":")
		if len(parts) > 1 {
			param := strings.Split(parts[0], ".")
			badParams[param[len(param)-1]] = strings.Join(parts[1:], ":")
		} else {
			badParams["bad_params"] = "json DTO not parse :" + errMsg
		}

		return badParams, ErrWrongParamsList
	}

	if d, ok := dto.(CheckDTO); (ok && !d.CheckParams(ctx, badParams)) || !route.CheckParams(ctx, badParams) {
		return badParams, ErrWrongParamsList
	}

	ctx.SetUserValue(JSONParams, dto)

	return route.Fnc(ctx)
}

// CheckParams check param of request
func (route *ApiRoute) CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool {
	for _, param := range route.Params {
		param.Check(ctx, badParams)
	}

	return len(badParams) == 0
}

func (route *ApiRoute) checkTypeAndConvertParam(ctx *fasthttp.RequestCtx, name string, values []string) (any, error) {
	// find param in InParams list & convert according to Type
	for _, param := range route.Params {
		if param.Name == name {

			if param.Type == nil {
				return values, nil
			}

			if param.Type.IsSlice() {
				slice, err := param.Type.ConvertSlice(ctx, values)
				logs.DebugLog(slice)
				return slice, err
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

const desc = "description"

func writeResponseForAuth(stream *jsoniter.Stream) {
	stream.WriteMore()
	writeResponse(stream, fasthttp.StatusUnauthorized, nil)
	stream.WriteMore()
	writeResponse(stream, fasthttp.StatusForbidden, nil)
}

func writeResponses(stream *jsoniter.Stream, respErrors []InParam, route *ApiRoute) {
	stream.WriteMore()
	stream.WriteObjectField("responses")
	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	writeResponse(stream, fasthttp.StatusOK, route.Resp)

	writeBadRequest(stream, respErrors, route)
	if route.NeedAuth {
		writeResponseForAuth(stream)
	}

	if resp, ok := route.Resp.(SwaggerParam); ok {
		for code, r := range resp {
			stream.WriteMore()
			if s, ok := r.(string); ok {
				stream.WriteObjectField(code)
				stream.WriteObjectStart()
				FirstFieldToJSON(stream, "description", s)
				stream.WriteObjectEnd()
			} else {
				props := NewReflectType(r)
				FirstObjectToJSON(stream, code, props.Props)
			}
		}
		return
	}
}

func writeBadRequest(stream *jsoniter.Stream, respErrors []InParam, route *ApiRoute) {
	if len(respErrors) > 0 {
		stream.WriteMore()
		//writeResponse(stream, fasthttp.StatusBadRequest, resp)
		stream.WriteObjectField("400")
		stream.WriteObjectStart()
		defer stream.WriteObjectEnd()
		FirstFieldToJSON(stream, "description", statusMsg(fasthttp.StatusBadRequest))
		stream.WriteMore()
		stream.WriteObjectField("content")
		stream.WriteObjectStart()
		defer stream.WriteObjectEnd()

		stream.WriteObjectField("application/json")
		stream.WriteObjectStart()
		defer stream.WriteObjectEnd()

		stream.WriteObjectField("schema")
		stream.WriteObjectStart()
		defer stream.WriteObjectEnd()
		FirstObjectToJSON(stream, "type", "object")

		stream.WriteMore()
		jParam := NewqInParam("body", nil, route.Multipart, respErrors...)
		jParam.WriteSwaggerProperties(stream)
	}
}

func writeResponse(stream *jsoniter.Stream, status int, resp any) {
	stream.WriteObjectField(strconv.Itoa(status))
	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()
	switch resp := resp.(type) {
	case nil:
		FirstFieldToJSON(stream, desc, statusMsg(status))
	case string:
		FirstFieldToJSON(stream, desc, resp)
	default:
		FirstObjectToJSON(stream, desc, resp)
	}
}

func statusMsg(status int) string {
	return fasthttp.StatusMessage(status)
}

// ApiRoutes is hair of APIRoute
type ApiRoutes map[string]*ApiRoute

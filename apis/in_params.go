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
	"runtime"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/logs"
)

type DefValueHeader struct {
	header   string
	defValue string
}

func NewDefValueHeader(header string, defValue string) DefValueHeader {
	return DefValueHeader{header: header, defValue: defValue}
}

func (d DefValueHeader) ConvertValue(ctx *fasthttp.RequestCtx) string {
	if ctx != nil {
		if cL := ctx.Request.Header.Peek(d.header); len(cL) > 0 {
			return gotools.BytesToString(cL)
		}
	}
	return d.defValue
}

func (d DefValueHeader) Expect() string {
	return fmt.Sprintf("string on header '%s'", d.header)
}

func (d DefValueHeader) FormatDoc() string {
	return "string"
}

func (d DefValueHeader) RequestType() string {
	return "string"
}

type DefValueCalcFnc func(ctx *fasthttp.RequestCtx) any

func (d DefValueCalcFnc) Expect() string {
	return "function func(ctx *fasthttp.RequestCtx) any"
}

func (d DefValueCalcFnc) FormatDoc() string {
	return "function func(ctx *fasthttp.RequestCtx) any"
}

func (d DefValueCalcFnc) RequestType() string {
	return "object"
}

// InParam implement params on request
type InParam struct {
	Name              string
	Desc              string
	Req               bool
	PartReq           []string
	Type              APIRouteParamsType
	DefValue          any
	IncompatibleWiths []string
	TestValue         string
}

func GetValue[T any](ctx *fasthttp.RequestCtx, param *InParam) T {
	return ctx.UserValue(param.Name).(T)
}
func (param *InParam) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		_, _ = fmt.Fprintf(s, `Name: "%s",  Desc: %q, Type: %v,`,
			param.Name,
			param.Desc,
			param.Type,
		)
		if len(param.PartReq) > 0 {
			_, _ = fmt.Fprintf(s, "PartReq: %v,", param.PartReq)
		}
		if len(param.IncompatibleWiths) > 0 {
			_, _ = fmt.Fprintf(s, "IncompatibleWiths: %v,", param.IncompatibleWiths)
		}
		if param.DefValue != nil {
			_, _ = fmt.Fprintf(s, "DefValue: %q,", param.DefValue)
		}
		if param.TestValue > "" {
			_, _ = fmt.Fprintf(s, "TestValue: %v,", param.TestValue)
		}
		if param.Req {
			_, _ = fmt.Fprintf(s, "Req: %v,", param.Req)
		}

	case 'g':
		_, _ = fmt.Fprintf(s,
			`{
				Name: "%s",
				Desc: %q,
				Type: %g,`,
			param.Name,
			param.Desc,
			param.Type,
		)
		if len(param.PartReq) > 0 {
			_, _ = fmt.Fprintf(s, `
				PartReq: %v,`,
				param.PartReq)
		}
		if len(param.IncompatibleWiths) > 0 {
			_, _ = fmt.Fprintf(s, "\r\t\t\t\t\t\tIncompatibleWiths: %v,", param.IncompatibleWiths)
		}
		if param.DefValue != nil {
			switch p := param.Type.(type) {
			case TypeInParam:
				if p.BasicKind == types.String {
					_, _ = fmt.Fprintf(s, "\r\t\t\t\t\t\tDefValue: %q,", param.DefValue)
				} else if d, ok := param.DefValue.(string); !ok || !strings.HasPrefix(d, "NULL") {
					_, _ = fmt.Fprintf(s, "\r\t\t\t\t\t\tDefValue: %s(%v),", types.Typ[p.BasicKind], param.DefValue)

				}
			}
		}
		if param.TestValue > "" {
			_, _ = fmt.Fprintf(s, "\r\t\t\t\t\t\tTestValue: %v,", param.TestValue)
		}
		if param.Req {
			_, _ = fmt.Fprintf(s, `
				Req:   %v,`, param.Req)
		}
		_, _ = fmt.Fprintf(s, "\n\t\t\t}")
	default:
		_, _ = s.Write(gotools.StringToBytes(fmt.Sprintf(`Name: "%s",  Desc: %q, Type: %g, Req: %v, DefValue: %q`,
			param.Name, param.Desc, param.Type, param.Req, param.DefValue)),
		)
	}
}

func (param *InParam) isPartReq() bool {
	return len(param.PartReq) > 0
}

func (param *InParam) WithNotRequired() *InParam {
	ret := new(InParam)
	*ret = *param
	ret.Req = false
	return ret
}

// Check params of ctx
func (param *InParam) Check(ctx *fasthttp.RequestCtx, badParams map[string]string) {
	value := ctx.UserValue(param.Name)
	if value == nil {
		// param is part of group required params
		if param.presentOtherRegParam(ctx) {
			return
		}

		value = param.defaultValueOfParams(ctx, badParams)
		//  not present required param
		if value != nil {
			ctx.SetUserValue(param.Name, value)
		} else if param.Req {
			badParams[param.Name] = PARAM_REQUIRED
		}
	} else if name, val := param.isHasIncompatibleParams(ctx); name > "" {
		// has present param which not compatible with 'param'
		badParams[param.Name] = fmt.Sprintf("incompatible params: %s=%s & %s=%s", param.Name, value, name, val)
	}
}

// found params incompatible with 'param'
func (param InParam) isHasIncompatibleParams(ctx *fasthttp.RequestCtx) (string, any) {
	for _, name := range param.IncompatibleWiths {
		val := ctx.FormValue(name)
		if len(val) > 0 {
			return name, val
		}
	}

	return "", nil
}

// check 'param' is one part of list required params AND one of other params is present
func (param InParam) presentOtherRegParam(ctx *fasthttp.RequestCtx) bool {
	// Looking for parameters associated with the original 'param'
	for _, name := range param.PartReq {
		// param 'name' is present
		if ctx.UserValue(name) != nil {
			return true
		}
	}

	return false
}

// defaultValueOfParams return value as default for param, it is only for single required param
func (param *InParam) defaultValueOfParams(ctx *fasthttp.RequestCtx, badParams map[string]string) any {
	switch def := param.DefValue.(type) {
	case DefValueHeader:
		value, err := param.Type.ConvertValue(ctx, def.ConvertValue(ctx))
		if err != nil {
			return nil
		}
		return value

	case DefValueCalcFnc:
		if ctx != nil {
			return def(ctx)
		}

		fnc := runtime.FuncForPC(reflect.ValueOf(def).Pointer())
		fName, line := fnc.FileLine(0)
		return fmt.Sprintf("%s:%d %s()", fName, line, getLastSegment(fnc.Name()))

	case ApisValues:
		if ctx != nil {
			key := string(def)
			value, ok := ctx.UserValue(key).(string)
			if !ok {
				return ctx.UserValue(key)
			}

			val, err := param.Type.ConvertValue(ctx, value)
			if err != nil {
				badParams[param.Name] = "wrong type, expected " + param.Type.String() + err.Error()
				logs.ErrorLog(err, "ConvertValue")
				return ctx.UserValue(key)
			}

			return val
		}

		return def

	default:
		return param.DefValue
	}
}

// inParamToJSON produces a human-friendly description of Apis.
// Based on real data of the executable application, does not require additional documentation.
func inParamToJSON(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	// todo: add description of the test-based return data
	param := (*InParam)(ptr)
	stream.WriteObjectStart()
	defer stream.WriteObjectEnd()

	FirstFieldToJSON(stream, "name", param.Name)
	AddFieldToJSON(stream, "description", param.Desc)

	if param.Type != nil {
		if t, ok := param.Type.(jsoniter.ValEncoder); ok {
			stream.WriteMore()
			t.Encode(unsafe.Pointer(&t), stream)
		} else {
			AddFieldToJSON(stream, "format", "formdata")
			t, ok := (param.Type).(TypeInParam)
			s := param.Type.String()
			if ok {
				switch {
				case t.BasicKind > types.Bool && t.BasicKind < types.Float32:
					AddFieldToJSON(stream, "type", "integer")
				case t.BasicKind == types.String:
					AddFieldToJSON(stream, "type", "string")
				case t.BasicKind > types.UnsafePointer:
					AddFieldToJSON(stream, "type", "untyped")
				//case t.BasicKind == typesExt.TStruct:
				//	AddFieldToJSON(stream, "type", t.TypeString(nil, 0))
				default:
					AddFieldToJSON(stream, "type", s)
				}
			} else {
				AddFieldToJSON(stream, "type", s)
			}
		}
	}

	if param.Req {
		if len(param.PartReq) > 0 {
			s := strings.Join(param.PartReq, ", ")
			AddFieldToJSON(stream, "required if one of", "{"+s+" and "+param.Name+"}")
		} else {
			AddObjectToJSON(stream, "required", true)
		}
	}

	if param.DefValue != nil {
		AddObjectToJSON(stream, "default", param.defaultValueOfParams(nil, nil))
	}

	if len(param.IncompatibleWiths) > 0 {
		s := strings.Join(param.IncompatibleWiths, ", ")
		AddFieldToJSON(stream, "IncompatibleWith", "only one of {"+s+" and "+param.Name+"} may use for request")
	}

}

func init() {
	jsoniter.RegisterTypeEncoderFunc("apis.InParam", inParamToJSON, func(pointer unsafe.Pointer) bool {
		return false
	})
}

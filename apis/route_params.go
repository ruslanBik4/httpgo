// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"fmt"
	"go/types"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"
)

type DefValueCalcFnc = func(ctx *fasthttp.RequestCtx) interface{}

// InParam implement params on request
type InParam struct {
	Name              string
	Desc              string
	Req               bool
	PartReq           []string
	Type              APIRouteParamsType
	DefValue          interface{}
	IncompatibleWiths []string
	TestValue         string
}

func (param *InParam) isPartReq() bool {
	return len(param.PartReq) > 0
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
func (param InParam) isHasIncompatibleParams(ctx *fasthttp.RequestCtx) (string, interface{}) {
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
func (param *InParam) defaultValueOfParams(ctx *fasthttp.RequestCtx, badParams map[string]string) interface{} {
	switch def := param.DefValue.(type) {
	case DefValueCalcFnc:
		if ctx != nil {
			return def(ctx)
		}

		fnc := runtime.FuncForPC(reflect.ValueOf(def).Pointer())
		fName, line := fnc.FileLine(0)
		return fmt.Sprintf("%s:%d %s()", fName, line, getLastSegment(fnc.Name()))

	case ApisValues:
		if ctx != nil {
			value, ok := ctx.UserValue(string(def)).(string)
			if !ok {
				return ctx.UserValue(string(def))
			}

			val, err := param.Type.ConvertValue(ctx, value)
			if err != nil {
				badParams[param.Name] = "wrong type, except " + param.Type.String() + err.Error()
				logs.ErrorLog(err, "ConvertValue")
				return ctx.UserValue(string(def))
			}

			return val
		}

		return def

	default:
		return param.DefValue
	}
}

// inParamToJSON produces a human-friendly description of Apis.
//Based on real data of the executable application, does not require additional documentation.
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
			t, ok := (param.Type).(TypeInParam)
			if ok {
				AddFieldToJSON(stream, "format", param.Type.String())
				switch {
				case t.BasicKind > types.Bool && t.BasicKind < types.Float32:
					AddFieldToJSON(stream, "type", "integer")
				case t.BasicKind == types.String:
					AddFieldToJSON(stream, "type", "string")
				case t.BasicKind > types.UnsafePointer:
					AddFieldToJSON(stream, "type", "untyped")
				default:
					AddFieldToJSON(stream, "type", param.Type.String())
				}
			} else {
				AddFieldToJSON(stream, "format", param.Type.String())
				AddFieldToJSON(stream, "type", param.Type.String())

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

/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
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
	"github.com/ruslanBik4/gotools/typesExt"
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

// GetValue for getting server value as its type
func GetValue[T any](ctx *fasthttp.RequestCtx, param *InParam) T {
	raw := ctx.UserValue(param.Name)
	v, ok := raw.(T)
	if ok {
		return v
	}

	// log the actual value & concrete type that failed the assertion - the
	// old "%#v", v here always logged T's own zero value (e.g. "" or 0),
	// which is useless for diagnosing a mismatch.
	logs.DebugLog("param %q: expected %T, got %#v (%T)", param.Name, v, raw, raw)
	return v
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
		const caret = "\r\t\t\t\t\t\t"
		_, _ = fmt.Fprintf(s,
			`{%sName: "%s",%[1]sDesc: %[3]q,%[1]sType: %[4]g,`,
			caret,
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
			_, _ = fmt.Fprintf(s, "%sIncompatibleWiths: %v,", caret, param.IncompatibleWiths)
		}
		if param.DefValue != nil {
			_, _ = fmt.Fprintf(s, "%sDefValue: ", caret)
			formatDefValueGo(s, param)
		}
		if param.TestValue > "" {
			_, _ = fmt.Fprintf(s, "%sTestValue: %v,", caret, param.TestValue)
		}
		if param.Req {
			_, _ = fmt.Fprintf(s, `%sReq: %v,`, caret, param.Req)
		}
		_, _ = fmt.Fprintf(s, "\n\t\t\t}")
	default:
		_, _ = fmt.Fprintf(s, `%cName: "%s",  Desc: %q, Type: %g, Req: %v, DefValue: %q`, verb,
			param.Name, param.Desc, param.Type, param.Req, param.DefValue,
		)
	}
}

// formatDefValueGo writes a Go source expression for param.DefValue, for
// embedding as the DefValue field of a generated apis.InParam struct
// literal. It must ALWAYS write a syntactically valid expression - the
// caller has already written the unconditional "DefValue: " field name, so
// writing nothing (as the previous switch-with-no-default did for any
// param.Type that wasn't TypeInParam, e.g. every StructInParam - dates,
// uuid, hstore, ranges, enums...) leaves a dangling "DefValue: " in the
// generated file with no value before the next field, which fails to
// compile.
func formatDefValueGo(s fmt.State, param *InParam) {
	p, ok := param.Type.(TypeInParam)
	if !ok {
		// APIRouteParamsType implementation we don't know how to build a
		// default-value expression for at all (not even the fallback below,
		// which is TypeInParam-specific) - drop the default rather than
		// emit broken source, and say so loudly at generation time.
		logs.DebugLog("param %q: no DefValue codegen support for %T (value %#v) - emitting nil",
			param.Name, param.Type, param.DefValue)
		_, _ = fmt.Fprint(s, "nil, // TODO: DefValue dropped, no codegen support for this param type")
		return
	}

	formatBasicDefValueGo(s, p, param.DefValue)
}

// formatBasicDefValueGo handles TypeInParam - this covers both plain scalar
// params (NewTypeInParam) and struct-wrapped ones (NewStructInParam, whose
// BasicKind is one of typesExt's extended kinds - TStruct/TMap/TArray/TAny -
// layered on top of go/types.BasicKind's numeric range: both constructors
// produce a TypeInParam, they just set BasicKind differently).
func formatBasicDefValueGo(s fmt.State, p TypeInParam, defValue any) {
	if isNullDefault(defValue) {
		// was: fmt.Fprintf(s, "(%t)(nil),", param.Type) - %t is the fmt
		// *boolean* verb; param.Type is never a bool, so this always
		// produced the literal garbage text "%!t(apis.TypeInParam=...)"
		// straight into the generated .go file.
		//
		// The follow-up fix (StringTypeKinds(p.BasicKind) unconditionally)
		// was ALSO wrong for any extended kind: typesExt.StringTypeKinds
		// returns a descriptive word for those, e.g. "struct" for
		// typesExt.TStruct, and "*struct" is not a valid Go type - it
		// produced exactly that broken literal for a NewStructInParam
		// default (e.g. a DateString column with a NULL default). Only cast
		// to a concrete pointer type for the plain go/types.BasicKind
		// values that actually have one; anything else falls back to a bare
		// nil, which is also the more correct default here since there's no
		// single real Go type to hand back a typed nil *of* for an
		// arbitrary wrapped struct.
		if goType, ok := basicGoTypeName(p.BasicKind); ok {
			_, _ = fmt.Fprintf(s, "(*%s)(nil),", goType)
		} else {
			_, _ = fmt.Fprint(s, "nil,")
		}
		return
	}

	if p.BasicKind == types.String {
		_, _ = fmt.Fprintf(s, "%q,", defValue)
		return
	}

	d, ok := defValue.(string)
	if !ok {
		_, _ = fmt.Fprintf(s, "%s(%v),", typesExt.StringTypeKinds(p.BasicKind), defValue)
		return
	}

	_, _ = fmt.Fprintf(s, "%s(%s),", typesExt.StringTypeKinds(p.BasicKind), d)
}

// basicGoTypeName returns the literal Go type-name spelling for a plain
// go/types.BasicKind (Bool, Int*, Uint*, Float*, String, Complex*) suitable
// for embedding in generated source as e.g. "(*<name>)(nil)". It reports
// false for anything typesExt overlays on top of BasicKind's numeric range
// (TStruct, TMap, TArray, TAny) - those describe a wrapped Go type that
// varies per column, not one fixed name, so there's no single valid cast to
// write here.
func basicGoTypeName(kind types.BasicKind) (string, bool) {
	switch kind {
	case types.Bool,
		types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
		types.Float32, types.Float64,
		types.Complex64, types.Complex128,
		types.String:
		return typesExt.StringTypeKinds(kind), true
	default:
		return "", false
	}
}

// isNullDefault reports whether defValue is catalog text that should have
// been resolved to a real nil already. Column.SetDefault (dbEngine/psql)
// already converts a plain "NULL" default to a genuine nil before this code
// ever sees it - EXCEPT for a parenthesized SQL-expression default such as
// "(NULL::text)", which SetDefault deliberately leaves un-stripped. This
// only recognizes that shape: a bare "NULL", optionally wrapped in one
// "(...)" and/or carrying a "::type" cast. The old check,
// strings.Contains(d, "NULL"), was far too broad - it would also fire on
// any ordinary non-null string default that merely contains the substring
// "NULL" (e.g. a status value like 'NULLABLE_FIELD'), silently turning a
// real default into nil.
func isNullDefault(defValue any) bool {
	d, ok := defValue.(string)
	if !ok {
		return false
	}

	d = strings.TrimSpace(d)
	d = strings.TrimSuffix(strings.TrimPrefix(d, "("), ")")
	d = strings.SplitN(d, "::", 2)[0]

	return strings.EqualFold(strings.TrimSpace(d), "NULL")
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
			key := (def)
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
			// was: unsafe.Pointer(&t) - &t is the address of the local
			// jsoniter.ValEncoder *interface variable* (a 2-word
			// type-descriptor+data header), not the address of the
			// concrete value it holds. Reinterpreting that header's
			// address as if it pointed at the concrete struct's first
			// field is undefined behaviour: it reads whatever bytes happen
			// to sit at that offset in the interface header rather than
			// the actual param.Type data, so Encode silently serializes
			// garbage instead of erroring. reflect.ValueOf(param.Type)
			// gets the pointer to the real underlying data (boxing it via
			// reflect.New first if param.Type's concrete type is held by
			// value rather than by pointer).
			rv := reflect.ValueOf(param.Type)
			if rv.Kind() != reflect.Ptr {
				boxed := reflect.New(rv.Type())
				boxed.Elem().Set(rv)
				rv = boxed
			}
			t.Encode(unsafe.Pointer(rv.Pointer()), stream)
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

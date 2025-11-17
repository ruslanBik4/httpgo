/*
 * Copyright (c) 2022-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
	"strconv"
	"strings"
	"unsafe"

	"github.com/jackc/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"github.com/valyala/quicktemplate"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ruslanBik4/gotools"
	. "github.com/ruslanBik4/httpgo/views/templates/json"
	"github.com/ruslanBik4/logs"

	"github.com/ruslanBik4/gotools/typesExt"
)

type ApisValues string

// APIRouteParamsType encapsulates types operation for apis parameters`
type APIRouteParamsType interface {
	fmt.Stringer
	CheckType(ctx *fasthttp.RequestCtx, value string) bool
	ConvertValue(ctx *fasthttp.RequestCtx, value string) (any, error)
	ConvertSlice(ctx *fasthttp.RequestCtx, values []string) (any, error)
	IsSlice() bool
}

// TypeInParam has type definition of params ApiRoute
type TypeInParam struct {
	types.BasicKind
	isSlice bool
	DTO     RouteDTO
}

// NewTypeInParam create TypeInParam
func NewTypeInParam(bk types.BasicKind) TypeInParam {

	return TypeInParam{
		BasicKind: bk,
	}
}

// NewStructInParam create TypeInParam for struct
func NewStructInParam(dto RouteDTO) TypeInParam {

	return TypeInParam{
		BasicKind: typesExt.TStruct,
		DTO:       dto,
	}
}

// NewSliceTypeInParam create TypeInParam for slice
func NewSliceTypeInParam(bk types.BasicKind) TypeInParam {
	return TypeInParam{bk, true, nil}
}

// CheckType check of value computable with the TypeInParam
func (t TypeInParam) CheckType(ctx *fasthttp.RequestCtx, value string) bool {
	switch t.BasicKind {
	case types.String:
		return true

	case types.Bool:
		value = strings.ToLower(value)
		return value == "true" || value == "false"

	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		_, err := strconv.ParseInt(value, 10, 64)
		return err == nil

	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		_, err := strconv.ParseUint(value, 10, 64)
		return err == nil

	case types.Float32, types.Float64:
		_, err := strconv.ParseFloat(value, 64)
		return err == nil

	case typesExt.TStruct:
		v := t.DTO.NewValue()
		err := Json.UnmarshalFromString(value, &v)
		if err != nil {
			logs.ErrorLog(err)
		}

		return err == nil

	default:
		return true
	}
}

// ConvertValue convert value according to TypeInParam's type
func (t TypeInParam) ConvertValue(ctx *fasthttp.RequestCtx, value string) (any, error) {
	switch t.BasicKind {
	case types.String:
		return value, nil

	case types.Bool:
		return strings.ToLower(value) == "true", nil

	case types.Int:
		return strconv.Atoi(value)

	case types.Int8:
		p, err := strconv.ParseInt(value, 10, 8)
		return int8(p), err

	case types.Int16:
		p, err := strconv.ParseInt(value, 10, 16)
		return int16(p), err
	case types.Int32:
		p, err := strconv.ParseInt(value, 10, 32)
		return int32(p), err

	case types.Int64:
		return strconv.ParseInt(value, 10, 64)

	case types.Uint8:
		p, err := strconv.ParseUint(value, 10, 8)
		return uint8(p), err

	case types.Uint16:
		p, err := strconv.ParseUint(value, 10, 16)
		return uint16(p), err

	case types.Uint32:
		p, err := strconv.ParseUint(value, 10, 32)
		return uint32(p), err

	case types.Uint64:
		return strconv.ParseUint(value, 10, 64)

	// 	check type convert float64
	case types.Float32, types.Float64, types.UntypedFloat:
		return strconv.ParseFloat(value, 64)

	case types.UnsafePointer:
		return nil, nil

	case typesExt.TMap:
		res := make(map[string]any, 0)
		err := Json.UnmarshalFromString(value, &res)
		if err != nil {
			return nil, errors.Wrap(err, "UnmarshalFromString")
		}
		return res, nil

	case typesExt.TArray:
		res := make([]any, 0)
		return t.ReadValue(value, res)

	case typesExt.TStruct:
		if t.DTO != nil {
			return t.ReadValue(value, t.DTO.NewValue())
		}

		return nil, errors.Wrapf(ErrWrongParamsList, "convert this type (%s) need DTO", t.String())

	default:
		return nil, errors.Wrapf(ErrWrongParamsList, "convert this type (%s) not implement", t.String())
	}
}

func (t TypeInParam) ReadValue(s string, res any) (any, error) {
	switch res := res.(type) {
	case pgtype.Value:
		err := res.Set(s)
		if err != nil {
			return nil, errors.Wrap(err, "Set s")
		}
		return res.Get(), nil

	case json.Unmarshaler:
		err := res.UnmarshalJSON(gotools.StringToBytes(s))
		if err != nil {
			return nil, errors.Wrap(err, "Unmarshal ")
		}
		return res, nil

	case Visit:
		val, err := fastjson.Parse(s)
		if err != nil {
			return nil, errors.Wrapf(err, "Parse '%s'", s)
		}
		val.GetObject().Visit(res.Each)
		return res.Result()

	default:
		err := Json.UnmarshalFromString(s, &res)
		if err != nil {
			return nil, errors.Wrap(err, "UnmarshalFromString")
		}
		return res, nil
	}
}

func (t TypeInParam) ConvertSlice(ctx *fasthttp.RequestCtx, values []string) (any, error) {
	switch tp := t.BasicKind; tp {
	case types.String:
		return values, nil

	case types.Int:
		arr := make([]int, len(values))
		for key, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[key] = v.(int)
		}
		return arr, nil

	case types.Int8:
		arr := make([]int8, len(values))
		for key, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[key] = v.(int8)
		}
		return arr, nil

	case types.Int16:
		arr := make([]int16, len(values))
		for key, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[key] = v.(int16)
		}
		return arr, nil

	case types.Int32:
		arr := make([]int32, len(values))
		for key, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[key] = v.(int32)
		}
		return arr, nil

	case types.Int64:
		arr := make([]int64, len(values))
		for key, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[key] = v.(int64)
		}
		return arr, nil

	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		arr := make([]int32, len(values))
		for key, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[key] = v.(int32)
		}
		return arr, nil

	case types.Float32, types.Float64, types.UntypedFloat:
		arr := make([]float32, len(values))
		for key, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[key] = v.(float32)
		}
		return arr, nil

	case types.UnsafePointer:
		return nil, nil

	default:
		arr := make([]any, len(values))
		for i, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[i] = v
		}

		return arr, nil

	}
}

func (t TypeInParam) IsSlice() bool {
	return t.isSlice
}

func (t TypeInParam) String() string {
	res := typesExt.StringTypeKinds(t.BasicKind)
	if t.isSlice {
		return "[]" + res
	}

	return res
}

func (t TypeInParam) StreamRequestType(w *quicktemplate.Writer) {
	if d, ok := t.DTO.(Docs); ok {
		w.N().S(d.RequestType())
		return
	}

	res := typesExt.StringTypeKinds(t.BasicKind)
	if t.isSlice {
		w.N().S(`array[` + res + "]")
	}

	w.N().S(res)
}

func (t TypeInParam) StreamFormat(w *quicktemplate.Writer) {
	if d, ok := t.DTO.(Docs); ok {
		w.N().S(d.FormatDoc())
		return
	}

	res := typesExt.StringTypeKinds(t.BasicKind)
	if t.isSlice {
		w.N().S(`array[` + res + "]")
	}

	w.N().S(res)
}

func (t TypeInParam) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v', 'g', 't':
		_, err := t.TypeString(s, verb)
		if err != nil {
			logs.ErrorLog(err)
		}
	default:
		_, err := t.TypeString(s, verb)
		if err != nil {
			logs.ErrorLog(err)
		}
	}
}

type streamState struct {
	*quicktemplate.QWriter
}

func (s *streamState) Width() (wid int, ok bool) {
	return -1, false
}

func (s *streamState) Precision() (prec int, ok bool) {
	return 0, false
}

func (s *streamState) Flag(c int) bool {
	return true
}

func (t TypeInParam) StreamTypeString(w *quicktemplate.Writer) {
	t.TypeString(&streamState{w.N()}, 's')
}

func (t TypeInParam) TypeString(s fmt.State, verb rune) (int, error) {
	res, namePackage := cases.Title(language.English, cases.NoLower).String(typesExt.StringTypeKinds(t.BasicKind)), ""
	if t.BasicKind < 0 {
		res = "T" + res
		namePackage = "Ext"
	}
	switch {
	case t.isSlice:
		if _, err := s.Write(gotools.StringToBytes(fmt.Sprintf("apis.NewSliceTypeInParam(types%s.%s)",
			namePackage,
			strings.ReplaceAll(res, ".", "")),
		)); err != nil {
			return -1, err
		}
	case t.DTO != nil:
		if _, err := s.Write([]byte("apis.NewStructInParam(")); err != nil {
			return -1, err
		}
		if f, ok := t.DTO.(fmt.Formatter); ok {
			f.Format(s, verb)
			//return strings.Replace(fmt.Sprintf(%s)", s.String()), "*", "&", 1)
		} else {
			if _, err := s.Write(gotools.StringToBytes(strings.Replace(fmt.Sprintf("%T{}", t.DTO), "*", "&", 1))); err != nil {
				return -1, err
			}
		}
		if _, err := s.Write([]byte(")")); err != nil {
			return -1, err
		}
		//return fmt.Sprintf("apis.NewStructInParam(%T{})", t.DTO)
	default:
		if _, err := s.Write(gotools.StringToBytes(fmt.Sprintf("apis.NewTypeInParam(types%s.%s)",
			namePackage,
			strings.ReplaceAll(res, ".", "")))); err != nil {
			return -1, err

		}
	}

	return 0, nil
}

type HeaderInParam struct {
	TypeInParam
	b bytes.Buffer
}

func (h *HeaderInParam) Write(b []byte) (n int, err error) {
	return h.b.Write(b)
}

func (h *HeaderInParam) Width() (wid int, ok bool) {
	return -1, false
}

func (h *HeaderInParam) Precision() (prec int, ok bool) {
	return 0, false
}

func (h *HeaderInParam) Flag(c int) bool {
	return true
}

func (h *HeaderInParam) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

func (h *HeaderInParam) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	stream.WriteObjectField("in")
	stream.WriteString("header")
	stream.WriteMore()
	stream.WriteObjectField("name")
	h.TypeString(h, 's')
	stream.WriteString(h.b.String())
	stream.WriteMore()
}

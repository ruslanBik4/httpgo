// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"go/types"
	"strconv"
)

// APIRouteParamsType encapsulates types operation for apis parameters
type APIRouteParamsType interface {
	fmt.Stringer
	CheckType(ctx *fasthttp.RequestCtx, value string) bool
	ConvertValue(ctx *fasthttp.RequestCtx, value string) (interface{}, error)
	ConvertSlice(ctx *fasthttp.RequestCtx, values []string) (interface{}, error)
	IsSlice() bool
}

// TypeInParam has type definition of params ApiRoute
type TypeInParam struct {
	types.BasicKind
	isSlice bool
}

// NewTypeInParam create TypeInParam
func NewTypeInParam(bk types.BasicKind) TypeInParam {
	return TypeInParam{
		BasicKind: bk}
}

// NewTypeInParam create TypeInParam
func NewSliceTypeInParam(bk types.BasicKind) TypeInParam {
	return TypeInParam{bk, true}
}

// CheckType check of value computable with the TypeInParam
func (t TypeInParam) CheckType(ctx *fasthttp.RequestCtx, value string) bool {
	switch t.BasicKind {
	case types.String:
		return true
	case types.Bool:
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

	}

	return true
}

// CheckType check of value compatable with the TypeInParam
func (t TypeInParam) ConvertValue(ctx *fasthttp.RequestCtx, value string) (interface{}, error) {
	switch t.BasicKind {
	case types.String:
		return value, nil
	case types.Bool:
		return value == "true", nil
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
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		return strconv.ParseUint(value, 10, 64)
	// 	check type convert float64
	case types.Float32, types.Float64:
		return strconv.ParseFloat(value, 64)
	default:
		return nil, errors.New("convert this type not implement")
	}

	return value, nil
}

func (t TypeInParam) ConvertSlice(ctx *fasthttp.RequestCtx, values []string) (interface{}, error) {
	switch t.BasicKind {
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
	case types.Float32, types.Float64:
		arr := make([]float32, len(values))
		for key, val := range values {
			v, err := t.ConvertValue(ctx, val)
			if err != nil {
				return nil, err
			}
			arr[key] = v.(float32)
		}
		return arr, nil
	default:
		panic(errors.New("convert this type not implement"))
	}

	return nil, nil
}

func (t TypeInParam) IsSlice() bool {
	return t.isSlice
}

func (t TypeInParam) String() string {
	res := stringTypeKinds[t.BasicKind]
	if t.isSlice {
		return "[]" + res
	}

	return res
}

var stringTypeKinds = map[types.BasicKind]string{
	types.Invalid: "Invalid",

	// predeclared types
	types.Bool:          "Bool",
	types.Int:           "Int",
	types.Int8:          "Int8",
	types.Int16:         "Int16",
	types.Int32:         "Int32",
	types.Int64:         "Int64",
	types.Uint:          "Uint",
	types.Uint8:         "Uint8",
	types.Uint16:        "Uint16",
	types.Uint32:        "Uint32",
	types.Uint64:        "Uint64",
	types.Uintptr:       "Uintptr",
	types.Float32:       "Float32",
	types.Float64:       "Float64",
	types.Complex64:     "Complex64",
	types.Complex128:    "Complex128",
	types.String:        "String",
	types.UnsafePointer: "UnsafePointer",

	// types for untyped values
	types.UntypedBool:    "bool",
	types.UntypedInt:     "int",
	types.UntypedRune:    "rune",
	types.UntypedFloat:   "float64",
	types.UntypedComplex: "complex128",
	types.UntypedString:  "string",
	types.UntypedNil:     "nil",
	26:                   "array",

	// aliases
	// types.Byte : "Byte",
	// types.Rune : "Rune",
}

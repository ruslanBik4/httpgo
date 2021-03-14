// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"fmt"
	"go/types"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/typesExt"
)

type ApisValues string

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
	// types.Struct
	isSlice bool
	DTO     RouteDTO
}

// NewTypeInParam create TypeInParam
func NewTypeInParam(bk types.BasicKind) TypeInParam {

	return TypeInParam{
		BasicKind: bk,
	}
}

func NewStructInParam(dto RouteDTO) TypeInParam {

	return TypeInParam{
		BasicKind: typesExt.TStruct,
		DTO:       dto,
	}
}

// NewTypeInParam create TypeInParam
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
		err := jsoniter.UnmarshalFromString(value, &v)
		if err != nil {
			logs.ErrorLog(err)
		}

		return err == nil

	default:
		return true
	}
}

// ConvertValue convert value according to TypeInParam's type
func (t TypeInParam) ConvertValue(ctx *fasthttp.RequestCtx, value string) (interface{}, error) {
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

	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		return strconv.ParseUint(value, 10, 64)

	// 	check type convert float64
	case types.Float32, types.Float64:
		return strconv.ParseFloat(value, 64)

	case types.UnsafePointer:
		return nil, nil

	case typesExt.TMap:
		res := make(map[string]interface{}, 0)
		err := jsoniter.UnmarshalFromString(value, &res)
		if err != nil {
			return nil, errors.Wrap(err, "UnmarshalFromString")
		}
		return res, nil

	case typesExt.TArray:
		res := make([]interface{}, 0)
		err := jsoniter.UnmarshalFromString(value, &res)
		if err != nil {
			logs.ErrorLog(err)
		}
		if err != nil {
			return nil, errors.Wrap(err, "UnmarshalFromString")
		}
		return res, nil

	case typesExt.TStruct:
		v := t.DTO.NewValue()
		err := jsoniter.UnmarshalFromString(value, &v)
		if err != nil {
			logs.ErrorLog(err)
		}
		if err != nil {
			return nil, errors.Wrap(err, "UnmarshalFromString")
		}
		return v, nil

	default:
		return nil, errors.Wrapf(ErrWrongParamsList, "convert this type (%s) not implement", t.String())
	}
}

func (t TypeInParam) ConvertSlice(ctx *fasthttp.RequestCtx, values []string) (interface{}, error) {
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

	case types.UnsafePointer:
		return nil, nil

	default:
		arr := make([]interface{}, len(values))
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

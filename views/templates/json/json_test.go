/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package json

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fastjson"
	qt422016 "github.com/valyala/quicktemplate"
)

func TestWriteAnyJSONOld(t *testing.T) {
	var input = map[string]any{
		"one":  []string{"1", "2"},
		"two":  true,
		"thre": 3.00,
		"5":    map[string]any{"1": "6", "4": "l", "5": "u"},
		"6":    map[int]any{1: "6", 4: "l", 5: "u"},
	}

	strJSON := bytes.NewBufferString("")
	StreamMap(qt422016.AcquireWriter(strJSON), input)
	value, err := fastjson.ParseBytes(strJSON.Bytes())
	if assert.Nil(t, err, "WriteAnyJSON=%v, stroka=%s", input, strJSON) {
		obj, err := value.Object()
		assert.Nil(t, err)
		result := make(map[string]any)
		obj.Visit(func(key []byte, v *fastjson.Value) {
			k := string(key)
			result[k], err = switchType(t, input[k], v)
			assert.Nil(t, err)
		})
		assert.Equal(t, input, result, strJSON)
	}

	t.Skipped()
}

func switchType(t *testing.T, input any, value *fastjson.Value) (any, error) {
	switch input.(type) {
	case bool:
		return value.Bool()
	case int:
		return value.Int()
	case float64:
		return value.Float64()
	case []string:
		return convertSlice[string](value.GetArray(), convertString()), nil

	case []int:
		return convertSlice(value.GetArray(), convertInt(t)), nil

	case []any:
		return convertSlice(value.GetArray(), convertAny()), nil

	case map[string]any:
		return convertMap[string, any](
				value.GetObject(),
				convertAny(),
				func(key []byte) string {
					return string(key)
				}),
			nil
	case map[int]any:
		return convertMap(value.GetObject(), convertAny(), ketToInt(t)), nil

	default:
		return value.String(), nil

	}
}

func convertAny() func(val *fastjson.Value) any {
	return func(val *fastjson.Value) any {
		return strings.Trim(val.String(), `"`)
	}
}

func convertString() func(val *fastjson.Value) string {
	return func(val *fastjson.Value) string { return strings.Trim(val.String(), `"`) }
}

func ketToInt(t *testing.T) func(key []byte) int {
	return func(key []byte) int {
		i, err := strconv.Atoi(string(key))
		assert.Nil(t, err)
		return i
	}
}

func convertInt(t *testing.T) func(val *fastjson.Value) int {
	return func(val *fastjson.Value) int {
		i, err := val.Int()
		assert.Nil(t, err)
		return i
	}
}

func convertSlice[T any](arr []*fastjson.Value, convert func(val *fastjson.Value) T) []T {
	res := make([]T, len(arr))
	for i, val := range arr {
		res[i] = convert(val)
	}

	return res
}

func convertMap[E comparable, T any](obj *fastjson.Object, convert func(val *fastjson.Value) T,
	keyConvert func(key []byte) E) map[E]T {
	res := make(map[E]T)
	obj.Visit(func(key []byte, v *fastjson.Value) {

		res[keyConvert(key)] = convert(v)
	})

	return res
}

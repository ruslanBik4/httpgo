// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package forms

import (
	"github.com/ruslanBik4/httpgo/dbEngine"
)

type ColumnDecor struct {
	dbEngine.Column
	IsHidden, IsReadOnly, isSlice bool
	PatternList                   dbEngine.Table
	Pattern                       string

	Value interface{}
}

func (col *ColumnDecor) GetValues() (values []interface{}) {

	switch val := col.Value.(type) {
	case []interface{}:
		return val
	case []string:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.isSlice = true

	case []int32:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.isSlice = true

	case []int64:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.isSlice = true

	case []float32:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.isSlice = true

	case []float64:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.isSlice = true

	default:
		values = append(values, val)
	}

	return
}

func (col *ColumnDecor) InputName(i int) string {
	if col.isSlice {
		return col.Name() + "[]"
	}

	return col.Name()
}

func (col *ColumnDecor) Label() string {
	if c := col.Comment(); c > "" {
		return c
	}

	return col.Name()
}

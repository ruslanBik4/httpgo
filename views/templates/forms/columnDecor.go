// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package forms

import (
	"context"
	"database/sql/driver"
	"go/types"
	"regexp"
	"strings"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/logs"
)

type ColumnDecor struct {
	dbEngine.Column
	IsHidden, IsReadOnly, isSlice bool
	PatternList                   dbEngine.Table
	PatternName                   string
	PlaceHolder                   string
	label                         string
	pattern                       string
	Value                         interface{}
}

var regPattern = regexp.MustCompile(`\{"pattern":\s*"([^"]+)"\}`)

func NewColumnDecor(column dbEngine.Column, patternList dbEngine.Table) *ColumnDecor {

	colDec := &ColumnDecor{Column: column, PatternList: patternList}
	comment := colDec.Comment()
	if m := regPattern.FindAllStringSubmatch(comment, -1); len(m) > 0 {
		colDec.pattern = m[0][1]
		colDec.label = strings.TrimRight(comment, m[0][0])
	} else if column.Comment() > "" {
		colDec.label = column.Comment()
	}

	return colDec
}

func (col *ColumnDecor) Placeholder() string {
	if col.PlaceHolder > "" {
		return col.PlaceHolder
	}

	if col.pattern > "" {
		return col.pattern
	}

	return col.Label()
}

func (col *ColumnDecor) Pattern() string {
	if col.pattern > "" {
		return col.pattern
	}

	if name := col.PatternName; name > "" {
		if col.PatternList != nil {
			err := col.PatternList.SelectAndRunEach(context.Background(),
				func(values []interface{}, columns []dbEngine.Column) error {
					col.pattern = values[0].(string)

					return nil
				},
				dbEngine.ColumnsForSelect("pattern"),
				dbEngine.WhereForSelect("name"),
				dbEngine.ArgsForSelect(name),
			)
			if err != nil {
				logs.ErrorLog(err, "")
			}
		}

		if col.pattern == "" {
			col.pattern = name
		}

	} else if col.BasicTypeInfo() == types.IsInteger {
		col.pattern = `[0-9]+`
	} else if col.BasicTypeInfo() == types.IsFloat {
		col.pattern = `[+-]?\d+(\.\d{2})?`
	} else if col.BasicTypeInfo() == types.IsComplex {
		col.pattern = `[+-]?\d+(\.\d{2})?`
	}

	return col.pattern
}
func (col *ColumnDecor) Type() string {
	const email = "email"
	const tel = "phone"

	if strings.HasPrefix(col.Name(), email) {
		return email
	} else if strings.HasPrefix(col.Name(), tel) {
		return "tel"
	}

	return col.Column.Type()
}

func (col *ColumnDecor) GetValues() (values []interface{}) {

	switch val := col.Value.(type) {
	case []interface{}:
		values = val
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

	case nil:
		if d := col.Default(); d > "" {
			values = append(values, d)
		}
	case driver.Valuer:
		v, err := val.Value()
		if err != nil {
			logs.ErrorLog(err, "val.Value")
		} else {
			values = append(values, v)
		}
	default:
		values = append(values, val)
	}

	if len(values) == 0 {
		values = append(values, nil)
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
	if c := col.label; c > "" {
		return c
	}

	return col.Name()
}

type Button struct {
	Title      string
	Position   bool
	buttonType string
}

type BlockColumns struct {
	Buttons            []Button
	Columns            []*ColumnDecor
	Id                 int
	Title, Description string
}

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

	"github.com/ruslanBik4/httpgo/typesExt"
	"github.com/ruslanBik4/logs"
)

type AttachmentList struct {
	Id  int32  `json:"id"`
	Url string `json:"url"`
}

type ColumnDecor struct {
	dbEngine.Column
	IsHidden, IsDisabled, IsReadOnly, IsSlice, IsNewPrimary,
	SelectWithNew bool
	InputType         string
	SpecialInputName  string
	DefaultInputValue string `json:"defaultInputValue,omitempty"`
	Attachments       []AttachmentList
	SelectOptions     map[string]string
	PatternList       dbEngine.Table
	PatternName       string
	PlaceHolder       string
	LinkNew           string
	Label             string
	pattern           string
	patternDesc       string
	Value             interface{}
	Suggestions       string
	SuggestionsParams map[string]interface{}
}

var regPattern = regexp.MustCompile(`{"pattern":\s*"([^"]+)"}`)

func NewColumnDecor(column dbEngine.Column, patternList dbEngine.Table) *ColumnDecor {

	comment := column.Comment()
	colDec := &ColumnDecor{
		Column:      column,
		IsReadOnly:  column.AutoIncrement() || strings.Contains(comment, "read_only"),
		PatternList: patternList,
	}
	if m := regPattern.FindAllStringSubmatch(comment, -1); len(m) > 0 {
		colDec.getPattern(m[0][1])
		colDec.Label = regPattern.ReplaceAllString(comment, "")
	} else if comment > "" {
		colDec.Label = comment
	} else {
		colDec.Label = column.Name()
	}

	colDec.IsHidden = column.Primary()
	colDec.InputType = colDec.inputType()

	return colDec
}

func (col *ColumnDecor) Copy() *ColumnDecor {
	return &ColumnDecor{
		Column:        col.Column,
		IsReadOnly:    col.IsReadOnly,
		IsSlice:       col.IsSlice,
		InputType:     col.InputType,
		SelectOptions: col.SelectOptions,
		PatternList:   col.PatternList,
		PatternName:   col.PatternName,
		PlaceHolder:   col.PlaceHolder,
		Suggestions:   col.Suggestions,
		Label:         col.Label,
		IsHidden:      col.IsHidden,
		LinkNew:       col.LinkNew,
		pattern:       col.pattern,
		patternDesc:   col.patternDesc,
		// todo must decide later
		// Value:         col.Value,
	}

}
func (col *ColumnDecor) Placeholder() string {
	if col.PlaceHolder > "" {
		return col.PlaceHolder
	}

	if col.patternDesc > "" {
		return col.patternDesc
	}

	if col.pattern > "" {
		return col.pattern
	}

	return col.Label
}

func (col *ColumnDecor) Pattern() string {
	if col.pattern > "" {
		return col.pattern
	}

	if name := col.PatternName; name > "" {
		col.getPattern(name)
	} else if col.BasicTypeInfo() == types.IsInteger && len(col.Attachments) == 0 {
		col.pattern = `^-?\d+$`
	} else if col.BasicTypeInfo() == types.IsFloat {
		col.pattern = `^-?\d+(\.\d{1,2})?$`
	} else if col.BasicTypeInfo() == types.IsComplex {
		col.pattern = `^[+±-]?\d+(\.\d{1,2})?$`
	}

	return col.pattern
}

func (col *ColumnDecor) GetFields(columns []dbEngine.Column) []interface{} {
	if len(columns) == 0 {
		return []interface{}{&col.Value}
	}

	v := make([]interface{}, len(columns))
	for i, c := range columns {
		switch c.Name() {
		case "pattern":
			v[i] = &col.pattern
		case "description":
			v[i] = &col.patternDesc
		}
	}

	return v
}

var regName = regexp.MustCompile(`\w+`)

func (col *ColumnDecor) getPattern(name string) {
	if col.PatternList != nil && regName.MatchString(name) {
		err := col.PatternList.SelectAndScanEach(context.Background(),
			nil,
			col,
			dbEngine.ColumnsForSelect("pattern", "description"),
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
}

const email = "email"
const tel = "phone"

func (col *ColumnDecor) Type() string {
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
		col.IsSlice = true
	case []string:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case []int32:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case []int64:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case []float32:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case []float64:
		values = make([]interface{}, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case nil:
		values = append(values, nil)
	case driver.Valuer:
		v, err := val.Value()
		if err != nil {
			logs.ErrorLog(err, "val.Value %v", val)
			values = append(values, nil)
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
	if col.SpecialInputName > "" {
		return col.SpecialInputName
	}

	if col.IsSlice {
		return col.Name() + "[]"
	}

	if col.IsNewPrimary {
		return "new." + col.Name()
	}

	return col.Name()
}

func (col *ColumnDecor) inputType() string {
	if col.Suggestions > "" {
		return "select"
	}

	if col.IsHidden {
		return "hidden"
	}

	switch col.Type() {
	case "daterange":
		return "date-range"
	case "date", "_date":
		return "date"
	case "datetime", "datetimetz", "timestamp", "timestamptz", "time", "_timestamp", "_timestamptz", "_time":
		return "datetime"
	case email, tel, "tel", "password", "url":
		return col.Type()
	case "text", "_text", "json", "jsonb":
		return "textarea"
	case "bytea", "_bytea":
		return "file"
	default:
		if typesExt.IsNumeric(col.BasicTypeInfo()) {
			return "number"
		}
		if col.BasicType() == types.Bool {
			return "checkbox"
		}

		return "text"
	}
}

type Button struct {
	Title      string
	Position   bool
	ButtonType string
}

type BlockColumns struct {
	Buttons            []Button
	Columns            []*ColumnDecor
	Id                 int
	Multiple           bool
	Title, Description string
}

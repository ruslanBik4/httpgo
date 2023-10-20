/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package forms

import (
	"context"
	"database/sql/driver"
	"go/types"
	"regexp"
	"strings"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/gotools"

	"github.com/ruslanBik4/httpgo/typesExt"
	"github.com/ruslanBik4/logs"
)

type AttachmentList struct {
	Id  int32  `json:"id"`
	Url string `json:"url"`
}

type SuggestionsParams struct {
	Key   string
	Value any
}

type SelectOption struct {
	Disabled, Selected bool
	Value              string
}

type ColumnDecor struct {
	dbEngine.Column
	IsHidden, IsDisabled, IsReadOnly, IsSlice, IsNewPrimary,
	SelectWithNew bool
	ExtProperties     *fastjson.Value
	InputType         string
	SpecialInputName  string
	DefaultInputValue string `json:"defaultInputValue,omitempty"`
	Attachments       []AttachmentList
	Events            map[string]string
	SelectOptions     map[string]SelectOption
	PatternList       dbEngine.Table
	PatternName       string
	PlaceHolder       string
	LinkNew           string
	Label             string
	Max               string
	Min               string
	multiple          bool
	pattern           string
	patternDesc       string
	Value             any
	Accept            string
	Suggestions       string
	SuggestionsParams map[string]any
}

var regPattern = regexp.MustCompile(`{(\s*['"][^"']+['"]:\s*(("[^"]+")|('[^']+')|([^"'{},]+)|({[^}]+})),?)+}`)
var regReadOnly = regexp.MustCompile(`\(*read_only\)*`)

func NewColumnDecor(col dbEngine.Column, patternList dbEngine.Table, suggestions ...SuggestionsParams) *ColumnDecor {

	comment := col.Comment()
	isReadOnly := regReadOnly.MatchString(comment)
	if isReadOnly {
		comment = regReadOnly.ReplaceAllString(comment, "")
	}
	colDec := &ColumnDecor{
		Column:            col,
		IsHidden:          col.Primary() && (col.AutoIncrement() || isReadOnly),
		IsReadOnly:        !col.Primary() && (col.AutoIncrement() || isReadOnly),
		PatternList:       patternList,
		SuggestionsParams: make(map[string]any),
	}
	for _, item := range suggestions {
		colDec.SuggestionsParams[item.Key] = item.Value
	}

	if m := regPattern.FindAllSubmatch([]byte(comment), -1); len(m) > 0 {
		parse, err := fastjson.ParseBytes(m[0][0])
		if err != nil {
			logs.ErrorLog(err, string(m[0][0]))
		} else {
			colDec.ExtProperties = parse
			p := parse.GetStringBytes("pattern")
			colDec.getPattern(gotools.BytesToString(p))
			p = parse.GetStringBytes("suggestions")
			if len(p) > 0 {
				colDec.Suggestions = gotools.BytesToString(p)
			}
			colDec.multiple = parse.GetBool("multiple")
			params := parse.GetObject("suggestions_params")
			params.Visit(func(key []byte, v *fastjson.Value) {
				b, err := v.StringBytes()
				if err != nil {
					logs.ErrorLog(err, key)
					return
				}
				colDec.SuggestionsParams[gotools.BytesToString(key)] = gotools.BytesToString(b)
			})
		}
		colDec.Label, _, _ = strings.Cut(comment, "{")
	} else if comment > "" {
		colDec.Label = comment
	} else {
		colDec.Label = col.Name()
	}

	colDec.InputType = colDec.inputType()

	return colDec
}

func NewColumnDecorFromJSON(val *fastjson.Value, patternList dbEngine.Table) *ColumnDecor {
	col := ColumnDecor{
		Events:      make(map[string]string),
		PatternList: patternList,
	}

	obj, err := val.Object()
	if err != nil {
		logs.ErrorLog(err)
		return nil
	}

	name, comment, errMsg := "", "", ""
	isRequired := false
	obj.Visit(func(key []byte, val *fastjson.Value) {
		switch key := gotools.BytesToString(key); key {
		case "max":
			col.Max = gotools.BytesToString(val.GetStringBytes())
		case "min":
			col.Min = gotools.BytesToString(val.GetStringBytes())
		case "name":
			name = gotools.BytesToString(val.GetStringBytes())
		case "title":
			comment = gotools.BytesToString(val.GetStringBytes())
			col.Label = comment
		case "required":
			b, err := val.Bool()
			if err != nil {
				logs.ErrorLog(err)
			}
			isRequired = b
		case "readOnly":
			col.IsReadOnly = true
		case "accept":
			col.Accept = gotools.BytesToString(val.GetStringBytes())
		case "suggestions":
			col.Suggestions = gotools.BytesToString(val.GetStringBytes())
		case "data":
			err := col.parseSelect(val)
			if err != nil {
				logs.ErrorLog(err)
			}
		case "hidden":
			col.IsHidden = true
			col.InputType = "hidden"
		case "multiple":
			col.multiple = true
		case "error":
			v := val.Get("message")
			errMsg = gotools.BytesToString(v.GetStringBytes())
		case "pattern":
			col.PatternName = gotools.BytesToString(val.GetStringBytes())
			col.getPattern(col.PatternName)
		case "type":
			col.InputType = gotools.BytesToString(val.GetStringBytes())
		case "defaultInputValue":
			col.DefaultInputValue = gotools.BytesToString(val.GetStringBytes())
		case "value":
			col.Value = gotools.BytesToString(val.GetStringBytes())
		default:
			if strings.HasPrefix(key, "on") {
				col.Events[key] = gotools.BytesToString(val.GetStringBytes())
			} else {
				logs.StatusLog("unknown field property '%s'", key)
			}
		}

	})

	if errMsg > "" && col.patternDesc == "" {
		col.patternDesc = errMsg
	}
	col.Column = dbEngine.NewStringColumn(name, comment, isRequired)

	return &col
}

func (col *ColumnDecor) Each(key []byte, val *fastjson.Value) {
	switch gotools.BytesToString(key) {
	case "name":
		//b., err = v.Int()
	case "required":
		col.Required()
	case "pattern":
		col.pattern = gotools.BytesToString(val.GetStringBytes())
	case "type":
		col.InputType = gotools.BytesToString(val.GetStringBytes())
	}

}

func (col *ColumnDecor) Result() (any, error) {
	return col, nil
}

func (col *ColumnDecor) parseSelect(val *fastjson.Value) error {
	blocks, err := val.Array()
	if err != nil {
		return err
	}

	col.SelectOptions = make(map[string]SelectOption, len(blocks))

	for _, val := range blocks {
		name := val.GetStringBytes("label")
		if len(name) == 0 {
			return errors.New("name is empty")
		}
		value := val.GetStringBytes("value")
		if len(value) == 0 {
			return errors.New("value is empty")
		}

		option := SelectOption{
			Value: gotools.BytesToString(value),
		}
		if val.Exists("disabled") {
			d, err := val.Get("disabled").Bool()
			if err != nil {
				logs.ErrorLog(err, "%s.disabled is wrong %s", name, val.GetStringBytes("disabled"))
			} else {
				option.Disabled = d
			}
		}

		if val.Exists("selected") {
			d, err := val.Get("selected").Bool()
			if err != nil {
				logs.ErrorLog(err, "%s.selected is wrong %s", name, val.GetStringBytes("selected"))
			} else {
				option.Selected = d
			}
		}

		col.SelectOptions[gotools.BytesToString(name)] = option
	}

	return nil
}

// Copy make dublicate of Column without Value
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
		col.patternDesc = "only digital"
	} else if col.BasicTypeInfo() == types.IsFloat {
		col.pattern = `^-?\d+(\.\d{1,2})?$`
		col.patternDesc = "float value"
	} else if col.BasicTypeInfo() == types.IsComplex {
		col.pattern = `^[+±-]?\d+(\.\d{1,2})?$`
		col.patternDesc = "complex value"
	}

	return col.pattern
}

func (col *ColumnDecor) GetFields(columns []dbEngine.Column) []any {
	if len(columns) == 0 {
		return []any{&col.Value}
	}

	v := make([]any, len(columns))
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
		err := col.PatternList.SelectAndScanEach(context.TODO(),
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

const (
	email = "email"
	tel   = "phone"
)

func (col *ColumnDecor) Type() string {
	if strings.HasPrefix(col.Name(), email) {
		return email
	} else if strings.HasPrefix(col.Name(), tel) {
		return "tel"
	}

	return col.Column.Type()
}

func (col *ColumnDecor) GetValues() (values []any) {

	switch val := col.Value.(type) {
	case []any:
		values = val
		col.IsSlice = true
	case []string:
		values = make([]any, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case []int32:
		values = make([]any, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case []int64:
		values = make([]any, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case []float32:
		values = make([]any, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case []float64:
		values = make([]any, len(val))
		for i, val := range val {
			values[i] = val
		}
		col.IsSlice = true

	case nil, pgtype.Date:
		values = append(values, val)

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

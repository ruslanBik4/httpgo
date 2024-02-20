/*
 * Copyright (c) 2023-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package forms

import (
	"github.com/valyala/fastjson"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/logs"
)

type Button struct {
	Classes  string
	Hidden   bool
	Position bool
	Title    string
	Type     string
	OnClick  string
}

type BlockColumns struct {
	Id                 int
	Buttons            []Button
	Columns            []*ColumnDecor
	Multiple           bool
	Title, Description string
}

type FormField struct {
	Title, Action, Method, Description string
	HideBlock                          any
	Blocks                             []BlockColumns
}

func NewBlockColumnsFromJSON(val *fastjson.Value, patternList dbEngine.Table) *BlockColumns {
	b := &BlockColumns{}
	obj, err := val.Object()
	if err != nil {
		logs.ErrorLog(err, val)
		return nil
	}

	obj.Visit(func(key []byte, v *fastjson.Value) {
		switch gotools.BytesToString(key) {
		case "id":
			b.Id, err = v.Int()
		case "title":
			b.Title = gotools.BytesToString(v.GetStringBytes())
		case "description":
			b.Description = gotools.BytesToString(v.GetStringBytes())
		case "multiple":
			b.Multiple, err = v.Bool()
		case "columns":
			b.Columns, err = parseField(v, patternList)
		case "buttons":
			b.Buttons, err = parseButtons(v)
		}
		if err != nil {
			logs.ErrorLog(err)
		}
	})
	return b
}

func parseField(val *fastjson.Value, patternList dbEngine.Table) (res []*ColumnDecor, err error) {
	blocks, err := val.Array()
	if err != nil {
		return nil, err
	}
	for _, val := range blocks {
		res = append(res, NewColumnDecorFromJSON(val, patternList))
	}
	return
}

func parseButtons(val *fastjson.Value) (res []Button, err error) {
	buttons, err := val.Array()
	if err != nil {
		return nil, err
	}
	for _, val := range buttons {
		var b Button

		obj := val.GetObject()
		obj.Visit(func(key []byte, v *fastjson.Value) {
			switch gotools.BytesToString(key) {
			case "class", "classes":
				b.Classes = gotools.BytesToString(v.GetStringBytes())
			case "hidden":
				b.Hidden = true
			case "title":
				b.Title = gotools.BytesToString(v.GetStringBytes())
			case "type":
				b.Type = gotools.BytesToString(v.GetStringBytes())
			case "OnClick", "onClick", "onclick", "on_click":
				b.OnClick = gotools.BytesToString(v.GetStringBytes())
			}
		})
		if b.Title > "" {
			res = append(res, b)
		}
	}
	return
}

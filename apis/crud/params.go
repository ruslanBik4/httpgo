/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package crud

import (
	"go/types"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
)

type DbApiParams struct {
	apis.InParam
	col dbEngine.Column
}

func NewDbApiParams(col dbEngine.Column) *DbApiParams {
	p := &DbApiParams{
		apis.InParam{
			Name:              col.Name(),
			Desc:              col.Comment(),
			DefValue:          col.Default(),
			Req:               col.Primary(),
			PartReq:           nil,
			IncompatibleWiths: nil,
			TestValue:         "",
		},
		col,
	}
	p.ConvertDbType(col)

	return p
}

func (p *DbApiParams) ConvertDbType(col dbEngine.Column) {
	if strings.HasPrefix(col.Type(), "_") {
		p.Type = apis.NewSliceTypeInParam(col.BasicType())
		p.Name += "[]"
	}
	switch col.Type() {
	case "date", "timestamp", "timestamptz", "time":
		// todo add new type of date/time
		// p.Type = apis.NewStructInParam(&DateTimeString{})
		p.Type = apis.NewTypeInParam(types.String)
	case "daterange": // col.BasicType() == typesExt.TStruct {
		p.Type = apis.NewStructInParam(&DateMarshal{})
	case "json":
		p.Type = apis.NewStructInParam(&DateMarshal{})

		// } else if col.Foreign() != nil {
		// 	p.Type
	default:
		p.Type = apis.NewTypeInParam(col.BasicType())
	}

}

func ToColDev(ctx *fasthttp.RequestCtx, DB *dbEngine.DB, patternList dbEngine.Table, col dbEngine.Column,
	value interface{}) *forms.ColumnDecor {

	colDec := forms.NewColumnDecor(col, patternList)
	colDec.IsDisabled = colDec.IsReadOnly && !(colDec.IsHidden)
	colDec.IsSlice = strings.HasPrefix(col.Type(), "_")
	colDec.Value = value

	if col.Primary() {
		colDec.IsHidden = true
		colDec.InputType = "hidden"
	} else if col.Type() == "text" {
		colDec.InputType = "textarea"
	} else if col.Name() == "id_photos" {
		colDec.InputType = "attachment"
	} else if col.Name() == "memo" {
		colDec.InputType = "markdown"
	}

	GetForeignOptions(ctx, DB, colDec, value)

	return colDec
}

func GetForeignOptions(ctx *fasthttp.RequestCtx, DB *dbEngine.DB, colDec *forms.ColumnDecor, id interface{}) {
	if f := colDec.Foreign(); f != nil && colDec.Suggestions == "" {
		colDec.Suggestions = "/search/" + f.Parent
		colDec.DefaultInputValue, _ = GetForeignName(ctx, DB, colDec, id).(string)
	}
}

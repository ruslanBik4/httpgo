/*
 * Copyright (c) 2022-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"go/types"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/gotools/typesExt"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
)

type DbApiParams struct {
	apis.InParam
	Col dbEngine.Column
}

func NewDbApiParams(col dbEngine.Column) *DbApiParams {
	param := apis.InParam{
		Name:              col.Name(),
		Desc:              strings.Split(col.Comment(), "{")[0],
		Req:               col.Primary(),
		PartReq:           nil,
		IncompatibleWiths: nil,
		TestValue:         "",
	}
	if !col.AutoIncrement() {
		param.DefValue = col.Default()
		if col.BasicTypeInfo() != types.IsString && param.DefValue == "" {
			param.DefValue = nil
		}
	}
	p := &DbApiParams{
		param,
		col,
	}
	p.ConvertDbType(col)

	return p
}

func (p *DbApiParams) ConvertDbType(col dbEngine.Column) {
	isArray := strings.HasPrefix(col.Type(), "_")
	switch col.Type() {
	case "date":
		p.Type = apis.NewStructInParam(&DateString{})
	case "timestamp", "time":
		p.Type = apis.NewStructInParam(&DateTimeString{})
	case "timestamptz":
		p.Type = apis.NewStructInParam(&DateTimeTZString{&DateTimeString{}})
	case "daterange":
		p.Type = apis.NewStructInParam(&DateRangeMarshal{})
	case "numrange":
		p.Type = apis.NewStructInParam(&NumrangeMarshal{})
	case "bytea":
		p.Type = apis.NewStructInParam(&DtoFileField{})
	case "json", "jsonb":
		p.Type = apis.NewStructInParam(&DtoField{})
	case "inet":
		p.Type = apis.NewStructInParam(&InetMarshal{})
	case "interval":
		p.Type = apis.NewStructInParam(&IntervalMarshal{})
	default:
		basicType := col.BasicType()
		if isArray {
			p.Type = apis.NewSliceTypeInParam(basicType)
			p.Name += "[]"
		} else if basicType == typesExt.TStruct {
			p.Type = apis.NewStructInParam(nil)
		} else {
			p.Type = apis.NewTypeInParam(basicType)
		}
	}
}

func ToColDev(ctx *fasthttp.RequestCtx, DB *dbEngine.DB, patternList dbEngine.Table, col dbEngine.Column,
	value any) *forms.ColumnDecor {

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

func GetForeignOptions(ctx *fasthttp.RequestCtx, DB *dbEngine.DB, colDec *forms.ColumnDecor, id any) {
	if f := colDec.Foreign(); f != nil && colDec.Suggestions == "" {
		colDec.Suggestions = "/search/" + f.Parent
		colDec.DefaultInputValue, _ = GetForeignName(ctx, DB, colDec, id).(string)
	}
}

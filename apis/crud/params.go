/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
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
	p.ConvertDbType()

	return p
}

func (p *DbApiParams) ConvertDbType() {
	if strings.HasPrefix(p.Col.Type(), "_") {
		p.convertArrayType()
		return
	}

	p.setType(p.Col.Type())
}

// setType handles type conversion based on PostgreSQL type name
func (p *DbApiParams) setType(t string, opts ...apis.InParamOptions) {
	switch t {
	case "date":
		p.Type = apis.NewStructInParam(&DateString{}, opts...)
	case "time":
		p.Type = apis.NewStructInParam(&DateTimeString{}, opts...)
	case "timestamp":
		p.Type = apis.NewStructInParam(NewTimestampString(), opts...)
	case "timestamptz":
		p.Type = apis.NewStructInParam(NewTzString(), opts...)
	case "daterange":
		p.Type = apis.NewStructInParam(NewDateRangeMarshal(), opts...)
	case "numrange":
		p.Type = apis.NewStructInParam(&NumrangeMarshal{}, opts...)
	case "bytea":
		p.Type = apis.NewStructInParam(&DtoFileField{}, opts...)
	case "json", "jsonb":
		p.Type = apis.NewStructInParam(&DtoField{}, opts...)
	case "inet":
		p.Type = apis.NewStructInParam(NewInetMarshal(), opts...)
	case "interval":
		p.Type = apis.NewStructInParam(NewIntervalMarshal(), opts...)
	case "uuid", "xml", "cidr", "macaddr", "macaddr8", "bit", "bit varying":
		p.Type = apis.NewTypeInParam(types.String, opts...)
	case "money":
		p.Type = apis.NewTypeInParam(types.Float64, opts...)
	case "hstore":
		p.Type = apis.NewStructInParam(&DtoField{}, opts...)
	// Geometric & full-text types — treated as composite (pgx has dedicated types)
	case "point":
		p.Type = apis.NewStructInParam(&PointString{}, opts...)
	case "line":
		p.Type = apis.NewStructInParam(&LineString{}, opts...)
	case "lseg":
		p.Type = apis.NewStructInParam(&LsegString{}, opts...)
	case "box":
		p.Type = apis.NewStructInParam(&BoxString{}, opts...)
	case "path":
		p.Type = apis.NewStructInParam(&PathString{}, opts...)
	case "polygon":
		p.Type = apis.NewStructInParam(&PolygonString{}, opts...)
	case "circle":
		p.Type = apis.NewStructInParam(&CircleString{}, opts...)
	case "tsvector":
		p.Type = apis.NewStructInParam(&TSVectorString{}, opts...)
	case "tsquery":
		p.Type = apis.NewStructInParam(&TSQueryString{}, opts...)

	default:
		basicType := p.Col.BasicType()
		if basicType == typesExt.TStruct {
			p.Type = apis.NewStructInParam(nil, opts...)
		} else {
			p.Type = apis.NewTypeInParam(basicType, opts...)
		}
	}
}

// convertArrayType handles array types by delegating to setType
func (p *DbApiParams) convertArrayType() {
	trimmedType := strings.TrimPrefix(p.Col.Type(), "_")
	p.setType(trimmedType, apis.SetSlice(true))
	p.Name += "[]"
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

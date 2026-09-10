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
	case "int4range":
		p.Type = apis.NewStructInParam(NewInt4RangeMarshal(), opts...)
	case "int8range":
		p.Type = apis.NewStructInParam(NewInt8RangeMarshal(), opts...)
	case "tsrange":
		p.Type = apis.NewStructInParam(NewTsRangeMarshal(), opts...)
	case "tstzrange":
		p.Type = apis.NewStructInParam(NewTsTzRangeMarshal(), opts...)
	case "numeric", "decimal":
		// pgtype.Numeric parses the decimal text itself - never through float64 -
		// so full precision survives, unlike the default basicType() branch below
		// (which this case pre-empts) that would otherwise map it to float64.
		p.Type = apis.NewStructInParam(&NumericString{}, opts...)
	case "bytea":
		//DtoFileField is array
		p.Type = apis.NewStructInParam(&DtoFileField{})
	case "json", "jsonb":
		p.Type = apis.NewStructInParam(&DtoField{}, opts...)
	case "inet", "cidr":
		// pgx's InetCodec explicitly handles both inet and cidr the same way
		// (both prefer netip.Prefix/netip.Addr), so cidr reuses inet's own wrapper
		// instead of falling through to a bare, unvalidated string like it used to.
		p.Type = apis.NewStructInParam(NewInetMarshal(), opts...)
	case "interval":
		p.Type = apis.NewStructInParam(NewIntervalMarshal(), opts...)
	case "uuid":
		// was: apis.NewTypeInParam(types.String) - a malformed UUID only surfaced
		// as an opaque Postgres error. UUIDString validates it up front.
		p.Type = apis.NewStructInParam(&UUIDString{}, opts...)
	case "macaddr", "macaddr8":
		// same reasoning as uuid above; pgx has no dedicated pgtype.Macaddr, it
		// scans/encodes both straight into net.HardwareAddr.
		p.Type = apis.NewStructInParam(&MacaddrString{}, opts...)
	case "xml", "bit", "bit varying":
		// left as plain text on purpose: xml has no meaningful client-side
		// validation to add here, and bit-string format checking is a low-value
		// edge case relative to the other gaps fixed in this pass.
		p.Type = apis.NewTypeInParam(types.String, opts...)
	case "money":
		// float64 for easier arithmetic on the Go/API side. Matches
		// UdtNameToType in dbEngine/psql/column.go, which already classifies
		// "money"/"_money" as types.Float64 at the base-type level - this
		// case exists only to short-circuit the generic basicType() fallback
		// below with the same result, so behavior now agrees with column.go
		// instead of silently overriding it to a string.
		p.Type = apis.NewTypeInParam(types.Float64, opts...)
	case "hstore":
		// was: apis.NewStructInParam(&DtoField{}) (map[string]any) - DtoField has
		// no GetPgxType at all and can't represent hstore's per-key NULLs the way
		// pgtype.Hstore (map[string]*string) does.
		p.Type = apis.NewStructInParam(NewHstoreMarshal(), opts...)
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
		// A user-defined enum column gets client-side label validation instead of
		// only finding out a bad label is invalid after a round trip to Postgres.
		// This also covers an array-of-enum column: convertArrayType (below) calls
		// setType with the trimmed element type name, which lands here the same way.
		if udt := p.Col.UserDefinedType(); udt != nil && len(udt.Enumerates) > 0 {
			p.Type = apis.NewStructInParam(NewEnumString(udt.Enumerates), opts...)
			return
		}

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

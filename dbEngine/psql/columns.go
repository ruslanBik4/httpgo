// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package psql

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/ruslanBik4/httpgo/dbEngine"
	"github.com/ruslanBik4/httpgo/typesExt"
)

type Column struct {
	Table                  dbEngine.Table `json:"-"`
	name                   string
	DataType               string
	ColumnDefault          string
	IsNullable             bool
	CharacterSetName       string
	comment                string
	UdtName                string
	characterMaximumLength int
	PrimaryKey             bool
	IsHidden               bool
}

func (c *Column) GetFields(columns []dbEngine.Column) []interface{} {
	v := make([]interface{}, len(columns))
	for i, col := range columns {
		switch name := col.Name(); name {
		case "data_type":
			v[i] = &c.DataType
		case "column_default":
			v[i] = &c.ColumnDefault
		case "is_nullable":
			v[i] = &c.IsNullable
		case "character_set_name":
			v[i] = &c.CharacterSetName
		case "character_maximum_length":
			v[i] = &c.characterMaximumLength
		case "udt_name":
			v[i] = &c.UdtName
		case "column_comment":
			v[i] = &c.comment
		default:
			panic("not implement scan for field " + name)
		}
	}

	return v
}

func NewColumnPone(name string, comment string, characterMaximumLength int) *Column {
	return &Column{name: name, comment: comment, characterMaximumLength: characterMaximumLength}
}

func NewColumn(table dbEngine.Table, name string, dataType string, columnDefault string, isNullable bool, characterSetName string, comment string, udtName string, characterMaximumLength int, primaryKey bool, isHidden bool) *Column {
	return &Column{
		Table:                  table,
		name:                   name,
		DataType:               dataType,
		ColumnDefault:          columnDefault,
		IsNullable:             isNullable,
		CharacterSetName:       characterSetName,
		comment:                comment,
		UdtName:                udtName,
		characterMaximumLength: characterMaximumLength,
		PrimaryKey:             primaryKey,
		IsHidden:               isHidden,
	}
}

func (c *Column) BasicTypeInfo() types.BasicInfo {
	switch c.BasicType() {
	case types.Bool:
		return types.IsBoolean
	case types.Int32, types.Int64:
		return types.IsInteger
	case types.Float32, types.Float64:
		return types.IsFloat
	case types.String:
		return types.IsString
	default:
		return types.IsUntyped
	}
}

func (c *Column) BasicType() types.BasicKind {
	switch c.UdtName {
	case "bool":
		return types.Bool
	case "int4", "_int4":
		return types.Int32
	case "int8", "_int8":
		return types.Int64
	case "float4", "_float4":
		return types.Float32
	case "float8", "_float8":
		return types.Float64
	case "numeric", "decimal":
		// todo add check field length
		return types.Float64
	case "date", "timestampt", "timestamptz", "time", "_date", "_timestampt", "_timestamptz", "_time":
		return types.String
	case "json":
		return typesExt.TMap
	case "timerange", "tsrange":
		// todo add check ranges
		return types.String
	case "varchar", "_varchar", "text":
		return types.String
	case "bytea", "_bytea":
		return types.UnsafePointer
	default:
		return types.Invalid
	}
}

const isNotNullable = "not null"

var dataTypeAlias = map[string][]string{
	"character varying":           {"varchar(255)", "varchar"},
	"character":                   {"char"},
	"integer":                     {"serial", "int"},
	"bigint":                      {"bigserial"},
	"double precision":            {"float", "real"},
	"timestamp without time zone": {"timestamp"},
	"timestamp with time zone":    {"timestamptz"},
	//todo: add check user-defined types
	"USER-DEFINED": {"timerange"},
	"ARRAY":        {"integer[]", "character varying[]"},
}

// todo: add check arrays
func (c *Column) CheckAttr(fieldDefine string) (res string) {
	fieldDefine = strings.ToLower(fieldDefine)
	isMayNull := strings.Contains(fieldDefine, isNotNullable)
	if c.IsNullable && isMayNull {
		res += " is nullable "
	} else if !c.IsNullable && !isMayNull {
		res += " is not nullable "
	}

	isTypeValid := strings.HasPrefix(fieldDefine, c.DataType)
	if !isTypeValid {
		for _, alias := range dataTypeAlias[c.DataType] {
			if isTypeValid = strings.HasPrefix(fieldDefine, alias); isTypeValid {
				break
			}
		}
	}

	if isTypeValid {
		l := c.CharacterMaximumLength()
		if strings.HasPrefix(c.DataType, "character") &&
			(l > 0) &&
			!strings.Contains(fieldDefine, fmt.Sprintf("char(%d)", l)) {
			res += fmt.Sprintf(" has length %d symbols", l)
		}
	} else {
		res += " has type " + c.DataType
	}

	return
}
func (c *Column) CharacterMaximumLength() int {
	return c.characterMaximumLength
}

func (c *Column) Comment() string {
	return c.comment
}

func (c *Column) Name() string {
	return c.name
}

func (c *Column) Primary() bool {
	return c.PrimaryKey
}

func (c *Column) Type() string {
	return c.UdtName
}

func (c *Column) Required() bool {
	return !c.IsNullable && ((c.ColumnDefault == "") || (c.ColumnDefault == "NULL"))
}

func (c *Column) SetNullable(f bool) {
	c.IsNullable = f
}

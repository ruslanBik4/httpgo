// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package psql

import (
	"github.com/ruslanBik4/httpgo/dbEngine"
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

func (c Column) CharacterMaximumLength() int {
	return c.characterMaximumLength
}

func (c Column) Comment() string {
	return c.comment
}

func (c Column) Name() string {
	return c.name
}

func (c Column) Type() string {
	return c.UdtName
}

func (c Column) Required() bool {
	return c.IsNullable && ((c.ColumnDefault == "") || (c.ColumnDefault == "NULL"))
}

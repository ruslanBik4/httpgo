// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbEngine

import (
	"go/types"
)

type NumberColumn struct {
	comment, name   string
	req, IsNullable bool
}

func (c *NumberColumn) AutoIncrement() bool {
	return false
}

func (c *NumberColumn) Default() string {
	return "0"
}

func NewNumberColumn(name, comment string, req bool) *NumberColumn {
	return &NumberColumn{comment: comment, name: name, req: req}
}

func (c *NumberColumn) CheckAttr(fieldDefine string) string {
	return ""
}

func (c *NumberColumn) Comment() string {
	return c.comment
}

func (c *NumberColumn) Primary() bool {
	return true
}

func (c *NumberColumn) Type() string {
	return "int"
}

func (c *NumberColumn) Required() bool {
	return c.req
}

func (c *NumberColumn) Name() string {
	return c.name
}

func (c *NumberColumn) CharacterMaximumLength() int {
	return 0
}

func (c *NumberColumn) BasicType() types.BasicKind {
	return types.Int
}

func (c *NumberColumn) BasicTypeInfo() types.BasicInfo {
	return types.IsInteger
}

func (c *NumberColumn) SetNullable(f bool) {
	c.IsNullable = f
}

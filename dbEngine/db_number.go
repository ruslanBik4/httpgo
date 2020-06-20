// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbEngine

import (
	"go/types"
)

type NumberColumn struct {
	comment, name string
	req           bool
}

func NewNumberColumn(name, comment string, req bool) *NumberColumn {
	return &NumberColumn{comment: comment, name: name, req: req}
}

func (s *NumberColumn) Comment() string {
	return s.comment
}

func (s *NumberColumn) Type() string {
	return "int"
}

func (s *NumberColumn) Required() bool {
	return s.req
}

func (s *NumberColumn) Name() string {
	return s.name
}

func (s *NumberColumn) CharacterMaximumLength() int {
	return 0
}

func (s *NumberColumn) BasicType() types.BasicKind	{
	return types.Int
}

func (s *NumberColumn) BasicTypeInfo() types.BasicInfo	{
	return types.IsInteger
}

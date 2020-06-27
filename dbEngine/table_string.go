// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbEngine

import (
	"golang.org/x/net/context"
)

type tableString struct {
}

func (t tableString) Columns() []Column {
	panic("implement me")
}

func (t tableString) FindColumn(name string) Column {
	panic("implement me")
}

func (t tableString) FindIndex(name string) *Index {
	panic("implement me")
}

func (t tableString) GetColumns(ctx context.Context) error {
	panic("implement me")
}

func (t tableString) Insert(ctx context.Context, Options ...BuildSqlOptions) error {
	panic("implement me")
}

func (t tableString) Update(ctx context.Context, Options ...BuildSqlOptions) error {
	panic("implement me")
}

func (t tableString) Name() string {
	return "StringTable"
}

func (t tableString) RereadColumn(name string) Column {
	panic("implement me")
}

func (t tableString) Select(ctx context.Context, Options ...BuildSqlOptions) error {
	panic("implement me")
}

func (t tableString) SelectAndScanEach(ctx context.Context, each func() error, rowValue RowScanner, Options ...BuildSqlOptions) error {
	panic("implement me")
}

func (t tableString) SelectAndRunEach(ctx context.Context, each FncEachRow, Options ...BuildSqlOptions) error {
	panic("implement me")
}

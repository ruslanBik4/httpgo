/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

// In-memory doubles of dbEngine.Column/Table for the tests of view.go and
// forms.go: the dbEngine test types (StringColumn, TableString) have a fixed
// type "string", no primary/foreign keys and panic on selects.

import (
	"context"
	"fmt"
	"go/types"
	"reflect"

	"github.com/jackc/pgx/v5"

	"github.com/ruslanBik4/dbEngine/dbEngine"
)

// fakeColumn is a column of any type with keys.
type fakeColumn struct {
	*dbEngine.StringColumn
	typ      string
	basic    types.BasicKind
	primary  bool
	autoInc  bool
	foreign  *dbEngine.ForeignKey
	table    dbEngine.Table
	userType *dbEngine.Types
}

func newCol(name, typ string, opts ...func(*fakeColumn)) *fakeColumn {
	c := &fakeColumn{StringColumn: dbEngine.NewStringColumn(name, "", false), typ: typ, basic: types.String}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func withComment(comment string) func(*fakeColumn) {
	return func(c *fakeColumn) {
		c.StringColumn = dbEngine.NewStringColumn(c.Name(), comment, c.Required())
	}
}

func asRequired(c *fakeColumn) {
	c.StringColumn = dbEngine.NewStringColumn(c.Name(), c.Comment(), true)
}

func asPrimary(c *fakeColumn) { c.primary, c.basic = true, types.Int32 }

func asAutoInc(c *fakeColumn) { c.autoInc = true }

func withBasic(kind types.BasicKind) func(*fakeColumn) {
	return func(c *fakeColumn) { c.basic = kind }
}

func withEnum(values ...string) func(*fakeColumn) {
	return func(c *fakeColumn) { c.userType = &dbEngine.Types{Name: c.typ, Enumerates: values} }
}

// refersTo makes the column a foreign key to column parentCol of table parent.
func refersTo(parent *fakeTable, parentCol string) func(*fakeColumn) {
	return func(c *fakeColumn) {
		c.foreign = &dbEngine.ForeignKey{
			Parent:     parent.Name(),
			Column:     parentCol,
			ForeignCol: parent.FindColumn(parentCol),
		}
	}
}

func (c *fakeColumn) Type() string                     { return c.typ }
func (c *fakeColumn) BasicType() types.BasicKind       { return c.basic }
func (c *fakeColumn) BasicTypeInfo() types.BasicInfo   { return types.Typ[c.basic].Info() }
func (c *fakeColumn) Primary() bool                    { return c.primary }
func (c *fakeColumn) AutoIncrement() bool              { return c.autoInc }
func (c *fakeColumn) Foreign() *dbEngine.ForeignKey    { return c.foreign }
func (c *fakeColumn) Table() dbEngine.Table            { return c.table }
func (c *fakeColumn) UserDefinedType() *dbEngine.Types { return c.userType }

// fakeTable answers selects from memory and records their SQL.
type fakeTable struct {
	*dbEngine.TableString
	// byArg: SelectOneAndScan returns byArg[fmt.Sprint(first arg)], pgx.ErrNoRows when absent
	byArg map[string]any
	// rows for SelectAndRunEach, with columns rowCols
	rows    [][]any
	rowCols []dbEngine.Column
	// sql of every select
	sql []string
}

func newTable(name, comment string, cols ...*fakeColumn) *fakeTable {
	columns := make([]dbEngine.Column, len(cols))
	t := &fakeTable{byArg: map[string]any{}}
	for i, c := range cols {
		columns[i] = c
	}
	t.TableString = dbEngine.NewTableString(name, comment, columns, nil, nil)
	for _, c := range cols {
		c.table = t
	}
	return t
}

func (t *fakeTable) record(opts []dbEngine.BuildSqlOptions) (*dbEngine.SQLBuilder, error) {
	b, err := dbEngine.NewSQLBuilder(t, opts...)
	if err != nil {
		return nil, err
	}
	sql, err := b.SelectSql()
	if err != nil {
		return nil, err
	}
	t.sql = append(t.sql, sql)
	return b, nil
}

func (t *fakeTable) SelectOneAndScan(_ context.Context, row any, opts ...dbEngine.BuildSqlOptions) error {
	b, err := t.record(opts)
	if err != nil {
		return err
	}
	if len(b.Args) == 0 {
		return pgx.ErrNoRows
	}
	v, ok := t.byArg[fmt.Sprint(b.Args[0])]
	if !ok {
		return pgx.ErrNoRows
	}
	reflect.ValueOf(row).Elem().Set(reflect.ValueOf(v))
	return nil
}

func (t *fakeTable) SelectAndRunEach(_ context.Context, each dbEngine.FncEachRow, opts ...dbEngine.BuildSqlOptions) error {
	if _, err := t.record(opts); err != nil {
		return err
	}
	for _, values := range t.rows {
		if err := each(values, t.rowCols); err != nil {
			return err
		}
	}
	return nil
}

func newDB(tables ...*fakeTable) *dbEngine.DB {
	db := &dbEngine.DB{Tables: make(map[string]dbEngine.Table, len(tables))}
	for _, t := range tables {
		db.Tables[t.Name()] = t
	}
	return db
}

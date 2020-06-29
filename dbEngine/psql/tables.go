// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package psql

import (
	"sync"

	"github.com/jackc/pgproto3/v2"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/dbEngine"
	"github.com/ruslanBik4/httpgo/logs"
)

// FieldsTable for fields parameters in form
type Table struct {
	conn       *Conn
	name, Type string
	ID         int
	Comment    string
	columns    []*Column
	PK         string
	lock       sync.RWMutex
}

func (t *Table) GetFields(columns []dbEngine.Column) []interface{} {
	if len(columns) == 0 {
		return []interface{}{&t.name, &t.Type, &t.Comment}
	}

	v := make([]interface{}, len(columns))
	for i, col := range columns {
		switch name := col.Name(); name {
		case "table_name":
			v[i] = &t.name
		case "table_type":
			v[i] = &t.Type
		case "comment":
			v[i] = &t.Comment
		case "oid":
			v[i] = &t.ID
		default:
			panic("not implement scan for field " + name)
		}
	}

	return v

}

func (t *Table) Columns() []dbEngine.Column {
	res := make([]dbEngine.Column, len(t.columns))
	for i, col := range t.columns {
		res[i] = col
	}

	return res
}

func (t *Table) Insert(ctx context.Context, Options ...dbEngine.BuildSqlOptions) error {
	b := &dbEngine.SQLBuilder{Table: t}
	for _, setOption := range Options {
		err := setOption(b)
		if err != nil {
			return errors.Wrap(err, "setOption")
		}
	}

	sql, err := b.InsertSql()
	if err != nil {
		return err
	}

	comTag, err := t.conn.Exec(ctx, sql, b.Args...)

	logs.DebugLog("%s", comTag)

	return errors.Wrap(err, sql)
}

func (t *Table) Update(ctx context.Context, Options ...dbEngine.BuildSqlOptions) error {
	b := &dbEngine.SQLBuilder{Table: t}
	for _, setOption := range Options {
		err := setOption(b)
		if err != nil {
			return errors.Wrap(err, "setOption")
		}
	}

	sql, err := b.UpdateSql()
	if err != nil {
		return err
	}

	comTag, err := t.conn.Exec(ctx, sql, b.Args...)

	logs.DebugLog("%s", comTag)

	return errors.Wrap(err, sql)
}

func (t *Table) Name() string {
	return t.name
}

func (t *Table) Select(ctx context.Context, Options ...dbEngine.BuildSqlOptions) error {
	b := &dbEngine.SQLBuilder{Table: t}
	for _, setOption := range Options {
		err := setOption(b)
		if err != nil {
			return errors.Wrap(err, "setOption")
		}
	}
	sql, err := b.SelectSql()
	if err != nil {
		return err
	}

	_, err = t.conn.Query(ctx, sql, b.Args...)

	return err
}

func (t *Table) SelectAndScanEach(ctx context.Context, each func() error, row dbEngine.RowScanner, Options ...dbEngine.BuildSqlOptions) error {

	b := &dbEngine.SQLBuilder{Table: t}
	for _, setOption := range Options {
		err := setOption(b)
		if err != nil {
			return errors.Wrap(err, "setOption")
		}
	}
	sql, err := b.SelectSql()
	if err != nil {
		return err
	}

	return t.conn.SelectAndScanEach(ctx, each, row, sql, b.Args...)
}

func (t *Table) SelectAndRunEach(ctx context.Context, each dbEngine.FncEachRow, Options ...dbEngine.BuildSqlOptions) error {
	b := &dbEngine.SQLBuilder{Table: t}
	for _, setOption := range Options {
		err := setOption(b)
		if err != nil {
			return errors.Wrap(err, "setOption")
		}
	}

	sql, err := b.SelectSql()
	if err != nil {
		return err
	}

	return t.conn.SelectAndRunEach(
		ctx,
		func(values []interface{}, columns []pgproto3.FieldDescription) error {
			return each(values, b.SelectColumns)
		},
		sql,
		b.Args...)
}

func (t *Table) FindColumn(name string) dbEngine.Column {
	return t.findColumn(name)
}

func (t *Table) findColumn(name string) *Column {
	for _, col := range t.columns {
		if col.Name() == name {
			return col
		}
	}

	return nil
}

// GetColumns получение значений полей для форматирования данных
// получение значений полей для таблицы
func (t *Table) GetColumns(ctx context.Context) error {

	err := t.conn.SelectAndRunEach(ctx, t.readColumnRow, sqlGetTablesColumns+" ORDER BY C.ordinal_position", t.name)
	if err != nil {
		return err
	}

	return nil
	// ind := &Index{}
	// return SelectAndScanEach(func() error {
	//
	// 	ind.columns = make([]*TableColumn, len(ind.col))
	// 	for i, col := range ind.col {
	// 		ind.columns[i] = t.FindColumn(col)
	// 	}
	//
	// 	t.Indexes = append(t.Indexes, &Index{
	// 		Name:   ind.Name,
	// 		columns: ind.columns,
	// 	})
	//
	// 	return nil
	// },
	// 	ind, sqlGetIndexes, t.Name)
}

func (t *Table) FindIndex(name string) *dbEngine.Index {
	// todo implements in future
	return nil
}

func (t *Table) RereadColumn(name string) dbEngine.Column {
	t.lock.RLock()
	defer t.lock.RUnlock()

	column := t.findColumn(name)
	if column == nil {
		column = NewColumnPone(
			name,
			"new column",
			0,
		)

		column.Table = t

		t.columns = append(t.columns, column)

		return column
	}

	// todo implement
	err := t.conn.SelectAndScanEach(
		context.TODO(),
		func() error {
			return nil
		},
		column, sqlGetColumnAttr, t.name, column.Name(),
	)
	if err != nil {
		logs.ErrorLog(err, sqlGetTablesColumns)
		return nil
	}

	return column
}

func (t *Table) readColumnRow(values []interface{}, columns []pgproto3.FieldDescription) error {

	pk, isPK := values[7].(string)
	if isPK {
		t.PK = pk
	}

	col := NewColumn(
		t,
		values[0].(string),
		//DataType:
		values[1].(string),
		//ColumnDefault:
		values[2].(string),
		//IsNullable:
		values[3].(bool),
		//CharacterSetName:
		values[4].(string),
		values[9].(string), //comment
		// udtname
		values[6].(string),
		//CharacterMaximumLength:
		int(values[5].(int32)),
		//PrimaryKey:
		isPK && values[8].(bool),
		false,
	)

	t.columns = append(t.columns, col)

	return nil
}

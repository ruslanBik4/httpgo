// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package psql

import (
	"sync"

	"github.com/jackc/pgproto3/v2"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/dbEngine"
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
		return []interface{}{&t.name, &t.Type}
	}
	panic("implement me")
}

func (t Table) Columns() []dbEngine.Column {
	panic("implement me")
}

func (t Table) Insert(ctx context.Context) {
	panic("implement me")
}

func (t Table) Name() string {
	return t.name
}

func (t Table) Select(ctx context.Context) {
	panic("implement me")
}

func (t Table) SelectAndScanEach(ctx context.Context, each func() error, rowValue dbEngine.RowScanner) error {
	panic("implement me")
}

func (t Table) SelectAndRunEach(ctx context.Context, each func(values []interface{}, columns []dbEngine.Column) error) error {
	panic("implement me")
}

func (t Table) FindField(name string) dbEngine.Column {
	for _, col := range t.columns {
		if col.Name() == name {
			return col
		}
	}

	return nil
}

// GetColumns получение значений полей для форматирования данных
// получение значений полей для таблицы
func (t Table) GetColumns(ctx context.Context) error {

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
	// 		ind.columns[i] = t.FindField(col)
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

func (table Table) FindIndex(name string) dbEngine.Index {
	return nil
}

func (table Table) RecacheField(nameColumn string) dbEngine.Column {
	table.lock.RLock()
	defer table.lock.RUnlock()

	column := table.FindField(nameColumn)
	if column == nil {
		column := NewColumnPone(
			nameColumn,
			"new column",
			0,
		)

		column.Table = table

		table.columns = append(table.columns, column)
	}

	// todo implement
	// var CharacterMaximumLength int
	// sql := sqlGetColumnAttr
	// rows := SelectToRow(sql, table.Name, nameColumn)
	// //todo chg len
	// err := rows.Scan(&column.DataType, &column.ColumnDefault, &column.IsNullable,
	// 	&column.CharacterSetName, &CharacterMaximumLength, &column.UdtName)
	// if err != nil {
	// 	logs.ErrorLog(err, "rows.Scan")
	// 	return nil
	// }

	return column
}

func (t Table) readColumnRow(values []interface{}, columns []pgproto3.FieldDescription) error {

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

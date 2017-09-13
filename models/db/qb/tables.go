// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"github.com/ruslanBik4/httpgo/models/db/schema"
)

// GetSchema return DB schema from table
func (table *QBTable) GetSchema() *schema.FieldsTable {
	return table.schema
}

func (table *QBTable) getFieldSchema(name string) *schema.FieldStructure {
	for _, field := range table.schema.Rows {
		if field.COLUMN_NAME == name {
			return field
		}
	}

	return nil
}

// AddTables adding table from map to list, returns qB
func (qb *QueryBuilder) AddTables(names map[string]string) *QueryBuilder {
	for alias, name := range names {
		qb.AddTable(alias, name)
	}

	return qb
}

// AddTable - add Table to list, returns object table
func (qb *QueryBuilder) AddTable(alias, name string) *QBTable {

	//if alias == ""  {
	//	alias = name
	//}
	table := &QBTable{Name: name, Alias: alias, qB: qb}
	defer schemaError()
	table.schema = schema.GetFieldsTable(table.Name)
	table.Fields = make(map[string]*QBField, 0)
	qb.Tables = append(qb.Tables, table)

	return table
}
// FindTable search table by {name} in list
func (qb *QueryBuilder) FindTable(name string) *QBTable {
	for _, table := range qb.Tables {
		if table.Name == name {
			return table
		}
	}

	return nil
}
func (table *QBTable) addAllFields() {

	if len(table.Fields) > 0 {
		return
	}
	for _, fieldStrc := range table.schema.Rows {

		//field := &QBField{Name: fieldStrc.COLUMN_NAME, Schema: fieldStrc, Table: table}
		table.AddField("", fieldStrc.COLUMN_NAME)
		//TODO: сделать одно место для добавления полей!
		table.qB.fields = append(table.qB.fields, table.Fields[fieldStrc.COLUMN_NAME])
		table.qB.Aliases = append(table.qB.Aliases, fieldStrc.COLUMN_NAME)
	}

}

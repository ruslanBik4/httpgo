// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import "github.com/ruslanBik4/httpgo/models/db/schema"

// getters
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

// add Tables list, returns qB
func (qb *QueryBuilder) AddTables(names map[string]string) *QueryBuilder {
	for alias, name := range names {
		qb.AddTable(alias, name)
	}

	return qb
}

//add Table, returns object table
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

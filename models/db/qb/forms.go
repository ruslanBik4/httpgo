// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"log"
	"github.com/ruslanBik4/httpgo/models/db/schema"
)

// return schema for render standart methods
func (qb *QueryBuilder) GetFields() (schTable QBTable) {

	schTable.Fields = make(map[string] *QBField, 0)

	if len(qb.fields) == 0 {
		for _, table := range qb.Tables {
			for _, fieldStrc := range table.schema.Rows {

				defer func() {
					result := recover()
					switch err := result.(type) {
					case schema.ErrNotFoundTable:
						log.Println(table, fieldStrc)
					case error:
						panic(err)
					case nil:
					}
				}()
				table.AddField("", fieldStrc.COLUMN_NAME)
				//field := &QBField{Name: fieldStrc.COLUMN_NAME, schema: fieldStrc, Table: table}
				//field.Alias = field.Name
				qb.fields = append(qb.fields, table.Fields[fieldStrc.COLUMN_NAME])
				//if fieldStrc.TABLEID {
				//	field.ChildQB = CreateEmpty()
				//	field.ChildQB.AddTable("p", field.schema.TableProps)
				//}
			}
		}

	}
	qb.checkSurrogateFields()

	for _, field := range qb.fields {
		schTable.Fields[field.Name] = field
	}

	for _, table := range qb.Tables {
		schTable.Name += " " + table.Join + table.Name
	}

	return schTable
}
func (qb *QueryBuilder) checkSurrogateFields() {
	for _, field := range qb.fields {
		if field.schema.IsHidden {
			continue
		} else if field.schema.SETID || field.schema.NODEID || field.schema.IdForeign {
			field.getSelectedValues()
		} else if field.schema.TABLEID {
			//field.ChildrenFields = schema.GetFieldsTable(field.schema.TableProps)
			//qb.checkSurrogateFields(&(*fields)[idx].ChildrenFields.Rows)

		}
	}
}


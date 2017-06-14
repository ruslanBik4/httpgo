// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"log"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/logs"
)

// return schema for render standart methods
func (qb *QueryBuilder) GetFields() (schTable QBTable) {


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
				qb.fields = append(qb.fields, table.Fields[fieldStrc.COLUMN_NAME])
			}
			logs.StatusLog(table)
		}
	}
	qb.checkSurrogateFields()

	schTable.Fields = make(map[string] *QBField, len(qb.fields))
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
			//field.checkSurrogateFields(&(*fields)[idx].ChildrenFields.Rows)

		}
	}
}


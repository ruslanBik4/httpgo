// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"strings"
)

type FieldsTable struct {
	Name string
	ID   int
	Comment string
	IsDadata bool
	Rows [] *FieldStructure
	Hiddens map[string] string
	SaveFormEvents 	map[string] string
	DataJSOM        map[string] interface{}
}

func (table *FieldsTable) FindField(name string) *FieldStructure {
	for _, field := range table.Rows {
		if field.COLUMN_NAME == name {
			return field
		}
	}

	return nil
}
func (table *FieldsTable) FillSurroggateFields(tableName string)  {

	for idx, fieldStrc := range table.Rows {

		//fieldStrc := &(table.Rows[idx])
		fieldStrc.ParseComment(fieldStrc.COLUMN_COMMENT)

		// TODO: refatoring this later - учитывать момент того, что попутных таблтиц еще может не быт в кеше
		if strings.HasPrefix(fieldStrc.COLUMN_NAME, "setid_") {
			fieldStrc.SETID = true
			fieldStrc.TableProps  = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "setid_")
			fieldStrc.TableValues = fieldStrc.Table.Name + "_" + table.Rows[idx].TableProps + "_has"
			fieldStrc.setEnumValues()

		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "nodeid_") {
			fieldStrc.NODEID = true
			fieldStrc.TableValues  = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "nodeid_")
			fieldStrc.setEnumValues()
		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "tableid_"){
			fieldStrc.TABLEID = true
			fieldStrc.TableProps  = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "tableid_")
		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "id_") {
			fieldStrc.IdForeign = true
			fieldStrc.TableProps  = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "id_")
			//table.Rows[idx].SQLforFORMList = "SELECT id, " + fieldStrc.GetForeignFields() + " FROM " + table.Rows[idx].TableProps
		}

	}
}
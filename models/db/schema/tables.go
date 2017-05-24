// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import "strings"

type FieldsTable struct {
	Name string
	ID   int
	Comment string
	IsDadata bool
	Rows [] FieldStructure
	Hiddens map[string] string
	SaveFormEvents 	map[string] string
	DataJSOM        map[string] interface{}
}

func (table *FieldsTable) FindField(name string) *FieldStructure {
	for idx, field := range table.Rows {
		if field.COLUMN_NAME == name {
			return &table.Rows[idx]
		}
	}

	return nil
}
func (table *FieldsTable) FillSurroggateFields()  {
	for idx, fieldStrc := range table.Rows {

		// TODO: refatoring this later - учитывать момент того, что попутных таблтиц еще может не быт в кеше
		if strings.HasPrefix(fieldStrc.COLUMN_NAME, "setid_") {
			table.Rows[idx].SETID = true
			table.Rows[idx].TableProps  = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "setid_")
			table.Rows[idx].TableValues = fieldStrc.Table.Name + "_" + table.Rows[idx].TableProps + "_has"
			table.Rows[idx].SelectValues = make(map[int] string, 0)
			table.Rows[idx].setEnumValues()
			table.Rows[idx].writeSQLbySETID()

		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "nodeid_") {
			table.Rows[idx].NODEID = true
			table.Rows[idx].TableValues  = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "nodeid_")
			table.Rows[idx].SelectValues = make(map[int] string, 0)
			table.Rows[idx].setEnumValues()
			table.Rows[idx].writeSQLByNodeID()
		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "tableid_"){
			table.Rows[idx].TABLEID = true
			table.Rows[idx].TableProps  = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "tableid_")
			table.Rows[idx].writeSQLByTableID()
		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "id_") {
			table.Rows[idx].IdForeign = true
			table.Rows[idx].TableProps  = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "id_")
			table.Rows[idx].SQLforFORMList = "SELECT id, " + fieldStrc.GetForeignFields() + " FROM " + table.Rows[idx].TableProps
			table.Rows[idx].SelectValues = make(map[int] string, 0)
		}

	}
}
// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"strings"
)

// FieldsTable for fields parameters in form
type FieldsTable struct {
	Name           string
	ID             int
	Comment        string
	IsDadata       bool
	Rows           []*FieldStructure
	Hiddens        map[string]string
	SaveFormEvents map[string]string
	DataJSOM       map[string]interface{}
}

// FindColumn search field in table by {name} & return structure field
func (table *FieldsTable) FindField(name string) *FieldStructure {
	for _, field := range table.Rows {
		if field.COLUMN_NAME == name {
			return field
		}
	}

	return nil
}

// FillSurroggateFields fill surrogate field property for API
// TODO: заменить строки имена таблиц на ссылку на схему
func (table *FieldsTable) FillSurroggateFields(tableName string) {

	for idx, fieldStrc := range table.Rows {

		fieldStrc.ParseComment(fieldStrc.COLUMN_COMMENT)

		// TODO: refatoring this later - учитывать момент того, что попутных таблтиц еще может не быт в кеше
		if strings.HasPrefix(fieldStrc.COLUMN_NAME, "setid_") {
			fieldStrc.SETID = true
			fieldStrc.TableProps = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "setid_")
			fieldStrc.TableValues = fieldStrc.Table.Name + "_" + table.Rows[idx].TableProps + "_has"
			fieldStrc.setEnumValues()

		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "nodeid_") {
			fieldStrc.NODEID = true
			fieldStrc.TableValues = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "nodeid_")

			TableValues := GetFieldsTable(fieldStrc.TableValues)
			//TODO: later refactoring - store values in field propertyes
			for _, field := range TableValues.Rows {
				if strings.HasPrefix(field.COLUMN_NAME, "id_") && (field.COLUMN_NAME != "id_"+fieldStrc.Table.Name) {
					fieldStrc.TableProps = field.COLUMN_NAME[3:]
					fieldStrc.ForeignFields = field.GetForeignFields()
					break
				}
			}

			if (fieldStrc.TableProps == "") || (fieldStrc.ForeignFields == "") {
				panic(ErrNotFoundTable{Table: fieldStrc.TableValues})
			}
			fieldStrc.setEnumValues()
		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "tableid_") {
			fieldStrc.TABLEID = true
			fieldStrc.TableProps = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "tableid_")
		} else if strings.HasPrefix(fieldStrc.COLUMN_NAME, "id_") {
			fieldStrc.IdForeign = true
			fieldStrc.TableProps = strings.TrimPrefix(fieldStrc.COLUMN_NAME, "id_")
		}

	}
}

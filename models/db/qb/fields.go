// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"strings"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/ruslanBik4/httpgo/models/db"
	"database/sql"
)
// getters
func (field *QBField) GetSchema() *schema.FieldStructure {
	return field.schema
}
// Scan implements the Scanner interface.
func (field *QBField) Scan(value interface{}) error {
	var temp sql.NullString

	if err := temp.Scan(value); err != nil {
		field.Value = ""
		return err
	}
	field.Value = temp.String

	return nil
}

// adding fields into qB
func (table *QBTable) AddFields(fields map[string] string) *QBTable {
	for alias, name := range fields {
		table.AddField(alias, name)
	}

	return table
}
// add field and returns table object
func (table *QBTable) AddField(alias, name string) *QBTable {

	if strings.Contains(name, " AS ") {
		pos := strings.Index(name, " AS ")
		alias = name[ pos + 4 : ]
		name  = name[: pos]
	} else if alias == ""  {
		alias = name
	}

	field := &QBField{Name: name, Alias: alias}
	table.Fields[alias] = field
	defer schemaError()

	field.schema = 	table.getFieldSchema(field.Name)

	if field.schema == nil {
		field.schema = &schema.FieldStructure{COLUMN_NAME: alias, COLUMN_TYPE: "calc"}
	} else {
		field.SelectValues = make(map[int] string, 0)

	}
	table.qB.fields = append(table.qB.fields, field)

	table.qB.Aliases = append(table.qB.Aliases, alias)

	return table
}



// return schema for render standart methods
func (qb *QueryBuilder) GetFields() (schTable QBTable) {

	schTable.Fields = make(map[string] *QBField, 0)

	if len(qb.fields) == 0 {
		for _, table := range qb.Tables {
			for _, fieldStrc := range table.schema.Rows {

				field := &QBField{Name: fieldStrc.COLUMN_NAME, schema: fieldStrc}
				qb.fields = append(qb.fields, field)
			}
		}

	}
	for _, field := range qb.fields {
		schTable.Fields[field.Name] = field
	}

	logs.StatusLog(schTable.Fields)
	for _, table := range qb.Tables {
		schTable.Name += " " + table.Join + table.Name
	}

	qb.checkSurrogateFields()
	return schTable
}
func (qb *QueryBuilder) checkSurrogateFields() {
	for _, field := range qb.fields {
		if field.schema.IsHidden {
			continue
		} else if field.schema.SETID || field.schema.NODEID || field.schema.IdForeign {
			//field.SelectValues = qb.putSelectValues(idx, field)
		} else if field.schema.TABLEID {
			//field.ChildrenFields = schema.GetFieldsTable(field.schema.TableProps)
			//qb.checkSurrogateFields(&(*fields)[idx].ChildrenFields.Rows)

		}
	}
}
func (field *QBField) putSelectValues(idx int) map[int] string {

		sqlCommand := field.SQLforFORMList
		comma      := " WHERE "
		for _, enumVal := range field.schema.EnumValues {
			if i := strings.Index(enumVal, ":"); i > 0 {
				// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
				if paramValue, ok := field.Table.qB.FieldsParams[enumVal[i+1:]]; ok  {
					enumVal = enumVal[:i] + fmt.Sprintf("%s", paramValue[0])
					sqlCommand += comma + enumVal
					comma = " OR "
				} else {
					continue
				}
			}

		}

		if field.schema.Where > "" {
			if i := strings.Index(field.schema.Where, ":"); i > 0 {
				// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
				param, suffix := field.schema.Where[i+1:], ""
				// считаем, что окончанием параметра могут быть символы ", )"
				j := strings.IndexAny(param, ", )")
				if j > 0 {
					suffix= param[j:]
					param = param[:j]
				}
				if paramValue, ok := field.Table.qB.FieldsParams[param]; ok {
					sqlCommand += comma + field.schema.Where[:i] + fmt.Sprintf("%s", paramValue[0]) + suffix
				}
			} else {
				sqlCommand += comma + field.schema.Where
			}

			logs.DebugLog("where for field " + field.schema.Where, sqlCommand)
		}
		//TODO: add where condition
		logs.DebugLog("sql for field " + field.Name, sqlCommand)
		rows, err := db.DoSelect(sqlCommand)
		if err != nil {
			logs.ErrorLog(err, field.SQLforFORMList)
		} else {

			defer rows.Close()
			for rows.Next() {
				var key int
				var title string
				if err := rows.Scan(&key, &title); err != nil {
					logs.ErrorLog(err, key)
				}

				field.SelectValues[key] = title
			}
		}


	return field.SelectValues
}
// put param
func (field *QBField) parseEnumValue(enumVal string) string {
	if i := strings.Index(enumVal, ":"); i > 0 {
		param, suffix := enumVal[i+1:], ""
		// считаем, что окончанием параметра могут быть символы ", )"
		j := strings.IndexAny(param, ", )")
		if j > 0 {
			suffix= param[j:]
			param = param[:j]
		}
		// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
		if paramField, ok := field.Table.Fields[param]; ok && (paramField.Value != "") {
			return enumVal[:i] + paramField.Value + suffix
		} else {
			if paramValue, ok := field.Table.qB.FieldsParams[param]; ok  {
				return enumVal[:i] + paramValue[0] + suffix
			}
		}
	}

	return ""
}
// todo: проверить работу
// create where for  query from SETID_ / NODEID_ / TABLEID_ fields
// условия вынимаем из определения поля типа SET
// и все условия оборачиваем в скобки для того, что бы потом можно было навесить еще усло
func (field *QBField) WhereFromSet() (result string) {

	defer schemaError()
	comma  := ""
	for _, enumVal := range field.schema.EnumValues {
		result += comma + field.parseEnumValue(enumVal)
		comma = " OR "
	}

	if (result > "") && (result != "1") {

		return " WHERE (" + result + ")"
	}

	return ""
}

func (field *QBField) writeSQLbySETID() error {

	where := field.WhereFromSet()
	fieldStrc := field.schema

	field.SQLforFORMList = fmt.Sprintf(`SELECT p.id, %s
		FROM %s p LEFT JOIN %s v
		ON (p.id = v.id_%[2]s) ` + where, fieldStrc.GetForeignFields(),
		fieldStrc.TableProps, fieldStrc.TableValues, fieldStrc.Table.Name)

	field.SQLforDATAList = fmt.Sprintf(`SELECT p.id, %s
		FROM %s p JOIN %s v
		ON (p.id = v.id_%[2]s AND v.id_%[4]s=?)` + where, fieldStrc.GetForeignFields(),
		fieldStrc.TableProps, fieldStrc.TableValues, fieldStrc.Table.Name)
	return nil
}
//getSQLFromNodeID(field *schema.FieldStructure) string
func (field *QBField) writeSQLByNodeID() (err error){
	var titleField string


	defer schemaError()

	//if field.schema.TableProps == "" {
	//	fieldsValues := GetFieldsTable(fieldStrc.TableValues)
	//
	//	for _, field := range fieldsValues.Rows {
	//		if strings.HasPrefix(field.COLUMN_NAME, "id_") && (field.COLUMN_NAME != "id_"+fieldStrc.Table.Name) {
	//			fieldStrc.TableProps = field.COLUMN_NAME[3:]
	//			titleField = field.GetForeignFields()
	//			break
	//		}
	//	}
	//}

	where := field.WhereFromSet()

	field.SQLforFORMList =  fmt.Sprintf(`SELECT p.id, %s
		FROM %s p LEFT JOIN %s v
		ON (p.id = v.id_%[2]s) ` + where,
		titleField,  field.schema.TableProps, field.schema.TableValues)

	field.SQLforDATAList =  fmt.Sprintf(`SELECT p.id, %s
		FROM %s p JOIN %s v
		ON (p.id = v.id_%[2]s AND v.id_%[4]s=?) ` + where,
		titleField, field.schema.TableProps, field.schema.TableValues, field.schema.Table.Name)

	return nil
}

func (field *QBField) writeSQLByTableID() error {


	where := field.WhereFromSet()
	if where > "" {
		where += " OR (id_%s=?)"
	} else {
		where = " WHERE (id_%s=?)"
	}

	field.SQLforFORMList =  fmt.Sprintf( `SELECT * FROM %s p ` + where, field.schema.TableProps, field.schema.Table.Name )

	return nil
}

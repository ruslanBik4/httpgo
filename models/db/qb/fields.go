// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"strings"
	"errors"
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

	field := &QBField{Name: name, Alias: alias, Table: table}
	table.Fields[alias] = field
	defer schemaError()

	field.schema = 	table.getFieldSchema(field.Name)

	if field.schema == nil {
		field.schema = &schema.FieldStructure{COLUMN_NAME: alias, COLUMN_TYPE: "calc"}
	} else {
		field.SelectValues = make(map[int] string, 0)
		// для TABLEID_ создадим таблицу свойств и заполним полями!
		if field.schema.TABLEID {
			field.ChildQB = CreateEmpty()
			field.ChildQB.AddTable("p", field.schema.TableProps)
		} else if field.schema.SETID {
			field.ChildQB = CreateEmpty()
			titleField := field.schema.GetForeignFields()

			field.ChildQB.AddTable( "p", field.schema.TableProps ).AddField("", "id").AddField("", titleField)

			onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.schema.TableProps, field.Table.Name )
			field.ChildQB.Join ( "v", field.schema.TableValues, onJoin ).AddField("", "id_" + field.Table.Name)

		} else if field.schema.NODEID {

			titleField := field.schema.GetForeignFields()
			field.ChildQB = CreateEmpty()
			field.ChildQB.AddTable( "p", field.schema.TableProps ).AddField("", "id").AddField("", titleField)

			onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.schema.TableProps, field.Table.Name )
			field.ChildQB.JoinTable ( "v", field.schema.TableValues, "JOIN", onJoin ).AddField("", "id_" + field.Table.Name)
		} else if field.schema.IdForeign {
				field.getSelectedValues()
		}

	}
	//table.qB.fields = append(table.qB.fields, field)


	return table
}

func (field *QBField) getSelectedValues() {

	defer func() {
		result := recover()
		switch err := result.(type) {
		case schema.ErrNotFoundTable:
			logs.ErrorLogHandler(err, err.Table, field.Name, field.Table.Name)
			panic(err)
		case nil:
		case error:
			panic(err)
		}

	}()
	field.ChildQB = CreateEmpty()
	titleField := field.schema.GetForeignFields()

	field.ChildQB.AddTable( "", field.schema.TableProps ).AddField("", "id").AddField("", titleField)

	rows, err := field.ChildQB.GetDataSql()
	if err != nil {
		logs.ErrorLog(err, field.ChildQB)
	} else {
		for rows.Next() {
			var id int
			var title string
			rows.Scan(&id, &title)
			field.SelectValues[id] = title
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
// put param Value
func (field *QBField) putEnumValueToArgs() error {
	for _, enumVal := range field.schema.EnumValues {
		if i := strings.Index(enumVal, ":"); i > 0 {
			param := enumVal[i+1:]
			// считаем, что окончанием параметра могут быть символы ", )"
			j := strings.IndexAny(param, ", )")
			if j > 0 {
				param = param[:j]
			}
			// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
			if paramField, ok := field.Table.Fields[param]; ok && (paramField.Value != "") {
				field.ChildQB.AddArgs( paramField.Value )
			} else if paramValue, ok := field.Table.qB.FieldsParams[param]; ok {
					field.ChildQB.AddArgs( paramValue[0] )
			} else {
				return errors.New( "not enougth parameter")
			}
		}
	}

	return nil
}
// parse enumValues & insert queryes parameters
func (field *QBField) parseEnumValue(enumVal string) string {
	if i := strings.Index(enumVal, ":"); i > 0 {
		param, suffix := enumVal[i+1:], ""
		// считаем, что окончанием параметра могут быть символы ", )"
		j := strings.IndexAny(param, ", )")
		if j > 0 {
			suffix= param[j:]
			param = param[:j]
		}
		return enumVal[:i] + "?" + suffix
	}

	return enumVal
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

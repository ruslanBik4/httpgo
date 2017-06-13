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
			field.ChildQB = Create(fmt.Sprintf( "id_%s=?", field.Table.Name ), "", "")
			field.ChildQB.AddTable("p", field.schema.TableProps)
			field.ChildQB.FieldsParams = table.qB.FieldsParams
			field.ChildQB.AddArg(0)
		} else if field.schema.SETID {
			field.ChildQB = CreateEmpty()
			titleField := field.schema.GetForeignFields()

			field.ChildQB.AddTable( "p", field.schema.TableProps ).AddField("", "id").AddField("", titleField)

			onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.schema.TableProps, field.Table.Name )
			field.ChildQB.Join ( "v", field.schema.TableValues, onJoin ).AddField("", "id_" + field.Table.Name)
			field.ChildQB.FieldsParams = table.qB.FieldsParams
			field.ChildQB.AddArg(0)

		} else if field.schema.NODEID {

			titleField := field.schema.GetForeignFields()
			field.ChildQB = CreateEmpty()
			field.ChildQB.AddTable( "p", field.schema.TableProps ).AddField("", "id").AddField("", titleField)

			onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.schema.TableProps, field.Table.Name )
			field.ChildQB.JoinTable ( "v", field.schema.TableValues, "JOIN", onJoin ).AddField("", "id_" + field.Table.Name)
			field.ChildQB.FieldsParams = table.qB.FieldsParams
			field.ChildQB.AddArg(0)
		} else if field.schema.IdForeign {
				// уже не нужно, но надо перепроверить!!!
				//field.getSelectedValues()
		}


	}
	//table.qB.fields = append(table.qB.fields, field)


	return table
}
// TODO: local qb
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

	field.ChildQB = Create( field.WhereFromSet(), "", "" )

	titleField := field.schema.GetForeignFields()

	field.ChildQB.AddTable( "", field.schema.TableProps ).AddField("", "id").AddField("", titleField)

	rows, err := field.ChildQB.GetDataSql()
	if err != nil {
		logs.ErrorLog(err, field.Name,  field.ChildQB)
	} else {
		for rows.Next() {
			var id int
			var title string
			rows.Scan(&id, &title)
			field.SelectValues[id] = title
		}
	}

}
func (field *QBField) putValueToArgs(param string) error {
		// считаем, что окончанием параметра могут быть символы ", )"
		// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
		if paramField, ok := field.Table.Fields[param]; ok && (paramField.Value != "") {
			field.ChildQB.AddArgs( paramField.Value )
		} else if paramValue, ok := field.Table.qB.FieldsParams[param]; ok {
			field.ChildQB.AddArgs( paramValue[0] )
		} else if param == "id_users" {
			field.ChildQB.AddArgs( 0 )
		} else {
			return errors.New( "not enougth parameter")
		}


	return nil
}
// parse enumValues & insert queryes parameters
func (field *QBField) parseEnumValue(enumVal string) (condition, param string) {
	if i := strings.Index(enumVal, ":"); i > 0 {
		paramName, suffix := enumVal[i+1:], ""
		// считаем, что окончанием параметра могут быть символы ", )"
		j := strings.IndexAny(paramName, ", )")
		if j > 0 {
			suffix    = paramName[j:]
			paramName = paramName[:j]
		}
		condition = enumVal[:i] + "?" + suffix
	} else {
		condition = enumVal
	}

	return condition, param
}
// todo: проверить работу
// create where for  query from SETID_ / NODEID_ / TABLEID_ fields
// условия вынимаем из определения поля типа SET
// и все условия оборачиваем в скобки для того, что бы потом можно было навесить еще условие
func (field *QBField) WhereFromSet() (result string) {

	defer schemaError()
	comma  := ""
	for _, enumVal := range field.schema.EnumValues {
		condition, param :=	field.parseEnumValue(enumVal)
		if param > "" {
			field.putValueToArgs(param)
		} else if condition == "1" {
			continue
		}
		result += comma + condition
		comma = " OR "
	}

	if field.schema.Where > "" {

		condition, param := field.parseEnumValue( field.schema.Where )
		if param > "" {
			field.putValueToArgs(param)
		}
		if (result > "") {

			result = "(" + result + ") AND "
		}

		return result + condition
	}

	return result
}

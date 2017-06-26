// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"database/sql"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/logs"
	"strings"
)

// for compatabilies interface logsType
func (field QBField) String() string {
	mess := "&QBField{Name: " + field.Name + ", Alias: " + field.Alias + ", Table: " + field.Table.Name + ", SelectValues: "
	for key, value := range field.SelectValues {
		mess += fmt.Sprintf("%d=%s", key, value)
	}
	mess += fmt.Sprintf(" SelectQB: %v", field.SelectQB)

	return mess + "}"
}

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
func (table *QBTable) AddFields(fields map[string]string) *QBTable {
	for alias, name := range fields {
		table.AddField(alias, name)
	}

	return table
}

// add field and returns table object
func (table *QBTable) AddField(alias, name string) *QBTable {

	if strings.Contains(name, " AS ") {
		pos := strings.Index(name, " AS ")
		alias = name[pos+4:]
		name = name[:pos]
	} else if alias == "" {
		alias = name
	}

	field := &QBField{Name: name, Alias: alias, Table: table}
	table.Fields[alias] = field
	defer schemaError()

	field.schema = table.getFieldSchema(field.Name)

	if field.schema == nil {
		field.schema = &schema.FieldStructure{COLUMN_NAME: alias, COLUMN_TYPE: "calc"}
	} else {
		// для TABLEID_ создадим таблицу свойств и заполним полями!
		if field.schema.TABLEID {
			field.ChildQB = Create(fmt.Sprintf("id_%s=?", field.Table.Name), "", "")
			tableProps := field.ChildQB.AddTable("p", field.schema.TableProps)
			for _, fieldStruct := range tableProps.schema.Rows {
				if fieldStruct.COLUMN_NAME == "id_"+field.Table.Name {
					continue
				}
				tableProps.AddField("", fieldStruct.COLUMN_NAME)
			}
		} else if field.schema.SETID {
			field.ChildQB = CreateEmpty()
			titleField := field.schema.GetForeignFields()

			field.ChildQB.AddTable("p", field.schema.TableProps).AddField("", "id").AddField("", titleField)

			onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.schema.TableProps, field.Table.Name)
			field.ChildQB.Join("v", field.schema.TableValues, onJoin).AddField("", "id_"+field.Table.Name)
		} else if field.schema.NODEID {

			titleField := field.schema.GetForeignFields()
			field.ChildQB = CreateEmpty()
			field.ChildQB.AddTable("p", field.schema.TableProps).AddField("", "id").AddField("", titleField)

			onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.schema.TableProps, field.Table.Name)
			field.ChildQB.JoinTable("v", field.schema.TableValues, "JOIN", onJoin).AddField("", "id_"+field.Table.Name)
		} else if field.schema.IdForeign {
			// уже не нужно, но надо перепроверить!!!
			//field.getSelectedValues()
		}

		if field.ChildQB != nil {

			field.ChildQB.PostParams = table.qB.PostParams
			field.ChildQB.parent = table.qB
			field.ChildQB.AddArg(0)
		}

	}

	return table
}

// TODO: local field
func (field *QBField) getSelectedValues() {

	defer func() {
		result := recover()
		switch err := result.(type) {
		case schema.ErrNotFoundTable:
			logs.ErrorLogHandler(err, err.Table, field.Name, field.Table.Name)
			panic(err)
		case nil:
		case ErrNotFoundParam:
			logs.ErrorLog(err, field, field.SelectQB)
		case error:
			logs.ErrorLog(err, field, field.SelectQB)
		}

	}()

	// создаем дочерний запрос
	field.SelectQB = CreateEmpty()

	titleField := field.schema.GetForeignFields()

	field.SelectQB.AddTable("", field.schema.TableProps).AddField("", "id").AddField("", titleField)

	// подключаем параметры POST-запроса от старшего запроса field
	field.SelectQB.PostParams = field.Table.qB.PostParams
	if field.SelectQB.PostParams == nil {
		logs.DebugLog(field.Name, field.Table)
	}
	// разбираем заменяемые параметры
	field.SelectQB.Where = field.parseWhereANDputArgs()

	rows, err := field.SelectQB.GetDataSql()
	if err != nil {
		logs.ErrorLog(err, field.Name, field.SelectQB)
	} else {
		field.SelectValues = make(map[int]string, 2)
		for rows.Next() {
			var id int
			var title string
			rows.Scan(&id, &title)
			field.SelectValues[id] = title
		}
	}

}

// ищет параметры в родительских запросах
func findParamInParent(QBparent *QueryBuilder, param string) string {
	if QBparent != nil {
		for _, table := range QBparent.Tables {
			if paramField, ok := table.Fields[param]; ok && (paramField.Value > "") {
				return paramField.Value
			}

			return findParamInParent(QBparent.parent, param)
		}
	}

	return ""
}

// locate in field table & post params PARAM & return her value
func (field *QBField) putValueToArgs(param string) string {
	// считаем, что окончанием параметра могут быть символы ", )"
	// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
	if paramField, ok := field.Table.Fields[param]; ok && (paramField.Value > "") {
		return paramField.Value
	} else if paramValue, ok := field.Table.qB.PostParams[param]; ok {
		return paramValue[0]
	} else if param == "id_users" {
		return "0"
	} else if paramValue := findParamInParent(field.Table.qB.parent, param); paramValue > "" {
		return paramValue
	} else {
		panic(&ErrNotFoundParam{Param: "not enougth parameter-" + param})
	}

	return ""
}

// parse enumValues & insert queryes parameters
func (field *QBField) parseEnumValue(enumVal string) string {
	if i := strings.Index(enumVal, ":"); i > 0 {
		param, suffix := enumVal[i+1:], ""
		// считаем, что окончанием параметра могут быть символы ", )"
		j := strings.IndexAny(param, ", )")
		if j > 0 {
			suffix = param[j:]
			param = param[:j]
		}
		field.SelectQB.AddArgs(field.putValueToArgs(param))

		return enumVal[:i] + "?" + suffix
	}
	return enumVal
}

// todo: проверить работу
// create where for  query from SETID_ / NODEID_ / TABLEID_ fields
// условия вынимаем из определения поля типа SET
// и все условия оборачиваем в скобки для того, что бы потом можно было навесить еще условие
func (field *QBField) parseWhereANDputArgs() (result string) {

	comma := ""
	for _, enumVal := range field.schema.EnumValues {
		if enumVal == "1" {
			continue
		}

		result += comma + field.parseEnumValue(enumVal)
		comma = " OR "
	}

	if field.schema.Where > "" {

		if result > "" {

			return "(" + result + ") AND " + field.parseEnumValue(field.schema.Where)
		}

		return field.parseEnumValue(field.schema.Where)
	}

	return result
}

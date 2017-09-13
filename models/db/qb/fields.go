// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"fmt"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/logs"
	"strconv"
	"strings"
)

// String for compatabilies interface LogsType
func (field QBField) String() string {
	mess := "&QBField{Name: " + field.Name + ", Alias: " + field.Alias + ", Table: " + field.Table.Name + ", SelectValues: "
	for key, value := range field.SelectValues {
		mess += fmt.Sprintf("%d=%s", key, value)
	}
	mess += fmt.Sprintf(" SelectQB: %v", field.SelectQB)

	return mess + "}"
}

// GetSchema возвращает схему поля, взятую из БД
func (field *QBField) GetSchema() *schema.FieldStructure {
	return field.Schema
}

// GetNativeValue возвращает значение поля типа соответствующего типы поля в БД
func (field QBField) GetNativeValue(tinyAsBool bool) interface{} {
	if field.Value == nil {
		return nil
	}
	switch dataType := field.Schema.DATA_TYPE; dataType {
	case "char", "varchar", "text", "date", "datetime", "timestamp", "time", "set", "enum":
		return string(field.Value)
	case "tinyint", "int", "uint", "int64":
		value, err := strconv.Atoi(string(field.Value))
		if err != nil {
			logs.ErrorLog(err, "convert RawBytes ", field.Value)
			return nil
		}
		// если нужно вернуть булево значение из поля типа tinyint
		if tinyAsBool && (dataType == "tinyint") {
			return value == 1
		}

		return value
	case "float", "double":
		value, err := strconv.ParseFloat(string(field.Value), 64)
		if err != nil {
			logs.ErrorLog(err, "convert RawBytes ", field.Value)
			return nil
		}
		return value
	}

	return field.Value

}

// AddFields adding fieldsfrom map into qB
func (table *QBTable) AddFields(fields map[string]string) *QBTable {
	for alias, name := range fields {
		table.AddField(alias, name)
	}

	return table
}

// AddField add field and returns table object
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

	field.Schema = table.getFieldSchema(field.Name)

	if field.Schema == nil {
		field.Schema = &schema.FieldStructure{COLUMN_NAME: alias, COLUMN_TYPE: "calc"}
		//	для агрегатных полей спрогнозируем тип
		if strings.Contains(name, "COUNT") {
			field.Schema.DATA_TYPE = "uint"
		} else if strings.Contains(name, "SUM") {
			field.Schema.DATA_TYPE = "double"
		} else {
			field.Schema.DATA_TYPE = "char"
		}
	} else {
		// для TABLEID_ создадим таблицу свойств и заполним полями!
		if field.Schema.TABLEID {
			field.ChildQB = Create(fmt.Sprintf("id_%s=?", field.Table.Name), "", "")
			tableProps := field.ChildQB.AddTable("p", field.Schema.TableProps)
			for _, fieldStruct := range tableProps.schema.Rows {
				if fieldStruct.COLUMN_NAME == "id_"+field.Table.Name {
					continue
				}
				tableProps.AddField("", fieldStruct.COLUMN_NAME)
			}
		} else if field.Schema.SETID {
			field.ChildQB = CreateEmpty()
			titleField := field.Schema.GetForeignFields()

			field.ChildQB.AddTable("p", field.Schema.TableProps).AddField("", "id").AddField("", titleField)

			onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.Schema.TableProps, field.Table.Name)
			field.ChildQB.Join("v", field.Schema.TableValues, onJoin).AddField("", "id_"+field.Table.Name)
		} else if field.Schema.NODEID {

			titleField := field.Schema.GetForeignFields()
			field.ChildQB = CreateEmpty()
			field.ChildQB.AddTable("p", field.Schema.TableProps).AddField("", "id").AddField("", titleField)

			onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.Schema.TableProps, field.Table.Name)
			field.ChildQB.JoinTable("v", field.Schema.TableValues, "JOIN", onJoin).AddField("", "id_"+field.Table.Name)
		} else if field.Schema.IdForeign {
			// уже не нужно, но надо перепроверить!!!
			//field.GetSelectedValues()
		}

		if field.ChildQB != nil {

			field.ChildQB.PostParams = table.qB.PostParams
			field.ChildQB.parent = table.qB
			field.ChildQB.AddArg(0)
		}

	}

	return table
}

// GetSelectedValues записываем лист значений для поля, чтобы показывать список на форме
func (field *QBField) GetSelectedValues() {

	defer func() {
		result := recover()
		switch err := result.(type) {
		case schema.ErrNotFoundTable:
			logs.ErrorLogHandler(err, err.Table, field.Name, field.Table.Name)
			panic(err)
		case nil:
		case *ErrNotFoundParam:
			logs.ErrorLog(err, field, field.SelectQB, field.Table.Name)
			logs.ErrorStack()
		case error:
			logs.ErrorLog(err, field, field.SelectQB)
		}

	}()

	titleField := field.Schema.GetForeignFields()
	// создаем дочерний запрос
	field.SelectQB = CreateFromSQL(fmt.Sprintf("SELECT id, %s FROM %s", titleField, field.Schema.TableProps))

	// подключаем параметры POST-запроса от старшего запроса field
	field.SelectQB.PostParams = field.Table.qB.PostParams
	if field.SelectQB.PostParams == nil {
		logs.DebugLog(field.Name, field.Table)
	}
	// разбираем заменяемые параметры
	field.SelectQB.SetWhere(field.parseWhereANDputArgs())

	if rows, err := field.SelectQB.GetDataSql(); err != nil {
		logs.ErrorLog(err, field.Name, field.SelectQB)
	} else {
		field.SelectValues = make(map[int]string, 0)
		for rows.Next() {
			var id int
			var title string
			err = rows.Scan(&id, &title)
			if err != nil {
				logs.ErrorLog(err, "get SelectedValues for field", field)
				field.SelectValues[0] = err.Error()

				if strings.Contains(err.Error(), "unsupported Scan") {
					break
				}
				continue
			}
			field.SelectValues[id] = title
		}
	}

}

// ищет параметры в родительских запросах
func findParamInParent(QBparent *QueryBuilder, param string) string {
	if QBparent != nil {
		for _, table := range QBparent.Tables {
			if paramField, ok := table.Fields[param]; ok && (paramField.Value != nil) {
				return string(paramField.Value)
			}

			return findParamInParent(QBparent.parent, param)
		}
	}

	return ""
}

// locate in field table & post params PARAM & return her value
// TODO: id_users get from session
func (field *QBField) putValueToArgs(param string) string {
	// считаем, что окончанием параметра могут быть символы ", )"
	// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
	if paramField, ok := field.Table.Fields[param]; ok && (paramField.Value != nil) {
		return string(paramField.Value)
	} else if paramValue, ok := field.Table.qB.PostParams[param]; ok {
		return paramValue[0]
	} else if param == "id_users" {
		return "1"
	} else if paramValue := findParamInParent(field.Table.qB.parent, param); paramValue > "" {
		return paramValue
	} else {
		panic(&ErrNotFoundParam{Param: "not enougth parameter-" + param})
	}

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
	for _, enumVal := range field.Schema.EnumValues {
		if enumVal == "1" {
			continue
		}

		result += comma + field.parseEnumValue(enumVal)
		comma = " OR "
	}

	if field.Schema.Where > "" {

		if result > "" {

			return "(" + result + ") AND " + field.parseEnumValue(field.Schema.Where)
		}

		return field.parseEnumValue(field.Schema.Where)
	}

	return result
}

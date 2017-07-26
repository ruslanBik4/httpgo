// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multiquery

import (
	"strings"
	"strconv"
	"github.com/ruslanBik4/httpgo/models/logs"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/db/schema"
)
// аргументы для запроса, формируются дирнамичекски по полученным данным
// для этого имеем несколько доп. полей для промежуточных результатов
type ArgsQuery struct {
	Comma, FieldList, Values string
	tableName, parentKey     string
	Args                     []interface{}
	TableValues				 map[int] map [string] []string
	Fields					 []string
	isNotContainParentKey    bool
}
// для подготовки запросов суррогатнызх полей
// их значения мы получаем в одном запросе вместе
//с данными основной таблицы
type MultiQuery struct {
	parentName				string
	Queryes 				map[string]*ArgsQuery
}
// найти в списке имя поля
func (query *ArgsQuery) findField(name string) bool {
	for _, val := range query.Fields {
		if val == name {
			return true
		}
	}

	return false
}
// добавляем запросы для мультиполей, различаем их по таблицам(куда будем делать вставки
// и по строкам (так как для tableid_ может прийти сразу несколько строк данных!
func (tableIDQueryes *MultiQuery) AddNewParam(key string, indSeparator int, val []string, field *schema.FieldStructure) {
	tableName := key[:indSeparator]
	query, ok := tableIDQueryes.Queryes[tableName]
	if !ok {
		query = &ArgsQuery{
			Comma:     "",
			FieldList: "",
			Values:    "",
			tableName: tableName,
			parentKey: "`id_" + tableIDQueryes.parentName + "`",
			TableValues: make( map[int] map [string] []string, 1),
		}
		parentTable := schema.GetParentTable(tableName)
		if parentTable != nil {
			query.parentKey = "`id_" + parentTable.Name + "`"
		}

	}
	fieldName := key[ indSeparator + 1: ]
	pos := strings.Index(fieldName, "[")
	row, err := strconv.Atoi( fieldName[pos+1:len(fieldName)-1] )
	if err != nil {
		logs.ErrorLog(err, fieldName)
		return
	}
	fieldName = "`" + fieldName[:pos] + "`"

	if !query.findField(fieldName) {
		query.Fields = append(query.Fields, fieldName)
	}

	if _, ok := query.TableValues[row]; !ok {
		query.TableValues[row] = make( map[string] []string, 1)
		logs.StatusLog(row)
	}
	query.TableValues[row][fieldName] = val
	query.Comma = ", "
	tableIDQueryes.Queryes[tableName] = query
}
// получаем запрос для вставки данных суррогатных полей
// далее он может быть использован внутри транзакции, например
func (query *ArgsQuery) GetUpdateSQL(idParent int) (string, []interface{}) {

	if !query.findField(query.parentKey) {
		query.Fields = append(query.Fields, query.parentKey)
	}
	params := "(?" + strings.Repeat(",?", len(query.Fields)-1) + ")"
	sqlCommand := fmt.Sprintf("insert into %s (%s) values ", query.tableName, strings.Join(query.Fields, ","))
	sqlCommand += params + strings.Repeat( "," + params, len(query.TableValues)-1) + " ON DUPLICATE KEY UPDATE "

	// готовим дубликаты для записи только неключыевых полей!
	comma := ""
	for _, field := range query.Fields {
		switch field {
		case "id", query.parentKey:
			continue
		default:
			sqlCommand += comma + field + "=VALUES(" + field + ")"
		}
		comma = ","
	}
	var args []interface{}

	for _, field := range query.TableValues {

		for _, name := range query.Fields {
			// последним добавляем вторичный ключ
			if name == query.parentKey {
				args = append(args, idParent)
			} else {
				if value, ok := field[name]; ok{
					args = append(args, value[0])
				} else {
					args = append(args, "DEFAULT" )
				}
			}
		}
	}
	logs.DebugLog(sqlCommand, args)
	return sqlCommand, args
}


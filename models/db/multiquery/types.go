// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multiquery

import (
	"strings"
	"strconv"
	"github.com/ruslanBik4/httpgo/models/logs"
	"fmt"
)

type ArgsQuery struct {
	Comma, FieldList, Values string
	tableName, parentKey     string
	Args                     []interface{}
	TableValues				 map[int] map [string] []string
	Fields					 []string
	isNotContainParentKey    bool
}
type MultiQuery struct {
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
func (tableIDQueryes *MultiQuery) AddNewParam(key string, indSeparator int, val []string) {
	tableName := key[:indSeparator]
	query, ok := tableIDQueryes.Queryes[tableName]
	if !ok {
		query = &ArgsQuery{
			Comma:     "",
			FieldList: "",
			Values:    "",
			tableName: tableName,
			parentKey: key,
			TableValues: make( map[int] map [string] []string, 1),
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
	logs.StatusLog(key)

}
func (query *ArgsQuery) GetUpdateSQL(lastInsertId int) (string, []interface{}) {

	if !query.findField(query.parentKey) {
		query.Fields = append(query.Fields, query.parentKey)
	}
	params := "(?" + strings.Repeat(",?", len(query.Fields)-1) + ")"
	sqlCommand := fmt.Sprintf("replace into %s (%s) values ", query.tableName, strings.Join(query.Fields, ","))
	sqlCommand += params + strings.Repeat( "," + params, len(query.TableValues)-1)
	var args []interface{}

	for _, field := range query.TableValues {

		for _, name := range query.Fields {
			// последним добавляем вторичный ключ
			if name == query.parentKey {
				args = append(args, lastInsertId)
			} else {
				args = append(args, field[name][0])
			}
		}
	}
	logs.DebugLog(sqlCommand, args)
	return sqlCommand, args
}

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
)

// return schema for render stadart methods
func (qb *QueryBuilder) GetFields() (schTable schema.FieldsTable) {

	schTable.Rows = qb.fields

	for _, table := range qb.Tables {
		schTable.Name += " " + table.Join + table.Name
	}

	qb.checkSurrogateFields(&qb.fields)
	return schTable
}
func (qb *QueryBuilder) checkSurrogateFields(fields * [] schema.FieldStructure ) {
	for idx, field := range *fields {
		if field.IsHidden {
			continue
		} else if field.SETID || field.NODEID || field.IdForeign {
			(*fields)[idx].SelectValues = qb.putSelectValues(idx, field)
		} else if field.TABLEID {
			(*fields)[idx].ChildrenFields = schema.GetFieldsTable(field.TableProps)
			qb.checkSurrogateFields(&(*fields)[idx].ChildrenFields.Rows)

		}
	}
}
func (qb *QueryBuilder) putSelectValues(idx int, field schema.FieldStructure) map[int] string {

		sqlCommand := field.SQLforFORMList
		comma      := " WHERE "
		for _, enumVal := range field.EnumValues {
			if i := strings.Index(enumVal, ":"); i > 0 {
				// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
				if paramValue, ok := qb.FieldsParams[enumVal[i+1:]]; ok  {
					enumVal = enumVal[:i] + fmt.Sprintf("%s", paramValue)
					sqlCommand += comma + enumVal
					comma = " OR "
				} else {
					continue
				}
			}

		}

		if field.Where > "" {
			if i := strings.Index(field.Where, ":"); i > 0 {
				// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
				param := field.Where[i+1:]
				// считаем, что окончанием параметра могут быть символы ", )"
				j := strings.IndexAny(param, ", )")
				if j > 0 {
					param = param[:j]
				}
				if paramValue, ok := qb.FieldsParams[param]; ok {
					sqlCommand += comma + field.Where[:i] + fmt.Sprintf("%s", paramValue) + field.Where[i+j+1:]
				}
			} else {
				sqlCommand += comma + field.Where
			}

			logs.DebugLog("where for field " + field.Where, sqlCommand)
		}
		//TODO: add where condition
		logs.DebugLog("sql for field " + field.COLUMN_NAME, sqlCommand)
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
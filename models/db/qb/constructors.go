// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"regexp"
)

// constructors
func Create(where, groupBy, orderBy string) *QueryBuilder{

	qb := &QueryBuilder{Where: where, OrderBy: orderBy, GroupBy: groupBy}
	return qb
}
func CreateEmpty() *QueryBuilder{

	qb := &QueryBuilder{}
	return qb
}
func CreateFromSQL(sqlCommand string) *QueryBuilder {
	qb := &QueryBuilder{sqlCommand: sqlCommand}
	var err error
	qb.Prepared, err = db.PrepareQuery(sqlCommand)
	if err != nil {
		panic(err)
	}
	qb.getFrom(sqlCommand)
	qb.getJoins(sqlCommand)
	return qb
}

var compRegEx = regexp.MustCompile(regFrom)

func (qb * QueryBuilder) getFrom(sql string)  {
	var match []string = compRegEx.FindStringSubmatch(sql)

	var tableName, Alias string
	for i, name := range compRegEx.SubexpNames() {
		if i > len(match) {
			break
		}
		if name == "table" {
			tableName = match[i]
		} else if name == "table_alias" {
			Alias = match[i]
		}
	}

	table := qb.AddTable(Alias, tableName)
	for _, fieldStrc := range table.schema.Rows {

		//field := &QBField{Name: fieldStrc.COLUMN_NAME, schema: fieldStrc, Table: table}
		table.AddField("", fieldStrc.COLUMN_NAME )
		//TODO: сделать одно место для добавления полей!
		qb.fields = append(qb.fields, table.Fields[fieldStrc.COLUMN_NAME])
		qb.Aliases = append(qb.Aliases, fieldStrc.COLUMN_NAME)
	}

}
var joinRegEx = regexp.MustCompile(regJoin)

func (qb * QueryBuilder) getJoins(sql string)  {
	var match [][]string = joinRegEx.FindAllStringSubmatch(sql, -1)
	var groupNames []string = joinRegEx.SubexpNames()
	for _, v := range match {
		if qb.getJoin(v, groupNames) {
			//result = append(result, join)
		} else {
			break
		}
	}
}
func (qb * QueryBuilder) getJoin(join []string, groupNames []string)  bool {
	var joinTableName, Alias, onLeftTable, onLeftField, onRightTable, onRightField string
	for i, name := range groupNames {
		switch name {

		case "join_table_name":
			joinTableName = join[i]
		case "join_table_alias":
			Alias = join[i]
		case "on_left_table":
			onLeftTable = join[i]
		case "on_left_field":
			onLeftField = join[i]
		case "on_right_table":
			onRightTable = join[i]
		case "on_right_field":
			onRightField = join[i]
		}
	}

	if (joinTableName > "") && (onLeftTable > "") && (onRightTable > "") && (onRightField > "") {
		qb.Join(Alias, joinTableName, "ON " + onLeftTable + "." + onLeftField + "=" + onRightTable + "." + onRightField)
		return true
	}

	return false
}




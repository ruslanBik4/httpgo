// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package qb has Query Builder for manipulate SQL-queryes & check in databases schema her parameters
package qb

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"regexp"
	"strings"
)

// Create - constructor from some parameters
func Create(where, groupBy, orderBy string) *QueryBuilder {

	qb := &QueryBuilder{Where: where, OrderBy: orderBy, GroupBy: groupBy}
	return qb
}

// CreateEmpty construct empty QueryBuilder
func CreateEmpty() *QueryBuilder {

	qb := &QueryBuilder{}
	return qb
}

// CreateFromSQL construct QueryBuilder from sql-query string
func CreateFromSQL(sqlCommand string) *QueryBuilder {
	qb := &QueryBuilder{sqlCommand: sqlCommand}
	var err error
	qb.Prepared, err = db.PrepareQuery(sqlCommand)
	if err != nil {
		panic(err)
	}

	qb.getFrom(sqlCommand)
	qb.getJoins(sqlCommand)

	if fieldsText, ok := getTextSelectFields(sqlCommand); ok {
		qb.getFields(fieldsText)
	}

	for _, table := range qb.Tables {
		table.addAllFields()
	}
	return qb
}

var compRegEx = regexp.MustCompile(regFrom)

func (qb *QueryBuilder) getFrom(sql string) *QBTable {
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

	return qb.AddTable(Alias, tableName)
}

var joinRegEx = regexp.MustCompile(regJoin)

func (qb *QueryBuilder) getJoins(sql string) {
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
func (qb *QueryBuilder) getJoin(join []string, groupNames []string) bool {
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
		qb.Join(Alias, joinTableName, "ON "+onLeftTable+"."+onLeftField+"="+onRightTable+"."+onRightField)
		return true
	}

	return false
}

var selectRegEx = regexp.MustCompile(regSelect)

func (qb *QueryBuilder) getTextSelectFields(sql string) {
	match := selectRegEx.FindStringSubmatch(sql)
	groupNames := selectRegEx.SubexpNames()

	for i, name := range groupNames {
		if name == "fields" {
			qb.getFields(match[i])
		}
	}

}

var reg = regexp.MustCompile(regField)

func (qb *QueryBuilder) getFields(textFields string) {
	fieldItems := strings.Split(textFields, ",")
	groupNames := reg.SubexpNames()

	for _, text := range fieldItems {
		if field := qb.getField(text, groupNames); field != nil {

			table := qb.Tables[0]
			if field.table > "" {
				table = qb.FindTable(field.table)
			}
			table.AddField(field.alias, field.name)
			if field.alias == "" {
				field.alias = field.name
			}
			qb.fields = append(qb.fields, table.Fields[field.alias])
			qb.Aliases = append(qb.Aliases, field.alias)
		}
	}

}

func (qb *QueryBuilder) getField(text string, groupNames []string) *tSqlField {
	var fieldNote string = strings.TrimSpace(text)
	var elements []string = reg.FindStringSubmatch(fieldNote)

	var funcName = ""
	var fieldTableName = ""
	var fieldName = ""
	var fieldAlias = ""

	for i, name := range groupNames {
		if name == "func_name" {
			if i > 0 && i <= len(elements) {
				funcName = strings.TrimRight(elements[i], ".")
			}
		}
		if name == "field_table" {
			if i > 0 && i <= len(elements) {
				fieldTableName = strings.TrimRight(elements[i], ".")
			}
		}
		if name == "field_name" {
			if i > 0 && i <= len(elements) {
				fieldName = elements[i]
			}
		}
		if name == "alias" {
			if i > 0 && i <= len(elements) {
				fieldAlias = elements[i]
			}
		}
	}

	if fieldTableName == "" && fieldName == "" {
		return nil
	}

	return &tSqlField{fun: funcName, table: fieldTableName, name: fieldName, alias: fieldAlias}
}

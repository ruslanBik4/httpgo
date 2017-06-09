// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// this module has more structures from creating sql-query with relation to db schema

package qb

import (
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"database/sql"
)

type QBField struct {
	Name           string
	Alias          string
        schema         *schema.FieldStructure
	Value          string
	SQLforFORMList string `отдаем в списках полей для формы`
	SQLforDATAList string `отдаем в составе данных`
	SelectValues   map[int] string
	Table          *QBTable
	ChildQB        *QueryBuilder
}
type QBTable struct {
	Name   string
	Alias  string
	Join   string
	Using  string
	Fields map[string] *QBField
	schema *schema.FieldsTable
	qB     *QueryBuilder
}
type QueryBuilder struct {
	Tables 		[] *QBTable
	Args 		[] interface{}
	fields 		[] *QBField
	Aliases 	[] string
	Prepared        *sql.Stmt
	FieldsParams 	map[string][]string
	sqlCommand, sqlSelect, sqlFrom string		`auto recalc`
	Where, GroupBy, OrderBy, Limits string	`may be defined outside`
	union *QueryBuilder
}
// addding arguments
func (qb *QueryBuilder) AddArg(arg interface{}) *QueryBuilder{
	qb.Args = append(qb.Args, arg)

	return qb
}
func (qb *QueryBuilder) AddArgs(args ... interface{}) *QueryBuilder{

	for _, arg := range args {
		qb.Args = append(qb.Args, arg)
	}

	return qb
}
// add Tables list, returns qB
func (qb *QueryBuilder) AddTables(names map[string] string) *QueryBuilder {
	for alias, name := range names {
		qb.AddTable(alias, name)
	}

	return qb
}
//add Table, returns object table
func (qb *QueryBuilder) AddTable(alias, name string) *QBTable {

	if alias == ""  {
		alias = name
	}
	table := &QBTable{Name: name, Alias: alias, qB: qb}
	table.Fields = make(map[string] *QBField, 0)
	defer schemaError()
	table.schema = schema.GetFieldsTable(table.Name)
	qb.Tables    = append(qb.Tables, table)

	return table
}
// add table with join
func (qb *QueryBuilder) JoinTable(alias, name, join, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join   = join
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) Join(alias, name, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join   = " JOIN "
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) LeftJoin(alias, name, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join   = " LEFT JOIN "
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) RightJoin(alias, name, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join   = " RIGHT JOIN "
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) InnerJoin(alias, name, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join   = " INNER JOIN "
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) AddUnion(union *QueryBuilder) *QueryBuilder {
	qb.union = union

	return qb
}

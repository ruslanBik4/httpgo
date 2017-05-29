// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// this module has more structures from creating sql-query with relation to db schema

package qb

import (
	"strings"
	"github.com/ruslanBik4/httpgo/models/db/schema"
)

type QBFields struct {
	Name  string
	Alias string

}
type QBTables struct {
	Name string
	Alias string
	Join string
	Using string
	Fields map[string] *QBFields
}
type QueryBuilder struct {
	Tables [] *QBTables
	Args [] interface{}
	fields [] schema.FieldStructure
	FieldsParams map[string][]string
	sql, Where, GroupBy, OrderBy, Limits string
	union string
}
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
	qb := &QueryBuilder{sql: sqlCommand}
	return qb
}
// addding arguments
func (qb *QueryBuilder) AddArgs(arg interface{}) *QueryBuilder{
	qb.Args = append(qb.Args, arg)

	return qb
}
// add Tables list, returns qb
func (qb *QueryBuilder) AddTables(names map[string] string) *QueryBuilder {
	for alias, name := range names {
		qb.AddTable(alias, name)
	}

	return qb
}
//add Table, returns object table
func (qb *QueryBuilder) AddTable(alias, name string) *QBTables {

	table := &QBTables{Name: name, Alias: alias}
	table.Fields = make(map[string] *QBFields, 0)
	qb.Tables    = append(qb.Tables, table)

	return table
}
// add table with join
func (qb *QueryBuilder) JoinTable(alias, name, join, usingOrOn string) *QBTables {

	table := qb.AddTable(alias, name)
	table.Join   = join
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) Join(alias, name, usingOrOn string) *QBTables {

	table := qb.AddTable(alias, name)
	table.Join   = " JOIN "
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) LeftJoin(alias, name, usingOrOn string) *QBTables {

	table := qb.AddTable(alias, name)
	table.Join   = " LEFT JOIN "
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) RightJoin(alias, name, usingOrOn string) *QBTables {

	table := qb.AddTable(alias, name)
	table.Join   = " RIGHT JOIN "
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) InnerJoin(alias, name, usingOrOn string) *QBTables {

	table := qb.AddTable(alias, name)
	table.Join   = " INNER JOIN "
	table.Using  = usingOrOn

	return table
}
func (qb *QueryBuilder) Union(sql string) *QueryBuilder {
	qb.union = sql

	return qb
}
// adding fields
func (table *QBTables) AddFields(fields map[string] string) *QBTables {
	for alias, name := range fields {
		table.AddField(alias, name)
	}

	return table
}
// add field and returns table object
func (table *QBTables) AddField(alias, name string) *QBTables {

	if strings.Contains(name, " AS ") {
		pos := strings.Index(name, " AS ")
		alias = name[ pos + 4 : ]
		name  = name[: pos]
	} else if alias == ""  {
		alias = name
	}

	field := &QBFields{Name: name}
	table.Fields[alias] = field

	return table
}


// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// this module has more structures from creating sql-query with relation to db Schema

package qb

import (
	"database/sql"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"strings"
)

// field in QB for incapsulate SQL & Schema propertyes
// may have children QB for getting data on child tables
type QBField struct {
	Name         string
	Alias        string
	Schema       *schema.FieldStructure
	Value        sql.RawBytes
	SelectValues map[int]string
	Table        *QBTable
	ChildQB      *QueryBuilder
	SelectQB     *QueryBuilder
}

// table in QB for incapsulate SQL & Schema propertyes
// ha map Fields as links field query
type QBTable struct {
	Name   string
	Alias  string
	Join   string
	Using  string
	Fields map[string]*QBField
	schema *schema.FieldsTable
	qB     *QueryBuilder
}

// inline SQL query
// recheck in DB Schema queryes tables&fields
// may be has parent - link to parent QB
type QueryBuilder struct {
	Tables                          []*QBTable
	Args                            []interface{}
	fields                          []*QBField
	Aliases                         []string
	Prepared                        *sql.Stmt
	PostParams                      map[string][]string
	sqlCommand, sqlSelect, sqlFrom  string // auto recalc
	Where, GroupBy, OrderBy, Limits string // may be defined outside
	union                           *QueryBuilder
	parent                          *QueryBuilder
}

// for compatabilies interface logsType
func (qb *QueryBuilder) PrintToLogs() string {

	mess := "&qb{sql: " + qb.sqlCommand + ", Where: " + qb.Where + ", Tables: "
	for _, table := range qb.Tables {
		mess += table.Name + ", "
	}
	mess += " Fields: "
	for _, alias := range qb.Aliases {
		mess += alias + ", "
	}
	mess += " Args: "
	for _, arg := range qb.Args {
		mess += fmt.Sprintf("%v, ", arg)
	}

	mess += " PostParams: "
	for _, arg := range qb.PostParams {
		mess += fmt.Sprintf("%v, ", arg)
	}
	return mess + "}"
}

// addding arguments
func (qb *QueryBuilder) AddArg(arg interface{}) *QueryBuilder {
	qb.Args = append(qb.Args, arg)

	return qb
}
func (qb *QueryBuilder) AddArgs(args ...interface{}) *QueryBuilder {

	for _, arg := range args {
		qb.Args = append(qb.Args, arg)
	}

	return qb
}
func (qb *QueryBuilder) SetArgs(args ...interface{}) *QueryBuilder {
	qb.Args = nil
	qb.AddArgs(args...)

	return qb
}

// replace where clause
func (qb *QueryBuilder) SetWhere(where string) {

	if qb.sqlCommand > "" {
		if qb.Where > "" {
			qb.sqlCommand = strings.Replace(qb.sqlCommand, qb.Where, where, -1)
		} else if where > "" {
			qb.sqlCommand += " WHERE " + where
		}
		if qb.Prepared != nil {
			qb.Prepared = nil
		}

	}

	qb.Where = where
}

// add table with join
func (qb *QueryBuilder) JoinTable(alias, name, join, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join = join
	table.Using = usingOrOn

	return table
}
func (qb *QueryBuilder) Join(alias, name, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join = " JOIN "
	table.Using = usingOrOn

	return table
}
func (qb *QueryBuilder) LeftJoin(alias, name, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join = " LEFT JOIN "
	table.Using = usingOrOn

	return table
}
func (qb *QueryBuilder) RightJoin(alias, name, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join = " RIGHT JOIN "
	table.Using = usingOrOn

	return table
}
func (qb *QueryBuilder) InnerJoin(alias, name, usingOrOn string) *QBTable {

	table := qb.AddTable(alias, name)
	table.Join = " INNER JOIN "
	table.Using = usingOrOn

	return table
}
func (qb *QueryBuilder) AddUnion(union *QueryBuilder) *QueryBuilder {
	qb.union = union

	return qb
}

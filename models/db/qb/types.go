// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// this module has more structures from creating sql-query with relation to db schema

package qb

type QBFields struct {
	Name  string
	Alias string

}
type QBTables struct {
	Name string
	Join string
	Fields map[string] *QBFields
}
type QueryBuilder struct {
	Tables map[string] *QBTables
	Where   string
	Args [] interface{}
	GroupBy string
	OrderBy string
}

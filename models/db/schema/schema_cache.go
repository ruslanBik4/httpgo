// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"fmt"

	"github.com/ruslanBik4/logs"
)

// SchemaCache хранит структуру полей - стоит продумать, как хранить еще и ключи
var SchemaCache map[string]*FieldsTable

// ErrNotFoundTable if not found table by name {Table}
type ErrNotFoundTable struct {
	Table string
}

func (err ErrNotFoundTable) Error() string {

	return fmt.Sprintf("Not table `%s` in schema ", err.Table)
}

// ErrNotFoundField if not found in table {Table} field by name {Column}
type ErrNotFoundField struct {
	Table     string
	FieldName string
}

func (err ErrNotFoundField) Error() string {

	return fmt.Sprintf("Not field `%s` for table `%s` in schema ", err.FieldName, err.Table)

}

// GetFieldsTable return schema table from cache
func GetFieldsTable(tableName string) *FieldsTable {
	table, ok := SchemaCache[tableName]
	if !ok {
		logs.ErrorLogHandler(ErrNotFoundTable{Table: tableName})
		panic(ErrNotFoundTable{Table: tableName})
	}
	return table
}

// GetParentTable return name parent table
func GetParentTable(tableName string) *FieldsTable {
	_, ok := SchemaCache[tableName]
	if !ok {
		logs.ErrorLogHandler(ErrNotFoundTable{Table: tableName})
		panic(ErrNotFoundTable{Table: tableName})
	}
	for _, fields := range SchemaCache {
		for _, field := range fields.Rows {
			if field.TableValues == tableName {
				return fields
			}
		}
	}
	return nil
}

func init() {
	SchemaCache = make(map[string]*FieldsTable, 0)
}

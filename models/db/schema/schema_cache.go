// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema
// хранит структуру полей - стоит продумать, как хранить еще и ключи
var SchemaCache map[string] *FieldsTable

func GetFieldsTable(tableName string) *FieldsTable {
	table, ok := SchemaCache[tableName]
	if !ok {
		return nil
	}
	return table
}
func init() {
	SchemaCache = make(map[string] *FieldsTable, 0)
}

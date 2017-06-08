// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import "github.com/ruslanBik4/httpgo/models/db/schema"

// getters
func (table *QBTable) GetSchema() *schema.FieldsTable {
	return table.schema
}

func (table *QBTable) getFieldSchema(name string) *schema.FieldStructure {
	for _, field := range table.schema.Rows {
		if field.COLUMN_NAME == name {
			return field
		}
	}

	return nil
}


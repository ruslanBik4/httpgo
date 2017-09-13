// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package api описание вспомогательных функций для роутеров API
package api

import (
	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"net/url"
)

// check params "fields" in Post request & add those in qBuilder table
func addFieldsFromPost(table *qb.QBTable, rForm url.Values) {

	if fields, ok := rForm["fields[]"]; ok {
		for _, val := range fields {
			table.AddField("", val)
		}
	}
}

func findField(key string, tables map[string]schema.FieldsTable) *schema.FieldStructure {

	for _, table := range tables {
		if field := table.FindField(key); field != nil {
			return field
		}
	}

	return nil
}

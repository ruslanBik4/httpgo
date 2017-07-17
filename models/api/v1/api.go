// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// описание вспомогательных функций для роутеров API
package api

import (
	"net/url"
	"github.com/ruslanBik4/httpgo/models/db/qb"
)
// check params "fields" in Post request & add those in qBuilder table
func addFieldsFromPost(table *qb.QBTable, rForm url.Values)  {

	if fields, ok := rForm["fields[]"]; ok {
		for _, val := range fields {
			table.AddField("", val)
		}
	}
}


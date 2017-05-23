// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import (
	"github.com/ruslanBik4/httpgo/models/server"
	"log"
)

func InitLists() {
	go func() {
		var tables RecordsTables
		err := tables.GetSelectTablesProp( "TABLE_SCHEMA='" + server.GetServerConfig().DBName() + " AND (RIGHT(table_name, 5) =  '_list') ")

		if err != nil {
			log.Println(err)
		}

	}()
}

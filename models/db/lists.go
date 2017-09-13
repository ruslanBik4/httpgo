// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import (
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/ruslanBik4/httpgo/models/server"
)

// InitLists - инициализация получения информации по справочникам
func InitLists() {
	var tables RecordsTables
	where := `TABLE_SCHEMA=? AND (RIGHT(table_name, 5) = ?)`
	err := tables.GetSelectTablesProp(where, server.GetServerConfig().DBName(), "_list")

	if err != nil {
		logs.ErrorLog(err, where)
	}

}

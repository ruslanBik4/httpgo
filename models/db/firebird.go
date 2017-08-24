// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import (
	_ "github.com/nakagami/firebirdsql"
	"database/sql"
	"github.com/ruslanBik4/httpgo/models/logs"
)
var (
	fbConn *sql.DB
)
func fbConnect() (err error) {
	if fbConn != nil {
		return nil
	}
	fbConn, err = sql.Open("firebirdsql", "sysdba:masterkey@/travel")
	if err != nil {
		logs.ErrorLog(err)
		return err
	} else if fbConn == nil {
		return sql.ErrNoRows
	}

	return nil
}
func FBSelect(sql string, args ...interface{}) (rows *sql.Rows, err error) {
	if err = fbConnect(); err != nil {
		return nil, err
	}

	return fbConn.Query(sql, args ...)

}
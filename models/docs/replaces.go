// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// tools for create document from templates

package docs

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/models/db"
)

func GetReplaces(person map[string]*sql.NullString, signSerf string) (replaces map[string]string) {

	rows, err := db.DoSelect("select * from doc_keywords_list")
	if err != nil {
		logs.ErrorLog(err)

		return nil
	}

	replaces = make(map[string]string, 0)

	defer rows.Close()

	today := time.Now()
	for rows.Next() {
		var id int
		var name, title string

		if err := rows.Scan(&id, &name, &title); err != nil {
			logs.ErrorLog(err)
			continue
		}

		name = GetKeyword(name)
		if val, ok := person[name]; ok && val.Valid {
			replaces[name] = val.String
		} else if name == "@docNumber" {
			replaces[name] = fmt.Sprintf("%d/%d", 1, today.Year())
		} else if name == "@docDate" {
			replaces[name] = today.Local().Format("02.01.2006")
		} else if name == "@signSerf" {
			replaces[name] = signSerf
		} else {
			//continue
			replaces[name] = title
		}
	}

	return replaces

}

// экранируем для поиска в изаммены в шаблоне
func GetKeyword(name string) string {
	return "@" + name
}

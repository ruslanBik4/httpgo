/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package tables

import (
	"database/sql"
	"regexp"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/logs"
)

var enumValidator = regexp.MustCompile(`(?:'([^,]+)',?)`)

// QueryStruct has property for form record in table view
type QueryStruct struct {
	HrefEdit   string
	Href       string
	row        []interface{}
	columns    []string
	fields     []dbEngine.Column
	Rows       *sql.Rows
	Tables     []dbEngine.Table
	widthTable int
	Order      string
	PostFields []dbEngine.Column
}

func (query *QueryStruct) findField(fieldName string) dbEngine.Column {
	for _, table := range query.Tables {
		if column := table.FindColumn(fieldName); column != nil {
			return column
		}
	}

	return nil

}
func (query *QueryStruct) beforeRender() (err error) {

	query.columns, err = query.Rows.Columns()
	if err != nil {
		logs.ErrorLog(err)
		return err
	}

	// mfields может не соответствовать набору столбцов, потому завязываем на имеющиеся, прочие - игнорируем
	for _, fieldName := range query.columns {
		if field := query.findField(fieldName); field == nil {
			query.row = append(query.row, field)
			query.fields = append(query.fields, field)
		}
	}

	return nil
}

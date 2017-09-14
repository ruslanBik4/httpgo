// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Creating %{DATE}

//Обслуживает кеш для справочников БД

package services

import (
	"database/sql"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/db/cache"
	DBschema "github.com/ruslanBik4/httpgo/models/db/schema"
	"strings"
	"time"
)

type dbListsService struct {
	name   string
	status string
	tables listTables
}
type listTables map[string]*listRows

type listRows struct {
	schema *DBschema.FieldsTable
	rows   []listRowData
}
type listRowData map[string]string
type rowField string

var (
	dbLists *dbListsService = &dbListsService{name: "DBLists", status: "create", tables: make(listTables, 0)}
)

func (lRows *listRows) addRows() error {
	rows, err := db.DoSelect("SELECT * FROM `" + lRows.schema.Name + "`")

	if err != nil {
		return err
	}

	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	scanArgs := make([]interface{}, len(columns))

	for idx := range scanArgs {
		scanArgs[idx] = &sql.RawBytes{}
	}
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return err
		}
		newRow := make(listRowData, len(columns))
		for idx, value := range columns {
			if scanArgs[idx] == nil {
				newRow[value.Name()] = "NULL"
			} else {
				newRow[value.Name()] = string(*scanArgs[idx].(*sql.RawBytes))
			}
		}

		lRows.rows = append(lRows.rows, newRow)
	}

	return nil
}
func (DBlists *dbListsService) Init() error {
	DBlists.status = "starting"

	for Status("schema") != "ready" {
		time.Sleep(50)
	}

	for tableName, fields := range DBschema.SchemaCache {
		if strings.HasSuffix(tableName, "_list") {
			DBlists.tables[tableName] = &listRows{schema: fields}
			err := DBlists.tables[tableName].addRows()
			if err != nil {
				DBlists.status = "crashing"
				return err
			}
		}
	}

	DBlists.status = "ready"

	return nil
}
func (DBlists *dbListsService) Send(messages ...interface{}) error {
	return nil

}
func (DBlists *dbListsService) Get(messages ...interface{}) (response interface{}, err error) {

	oper, ok := messages[0].(string)
	if !ok {
		return nil, &ErrServiceNotCorrectParamType{Name: schema.name, Param: messages[0]}
	}

	switch oper {

	case "all-list":

		result := make([]string, 0, len(DBlists.tables))
		for name, _ := range DBlists.tables {
			result = append(result, name)
		}
		return result, nil
	case "one":
		list, ok := DBlists.tables[messages[1].(string)]
		if !ok {
			return nil, ErrServiceWrongIndex{Name: messages[1].(string)}
		}
		result := make([]map[string]interface{}, len(list.rows))
		for idx, row := range list.rows {
			result[idx] = make(map[string]interface{}, 0)
			for key, value := range row {
				result[idx][key] = value
			}
		}
		return result, nil
	//	сокращенный вариант вызова - только имя таблицы
	default:
		return cache.GetListRecord(oper), nil
	}

}
func (DBlists *dbListsService) Connect(in <-chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (DBlists *dbListsService) Close(out chan<- interface{}) error {

	return nil
}
func (DBlists *dbListsService) Status() string {

	return DBlists.status
}

func init() {
	AddService(dbLists.name, dbLists)
}

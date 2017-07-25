// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Creating %{DATE}

//Обслуживает кеш для справочников БД

package services

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/db/cache"
	DBschema "github.com/ruslanBik4/httpgo/models/db/schema"
	"time"
	"strings"
	"database/sql"
)

type DBlistsService struct {
	name   string
	status string
	tables listTables
}
type listTables map[string] *listRows

type listRows struct {
	schema *DBschema.FieldsTable
	rows [] listRowData
}
type listRowData map[string] string
type rowField string
var (
	DBlists *DBlistsService = &DBlistsService{name: "DBlists", status: "create", tables:make(listTables,0)}
)


func (lRows *listRows) addRows() error {
	rows, err := db.DoSelect("SELECT * FROM `" + lRows.schema.Name + "`")

	if err != nil {
		return err
	}

	defer rows.Close()

	columns, err := rows.ColumnTypes()
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
				newRow[value.Name()] = string( *scanArgs[idx].(*sql.RawBytes) )
			}
		}

		lRows.rows = append(lRows.rows, newRow)
	}

	return nil
}
func (DBlists *DBlistsService) Init() error {
	DBlists.status = "starting"

	for Status("schema") != "ready" {
		time.Sleep(50)
	}

	for tableName, fields := range DBschema.SchemaCache {
		if strings.HasSuffix(tableName, "_list") {
			DBlists.tables[tableName] = &listRows{ schema: fields }
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
func (DBlists *DBlistsService) Send(messages ...interface{}) error {
	return nil

}
func (DBlists *DBlistsService) Get(messages ...interface{}) (responce interface{}, err error) {
	oper := messages[0].(string)
	switch oper {

	case "all-list":

		result := make([] string, 0, len(DBlists.tables))
		for name, _ := range DBlists.tables {
			result = append(result, name)
		}
		return result, nil
	case "one":
		list, ok := DBlists.tables[messages[1].(string)]
		if !ok {
			return nil, ErrServiceWrongIndex{Name: messages[1].(string)}
		}
		result := make([]map[string] interface{}, len(list.rows))
		for idx, row := range list.rows {
			result[idx] = make(map[string] interface{}, 0)
			for key, value := range row {
				result[idx][key] = value
			}
		}
		return result, nil
	}
	switch tableName := messages[0].(type) {
	case string:
		return cache.GetListRecord(tableName), nil
	default:
		return nil, &ErrServiceNotCorrectParamType{Name: schema.name, Param: messages[0]}
	}

	return nil, nil

}
func (DBlists *DBlistsService) Connect(in <-chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (DBlists *DBlistsService) Close(out chan<- interface{}) error {

	return nil
}
func (DBlists *DBlistsService) Status() string {

	return DBlists.status
}

func init() {
	AddService(DBlists.name, DBlists)
}

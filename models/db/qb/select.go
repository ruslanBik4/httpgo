// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"strings"
	"strconv"
	"log"
)

func (qb * QueryBuilder) SelectToMultidimension() ( arrJSON [] map[string] interface {}, err error ) {

	var fields [] schema.FieldStructure
	var qFields, sql, qFrom string

	commaTbl, commaFld := "", ""
	for alias, table := range qb.Tables {
		tableStrc := schema.GetFieldsTable(table.Name)
		if fields == nil {
			panic("Not table in schema")
		}
		// temporary not validate first table on  having JOIN property
		// TODO: add checking join if first table as error!!!
		if table.Join > "" {
			commaTbl = " JOIN "
		}
		qFrom += commaTbl + table.Name + " " + alias + table.Join
		commaTbl = ", "
		for alias, field := range table.Fields {
			fieldStrc := tableStrc.FindField(field.Name)
			if fieldStrc == nil {
				panic("not field in table " + qFrom)
			}
			fields = append(fields, *fieldStrc)
			qFields += commaFld + field.Name + " " + alias
			commaFld = ", "
		}
	}

	if qb.GroupBy > "" {
		sql += " group by " + qb.GroupBy
	}
	if qb.OrderBy > "" {
		sql += " group by " + qb.OrderBy
	}

	rows, err := db.DoSelect("select " + qFields + " from " + qFrom + sql, qb.Args...)


	if err != nil {
		//log.Println("mysql.go,","string 306,", err, sql)
		return nil, err
	}

	defer rows.Close()

	var valuePtrs []interface{}
	var fieldID *schema.FieldStructure

	for _, field := range fields {
		valuePtrs = append(valuePtrs, &field )
	}


	for rows.Next() {
		values := make(map[string] interface{}, len(fields) )
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Println(err)
			continue
		}


		for _, field := range fields {

			values[field.COLUMN_NAME] = field.Value
			//TODO для полей типа tableid_, setid_, nodeid_ придумать механизм для блока WHERE
			// (по ключу родительской таблицы и патетрну из свойств поля для полей типа set)
			//TODO для полей типа setid_ формировать название таблицы
			//TODO также на уровне функции продумать менанизм, который позволит выбирать НЕ ВСЕ поля из третей таблицы
			if strings.HasPrefix(field.COLUMN_NAME, "setid_")  {
				values[field.COLUMN_NAME], err = db.SelectToMultidimension( field.SQLforSelect, fieldID.Value )
				if err != nil {
					log.Println(err)
					values[field.COLUMN_NAME] = err.Error()
				}
				continue
			} else if strings.HasPrefix(field.COLUMN_NAME, "nodeid_"){
				values[field.COLUMN_NAME], err = db.SelectToMultidimension( field.SQLforSelect, fieldID.Value )
				if err != nil {
					log.Println(err)
					values[field.COLUMN_NAME] = err.Error()
				}
				continue
			} else if strings.HasPrefix(field.COLUMN_NAME, "tableid_"){
				values[field.COLUMN_NAME], err = db.SelectToMultidimension( field.SQLforSelect, fieldID.Value)
				if err != nil {
					log.Println(err)
					values[field.COLUMN_NAME] = err.Error()
				}
				continue

			}

			switch field.COLUMN_TYPE {
			case "varchar", "date", "datetime":
				values[field.COLUMN_NAME] = field.Value
			case "tinyint":
				if field.Value == "1" {
					values[field.COLUMN_NAME] = true
				} else {
					values[field.COLUMN_NAME] = false

				}
			case "int", "int64", "float":
				values[field.COLUMN_NAME], _ = strconv.Atoi(field.Value)
			default:
				values[field.COLUMN_NAME] = field.Value
			}
		}

		arrJSON = append(arrJSON, values)
	}

	return arrJSON, nil

}


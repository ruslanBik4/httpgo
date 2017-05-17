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

func (qb * QueryBuilder) createSQL() ( sql string, fields [] schema.FieldStructure, err error ) {

	var qFields, qFrom string

	commaTbl, commaFld := "", ""
	for aliasTable, table := range qb.Tables {
		tableStrc := schema.GetFieldsTable(table.Name)
		if tableStrc == nil {
			panic( &schema.ErrNotFoundTable{ Table:table.Name} )
		}
		// temporary not validate first table on  having JOIN property
		// TODO: add checking join if first table as error!!!
		if table.Join > "" {
			qFrom += " " + table.Join + " " + table.Name + " " + aliasTable + " " + table.Using
		} else {
			qFrom += commaTbl + " " + table.Name + " " + aliasTable + " " + table.Join
		}
		commaTbl = ", "
		if len(table.Fields) == 0 {
			qFields += commaFld + aliasTable + ".*"
			commaFld = ", "

			for _, fieldStrc := range tableStrc.Rows {

				fields = append(fields, fieldStrc)
			}
		} else {
			for alias, field := range table.Fields {
				var queryName string
				fieldStrc := tableStrc.FindField(field.Name)
				if alias > "" {
					queryName = ` as "` + alias + `"`
				}
				if fieldStrc == nil {
					fieldStrc = &schema.FieldStructure{COLUMN_NAME: alias}
					qFields += commaFld + field.Name + queryName
				} else {

					qFields += commaFld + aliasTable + "." + field.Name + queryName
				}
				fields = append(fields, *fieldStrc)
				commaFld = ", "
			}
		}
	}

	if qb.Where > "" {
		sql += " where " + qb.Where
	}
	if qb.GroupBy > "" {
		sql += " group by " + qb.GroupBy
	}
	if qb.OrderBy > "" {
		sql += " order by " + qb.OrderBy
	}

	return "select " + qFields + " from " + qFrom + sql, fields, nil

}
func (qb * QueryBuilder) SelectToMultidimension() ( arrJSON [] map[string] interface {}, err error ) {

	sql, fields, err := qb.createSQL()

	log.Println(sql)
	rows, err := db.DoSelect(sql, qb.Args...)


	if err != nil {
		//log.Println("mysql.go,","string 306,", err, sql)
		return nil, err
	}

	defer rows.Close()

	var valuePtrs []interface{}
	//var fieldID *schema.FieldStructure

	for ind, _ := range fields {
		valuePtrs = append(valuePtrs, &fields[ind] )
	}

	columns, _ := rows.Columns()
	for rows.Next() {
		var fieldID string
		values := make(map[string] interface{}, len(fields) )
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Println(err)
			continue
		}


		for idx, fieldName := range columns {

			field := fields[idx]
			if fieldName == "id" {
				fieldID = field.Value
			}
			if strings.HasPrefix(fieldName, "setid_")  {
				values[fieldName], err = db.SelectToMultidimension( field.SQLforSelect, fieldID )
				if err != nil {
					log.Println(err, field.SQLforSelect)
					values[fieldName] = err.Error()
				}
				continue
			} else if strings.HasPrefix(fieldName, "nodeid_"){
				values[fieldName], err = db.SelectToMultidimension( field.SQLforSelect, fieldID )
				if err != nil {
					log.Println(err, field.SQLforSelect)
					values[fieldName] = err.Error()
				}
				continue
			} else if strings.HasPrefix(fieldName, "tableid_"){
				values[fieldName], err = db.SelectToMultidimension( field.SQLforSelect, fieldID)
				if err != nil {
					log.Println(err, field.SQLforSelect)
					values[fieldName] = err.Error()
				}
				continue

			}

			switch field.COLUMN_TYPE {
			case "varchar", "date", "datetime":
				values[fieldName] = field.Value
			case "tinyint":
				if field.Value == "1" {
					values[fieldName] = true
				} else {
					values[fieldName] = false

				}
			case "int", "int64", "float":
				values[fieldName], _ = strconv.Atoi(field.Value)
			default:
				values[fieldName] = field.Value
			}
		}

		arrJSON = append(arrJSON, values)
	}

	return arrJSON, nil

}


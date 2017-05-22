// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"strconv"
	"log"
	"strings"
	"fmt"
)

func (qb * QueryBuilder) createSQL() ( sql string, fields [] schema.FieldStructure, err error ) {

	var qFields, qFrom string

	commaTbl, commaFld := "", ""
	for _, table := range qb.Tables {
		tableStrc := schema.GetFieldsTable(table.Name)
		aliasTable:= table.Alias
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
		if len(table.Fields) > 0 {
			for alias, field := range table.Fields {
				var queryName string
				fieldStrc := tableStrc.FindField(field.Name)
				if alias > "" {
					queryName = ` AS "` + alias + `"`
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
		} else if table.Join == ""{
			qFields += commaFld + aliasTable + ".*"
			commaFld = ", "

			for _, fieldStrc := range tableStrc.Rows {

				fields = append(fields, fieldStrc)
			}
		}
	}

	if qb.Where > "" {
		if strings.Contains(qb.Where, "WHERE") {
			sql += qb.Where
		} else {
			sql += " WHERE " + qb.Where
		}
	}
	if qb.GroupBy > "" {
		sql += " GROUP BY " + qb.GroupBy
	}
	if qb.OrderBy > "" {
		sql += " ORDER BY " + qb.OrderBy
	}

	return "SELECT " + qFields + " FROM " + qFrom + sql, fields, nil

}
func (qb * QueryBuilder) SelectToMultidimension() ( arrJSON [] map[string] interface {}, err error ) {

	sql, fields, err := qb.createSQL()

	log.Println(sql)
	rows, err := db.DoSelect(sql, qb.Args...)


	if err != nil {
		log.Println("mysql.go,","string 306,", err, sql)
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
			fieldTableName := field.Table.Name
			if fieldName == "id" {
				fieldID = field.Value
			}
			if field.SETID  {
				gChild := Create(field.WhereFromSet(field.Table),"", "")
				tableProps := strings.TrimPrefix(field.COLUMN_NAME, "setid_")
				titleField := field.GetForeignFields()

				gChild.AddTable( "p", tableProps ).AddField("", "id").AddField("", titleField)

				tableValue  := field.Table.Name + "_" + tableProps + "_has"
				onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", tableProps, fieldTableName )
				gChild.JoinTable ( "v", tableValue, "JOIN", onJoin )

				gChild.AddArgs(fieldID)

				values[fieldName], err = gChild.SelectToMultidimension()
				if err != nil {
					log.Println(err, field.SQLforChieldList)
					values[fieldName] = err.Error()
				}
				continue
			} else if field.NODEID {
				gChild := Create(field.WhereFromSet(field.Table),"", "")

				var tableProps, titleField string

				tableValue  := strings.TrimPrefix(field.COLUMN_NAME, "nodeid_")
				fieldsValues := schema.GetFieldsTable(tableValue)

				if fieldsValues == nil {
					log.Println( schema.ErrNotFoundTable{Table:tableValue} )
					continue
				}
				//TODO: later refactoring - store values in field propertyes
				for _, field := range fieldsValues.Rows {
					if strings.HasPrefix(field.COLUMN_NAME, "id_") && (field.COLUMN_NAME != "id_" + fieldTableName) {
						tableProps = field.COLUMN_NAME[3:]
						titleField = field.GetForeignFields()
						break
					}
				}

				if (tableProps == "") || (titleField == "") {
					continue
				}

				gChild.AddTable( "p", tableProps ).AddField("", "id").AddField("", titleField)

				onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", tableProps, fieldTableName )
				gChild.JoinTable ( "v", tableValue, "JOIN", onJoin ).AddField("", "id_" + fieldTableName)
				gChild.AddArgs(fieldID)

				values[fieldName], err = gChild.SelectToMultidimension()
				if err != nil {
					log.Println(err, field.SQLforChieldList)
					values[fieldName] = err.Error()
				}
				continue
			} else if field.TABLEID {
				values[fieldName], err = db.SelectToMultidimension( field.SQLforChieldList, fieldID)
				if err != nil {
					log.Println(err, field.SQLforChieldList)
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


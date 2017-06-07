// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"database/sql"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"strconv"
	"github.com/ruslanBik4/httpgo/models/logs"
	"strings"
	"fmt"
)

func (qb * QueryBuilder) createSQL() ( sql string, err error ) {

	var qFields, qFrom string

	commaTbl, commaFld := "", ""

	qb.fields = make([] schema.FieldStructure, 0)
	for _, table := range qb.Tables {

		defer func() {
			result := recover()
			switch err1 := result.(type) {
			case schema.ErrNotFoundTable:
				logs.ErrorLog(err1, err1.Table)
				err = err1
			case nil:
			default:
				panic(err)
			}
		}()
		tableStrc := schema.GetFieldsTable(table.Name)
		aliasTable:= table.Alias
		// temporary not validate first table on  having JOIN property
		// TODO: add checking join if first table as error!!!
		if table.Join > "" {
			qFrom += " " + table.Join + " " + table.Name + " " + aliasTable + " " + table.Using
		} else {
			qFrom += commaTbl + table.Name + " " + aliasTable
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
				qb.fields = append(qb.fields, *fieldStrc)
				qb.Aliases = append(qb.Aliases, alias)
				commaFld = ", "
			}
		} else if table.Join == ""{
			qFields += commaFld + aliasTable + ".*"
			commaFld = ", "

			for _, fieldStrc := range tableStrc.Rows {

				qb.fields = append(qb.fields, fieldStrc)
			}
		}
	}

	sql += qb.getWhere()

	if qb.union != nil {
		sql += qb.unionSQL()
	}
	if qb.GroupBy > "" {
		sql += " GROUP BY " + qb.GroupBy
	}
	if qb.OrderBy > "" {
		sql += " ORDER BY " + qb.OrderBy
	}
	if qb.Limits > "" {
		sql += " LIMIT " + qb.Limits
	}

	return "SELECT " + qFields + " FROM " + qFrom + sql, nil

}
func (qb * QueryBuilder) getWhere() string {
	if qb.Where > "" {
		if strings.Contains(qb.Where, "WHERE") {
			return qb.Where
		} else {
			return " WHERE " + qb.Where
		}
	}
	return ""
}
func (qb * QueryBuilder) unionSQL() string {
	var qFields, qFrom string

	commaTbl, commaFld := "", ""
	for _, table := range qb.union.Tables {
		if table.Join > "" {
			qFrom += " " + table.Join + " " + table.Name + " " + table.Alias + " " + table.Using
		} else {
			qFrom += commaTbl + table.Name + " " + table.Alias
		}
		commaTbl = ", "

	}
	for _, alias := range qb.Aliases {
		for _, table := range qb.union.Tables {
			if field, ok := table.Fields[alias]; ok {

				tableStrc := schema.GetFieldsTable(table.Name)
				fieldStrc := tableStrc.FindField(field.Name)
				if fieldStrc != nil {
					qFields += commaFld + table.Alias + "." + field.Name + ` AS "` + field.Alias + `"`
				} else {
					qFields += commaFld + field.Name + ` AS "` + field.Alias + `"`

				}
				commaFld = ", "
				break
			}
		}
	}

	return " UNION SELECT " + qFields + " FROM " + qFrom + qb.union.getWhere()
}
func getSETID_Values(field schema.FieldStructure, fieldID string) (arrJSON [] map[string] interface {}, err error ){

	where := field.WhereFromSet(field.Table)
	gChild := Create(where,"", "")
	titleField := field.GetForeignFields()

	gChild.AddTable( "p", field.TableProps ).AddField("", "id").AddField("", titleField)

	onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.TableProps, field.Table.Name )
	gChild.JoinTable ( "v", field.TableValues, "JOIN", onJoin )

	gChild.AddArg(fieldID)

	return gChild.SelectToMultidimension()

}
func getNODEID_Values(field schema.FieldStructure, fieldID string) (arrJSON [] map[string] interface {}, err error ) {

	fieldTableName := field.Table.Name
	where := field.WhereFromSet(field.Table)

	gChild := Create(where,"", "")

	var tableProps, titleField string

	defer func() {
		result := recover()
		switch err1 := result.(type) {
		case schema.ErrNotFoundTable:
			logs.ErrorLog(err1, field.TableValues)
			err = err1
		case nil:
		default:
			panic(err)
		}
	}()
	fieldsValues := schema.GetFieldsTable(field.TableValues)

	//TODO: later refactoring - store values in field propertyes
	for _, field := range fieldsValues.Rows {
		if strings.HasPrefix(field.COLUMN_NAME, "id_") && (field.COLUMN_NAME != "id_" + fieldTableName) {
			tableProps = field.COLUMN_NAME[3:]
			titleField = field.GetForeignFields()
			break
		}
	}

	if (tableProps == "") || (titleField == "") {
		return nil, schema.ErrNotFoundTable{Table: field.TableValues}
	}

	gChild.AddTable( "p", tableProps ).AddField("", "id").AddField("", titleField)

	onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.TableProps, fieldTableName )
	gChild.JoinTable ( "v", field.TableValues, "JOIN", onJoin ).AddField("", "id_" + fieldTableName)
	gChild.AddArg(fieldID)

	return gChild.SelectToMultidimension()

}
func getTABLEID_Values(field schema.FieldStructure, fieldID string) (arrJSON [] map[string] interface {}, err error ){

	where := field.WhereFromSet(field.Table)
	if where > "" {
		where += fmt.Sprintf( " AND (id_%s=?)", field.Table.Name )
	} else {
		where = fmt.Sprintf( " WHERE (id_%s=?)", field.Table.Name )
	}
	gChild := Create(where,"", "")
	gChild.AddTable( "p", field.TableProps )

	gChild.AddArg(fieldID)

	return gChild.SelectToMultidimension()

}

func (qb * QueryBuilder) GetDataSql() (rows *sql.Rows, err error)  {
	//var rows  *extsql.Rows
	var sqlQuery string
	sqlQuery, err = qb.createSQL()
	rows, err = db.DoSelect(sqlQuery, qb.Args...)

	logs.DebugLog("rows=", rows)
	if err != nil {
		logs.ErrorLog(err, sqlQuery)
		return nil, err
	}
	return rows, nil
}


func (qb * QueryBuilder) SelectToMultidimension() ( arrJSON [] map[string] interface {}, err error ) {

	sql, err := qb.createSQL()

	rows, err := db.DoSelect(sql, qb.Args...)


	if err != nil {
		logs.ErrorLog(err, sql)
		return nil, err
	}

	defer rows.Close()

	var valuePtrs []interface{}

	for idx, _ := range qb.fields {
		valuePtrs = append(valuePtrs, &qb.fields[idx] )
	}

	columns, _ := rows.Columns()
	for rows.Next() {
		var fieldID string
		values := make(map[string] interface{}, len(qb.fields) )
		if err := rows.Scan(valuePtrs...); err != nil {
			logs.ErrorLog(err, valuePtrs)
			continue
		}


		for idx, fieldName := range columns {

			field := qb.fields[idx]
			if fieldName == "id" {
				fieldID = field.Value
			}
			if field.SETID  {
				values[fieldName], err = getSETID_Values(field, fieldID)
				if err != nil {
					logs.ErrorLog(err, field.SQLforFORMList)
					values[fieldName] = err.Error()
				}
				continue
			} else if field.NODEID {

				values[fieldName], err = getNODEID_Values(field, fieldID)
				if err != nil {
					logs.ErrorLog(err, field.SQLforFORMList)
					values[fieldName] = err.Error()
				}
				continue
			} else if field.TABLEID {
				values[fieldName], err = getTABLEID_Values(field, fieldID)
				if err != nil {
					logs.ErrorLog(err, field.SQLforFORMList)
					values[fieldName] = err.Error()
				}
				continue

			}

			switch field.DATA_TYPE {
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


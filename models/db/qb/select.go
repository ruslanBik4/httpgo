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
//SelectToMultidimension(sql string, args ...interface
//@version 1.10 Sergey Litvinov 2017-05-25 15:15
func SelectToMultidimension(sql string, args ...interface{}) ( arrJSON [] map[string] interface {}, err error ) {
	qBuilder := CreateFromSQL(sql)
	qBuilder.AddArgs(args)

	return qBuilder.SelectToMultidimension()
}

func (qb * QueryBuilder) createSQL() ( sql string, err error ) {

	commaTbl, commaFld := "", ""
	for idx, table := range qb.Tables {

		aliasTable:= table.Alias
		//first table must'n to having JOIN property
		if (idx > 0) && (table.Join > "") {
			qb.sqlFrom += " " + table.Join + " " + table.Name + " " + aliasTable + " " + table.Using
		} else {
			qb.sqlFrom += commaTbl + table.Name + " " + aliasTable
		}
		commaTbl = ", "
		if len(table.Fields) > 0 {
			for alias, field := range table.Fields {
				var queryName string
				if alias > "" {
					queryName = ` AS "` + alias + `"`
				}
				if field.schema.COLUMN_TYPE == "calc" {
					qb.sqlSelect += commaFld + field.Name + queryName
				} else {

					qb.sqlSelect += commaFld + aliasTable + "." + field.Name + queryName

				}
				commaFld = ", "
			}
		} else if table.Join == "" {
			qb.sqlSelect += commaFld + aliasTable + ".*"
			commaFld = ", "

			for _, fieldStrc := range table.schema.Rows {

				field := &QBField{Name: fieldStrc.COLUMN_NAME, schema: fieldStrc}
				qb.fields = append(qb.fields, field)
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

	return "SELECT " + qb.sqlSelect + " FROM " + qb.sqlFrom + sql, nil

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

				if field.schema.COLUMN_TYPE == "calc" {
					qFields += commaFld + field.Name + ` AS "` + field.Alias + `"`
				} else {
					qFields += commaFld + table.Alias + "." + field.Name + ` AS "` + field.Alias + `"`

				}
				commaFld = ", "
				break
			}
		}
	}

	return " UNION SELECT " + qFields + " FROM " + qFrom + qb.union.getWhere()
}
func getSETID_Values(field *schema.FieldStructure, where, fieldID string) (arrJSON [] map[string] interface {}, err error ){

	gChild := Create(where,"", "")
	titleField := field.GetForeignFields()

	gChild.AddTable( "p", field.TableProps ).AddField("", "id").AddField("", titleField)

	onJoin := fmt.Sprintf("ON (p.id = v.id_%s AND id_%s = ?)", field.TableProps, field.Table.Name )
	gChild.Join ( "v", field.TableValues, onJoin )

	gChild.AddArg(fieldID)

	return gChild.SelectToMultidimension()

}
func getNODEID_Values(field *schema.FieldStructure, where, fieldID string) (arrJSON [] map[string] interface {}, err error ) {

	fieldTableName := field.Table.Name

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
func getTABLEID_Values(field *schema.FieldStructure, where, fieldID string) (arrJSON [] map[string] interface {}, err error ){

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

	if err != nil {
		logs.ErrorLog(err, sqlQuery)
		return nil, err
	}
	return rows, nil
}

func (qb * QueryBuilder) SelectToMultidimension() ( arrJSON [] map[string] interface {}, err error ) {

	rows, err := qb.GetDataSql()
	if err != nil {
		//logs.ErrorLog(err) //errors output in qb.GetDataSql()
		return nil, err
	}

	defer rows.Close()

	return qb.ConvertDataToJson(rows)
}

func (qb * QueryBuilder) ConvertDataToJson(rows *sql.Rows) ( arrJSON [] map[string] interface {}, err error ) {


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
			schema:= field.schema
			if fieldName == "id" {
				fieldID = field.Value
			}
			if schema.SETID  {
				where := field.WhereFromSet()
				values[fieldName], err = getSETID_Values(schema, where, fieldID)
				if err != nil {
					logs.ErrorLog(err, field.SQLforFORMList)
					values[fieldName] = err.Error()
				}
				continue
			} else if schema.NODEID {

				where := field.WhereFromSet()
				values[fieldName], err = getNODEID_Values(schema, where, fieldID)
				if err != nil {
					logs.ErrorLog(err, field.SQLforFORMList)
					values[fieldName] = err.Error()
				}
				continue
			} else if schema.TABLEID {
				where := field.WhereFromSet()
				values[fieldName], err = getTABLEID_Values(schema, where, fieldID)
				if err != nil {
					logs.ErrorLog(err, field.SQLforFORMList)
					values[fieldName] = err.Error()
				}
				continue

			}

			switch schema.DATA_TYPE {
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


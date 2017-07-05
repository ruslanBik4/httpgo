// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"database/sql"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/logs"
	"strings"
)

//SelectToMultidimension(sql string, args ...interface
//@version 1.10 Sergey Litvinov 2017-05-25 15:15
func SelectToMultidimension(sql string, args ...interface{}) (arrJSON []map[string]interface{}, err error) {
	qBuilder := CreateFromSQL(sql)
	qBuilder.AddArgs(args)

	return qBuilder.SelectToMultidimension()
}

func (qb *QueryBuilder) createSQL() (sql string, err error) {

	commaTbl, commaFld := "", ""
	for idx, table := range qb.Tables {

		aliasTable := table.Alias
		//first table must'n to having JOIN property
		if (idx > 0) && (table.Join > "") {
			qb.sqlFrom += " " + table.Join + " " + table.Name + " " + aliasTable + " " + table.Using
		} else {
			qb.sqlFrom += commaTbl + table.Name + " " + aliasTable
		}
		commaTbl = ", "
		if aliasTable > "" {
			aliasTable += "."
		}
		if len(table.Fields) > 0 {
			for alias, field := range table.Fields {
				var queryName string
				if (alias > "") && (alias != field.Name) {
					queryName = ` AS "` + alias + `"`
				}
				if field.schema.COLUMN_TYPE == "calc" {
					qb.sqlSelect += commaFld + field.Name + queryName
				} else {
					qb.sqlSelect += commaFld + aliasTable + field.Name + queryName
				}
				qb.fields = append(qb.fields, table.Fields[alias])
				qb.Aliases = append(qb.Aliases, alias)
				commaFld = ", "
			}
		} else if table.Join == "" {
			qb.sqlSelect += commaFld + aliasTable + "*"
			commaFld = ", "

			for _, fieldStrc := range table.schema.Rows {

				//field := &QBField{Name: fieldStrc.COLUMN_NAME, schema: fieldStrc, Table: table}
				table.AddField("", fieldStrc.COLUMN_NAME)
				//TODO: сделать одно место для добавления полей!
				qb.fields = append(qb.fields, table.Fields[fieldStrc.COLUMN_NAME])
				qb.Aliases = append(qb.Aliases, fieldStrc.COLUMN_NAME)
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
func (qb *QueryBuilder) getWhere() string {
	if qb.Where > "" {
		if strings.Contains(qb.Where, "WHERE") {
			return qb.Where
		} else {
			return " WHERE " + qb.Where
		}
	}
	return ""
}
func (qb *QueryBuilder) unionSQL() string {
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
func (qb *QueryBuilder) GetDataSql() (rows *sql.Rows, err error) {

	if qb.Prepared == nil {
		qb.sqlCommand, err = qb.createSQL()
		logs.DebugLog("sql=", qb.sqlCommand)
		if err == nil {
			qb.Prepared, err = db.PrepareQuery(qb.sqlCommand)
		}
		// здесь отловим все ошибки и запротоколируем!
		if err != nil {
			logs.ErrorLog(err, qb.sqlCommand)
			return nil, err
		}
	}

	return qb.Prepared.Query(qb.Args...)
}
func (qb *QueryBuilder) SelectRunFunc(onReadRow func(columns []string, values []sql.RawBytes, rows *sql.Rows) error) error {

	rows, err := qb.GetDataSql()
	if err != nil {
		logs.ErrorLog(err, qb)
		return err
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values   := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			continue
		}
		if err := onReadRow(columns, values, rows); err != nil {
			return err
		}
	}

	return nil
}

// предназначен для получения данных в формате JSON
func (qb *QueryBuilder) SelectToMultidimension() (arrJSON []map[string]interface{}, err error) {

	rows, err := qb.GetDataSql()
	if err != nil {
		logs.ErrorLog(err, qb)
		return nil, err
	}

	defer rows.Close()

	return qb.ConvertDataToJson(rows)
}

//@func (field * QueryBuilder) ConvertDataToJson(rows *sql.Rows) ( arrJSON [] map[string] interface {}, err error ) {
//@author Sergey Litvinov
//@version 1.00 2017-06-12
func (qb *QueryBuilder) ConvertDataToJson(rows *sql.Rows) (arrJSON []map[string]interface{}, err error) {

	var valuePtrs []interface{}

	for _, field := range qb.fields {
		valuePtrs = append(valuePtrs, &field.Value)
	}

	for rows.Next() {

		values := make(map[string]interface{}, len(qb.fields))
		if err := rows.Scan(valuePtrs...); err != nil {
			logs.ErrorLog(err, valuePtrs, qb)
			continue
		}

		for _, field := range qb.fields {

			fieldName := field.Alias
			// all inline field has QB & we run thiq QB & store result in map
			if field.ChildQB != nil {
				if fieldID, ok := field.Table.Fields["id"]; ok {
					field.ChildQB.Args[0] = string(fieldID.Value)
				} else {
					// проставляем 0 на случай, если в выборке нет ID
					field.ChildQB.Args[0] = 0
				}

				values[fieldName], err = field.ChildQB.SelectToMultidimension()
				if err != nil {
					logs.ErrorLog(err, field.ChildQB)
					values[fieldName] = err.Error()
				}
				continue
			}

			values[fieldName] = field.getNativeValue(true)
		}

		arrJSON = append(arrJSON, values)
	}

	return arrJSON, nil
}

//(field * QueryBuilder) ConvertDataNotChangeType(rows *sql.Rows) ( arrJSON [] map[string] interface {}, err error )
//Not Convert BooleanType
//@author Sergey Litvinov
func (qb *QueryBuilder) ConvertDataNotChangeType(rows *sql.Rows) (arrJSON []map[string]interface{}, err error) {

	var valuePtrs []interface{}

	for _, field := range qb.fields {
		valuePtrs = append(valuePtrs, &field.Value)
	}

	columns, _ := rows.Columns()
	for rows.Next() {

		values := make(map[string]interface{}, len(qb.fields))
		if err := rows.Scan(valuePtrs...); err != nil {
			logs.ErrorLog(err, valuePtrs)
			continue
		}

		for idx, fieldName := range columns {

			field := qb.fields[idx]
			if field == nil {
				logs.DebugLog("nil field", idx)
				continue

			}

			if field.ChildQB != nil {
				if fieldID, ok := field.Table.Fields["id"]; ok {
					field.ChildQB.Args[0] = fieldID.Value
				} else {
					// проставляем 0 на случай, если в выборке нет ID
					field.ChildQB.Args[0] = 0
				}

				values[fieldName], err = field.ChildQB.SelectToMultidimension()
				if err != nil {
					logs.ErrorLog(err, field.ChildQB)
					values[fieldName] = err.Error()
				}
				continue
			}

			values[fieldName] = field.getNativeValue(false)
		}

		arrJSON = append(arrJSON, values)
	}

	return arrJSON, nil
}

//@func (field * QueryBuilder) SelectToNotChangeBoolean() ( arrJSON [] map[string] interface {}, err error )
// Get rows not convert tinyInt fields
//@author Sergey Litvinov
func (qb *QueryBuilder) GetSelectToNotChangeBoolean() (arrJSON []map[string]interface{}, err error) {

	rows, err := qb.GetDataSql()
	if err != nil {
		logs.ErrorLog(err, qb)
		return nil, err
	}

	defer rows.Close()
	return qb.ConvertDataNotChangeType(rows)
}

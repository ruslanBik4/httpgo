// Copyright 2017
// 	Author's: Mykhailo Sizov sizov.mykhailo@gmail.com
// All rights reserved.
// version 1.0
// Базовый функционал для работы с транзакциями.
// Важно : откат изменений происходит только в рамках данной транзакции!
package db

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/ruslanBik4/httpgo/models/logs"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type TxConnect struct {
	tx *sql.Tx
}

func (conn *TxConnect) PrepareQuery(sql string) (*sql.Stmt, error) {
	return conn.tx.Prepare(sql)
}

func StartTransaction() (*TxConnect, error) {
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	return &TxConnect{tx: tx}, nil
}
func (conn *TxConnect) CommitTransaction() {
	conn.tx.Commit()
}
func (conn *TxConnect) RollbackTransaction() {
	conn.tx.Rollback()
}
func (conn *TxConnect) prepareQuery(sql string) (*sql.Stmt, error) {
	return conn.tx.Prepare(sql)
}
func (conn *TxConnect) DoInsert(sql string, args ...interface{}) (int, error) {

	resultSQL, err := conn.tx.Exec(sql, args...)

	if err != nil {
		logs.ErrorLog(err, sql)
		return -1, err
	} else {
		lastInsertId, err := resultSQL.LastInsertId()
		return int(lastInsertId), err
	}
}

func (conn *TxConnect) DoUpdate(sql string, args ...interface{}) (int, error) {
	resultSQL, err := conn.tx.Exec(sql, args...)

	if err != nil {
		return -1, err
	} else {
		RowsAffected, err := resultSQL.RowsAffected()
		return int(RowsAffected), err
	}
}
func (conn *TxConnect) DoSelect(sql string, args ...interface{}) (*sql.Rows, error) {
	if SQLvalidator.MatchString(strings.ToLower(sql)) {
		return conn.tx.Query(sql, args...)
	} else {
		return nil, mysql.ErrMalformPkt
	}
}
func (conn *TxConnect) DoQuery(sql string, args ...interface{}) *sql.Rows {
	var result bytes.Buffer

	w := io.Writer(&result)
	Encode := json.NewEncoder(w)

	if strings.HasPrefix(sql, "insert") {
		resultSQL, err := conn.tx.Exec(sql, args...)

		if err != nil {
			Encode.Encode(err)
		} else {
			lastInsertId, _ := resultSQL.LastInsertId()
			Encode.Encode(lastInsertId)
		}

		logs.DebugLog(result.String())

		return nil
	}

	if strings.HasPrefix(sql, "update") {
		resultSQL, err := conn.tx.Exec(sql, args...)

		if err != nil {
			Encode.Encode(err)
		} else {
			RowsAffected, _ := resultSQL.RowsAffected()
			Encode.Encode(RowsAffected)
		}

		logs.DebugLog(result.String())

		return nil
	}

	rows, err := conn.tx.Query(sql, args...)

	if err != nil {
		logs.ErrorLog(err, result.String())
		return nil
	}

	return rows
}

type MultiQueryTransact struct {
	tx      *TxConnect
	Queryes map[string]*ArgsQuery
}

const _2K = (1 << 10) * 2

//выполняет запрос согласно переданным данным в POST,
//для суррогатных полей готовит запросы для изменения связанных полей
//возвращает id новой записи
func (conn *TxConnect) DoInsertFromForm(r *http.Request, userID string) (lastInsertId int, err error) {

	r.ParseMultipartForm(_2K)

	if r.FormValue("table") == "" {
		logs.ErrorLog(errors.New("not table name"))
		return -1, http.ErrNotSupported
	}

	tableName := r.FormValue("table")

	var row argsRAW
	var tableIDQueryes MultiQueryTransact
	tableIDQueryes.tx = conn
	tableIDQueryes.Queryes = make(map[string]*ArgsQuery, 0)

	comma, sqlCommand, values := "", "insert into "+tableName+"(", "values ("

	for key, val := range r.Form {

		indSeparator := strings.Index(key, ":")

		if key == "table" {
			continue
		} else if strings.HasPrefix(key, "setid_") {
			tableProps := getTableProps(key, "setid_")
			defer func(tableName, tableProps string, values []string) {
				if err == nil {
					//TODO принять функцию
					err = conn.insertMultiSet(tableName, tableProps,
						tableName+"_"+tableProps+"_has", userID, values, lastInsertId)
				}
			}(tableName, tableProps, val)
			continue
		} else if strings.HasPrefix(key, "nodeid_") {
			tableProps := getTableProps(key, "nodeid_")
			defer func(tableName, tableValues string, values []string) {
				if err == nil {
					tableProps := GetNameTableProps(tableValues, tableName)
					if tableProps == "" {
						logs.DebugLog("Empty tableProps! ", tableValues)
					}
					err = conn.insertMultiSet(tableName, tableProps, tableValues, userID, values, lastInsertId)
				}
			}(tableName, tableProps, val)
			continue
		} else if key == "id_users" {

			sqlCommand += comma + "`" + key + "`"
			row = append(row, userID)

		} else if strings.Contains(key, "[]") {
			sqlCommand += comma + "`" + strings.TrimRight(key, "[]") + "`"
			str, comma := "", ""
			for _, value := range val {
				str += comma + value
				comma = ","
			}
			row = append(row, str)
		} else if (indSeparator > 1) && strings.Contains(key, "[") {
			tableIDQueryes.addNewParam(key, indSeparator, val)
			continue
		} else {
			sqlCommand += comma + "`" + key + "`"
			row = append(row, val[0])
		}
		values += comma + "?"
		comma = ", "

	}

	// если будут дополнительные запросы
	if len(tableIDQueryes.Queryes) > 0 {
		// исполнить по завершению функции, чтобы получить lastInsertId
		defer func() {
			if err == nil {
				err = tableIDQueryes.runQueryes(tableName, lastInsertId, tableIDQueryes.Queryes)
			}
		}()

	}
	return conn.DoInsert(sqlCommand+") "+values+")", row...)

}

//выполняет запрос согласно переданным данным в POST,
//для суррогатных полей готовит запросы для изменения связанных полей
//возвращает количество измененных записей
func (conn *TxConnect) DoUpdateFromForm(r *http.Request, userID string) (RowsAffected int, err error) {

	r.ParseMultipartForm(_2K)

	if r.FormValue("table") == "" {

		logs.ErrorLog(errors.New("not table name"))
		return -1, http.ErrNotSupported
	}

	tableName := r.FormValue("table")
	var row argsRAW
	var id int
	var tableIDQueryes MultiQueryTransact
	tableIDQueryes.Queryes = make(map[string]*ArgsQuery, 0)

	comma, sqlCommand, where := "", "update "+tableName+" set ", " where id="

	for key, val := range r.Form {

		indSeparator := strings.Index(key, ":")
		if key == "table" {
			continue
		} else if key == "id" {
			where += val[0]
			id, _ = strconv.Atoi(val[0])
			continue
		} else if strings.HasPrefix(key, "setid_") {
			tableProps := getTableProps(key, "setid_")
			defer func(tableProps string, values []string) {
				if err == nil {
					err = conn.insertMultiSet(tableName, tableProps,
						tableName+"_"+tableProps+"_has", userID, values, id)
				} else {
					logs.ErrorLog(err)
				}
			}(tableProps, val)
			continue
		} else if strings.HasPrefix(key, "nodeid_") {
			tableProps := getTableProps(key, "nodeid_")
			defer func(tableValues string, values []string) {
				if err == nil {
					tableProps := GetNameTableProps(tableValues, tableName)
					if tableProps == "" {
						logs.DebugLog("Empty tableProps! ", tableValues)
					}
					err = conn.insertMultiSet(tableName, tableProps, tableValues, userID, values, id)
				} else {
					logs.ErrorLog(err)
				}
			}(tableProps, val)
			continue
		} else if key == "id_users" {

			sqlCommand += comma + "`" + key + "`=?"
			row = append(row, userID)

		} else if strings.Contains(key, "[]") {
			sqlCommand += comma + "`" + strings.TrimRight(key, "[]") + "`=?"
			str, comma := "", ""
			for _, value := range val {
				str += comma + value
				comma = ","
			}
			row = append(row, str)
		} else if (indSeparator > 1) && strings.Contains(key, "[") {
			tableIDQueryes.addNewParam(key, indSeparator, val)
			continue
		} else {
			sqlCommand += comma + "`" + key + "`=?"
			row = append(row, val[0])
		}
		comma = ", "

	}
	// если будут дополнительные запросы
	if len(tableIDQueryes.Queryes) > 0 {
		// исполнить по завершению функции, чтобы получить lastInsertId
		defer func() {
			if err == nil {
				err = tableIDQueryes.runQueryes(tableName, id, tableIDQueryes.Queryes)
			}
		}()

	}
	return conn.DoUpdate(sqlCommand+where, row...)

}

func (conn *TxConnect) addNewItem(tableProps, value, userID string) (int, error) {

	if newId, err := conn.DoInsert("insert into "+tableProps+"(title, id_users) values (?, ?)", value, userID); err != nil {
		return -1, err
	} else {
		return newId, nil
	}

}

func (conn *TxConnect) insertMultiSet(tableName, tableProps, tableValues, userID string, values []string, id int) (err error) {

	sqlCommand := fmt.Sprintf("insert IGNORE into %s (id_%s, id_%s) values (%d, ?)",
		tableValues, tableName, tableProps, id)
	smtp, err := conn.PrepareQuery(sqlCommand)
	if err != nil {
		logs.ErrorLog(err)
		return err
	}
	var params, comma string
	var valParams argsRAW

	for _, value := range values {

		// если не числовое значение - стало быть, это новое свойство и его добавим в таблицу свойств
		if !DigitsValidator.MatchString(value) {
			newId, err := conn.addNewItem(tableProps, value, userID)
			if err != nil {
				logs.ErrorLog(err)
				continue
			}
			value = strconv.Itoa(newId)
		}
		if resultSQL, err := smtp.Exec(value); err != nil {
			logs.ErrorLog(err, sqlCommand)
		} else {
			logs.DebugLog(resultSQL)
		}
		params += comma + "?"
		valParams = append(valParams, value)
		comma = ","
	}
	sqlCommand = fmt.Sprintf("delete from %s where id_%s = %d AND id_%s not in (%s)",
		tableValues, tableName, id, tableProps, params)

	if smtp, err = conn.PrepareQuery(sqlCommand); err != nil {
		logs.ErrorLog(err)
		return err
	}

	if resultSQL, err := smtp.Exec(valParams...); err != nil {
		logs.ErrorLog(err, sqlCommand)
		return err
	} else {
		logs.DebugLog(resultSQL)
	}

	return err
}

func (tableIDQueryes *MultiQueryTransact) addNewParam(key string, indSeparator int, val []string) {
	tableName := key[:indSeparator]
	query, ok := tableIDQueryes.Queryes[tableName]
	if !ok {
		query = &ArgsQuery{
			Comma:     "",
			FieldList: "",
			Values:    "",
		}
	}
	fieldName := key[strings.Index(key, ":")+1:]
	pos := strings.Index(fieldName, "[")
	fieldName = "`" + fieldName[:pos] + "`"

	// пока беда в том, что количество должно точно соответствовать!
	//если первый  - то создаем новый список параметров для вставки
	if strings.HasPrefix(query.FieldList, fieldName) {
		query.Comma = "), ("
	} else if !strings.Contains(query.FieldList, fieldName) {
		query.FieldList += query.Comma + fieldName
	}

	query.Values += query.Comma + "?"
	query.Comma = ", "
	query.Args = append(query.Args, val)
	tableIDQueryes.Queryes[tableName] = query

}

func (tableIDQueryes *MultiQueryTransact) runQueryes(tableName string, lastInsertId int, Queryes map[string]*ArgsQuery) (err error) {

	parentKey := "id_" + tableName
	for childTableName, query := range Queryes {

		isNotContainParentKey := !strings.Contains(query.FieldList, parentKey)
		if isNotContainParentKey {
			query.FieldList += query.Comma + parentKey
			query.Values += query.Comma + "?"
		}
		fullCommand := fmt.Sprintf("replace into %s (%s) values (%s)", childTableName, query.FieldList, query.Values)

		var args []interface{}

		for i := range query.Args[0].([]string) {
			if i > 0 {
				fullCommand += ",(" + query.Values + ")"
			}
			for _, valArr := range query.Args {
				switch valArr.(type) {
				case []string:
					args = append(args, valArr.([]string)[i])
				default:
					args = append(args, valArr)

				}
			}
			// последним добавляем вторичный ключ
			if isNotContainParentKey {
				args = append(args, lastInsertId)
			}
		}
		if id, err := tableIDQueryes.tx.DoInsert(fullCommand, args...); err != nil {
			logs.ErrorLog(err)
		} else {
			logs.DebugLog(fullCommand, id)
		}
	}
	return err
}

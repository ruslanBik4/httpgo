// Package db
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
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/ruslanBik4/httpgo/models/logs"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// тип данных, что хранит транзакцию
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

//выполняет запрос согласно переданным данным в POST,
//для суррогатных полей готовит запросы для изменения связанных полей
//возвращает id новой записи
func (conn *TxConnect) DoInsertFromForm(r *http.Request, userID string) (lastInsertId int, err error) {

	return DoInsertFromForm(r, userID, conn)

}

//выполняет запрос согласно переданным данным в POST,
//для суррогатных полей готовит запросы для изменения связанных полей
//возвращает количество измененных записей
func (conn *TxConnect) DoUpdateFromForm(r *http.Request, userID string) (RowsAffected int, err error) {
	return DoUpdateFromForm(r, userID, conn)
}

func (conn *TxConnect) addNewItem(tableProps, value, userID string) (int, error) {

	if newId, err := conn.DoInsert("insert into "+tableProps+"(title, id_users) values (?, ?)", value, userID); err != nil {
		return -1, err
	} else {
		return newId, nil
	}

}

//TODO: добавить запись для мультиполей (setid_)
func (conn *TxConnect) insertMultiSet(tableName, tableProps, tableValues, userID string, values []string, id int) (err error) {

	// для обновление связей полей пытаемся вставить новую связку
	// родительской таблицы с таблицей свойств
	// игнорируем6 ЕСЛИ УЖЕ ЕСТЬ ТАКАЯ СВЯЗКА
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

		id, err := strconv.Atoi(value)
		// если не числовое значение - стало быть, это новое свойство и его добавим в таблицу свойств
		if err != nil {
			id, err = conn.addNewItem(tableProps, value, userID)
			if err != nil {
				logs.ErrorLog(err, value)
				return err
			}
			//value = strconv.Itoa(newId)
		}
		if resultSQL, err := smtp.Exec(value); err != nil {
			logs.ErrorLog(err)
		} else {
			logs.DebugLog(resultSQL.LastInsertId())
		}
		params += comma + "?"
		valParams = append(valParams, id)
		comma = ","
	}
	// теперь удалим все записи, которые НЕ пришли в запросе
	sqlCommand = fmt.Sprintf("delete from %s where id_%s = %d AND id_%s not in (%s)",
		tableValues, tableName, id, tableProps, params)

	if smtp, err = conn.PrepareQuery(sqlCommand); err != nil {
		logs.ErrorLog(err)
		return err
	}

	if resultSQL, err := smtp.Exec(valParams...); err != nil {
		logs.ErrorLog(err, valParams)
		return err
	} else {
		logs.DebugLog(resultSQL)
	}

	return err

}

//GetDataCustom get data with custom sql query
//Give  begSQL + tableName + endSQL, split, run sql
//DoUpdateFromMap - function generate sql query from data map
package db

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/go-sql-driver/mysql"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/server"
)

var (
	mysqlMu      sync.RWMutex
	dbConn       *sql.DB
	SQLvalidator = regexp.MustCompile(`^(\s*)?(\()?select(\s+.+\s)+(\s*)?from\s+`)
	//регулярное выражение вытаскивающее имя таблицы из запроса
	//TODO не отрабатывает конструкцию FROM table1, table2
	tableNameFromSQL = regexp.MustCompile(`(?is)(?:from|into|update|join)\s+(\w+)`)
)

//SqlCustom На основе этой структуры формируется запрос вида sqlBeg + table + sqlEnd
type SqlCustom struct {
	Table  string
	SqlBeg string
	SqlEnd string
	Sql    string
}

type ErrBadSelectQuery struct {
	Sql, Message string
}

func (err ErrBadSelectQuery) Error() string {
	return "Bad query for select - " + err.Sql
}
func PrepareQuery(sql string) (*sql.Stmt, error) {
	return dbConn.Prepare(sql)
}

//doConnect() error
func doConnect() error {
	var DriveName mysql.MySQLDriver
	var err error

	if dbConn != nil {
		return nil
	}
	// lock for prepare dublicate DB call
	mysqlMu.Lock()
	defer func() {
		mysqlMu.Unlock()
	}()
	serverConfig := server.GetServerConfig()
	dbConn, err = sql.Open("mysql", serverConfig.DNSConnection())
	if err != nil {
		logs.ErrorLog(err)
		return err
	} else if dbConn == nil {
		logs.DebugLog(DriveName)
		return mysql.ErrInvalidConn
	}
	dbConn.Ping()

	dbConn.SetMaxOpenConns(250)
	logs.StatusLog(dbConn)

	return nil

}

//DoInsert(sql string, args ...interface
func DoInsert(sql string, args ...interface{}) (int, error) {

	logs.DebugLog(sql)
	if err := doConnect(); err != nil {
		return -1, err
	}

	resultSQL, err := dbConn.Exec(sql, args...)

	if err != nil {
		logs.ErrorLog(err, sql)
		logs.ErrorStack(err)
		return -1, err
	} else {
		lastInsertId, err := resultSQL.LastInsertId()
		return int(lastInsertId), err
	}
}

//DoUpdate(sql string, args ...interface
func DoUpdate(sql string, args ...interface{}) (int, error) {

	if err := doConnect(); err != nil {
		return -1, err
	}

	resultSQL, err := dbConn.Exec(sql, args...)

	if err != nil {
		return -1, err
	} else {
		RowsAffected, err := resultSQL.RowsAffected()
		return int(RowsAffected), err
	}
}

//DoSelect(sql string, args ...interface
func DoSelect(sql string, args ...interface{}) (rows *sql.Rows, err error) {

	if err = doConnect(); err != nil {
		return nil, err
	}
	if SQLvalidator.MatchString(strings.ToLower(sql)) {
		return dbConn.Query(sql, args...)
	} else {
		logs.ErrorLog(&ErrBadSelectQuery{Message: "Bad query for select -", Sql: sql})
		logs.ErrorStack(err)
		return nil, err
	}
}
func DoQuery(sql string, args ...interface{}) *sql.Rows {

	if err := doConnect(); err != nil {
		return nil
	}

	var result bytes.Buffer

	w := io.Writer(&result)
	Encode := json.NewEncoder(w)

	if strings.HasPrefix(sql, "insert") {
		resultSQL, err := dbConn.Exec(sql, args...)

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
		resultSQL, err := dbConn.Exec(sql, args...)

		if err != nil {
			Encode.Encode(err)
		} else {
			RowsAffected, _ := resultSQL.RowsAffected()
			Encode.Encode(RowsAffected)
		}

		logs.DebugLog(result.String())

		return nil
	}

	rows, err := dbConn.Query(sql, args...)

	if err != nil {
		logs.ErrorLog(err, result.String())
		return nil
	}

	return rows
}

type rowFields struct {
	row map[string]string
}

// func (rows *Myrow)  Scan(value interface{}) err {
//
//     return true
//
// }
//подготовка для цикла чтения записей, формирует row для сканирования записи,rowField - для выборки значение и массив типов для последующей обработки
func PrepareRowsToReading(rows *sql.Rows) (row []interface{}, rowField map[string]*sql.NullString,
	columns []string, colTypes []*sql.ColumnType) {

	columns, err := rows.Columns()

	if err != nil {
		panic(err)
	}

	colTypes, err = rows.ColumnTypes()
	if err != nil {
		panic(err)
	}

	rowField = make(map[string]*sql.NullString, len(columns))

	for _, val := range columns {

		rowField[val] = new(sql.NullString)
		row = append(row, rowField[val])
	}

	return row, rowField, columns, colTypes

}

//GetResultToJSON (rows *sql.Rows) []byte
func GetResultToJSON(rows *sql.Rows) []byte {

	var rowOutput []map[string]string

	row, rowField, _, _ := PrepareRowsToReading(rows)

	var result bytes.Buffer
	w := io.Writer(&result)
	Encode := json.NewEncoder(w)

	defer rows.Close()
	for rows.Next() {

		if err := rows.Scan(row...); err != nil {
			fmt.Println("err:", err)
			continue
		}

		output := make(map[string]string)

		for name, field := range rowField {
			if field.Valid {
				output[name] = field.String
			} else {
				output[name] = "NULL"
			}
		}

		rowOutput = append(rowOutput, output)
	}

	if err := Encode.Encode(rowOutput); err != nil {
		Encode.Encode(err)
	}

	return result.Bytes()
}

//getValue(fieldValue *sql.NullString) string
func getValue(fieldValue *sql.NullString) string {
	if fieldValue.Valid {
		return fieldValue.String
	}

	return "NULL"
}

func getTableFromSchema(tableName string) *schema.FieldsTable {

	defer func() {
		result := recover()
		switch err := result.(type) {
		case schema.ErrNotFoundTable:
			logs.ErrorLog(err, "tableName", tableName)
			//err = schema.ErrNotFoundTable{Table:tablePart[1]}
		case nil:
		case error:
			panic(err)
		}
	}()

	return schema.GetFieldsTable(tableName)
}

// get table names from sql-query
func getTablesFromSQL(sql string) (tables []*schema.FieldsTable) {

	arrTables := tableNameFromSQL.FindAllStringSubmatch(sql, -1)
	for _, tablePart := range arrTables {

		//for _, tableName := range tablePart {

		fields := getTableFromSchema(tablePart[1])
		tables = append(tables, fields)
		//}
	}

	return tables
}

//GetDataPrepareRowsToReading - function get rows with structure field
//@author Sergey Litvinov
func GetDataPrepareRowsToReading(sql string, args ...interface{}) (rows *sql.Rows, row []interface{}, rowField map[string]*sql.NullString,
	columns []string, colTypes []*sql.ColumnType, err error) {
	rows, err = DoSelect(sql, args...)

	if err != nil {
		logs.ErrorLog(err)
		return nil, nil, nil, nil, nil, err
	}

	row, rowField, columns, colTypes = PrepareRowsToReading(rows)
	return rows, row, rowField, columns, colTypes, nil
}

//ConvertPrepareRowsToJson convert many rows to json
//@author Sergey Litvinov
func ConvertPrepareRowsToJson(rows *sql.Rows, row []interface{}, rowField map[string]*sql.NullString,
	columns []string, colTypes []*sql.ColumnType) (arrJSON []map[string]interface{}, err error) {

	logs.DebugLog(rows)

	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			logs.ErrorLog(err)
			continue
		}

		logs.DebugLog(row...)

		values := make(map[string]interface{}, len(columns))

		for _, colType := range colTypes {

			fieldName := colType.Name()
			fieldValue, ok := rowField[fieldName]
			if !ok {
				logs.ErrorLog(err)
				continue
			}
			//logs.DebugLog(colType.Length())
			switch colType.DatabaseTypeName() {
			case "varchar", "date", "datetime":
				if fieldValue.Valid {
					values[fieldName] = fieldValue.String
				} else {
					values[fieldName] = nil
				}
			case "tinyint":
				if getValue(fieldValue) == "1" {
					values[fieldName] = true

				} else {
					values[fieldName] = false

				}
			case "int", "int64", "float":
				values[fieldName], _ = strconv.Atoi(getValue(fieldValue))
			default:
				if fieldValue.Valid {
					values[fieldName] = fieldValue.String
				} else {
					values[fieldName] = nil
				}
			}
		}

		arrJSON = append(arrJSON, values)
	}

	return arrJSON, nil
}

//ConvertPrepareRowToJson convert one row to json
//@author Sergey Litvinov
func ConvertPrepareRowToJson(rowField map[string]*sql.NullString, columns []string,
	colTypes []*sql.ColumnType) (id int, arrJSON map[string]interface{}, err error) {
	id = 0
	values := make(map[string]interface{}, len(columns))

	for _, colType := range colTypes {

		fieldName := colType.Name()
		fieldValue, ok := rowField[fieldName]
		if fieldName == "id" {
			id, _ = strconv.Atoi(getValue(fieldValue))
		}
		if !ok {
			logs.ErrorLog(err)
			continue
		}
		//logs.DebugLog(colType.Length())
		switch colType.DatabaseTypeName() {
		case "varchar", "date", "datetime":
			if fieldValue.Valid {
				values[fieldName] = fieldValue.String
			} else {
				values[fieldName] = nil
			}
		case "tinyint":
			if getValue(fieldValue) == "1" {
				values[fieldName] = true

			} else {
				values[fieldName] = false

			}
		case "int", "int64", "float":
			values[fieldName], _ = strconv.Atoi(getValue(fieldValue))

		default:
			if fieldValue.Valid {
				values[fieldName] = fieldValue.String
			} else {
				values[fieldName] = nil
			}
		}
	}

	arrJSON = values

	return id, arrJSON, nil
}

//GetDataCustom get data with custom sql query
//Give  begSQL + tableName + endSQL, split, run sql
//@version 2.00 2017-04-20
//@author Sergey Litvinov
func GetDataCustom(sqlParam SqlCustom, args ...interface{}) (rows *sql.Rows,
	row []interface{}, rowField map[string]*sql.NullString, columns []string,
	colTypes []*sql.ColumnType, err error) {

	if sqlParam.SqlBeg == "" {
		sqlParam.SqlBeg = "SELECT * FROM "
	}

	if sqlParam.SqlEnd == "" {
		sqlParam.SqlEnd = " WHERE id=?"
	}
	if sqlParam.Sql == "" && sqlParam.Table != "" {
		sqlParam.Sql = sqlParam.SqlBeg + sqlParam.Table + sqlParam.SqlEnd
	}
	if sqlParam.Sql != "" {
		logs.DebugLog("sqlParam.Sql=", sqlParam.Sql)
		rows, row, rowField, columns, colTypes, err = GetDataPrepareRowsToReading(sqlParam.Sql, args...)

		if err != nil {
			logs.ErrorLog(err)
			return nil, nil, nil, nil, nil, err
		}

		return rows, row, rowField, columns, colTypes, nil
	} else {
		err = errors.New("Error. Not enough parameters for the function GetDataCustom")
		return nil, nil, nil, nil, nil, err
	}

}

//DoUpdateFromMap - function generate sql update query from data map with current id
//@author Sergey Litvinov
//@version 1.10 2017-06-15 Add new validation type(Sergey Litvinov)
func DoUpdateFromMap(table string, mapData map[string]interface{}) (RowsAffected int, err error) {
	var row argsRAW
	var id int

	comma, sqlCommand, where := "", "UPDATE "+table+" SET ", " WHERE "

	for key, val := range mapData {
		if key == "id" {
			where += "`" + key + "`=?"
			switch val.(type) {
			case int:
				id = val.(int)
			case string:
				id, _ = strconv.Atoi(val.(string))
			default:
				logs.ErrorLog(errors.New("Not correct fornmat id"))
			}
			logs.DebugLog(val)
			continue
		} else {
			sqlCommand += comma + "`" + key + "`=?"
			row = append(row, val)
		}
		comma = ", "

	}
	row = append(row, id)
	sqlCommand = sqlCommand + where
	//logs.DebugLog("sql ",  sqlCommand)
	//logs.DebugLog(  row... )

	RowsAffected, err = DoUpdate(sqlCommand, row...)
	return RowsAffected, err
}

// Функция возвращает результат выполнения запроса в заданой структуре
func PerformSelectQuery(sql string, args ...interface{}) (arrJSON []map[string]interface{}, err error) {

	rows, err := DoSelect(sql, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	valuePtrs, rowValues, columns, colTypes := PrepareRowsToReading(rows)

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			logs.ErrorLog(err)
			continue
		}
		values := make(map[string]interface{}, len(columns))
		for _, colType := range colTypes {

			fieldName := colType.Name()
			fieldValue, ok := rowValues[fieldName]
			if !ok {
				logs.ErrorLog(err)
				continue
			}
			if fieldValue.Valid {
				values[fieldName] = fieldValue.String
			} else {
				values[fieldName] = nil
			}
		}
		arrJSON = append(arrJSON, values)
	}
	return arrJSON, nil
}

//GetDataCustom get data with custom sql query
//Give  begSQL + tableName + endSQL, split, run sql
//DoUpdateFromMap - function generate sql query from data map
package db

import (
	"fmt"
	"regexp"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"encoding/json"
	"bytes"
	"io"
	"log"
	"strings"
	"strconv"
	"github.com/ruslanBik4/httpgo/models/server"
	"errors"
)
var (
	dbConn *sql.DB
	SQLvalidator = regexp.MustCompile(`^select\s+.+\s*from\s+`)
)
//SqlCustom На основе этой структуры формируется запрос вида sqlBeg + table + sqlEnd
type SqlCustom struct {
    Table  string;
    SqlBeg string;
    SqlEnd string;
    Sql    string;
}


func prepareQuery(sql string) (*sql.Stmt, error){
	return dbConn.Prepare(sql)
}
func doConnect() error {
	var DriveName mysql.MySQLDriver
	var err error

	if dbConn != nil {
		return nil
	}
	serverConfig := server.GetServerConfig()
	dbConn, err = sql.Open( "mysql", serverConfig.DNSConnection() )
	if err != nil {
		log.Println(err)
		return err
	} else if dbConn == nil {
		log.Println( DriveName )
		return sql.ErrTxDone
	}

	return nil

}
func DoInsert(sql string, args ...interface{}) (int, error) {

	if err := doConnect(); err != nil {
		return -1, err
	}

	resultSQL, err := dbConn.Exec( sql, args ...)

	if err != nil {
		log.Println(sql)
		return -1, err
	} else {
		lastInsertId, err := resultSQL.LastInsertId()
		return int(lastInsertId), err
	}
}
func DoUpdate(sql string, args ...interface{}) (int, error) {

	if err := doConnect(); err != nil {
		return -1, err
	}

	resultSQL, err := dbConn.Exec( sql, args ...)

	if err != nil {
		return -1, err
	} else {
		RowsAffected, err:= resultSQL.RowsAffected()
		return int(RowsAffected), err
	}
}
func DoSelect(sql string, args ...interface{})  (*sql.Rows, error) {

	if err := doConnect(); err != nil {
		return nil, err
	}
	if SQLvalidator.MatchString(strings.ToLower(sql)) {
		return dbConn.Query(sql, args ...)
	} else {
		return nil, mysql.ErrMalformPkt
	}
}
func DoQuery(sql string, args ...interface{})  *sql.Rows {

	if err := doConnect(); err != nil {
		return nil
	}

	var result bytes.Buffer

	w := io.Writer(&result)
	Encode := json.NewEncoder(w)

	if strings.HasPrefix(sql, "insert") {
		resultSQL, err := dbConn.Exec( sql, args ...)

		if err != nil {
			Encode.Encode(err)
		} else {
			lastInsertId, _:= resultSQL.LastInsertId()
			Encode.Encode(lastInsertId)
		}

		log.Print( result.String() )

		return nil
	}

	if strings.HasPrefix(sql, "update") {
		resultSQL, err := dbConn.Exec( sql, args ...)

		if err != nil {
			Encode.Encode(err)
		} else {
			RowsAffected, _:= resultSQL.RowsAffected()
			Encode.Encode(RowsAffected)
		}

		log.Print( result.String() )

		return nil
	}

	rows, err :=  dbConn.Query( sql, args ...)

	if err != nil {
		log.Print( result.String() )
		return nil
	}

	return rows
}
type rowFields struct {
	row map[string] string
}
// func (rows *Myrow)  Scan(value interface{}) err {
//
//     return true
//
// }
//подготовка для цикла чтения записей, формирует row для сканирования записи,rowField - для выборки значение и массив типов для последующей обработки
func PrepareRowsToReading(rows *sql.Rows) (row [] interface {}, rowField map[string] *sql.NullString, columns [] string, colTypes [] *sql.ColumnType) {

	columns, err := rows.Columns()

	if err != nil {
		panic(err)
	}

	colTypes, err = rows.ColumnTypes()
	if err != nil {
		panic(err)
	}

	rowField = make( map[string] *sql.NullString, len(columns) )

	for _, val := range columns {

		rowField[val] = new(sql.NullString)
		row = append( row, rowField[val] )
	}


	return row, rowField, columns, colTypes

}
func GetResultToJSON (rows *sql.Rows) []byte{

	var rowOutput [] map[string] string

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

		output := make( map[string] string )

		for name, field := range rowField {
			if field.Valid {
				output[name] = field.String
			} else {
				output[name] = "NULL"
			}
		}

		rowOutput = append( rowOutput, output)
	}

	if err := Encode.Encode(rowOutput); err != nil {
		Encode.Encode(err)
	}


	return result.Bytes();
}
type TableOptions struct {
	TABLE_NAME   string
	TABLE_TYPE string
	ENGINE string
	TABLE_COMMENT string
}
type RecordsTables struct {
	Rows [] TableOptions
}
// получение данных для одной таблицы
func (ns *TableOptions) GetTableProp(tableName string) error {

	rows := DoQuery("SELECT TABLE_NAME, TABLE_TYPE, ENGINE, " +
		"IF (TABLE_COMMENT = NULL OR TABLE_COMMENT = '', TABLE_NAME, TABLE_COMMENT) " +
		"FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME=? order by TABLE_COMMENT", tableName)

	for rows.Next() {

		err := rows.Scan( &ns.TABLE_NAME, &ns.TABLE_TYPE, &ns.ENGINE, &ns.TABLE_COMMENT)

		if err != nil {
			log.Println(err)
			return err
		}

	}

	return nil
}
// получение таблиц
func (ns *RecordsTables) GetTablesProp(bd_name string)  error {

	return ns.GetSelectTablesProp( "TABLE_SCHEMA='" + bd_name + "'")

}
func (ns *RecordsTables) GetSelectTablesProp(where string)  error {

	rows := DoQuery("SELECT TABLE_NAME, TABLE_TYPE, ENGINE, " +
		"IF (TABLE_COMMENT = NULL OR TABLE_COMMENT = '', TABLE_NAME, TABLE_COMMENT) " +
		"FROM INFORMATION_SCHEMA.TABLES WHERE " + where + " order by TABLE_COMMENT")

	for rows.Next() {

		var row TableOptions
		err := rows.Scan( &row.TABLE_NAME, &row.TABLE_TYPE, &row.ENGINE, &row.TABLE_COMMENT)

		if err != nil {
			log.Println(err)
			continue
		}

		ns.Rows = append(ns.Rows, row)

	}

	return nil

}
type FieldStructure struct {
	COLUMN_NAME   string
	DATA_TYPE string
	COLUMN_DEFAULT sql.NullString
	IS_NULLABLE string
	CHARACTER_SET_NAME sql.NullString
	COLUMN_COMMENT sql.NullString
	COLUMN_TYPE string
	CHARACTER_MAXIMUM_LENGTH sql.NullInt64
	//TITLE string
	//TYPE_INPUT string
	//IS_VIEW []uint8

}
type FieldsTable struct {
	Rows [] FieldStructure
	Options TableOptions
}

// получение значений полей для форматирования данных
// получение значений полей для таблицы
// @param args []int{} - можно передавать ограничения выводимых полей.
// Действует как LIMIT a или LIMIT a, b
//
func (ns *FieldsTable) GetColumnsProp(table_name string, args ...int) error {

	valuesText := []string{}
	for _, arg := range args {
		valuesText = append(valuesText, strconv.Itoa(arg))
	}

	limiter := strings.Join(valuesText, ", ")
	if limiter != "" {
		limiter = " LIMIT " + limiter
	}


	rows, err := DoSelect("SELECT COLUMN_NAME, DATA_TYPE, COLUMN_DEFAULT, " +
		"IS_NULLABLE, CHARACTER_SET_NAME, COLUMN_COMMENT, COLUMN_TYPE, CHARACTER_MAXIMUM_LENGTH " +
		"FROM INFORMATION_SCHEMA.COLUMNS C " +
		"WHERE TABLE_SCHEMA=? AND TABLE_NAME=? ORDER BY ORDINAL_POSITION" + limiter,
		server.GetServerConfig().DBName(), table_name)
	if err != nil {
		return err
	}

	if ns.Rows == nil {
		ns.Rows = make([] FieldStructure, 0)
	}
	for rows.Next() {
		var row FieldStructure

		err := rows.Scan( &row.COLUMN_NAME, &row.DATA_TYPE, &row.COLUMN_DEFAULT, &row.IS_NULLABLE,
					&row.CHARACTER_SET_NAME, &row.COLUMN_COMMENT, &row.COLUMN_TYPE,
					&row.CHARACTER_MAXIMUM_LENGTH )
			//&row.TITLE, &row.TYPE_INPUT, &row.IS_VIEW,

		if err != nil {
			log.Println(err)
			continue
		}

		ns.Rows = append(ns.Rows, row)

	}

	return nil
}
func getValue(fieldValue *sql.NullString) string {
	if fieldValue.Valid {
		return fieldValue.String
	}

	return "NULL"
}
func getSQLFromSETID(key, parentTable string) string{
	tableProps := strings.TrimPrefix(key, "setid_")
	tableValue := parentTable + "_" + tableProps + "_has"

	titleField := "title" //field.getForeignFields(tableProps)
	if titleField == "" {
		return ""
	}

	return fmt.Sprintf( `SELECT p.id, %s, id_%s
	FROM %s p LEFT JOIN %s v ON (p.id=v.id_%[3]s AND id_%[2]s=?) `,
		titleField, parentTable,
		tableProps, tableValue)

}
//func getSQLFromNodeID(key, parentTable string) string{
//	var tableProps, titleField string
//
//	tableValue := strings.TrimPrefix(key, "nodeid_")
//
//	var ns FieldsTable
//	ns.GetColumnsProp(tableValue)
//
//	var fields FieldsTable
//
//	fields.PutDataFrom(ns)
//
//	for _, field := range fields.Rows {
//		if (field.COLUMN_NAME != "id_" + parentTable) && strings.HasPrefix(field.COLUMN_NAME, "id_") {
//			tableProps = field.COLUMN_NAME[3:]
//			titleField = field.getForeignFields(tableProps)
//			break
//		}
//	}
//
//	if titleField == "" {
//		return ""
//	}
//
//	return fmt.Sprintf( `SELECT p.id, %s, id_%s
//	FROM %s p LEFT JOIN %s v ON (p.id=v.id_%[3]s AND id_%[2]s=?) `,
//		titleField, parentTable,
//		tableProps, tableValue)
//
//}

func SelectToMultidimension(sql string, args ...interface{}) ( arrJSON [] map[string] interface {}, err error ) {

	rows, err := DoSelect(sql, args...)

	if err != nil {
		log.Println(err, sql)
		return nil, err
	}

	defer rows.Close()

	valuePtrs, rowValues, columns, colTypes := PrepareRowsToReading(rows)

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Println(err)
			continue
		}

		values := make(map[string] interface{}, len(columns) )

		id, ok := rowValues["id"]

		if !ok {
			log.Println("not id")
		}

		for _, colType := range colTypes {

			fieldName := colType.Name()
			fieldValue, ok := rowValues[fieldName]
			if !ok {
				log.Println(err)
				continue
			}
			//TODO для полей типа tableid_, setid_, nodeid_ придумать механизм для блока WHERE
			// (по ключу родительской таблицы и патетрну из свойств поля для полей типа set)
			//TODO для полей типа setid_ формировать название таблицы
			//TODO также на уровне функции продумать менанизм, который позволит выбирать НЕ ВСЕ поля из третей таблицы
			if strings.HasPrefix(fieldName, "setid_") || strings.HasPrefix(fieldName, "nodeid_") {
				values[fieldName], err = SelectToMultidimension( getSQLFromSETID(fieldName, "rooms"), id )
				if err != nil {
					log.Println(err)
					values[fieldName] = err.Error()
				}
				continue
			} else if strings.HasPrefix(fieldName, "tableid_"){
				values[fieldName], err = SelectToMultidimension( "SELECT * FROM " + fieldName[ len("tableid_") : ])
				if err != nil {
					log.Println(err)
					values[fieldName] = err.Error()
				}
				continue

			}

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

//GetDataPrepareRowsToReading - function get rows with structure field
//@aurhor Sergey Litvinov
func GetDataPrepareRowsToReading(sql string, args ...interface{})  (rows *sql.Rows, row [] interface {}, rowField map[string] *sql.NullString,
							columns [] string, colTypes [] *sql.ColumnType, err error )  {
	rows, err = DoSelect(sql, args...)

	if err != nil {
		log.Println(err)
		return nil, nil, nil, nil, nil, err
	}

	row, rowField, columns, colTypes = PrepareRowsToReading(rows)
	return rows, row, rowField, columns, colTypes, nil
}

//ConvertPrepareRowsToJson convert many rows to json
//@aurhor Sergey Litvinov
func ConvertPrepareRowsToJson(rows *sql.Rows, row [] interface {}, rowField map[string] *sql.NullString,
			columns [] string, colTypes [] *sql.ColumnType) ( arrJSON [] map[string] interface {}, err error ) {

    log.Println( rows)

	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			log.Println(err)
			continue
		}
		log.Println("row=")
		log.Println(row...)

		values := make(map[string] interface{}, len(columns) )

		for _, colType := range colTypes {

			fieldName := colType.Name()
			fieldValue, ok := rowField[fieldName]
			if !ok {
				log.Println(err)
				continue
			}
			//log.Println(colType.Length())
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
//@aurhor Sergey Litvinov
func ConvertPrepareRowToJson(rowField map[string] *sql.NullString, columns [] string,
		colTypes [] *sql.ColumnType) (id int, arrJSON map[string] interface {},   err error ) {
		id = 0;
		values := make(map[string] interface{}, len(columns) )

		for _, colType := range colTypes {

			fieldName := colType.Name()
			fieldValue, ok := rowField[fieldName]
			if (fieldName == "id"){
				id, _ = strconv.Atoi(getValue(fieldValue))
			}
			if !ok {
				log.Println(err)
				continue
			}
			//log.Println(colType.Length())
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
//@aurhor Sergey Litvinov
func GetDataCustom(sqlParam SqlCustom, args ...interface{}) (rows *sql.Rows,
		row [] interface {}, rowField map[string] *sql.NullString, columns [] string,
		colTypes [] *sql.ColumnType, err error ){

	if(sqlParam.SqlBeg == ""){
		sqlParam.SqlBeg = "SELECT * FROM "
	}

	if(sqlParam.SqlEnd == ""){
		sqlParam.SqlEnd = " WHERE id=?"
	}
	if (sqlParam.Sql == "" &&  sqlParam.Table !=""){
		sqlParam.Sql = sqlParam.SqlBeg + sqlParam.Table + sqlParam.SqlEnd
	}
	if (sqlParam.Sql != ""){
        log.Print("sqlParam.Sql=", sqlParam.Sql)
		rows, row, rowField, columns, colTypes, err = GetDataPrepareRowsToReading(sqlParam.Sql, args...)

		if err != nil {
			log.Println(err)
			return nil, nil, nil, nil, nil, err
		}

		return rows, row, rowField, columns, colTypes, nil
	} else{
		err = errors.New("Error. Not enough parameters for the function GetDataCustom")
		return nil, nil, nil, nil, nil, err
	}


}


//DoUpdateFromMap - function generate sql query from data map
//@version 1.00 2017-04-11
//@aurhor Sergey Litvinov
func DoUpdateFromMap(table string, mapData map[string] interface{}) (RowsAffected int, err error) {

	var row argsRAW
	var id int
	var tableIDQueryes MultiQuery
	tableIDQueryes.Queryes = make(map[string] *ArgsQuery, 0)

	comma, sqlCommand, where := "", "UPDATE " + table + " SET ", " WHERE id="

	for key, val := range mapData {
		if key == "id" {
			where += val.(string)
			id, _ = strconv.Atoi(val.(string))
			continue
		}  else {
			sqlCommand += comma + "`" + key + "`=?"
			row = append( row, val )
		}
		comma = ", "

	}
	row = append( row, id )
    RowsAffected, err = DoUpdate(sqlCommand + where, row ... )
	return RowsAffected, err

}

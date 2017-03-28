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
)
var (
	dbConn *sql.DB
	SQLvalidator = regexp.MustCompile(`^select\s+.+\s*from\s+`)

)
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
		notTable, error := regexp.MatchString( "'phpacademy.category' doesn't exist", err.Error() )
		if notTable {
			rowsAffected, err := dbConn.Exec("create table `category` (  `key_category` int(11) unsigned NOT NULL AUTO_INCREMENT,  `name` varchar(255) NOT NULL,  `key_parent` int(11) NOT NULL DEFAULT '-1',  `short_text` mediumtext,  `long_text` longtext,  `text_task` mediumtext,  `reg_expr` varchar(255) NOT NULL COMMENT 'проверочное выражение для задания',  `is_view` int(11) NOT NULL DEFAULT '1',  `video` varchar(255) NOT NULL,  `date_sys` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,  `leaf` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Признак элемента, завершающего ветку (лист)',  PRIMARY KEY (`key_category`),  KEY `name` (`name`) ) ENGINE=InnoDB AUTO_INCREMENT=83 DEFAULT CHARSET=utf8" )
			if err != nil {
				Encode.Encode(err)
			} else {
				//                  fmt.Printf( "#%d %s ", rowsAffected, "Not table category, i create!" )
				lastInsertId, _:= rowsAffected.LastInsertId()
				Encode.Encode(lastInsertId)
			}
		} else {
			Encode.Encode(sql)
			Encode.Encode(err)
			if error != nil {
				Encode.Encode(error)
			}

		}
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
func SelectToMultidimension(sql string, args ...interface{}) ( arrJSON [] map[string] interface {}, err error ) {

	rows, err := DoSelect(sql, args...)

	if err != nil {
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

		for _, colType := range colTypes {

			fieldName := colType.Name()
			fieldValue, ok := rowValues[fieldName]
			if !ok {
				log.Println(err)
				continue
			}
			log.Println(colType.Length())
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
func ConvertPrepareRowsToJson(rows *sql.Rows, row [] interface {}, rowField map[string] *sql.NullString,
			columns [] string, colTypes [] *sql.ColumnType) ( arrJSON [] map[string] interface {}, err error ) {

	log.Println("\n", rows)

	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			log.Println(err)
			continue
		}
		log.Println("\n")
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
func ConvertPrepareRowToJson(row [] interface {}, rowField map[string] *sql.NullString, columns [] string,
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


//GetDataCustom get data with custom sql
func GetDataCustom(tableName string, begSql string, endSql string, args ...interface{}) (rows *sql.Rows,
		row [] interface {}, rowField map[string] *sql.NullString, columns [] string,
		colTypes [] *sql.ColumnType, err error ){

	if(begSql == ""){
		begSql = "SELECT * FROM "
	}

	if(endSql == ""){
		endSql = " WHERE id=?"
	}

	sql:= begSql + tableName + endSql

	rows, row, rowField, columns, colTypes, err = GetDataPrepareRowsToReading(sql, args...)


	if err != nil {
		log.Println(err)
		return nil, nil, nil, nil, nil, err
	}

	return rows, row, rowField, columns, colTypes, nil

}
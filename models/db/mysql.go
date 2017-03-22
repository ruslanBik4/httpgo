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
	httpgoJson "github.com/ruslanBik4/httpgo/views/templates/json"
)
const dbName = "travel"
const dbUser = "travel"
const dbPass = "3216732167"
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
	dbConn, err = sql.Open( "mysql", fmt.Sprintf("%s:%s@/%s?persistent", dbUser, dbPass, dbName) )
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
func GetResultToJSON (rows *sql.Rows) []byte{

	var row [] interface {}
	var rowOutput [] map[string] string

	rowField := make( map[string] *sql.NullString )

	if columns, err := rows.Columns(); err != nil {
		return nil
	} else {

		for _, val := range columns {

			rowField[val] = new(sql.NullString)
			row = append( row, rowField[val] )
		}
	}

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


	rows := DoQuery("SELECT COLUMN_NAME, DATA_TYPE, COLUMN_DEFAULT, " +
		"IS_NULLABLE, CHARACTER_SET_NAME, COLUMN_COMMENT, COLUMN_TYPE, CHARACTER_MAXIMUM_LENGTH " +
		"FROM INFORMATION_SCHEMA.COLUMNS C " +
		"WHERE TABLE_SCHEMA=? AND TABLE_NAME=? ORDER BY ORDINAL_POSITION" + limiter,
		dbName, table_name);
	//, form.html_name AS form_html_name, form.html_id AS form_html_id, form.js_func_onsubmit, field.db_table_name, field.db_field_name, field.label, field.html_type, field.html_class, field.html_name, field.html_id, field.html_value, c.name AS constraint_name, rc.value AS constraint_value, rc.relative_html_input_name

	//left join  ui_input_forms form
	//JOIN ui_input_fields field ON form.id=field.id_ui_input_forms AND form.html_name='client_registration'
	//LEFT JOIN ui_input_fields_rules rule ON rule.id=field.id_ui_input_fields_rules
	//LEFT JOIN ui_input_fields_rules_constraints rc ON rule.id=rc.id_ui_input_fields_rules
	//LEFT JOIN ui_input_fields_constraints c ON c.id = rc.id_ui_input_fields_constraints


	// 	select IFNULL(F_N.title, ''), IFNULL(F_N.type_input, ''), IFNULL(F_N.is_view, ''), COLUMN_NAME, DATA_TYPE, IFNULL( COLUMN_DEFAULT, ''), IS_NULLABLE, IFNULL(CHARACTER_SET_NAME, ''), IFNULL( C.COLUMN_COMMENT, '') from INFORMATION_SCHEMA.COLUMNS C left join allservi.field_names F_N on (F_N.field_name = C.COLUMN_NAME) where C.TABLE_NAME = ?

	if(rows == nil) {
		return sql.ErrNoRows
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

func SelectToMultidimension(sql string,args ...interface{}) map[int] httpgoJson.MultiDimension{

	println(args)


	rows, err := DoSelect(sql,args...)

	defer rows.Close()
	if err != nil {
		println(err)
	}

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	final_result := map[int] httpgoJson.MultiDimension{}
	result_id := 0
	for rows.Next() {
		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		tmp_struct := httpgoJson.MultiDimension{}

		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if (ok) {
				v = string(b)
			} else {
				v = val
			}
			tmp_struct[col] = fmt.Sprintf("%v",v)
		}

		final_result[ result_id ] = tmp_struct
		result_id++
	}

	return final_result



}

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
	"github.com/ruslanBik4/httpgo/models/db/schema"
)
var (
	dbConn *sql.DB
	SQLvalidator = regexp.MustCompile(`^select\s+.+\s*from\s+`)
	tableNameFromSQL = regexp.MustCompile(`(?i)\bfrom\b\s*\W?(\w+)`)
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
func getValue(fieldValue *sql.NullString) string {
	if fieldValue.Valid {
		return fieldValue.String
	}

	return "NULL"
}
func getSQLFromSETID(key, parentTable string) string{
	tableProps := strings.TrimPrefix(key, "setid_")
	tableValue := parentTable + "_" + tableProps + "_has"

	fields := schema.GetFieldsTable(parentTable)
	field  := fields.FindField(key)

	titleField := field.GetForeignFields()
	if titleField == "" {
		return ""
	}

	return fmt.Sprintf(`SELECT p.id, p.%s, v.id_%s
		FROM %s v
		JOIN %s p
		ON p.id = v.id_%[4]s AND id_%[2]s = ?
		ORDER BY p.id `,
		titleField, parentTable, tableValue, tableProps)

}
func getSQLFromNodeID(key, parentTable string) string{
	var tableProps, titleField string

	tableValue := strings.TrimPrefix(key, "nodeid_")

	fields := schema.GetFieldsTable(tableValue)

	for _, field := range fields.Rows {
		if strings.HasPrefix(field.COLUMN_NAME, "id_") && (field.COLUMN_NAME != "id_" + parentTable) {
			tableProps = field.COLUMN_NAME[3:]
			titleField = field.GetForeignFields()
			break
		}
	}

	if titleField == "" {
		return ""
	}

	return fmt.Sprintf(`SELECT p.id, p.%s, v.id_%s
		FROM %s v
		JOIN %s p
		ON p.id = v.id_%[4]s AND id_%[2]s = ?
		ORDER BY p.id `,
		titleField, parentTable, tableValue, tableProps)

}

func getSQLFromTableID(fields *schema.FieldsTable, key, parentTable string) string {


	tableProps := strings.TrimPrefix(key, "tableid_")
	//TODO структуру таблдицы вынести в ф-цию selectToMulti.. и также поле

	field := fields.FindField(key)
	where := field.WhereFromSet(fields)
	if where > "" {
		where += " AND (id_%s=?)"
	} else {
		where = " WHERE (id_%s=?)"
	}

	 return fmt.Sprintf( `SELECT * FROM %s p ` + where, tableProps, parentTable )

}
func SelectToMultidimension(sql string, args ...interface{}) ( arrJSON [] map[string] interface {}, err error ) {

	rows, err := DoSelect(sql, args...)

	arrSTR := tableNameFromSQL.FindStringSubmatch(sql)

	tableName := "rooms"
	if len(arrSTR) < 1 {
		log.Println(arrSTR)
	} else {
		tableName = arrSTR[1]
	}

	if err != nil {
		log.Println(err, sql)
		return nil, err
	}

	defer rows.Close()

	valuePtrs, rowValues, columns, colTypes := PrepareRowsToReading(rows)
	fields := schema.GetFieldsTable(tableName)

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Println(err)
			continue
		}
		//заполняем структуру занчениями
		for _, colType := range colTypes {
			fieldName := colType.Name()
			fieldValue, ok := rowValues[fieldName]
			if !ok {
				log.Println(err)
				continue
			}
			if fieldValue.Valid {
				fields.FindField(fieldName).Value = fieldValue.String
			} else {
				fields.FindField(fieldName).Value = ""
			}
		}

		values := make(map[string] interface{}, len(columns) )

		//TODO предусмотреть ситуацию, когда в таблице нету поля id - для рекурсии
		id, ok := rowValues["id"]

		if !ok || !id.Valid {
			log.Println("record ID not found")
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
			if strings.HasPrefix(fieldName, "setid_")  {
				values[fieldName], err = SelectToMultidimension( getSQLFromSETID(fieldName, tableName), id.String )
				if err != nil {
					log.Println(err)
					values[fieldName] = err.Error()
				}
				continue
			} else if strings.HasPrefix(fieldName, "nodeid_"){
				values[fieldName], err = SelectToMultidimension( getSQLFromNodeID(fieldName, tableName), id.String )
				if err != nil {
					log.Println(err)
					values[fieldName] = err.Error()
				}
				continue
			} else if strings.HasPrefix(fieldName, "tableid_"){
				log.Println(getSQLFromTableID(fields, fieldName, tableName))
				values[fieldName], err = SelectToMultidimension( getSQLFromTableID(fields, fieldName, tableName), id.String)
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

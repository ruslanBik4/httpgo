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
func getSQLFromSETID(fields *schema.FieldsTable, field *schema.FieldStructure, parentTable string) string{
	tableProps := strings.TrimPrefix(field.COLUMN_NAME, "setid_")
	tableValue := parentTable + "_" + tableProps + "_has"

	//titleField := field.GetForeignFields()

	where := field.WhereFromSet(fields)

	return fmt.Sprintf(`SELECT p.id
		FROM  %s p
		JOIN	%s v
		ON (p.id = v.id_%[1]s AND id_%[3]s = ?) ` + where,
		tableProps, tableValue, parentTable)

}
func getSQLFromNodeID(fields *schema.FieldsTable, field *schema.FieldStructure,parentTable string) string{
	var tableProps, titleField string

	tableValue := strings.TrimPrefix(field.COLUMN_NAME, "nodeid_")

	fieldsValues := schema.GetFieldsTable(tableValue)

	for _, field := range fieldsValues.Rows {
		if strings.HasPrefix(field.COLUMN_NAME, "id_") && (field.COLUMN_NAME != "id_" + parentTable) {
			tableProps = field.COLUMN_NAME[3:]
			titleField = field.GetForeignFields()
			break
		}
	}

	where := field.WhereFromSet(fields)

	return fmt.Sprintf(`SELECT p.id, %s, id_%s
		FROM %s v
		JOIN %s p
		ON (p.id = v.id_%[4]s AND id_%[2]s = ?) ` + where,
		titleField, parentTable, tableValue, tableProps)

}

func getSQLFromTableID(fields *schema.FieldsTable, field *schema.FieldStructure, parentTable string) string {


	tableProps := strings.TrimPrefix(field.COLUMN_NAME, "tableid_")
	//TODO структуру таблдицы вынести в ф-цию selectToMulti.. и также поле

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

	fields := schema.GetFieldsTable(tableName)
	defer rows.Close()

	//_, rowValues, columns, colTypes := PrepareRowsToReading(rows)

	columns, err := rows.Columns()

	var valuePtrs []interface{}
	for _, fieldName := range columns {
		valuePtrs = append(valuePtrs, fields.FindField(fieldName) )
	}


	for rows.Next() {
		values := make(map[string] interface{}, len(columns) )
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Println(err)
			continue
		}


		//TODO предусмотреть ситуацию, когда в таблице нету поля id - для рекурсии
		fieldID := fields.FindField("id")

		if fieldID != nil {
			fields.ID, _ = strconv.Atoi( fieldID.Value )
		}

		for _, fieldName := range columns {

			field := fields.FindField(fieldName)
			//if field == nil {
			//	values[fieldName] =
			//}
			values[field.COLUMN_NAME] = field.Value
			//TODO для полей типа tableid_, setid_, nodeid_ придумать механизм для блока WHERE
			// (по ключу родительской таблицы и патетрну из свойств поля для полей типа set)
			//TODO для полей типа setid_ формировать название таблицы
			//TODO также на уровне функции продумать менанизм, который позволит выбирать НЕ ВСЕ поля из третей таблицы
			if strings.HasPrefix(field.COLUMN_NAME, "setid_")  {
				sqlCommand := getSQLFromSETID(fields, field, tableName)
				log.Println(sqlCommand)
				values[field.COLUMN_NAME], err = SelectToMultidimension( sqlCommand, fields.ID )
				if err != nil {
					log.Println(err)
					values[field.COLUMN_NAME] = err.Error()
				}
				continue
			} else if strings.HasPrefix(field.COLUMN_NAME, "nodeid_"){
				sqlCommand := getSQLFromNodeID(fields, field, tableName)
				log.Println(sqlCommand)
				values[field.COLUMN_NAME], err = SelectToMultidimension( sqlCommand, fields.ID )
				if err != nil {
					log.Println(err)
					values[field.COLUMN_NAME] = err.Error()
				}
				continue
			} else if strings.HasPrefix(field.COLUMN_NAME, "tableid_"){
				sqlCommand := getSQLFromTableID(fields, field, tableName)
				log.Println(sqlCommand)
				values[field.COLUMN_NAME], err = SelectToMultidimension( sqlCommand, fields.ID)
				if err != nil {
					log.Println(err)
					values[field.COLUMN_NAME] = err.Error()
				}
				continue

			}

			switch field.COLUMN_TYPE {
			case "varchar", "date", "datetime":
				values[field.COLUMN_NAME] = field.Value
			case "tinyint":
				if field.Value == "1" {
					values[field.COLUMN_NAME] = true
				} else {
					values[field.COLUMN_NAME] = false

				}
			case "int", "int64", "float":
				values[field.COLUMN_NAME], _ = strconv.Atoi(field.Value)
			default:
				values[field.COLUMN_NAME] = field.Value
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

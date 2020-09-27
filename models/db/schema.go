package db

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/logs"
)

// TableOptions get table properties
type TableOptions struct {
	TABLE_NAME    string
	TABLE_TYPE    string
	ENGINE        string
	TABLE_COMMENT string
}

// RecordsTables get field properties
type RecordsTables struct {
	Rows []TableOptions
}

// GetTableProp получение данных для одной таблицы
func (ns *TableOptions) GetTableProp(tableName string) error {

	rows := DoQuery("SELECT TABLE_NAME, TABLE_TYPE, ENGINE, "+
		"IF (TABLE_COMMENT = NULL OR TABLE_COMMENT = '', TABLE_NAME, TABLE_COMMENT) "+
		"FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME=? order by TABLE_COMMENT", tableName)

	for rows.Next() {

		err := rows.Scan(&ns.TABLE_NAME, &ns.TABLE_TYPE, &ns.ENGINE, &ns.TABLE_COMMENT)

		if err != nil {
			logs.ErrorLog(err)
			return err
		}

	}

	return nil
}

// GetTablesProp получение таблиц
func (ns *RecordsTables) GetTablesProp(nameDB string) error {

	return ns.GetSelectTablesProp("TABLE_SCHEMA=?", nameDB)

}

// GetSelectTablesProp получение данных таблиц по условию
func (ns *RecordsTables) GetSelectTablesProp(where string, args ...interface{}) error {

	sqlCommand := `select TABLE_NAME, IFNULL(TABLE_TYPE, 'VIEW'), IFNULL(ENGINE, 'VIEW'),
			IFNULL(TABLE_COMMENT, TABLE_NAME) FROM INFORMATION_SCHEMA.TABLES
			 WHERE ` + where + " order by TABLE_COMMENT"
	rows, err := DoSelect(sqlCommand, args...)

	if err != nil {
		return err
	}

	for rows.Next() {

		var row TableOptions
		err := rows.Scan(&row.TABLE_NAME, &row.TABLE_TYPE, &row.ENGINE, &row.TABLE_COMMENT)

		if err != nil {
			logs.ErrorLog(err, row.TABLE_NAME, row.TABLE_TYPE)
			continue
		}

		ns.Rows = append(ns.Rows, row)

	}

	return nil

}

// FieldStructure get field properties
type FieldStructure struct {
	COLUMN_NAME              string
	DATA_TYPE                string
	COLUMN_DEFAULT           sql.NullString
	IS_NULLABLE              string
	CHARACTER_SET_NAME       sql.NullString
	COLUMN_COMMENT           sql.NullString
	COLUMN_TYPE              string
	COLUMN_KEY               string
	EXTRA                    string
	CHARACTER_MAXIMUM_LENGTH sql.NullInt64
	//TITLE string
	//TYPE_INPUT string
	//IS_VIEW []uint8

}

// FieldsTable  get fields from table & options
type FieldsTable struct {
	Rows    []FieldStructure
	Options TableOptions
}

// GetColumnsProp получение значений полей для форматирования данных
// получение значений полей для таблицы
// @param args []int{} - можно передавать ограничения выводимых полей.
// Действует как LIMIT a или LIMIT a, b
func (ns *FieldsTable) GetColumnsProp(tableName string, args ...int) error {

	valuesText := []string{}
	for _, arg := range args {
		valuesText = append(valuesText, strconv.Itoa(arg))
	}

	limiter := strings.Join(valuesText, ", ")
	if limiter != "" {
		limiter = " LIMIT " + limiter
	}

	rows, err := DoSelect(`SELECT COLUMN_NAME, DATA_TYPE, COLUMN_DEFAULT,
		IS_NULLABLE, CHARACTER_SET_NAME, COLUMN_COMMENT, COLUMN_TYPE, COLUMN_KEY, EXTRA, CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.COLUMNS C
		WHERE TABLE_SCHEMA=? AND TABLE_NAME=? ORDER BY ORDINAL_POSITION`+limiter,
		server.GetServerConfig().DBName(), tableName)
	if err != nil {
		return err
	}

	if ns.Rows == nil {
		ns.Rows = make([]FieldStructure, 0)
	}
	for rows.Next() {
		var row FieldStructure

		err := rows.Scan(&row.COLUMN_NAME, &row.DATA_TYPE, &row.COLUMN_DEFAULT, &row.IS_NULLABLE,
			&row.CHARACTER_SET_NAME, &row.COLUMN_COMMENT, &row.COLUMN_TYPE, &row.COLUMN_KEY, &row.EXTRA,
			&row.CHARACTER_MAXIMUM_LENGTH)

		if err != nil {
			logs.ErrorLog(err)
			continue
		}

		ns.Rows = append(ns.Rows, row)

	}

	return nil
}

// PutDataFrom заполняет структуру для формы данными, взятыми из структуры БД
func (ns *FieldsTable) PutDataFrom(tableName string) (fields *schema.FieldsTable) {

	fields = &schema.FieldsTable{Name: tableName}
	fields.Rows = make([]*schema.FieldStructure, len(ns.Rows))
	for i, field := range ns.Rows {
		fields.Rows[i] = &schema.FieldStructure{
			COLUMN_NAME: field.COLUMN_NAME,
			DATA_TYPE:   field.DATA_TYPE,
			IS_NULLABLE: field.IS_NULLABLE,
			COLUMN_TYPE: field.COLUMN_TYPE,
			Events:      make(map[string]string, 0),
			DataJSOM:    make(map[string]interface{}, 0),
			Table:       fields,
			// TODO: продумать позже механизм для READ-ONLY полей
			//IsHidden:    (field.COLUMN_KEY=="PRI") || (field.EXTRA=="on update CURRENT_TIMESTAMP"),
			PrimaryKey: field.COLUMN_KEY == "PRI",
		}
		if field.CHARACTER_SET_NAME.Valid {
			fields.Rows[i].CHARACTER_SET_NAME = field.CHARACTER_SET_NAME.String
		}

		if field.CHARACTER_MAXIMUM_LENGTH.Valid {
			fields.Rows[i].CHARACTER_MAXIMUM_LENGTH = int(field.CHARACTER_MAXIMUM_LENGTH.Int64)
		}
		if field.COLUMN_DEFAULT.Valid {
			fields.Rows[i].COLUMN_DEFAULT = field.COLUMN_DEFAULT.String
		}

		if field.COLUMN_COMMENT.Valid {
			fields.Rows[i].COLUMN_COMMENT = field.COLUMN_COMMENT.String
		}

		//if (field.DATA_TYPE == "set") || (field.DATA_TYPE == "enum") {
		//	fieldStrc.
		//}
		//fields.Rows[i] = fieldStrc
	}

	fields.SaveFormEvents = make(map[string]string, 0)

	if pos := strings.Index(ns.Options.TABLE_COMMENT, "onload:"); pos > 0 {
		fields.Comment = ns.Options.TABLE_COMMENT[:pos]
		fields.DataJSOM = make(map[string]interface{}, 0)

		fields.DataJSOM["onload"] = ns.Options.TABLE_COMMENT[pos+len("onload:"):]
	} else {
		fields.Comment = ns.Options.TABLE_COMMENT
	}

	return fields
}

var schemaReady bool

// InitSchema read schema from DB & fill SchemaCache
func InitSchema() {
	// TODO: предусмотреть флаг, обозначающий, что кеширование данных не закончено
	//go func() {
	var tables RecordsTables
	tables.GetTablesProp(server.GetServerConfig().DBName())

	// первый проход заполняет в кеши данными полей первого уровня
	for _, table := range tables.Rows {
		var fields FieldsTable
		fields.GetColumnsProp(table.TABLE_NAME)

		schema.SchemaCache[table.TABLE_NAME] = fields.PutDataFrom(table.TABLE_NAME)
	}
	// теперь заполняем данные второго уровня - которые зависят от других таблиц
	for tableName, fields := range schema.SchemaCache {
		fields.FillSurroggateFields(tableName)
		//for _, field := range fields.Rows {
		//
		//}
		//logs.StatusLog(tableName, fields)
	}

	schemaReady = true

	//}()
}

package db

import (
	"strings"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"database/sql"
	"strconv"
	"log"
)
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

	rows, err := DoSelect("SELECT TABLE_NAME, TABLE_TYPE, ENGINE, " +
		"IF (TABLE_COMMENT = NULL OR TABLE_COMMENT = '', TABLE_NAME, TABLE_COMMENT) " +
		"FROM INFORMATION_SCHEMA.TABLES WHERE " + where + " order by TABLE_COMMENT")

	if err != nil {
		return err
	}

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

// заполняет структуру для формы данными, взятыми из структуры БД
func (ns *FieldsTable) PutDataFrom() (fields *schema.FieldsTable) {

	fields = &schema.FieldsTable{Name: ns.Options.TABLE_NAME}
	fields.Rows = make([] schema.FieldStructure, len(ns.Rows) )
	for i, field := range ns.Rows {
		fieldStrc := &schema.FieldStructure{
			COLUMN_NAME: field.COLUMN_NAME,
			DATA_TYPE  : field.DATA_TYPE,
			IS_NULLABLE: field.IS_NULLABLE,
			COLUMN_TYPE: field.COLUMN_TYPE,
			Events     : make(map[string] string, 0),
			DataJSOM   : make(map[string] interface{}, 0),
			Table	   : fields,
			IsHidden   : false,
		}
		if field.CHARACTER_SET_NAME.Valid {
			fieldStrc.CHARACTER_SET_NAME = field.CHARACTER_SET_NAME.String
		}
		fieldStrc.GetTitle(field.COLUMN_NAME)

		if field.CHARACTER_MAXIMUM_LENGTH.Valid {
			fieldStrc.CHARACTER_MAXIMUM_LENGTH = int(field.CHARACTER_MAXIMUM_LENGTH.Int64)
		}
		if field.COLUMN_DEFAULT.Valid {
			fieldStrc.COLUMN_DEFAULT = field.COLUMN_DEFAULT.String
		}

		if field.COLUMN_COMMENT.Valid {
			fieldStrc.GetTitle(field.COLUMN_COMMENT.String)
		}
		fields.Rows[i] = *fieldStrc
	}

	fields.SaveFormEvents = make(map[string] string, 0)

	if pos := strings.Index(ns.Options.TABLE_COMMENT, "onload:"); pos > 0 {
		fields.Comment = ns.Options.TABLE_COMMENT[:pos]
		fields.DataJSOM = make( map[string] interface{}, 0 )

		fields.DataJSOM["onload"] = ns.Options.TABLE_COMMENT[pos + len("onload:"): ]
	} else {
		fields.Comment = ns.Options.TABLE_COMMENT
	}

	return fields
}

func InitSchema() {
	go func() {
		var tables RecordsTables
		tables.GetTablesProp(server.GetServerConfig().DBName() )

		for _, table := range tables.Rows {
			var fields FieldsTable
			fields.GetColumnsProp(table.TABLE_NAME)

			schema.SchemaCache[table.TABLE_NAME] = fields.PutDataFrom()
			log.Println(schema.SchemaCache[table.TABLE_NAME].Name)
		}

	}()
}
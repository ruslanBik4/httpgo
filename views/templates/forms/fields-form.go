// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// генерация форм на основе типов полей таблиц БД
package forms

//todo необходима вариативность вывода в input select значений из енумов и справочников,пример - есть енум из 9-ти позиций а вывести нужно только 1,5,6(соответственно и юзер может что либо делать только с ними а не со всем списком 1-9)

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	_ "strconv"
	"strings"
	"time"

	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/logs"
)

var (
	enumValidator = regexp.MustCompile(`(?:'([^,]+)',?)`)
)

//Сия структура нужна для подготовки к отображению поля на форме (возможо, в таблице и еще других компонентах веб-старницы)
//На данный момент создается на лету, в будущем
//TODO: перенести в сервис отдачи структур и сделать независемым от реализации СУБД
type FieldStructure struct {
	Table                    *FieldsTable
	COLUMN_NAME              string
	DATA_TYPE                string
	COLUMN_DEFAULT           string
	IS_NULLABLE              string
	CHARACTER_SET_NAME       string
	COLUMN_COMMENT           string
	COLUMN_TYPE              string
	CHARACTER_MAXIMUM_LENGTH int
	Value                    string
	IsHidden                 bool
	InputType                string
	CSSClass                 string
	CSSStyle                 string
	TableName                string
	Events                   map[string]string
	Where                    string
	Figure                   string
	Placeholder              string
	Pattern                  string
	MinDate                  string
	MaxDate                  string
	BeforeHtml               string
	Html                     string
	AfterHtml                string
	ForeignFields            string
	LinkTD                   string
	DataJSOM                 map[string]interface{}
	EnumValues               []string
}

type FieldsTable struct {
	Name           string
	ID             int
	Comment        string
	IsDadata       bool
	Rows           []FieldStructure
	Hiddens        map[string]string
	SaveFormEvents map[string]string
	DataJSOM       map[string]interface{}
}

func (field *FieldStructure) setEnumValues() {
	if len(field.EnumValues) > 0 {
		return
	}
	fields := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)
	for _, title := range fields {
		field.EnumValues = append(field.EnumValues, title[len(title)-1])
	}
}

// стиль показа для разных типов полей
// новый метод, еще обдумываю
func (field *FieldStructure) TypeInput() string {
	if (field.COLUMN_NAME == "id") || (field.COLUMN_NAME == "date_sys") {
		//ns.ID, _ = strconv.Atoi(val)
		//возможно, тут стоит предусмотреть некоторые действия
		return "hidden"
	}
	if field.COLUMN_NAME == "isDel" {
		return "button"
	}
	if strings.HasPrefix(field.COLUMN_NAME, "id_") {
		return "ForeignSelect"
	}
	if strings.HasPrefix(field.COLUMN_NAME, "setid_") || strings.HasPrefix(field.COLUMN_NAME, "nodeid_") {
		return "set"
	}
	if strings.HasPrefix(field.COLUMN_NAME, "tableid_") {
		return "table"
	}
	if field.InputType == "" {
		switch field.DATA_TYPE {
		case "varchar":
			field.InputType = "text"
		case "set":
			field.setEnumValues()
			field.InputType = "set"
		case "enum":
			field.setEnumValues()
			if len(field.EnumValues) > 2 {
				field.InputType = "select"
			} else {
				field.InputType = "enum"
			}
		case "tinyint":
			field.InputType = "checkbox"
		case "int", "double":
			field.InputType = "number"
		case "date":
			field.InputType = "date"
		case "time":
			field.InputType = "time"
		case "timestamp", "datetime":
			field.InputType = "datetime"
		case "text":
			field.InputType = "textarea"
		case "blob":
			field.InputType = "file"
		default:
			field.InputType = "text"
		}
	}

	return field.InputType

}

//старый метод, обсолете, буду избавляться
//TODO старый метод, обсолете, буду избавляться
func StyleInput(dataType string) string {
	switch dataType {
	case "varchar":
		return "search"
	case "set", "enum":
		return "select"
	case "tinyint":
		return "checkbox"
	case "int":
		return "number"
	case "date":
		return "date"
	case "timestamp", "datetime":
		return "datetime"
	case "text":
		return "textarea"
	case "blob":
		return "file"
	}

	return "text"

}

// минимальный размер поля для разных типов полей
func GetLengthFromType(dataType string) (width int, size int) {
	switch dataType {
	case "select":
		return 120, 50
	case "checkbox":
		return 50, 15
	case "number":
		return 70, 50
	case "date":
		return 110, 50
	case "datetime":
		return 140, 50
	case "timestamp":
		return 140, 50
	case "textarea":
		return 120, 50
	}

	return 120, 50

}

func getFields(tableName string) (fields FieldsTable, err error) {
	var ns db.FieldsTable

	ns.GetColumnsProp(tableName)

	fields.PutDataFrom(ns)
	fields.Name = tableName

	return fields, nil
}

// Scan implements the Scanner interface.
func (field *FieldStructure) Scan(value interface{}) error {
	var temp sql.NullString

	if err := temp.Scan(value); err != nil {
		field.Value = ""
		return err
	}
	field.Value = temp.String

	return nil
}

func (ns *FieldsTable) FindField(name string) *FieldStructure {
	for idx, field := range ns.Rows {
		if field.COLUMN_NAME == name {
			return &ns.Rows[idx]
		}
	}

	return nil
}

// create where for  query from SETID_ / NODEID_ fields
func (field *FieldStructure) whereFromSet(ns *FieldsTable) (result string) {
	fields := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)
	comma := " WHERE "
	for _, title := range fields {
		enumVal := title[len(title)-1]
		if i := strings.Index(enumVal, ":"); i > 0 {
			param := ""
			// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
			if paramField := ns.FindField(enumVal[i+1:]); (paramField != nil) && (paramField.Value != "") {
				param = paramField.Value
				enumVal = enumVal[:i] + fmt.Sprintf("%s", param)
			} else {
				continue
			}
		}
		result += comma + enumVal
		comma = " OR "
	}

	return result
}

//получаем связанную таблицу с полями для поля типа TABLEID_
func (field *FieldStructure) getTableFrom(ns *FieldsTable, tablePrefix, key string) {
	//key := field.COLUMN_NAME
	tableProps := key[len("tableid_"):]

	fields, err := getFields(tableProps)
	if err != nil {
		field.Html += "<td>Error during readin table schema!" + tableProps + "</td>"
		return
	}

	where := field.whereFromSet(ns)
	if where > "" {
		where += " AND (id_%s=?)"
	} else {
		where = " WHERE (id_%s=?)"
	}
	sqlCommand := fmt.Sprintf(`SELECT * FROM %s p `+where, tableProps, ns.Name)

	rows, err := db.DoSelect(sqlCommand, ns.ID)
	if err != nil {
		logs.ErrorLog(err, sqlCommand)
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		logs.ErrorLog(err)
	}

	field.Html = "<thead> <tr>"

	var row []interface{}
	var newRow string

	rowField := make([]*sql.NullString, len(columns))
	for idx, fieldName := range columns {

		if (fieldName == "id") || (fieldName == "id_"+ns.Name) {
			//newRow += "<td></td>"
		} else {
			fieldStruct := fields.FindField(fieldName)
			field.Html += "<td>" + fieldStruct.COLUMN_COMMENT + "</td>"
			newRow += getTD(tableProps, fieldName, "", "id_"+ns.Name, 0, fieldStruct)
		}

		rowField[idx] = new(sql.NullString)
		row = append(row, rowField[idx])
	}
	field.Html += "</tr></thead><tbody>"

	idx := 0

	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			logs.ErrorLog(err)
			continue
		}
		idx++
		field.Html += "<tr>"
		for i, value := range rowField {
			field.Html += getTD(tableProps, columns[i], value.String, "id_"+ns.Name, idx, fields.FindField(columns[i]))
		}

		field.Html += "</tr>"

	}
	field.Html += fmt.Sprintf(`<tr id="tr%s">%s</tr></tbody>`, tablePrefix+key, newRow)
}

const CELL_TABLE = `<td class="%s"><input type="%s" name="%s:%s" value="%s"/></td>`
const CELL_SELECT = `<td class="%s"><select name="%s:%s" class="">%s</select></td>`

func getTD(tableProps, fieldName, value, parentField string, idx int, fieldStruct *FieldStructure) (html string) {
	inputName := fieldName + fmt.Sprintf("[%d]", idx)

	required, events, dataJson := "", "", ""

	if fieldStruct.IS_NULLABLE == "NO" {
		required = "required"
	}

	if (fieldName == "id") || (fieldName == parentField) {
		if value > "" {
			html += fmt.Sprintf(`<input type="%s" name="%s:%s" value="%s"/>`, "hidden", tableProps, inputName, value)
		}
	} else if strings.HasPrefix(fieldName, "id_") {
		fieldStruct.GetOptions(fieldName[3:], value)
		html += fmt.Sprintf(CELL_SELECT, fieldStruct.CSSClass, tableProps, inputName, fieldStruct.Html)
	} else if strings.HasPrefix(fieldName, "setid_") || strings.HasPrefix(fieldName, "nodeid_") {
		html += "<td>" + fieldStruct.RenderMultiSelect(nil, tableProps+":", fieldName, value, "", required) + "</td>"
	} else {
		fullInputName := tableProps + ":" + inputName
		switch fieldStruct.DATA_TYPE {
		case "enum":
			html += "<td class='" + fieldStruct.CSSClass + "'>" + fieldStruct.RenderEnum(fullInputName, value, required, events, dataJson) + "</td>"
		case "set":
			html += "<td class='" + fieldStruct.CSSClass + "'>" + fieldStruct.RenderSet(fullInputName, value, required, events, dataJson) + "</td>"
		case "tinyint":
			checked := ""
			if value == "1" {
				checked = "checked"
			}
			html += "<td class='" + fieldStruct.CSSClass + "'>" +
				RenderCheckBox(fullInputName, fieldName, "", 1, checked, required, events, dataJson) + "</td>"
		default:
			html += fmt.Sprintf(CELL_TABLE, fieldStruct.CSSClass, StyleInput(fieldStruct.DATA_TYPE), tableProps, inputName, value)
		}
	}

	return html
}

func (field *FieldStructure) getSQLFromSETID(key, parentTable string) string {
	tableProps := strings.TrimPrefix(key, "setid_")
	tableValue := parentTable + "_" + tableProps + "_has"

	titleField := field.getForeignFields(tableProps)
	if titleField == "" {
		return ""
	}

	return fmt.Sprintf(`SELECT p.id, %s, id_%s
	FROM %s p LEFT JOIN %s v ON (p.id=v.id_%[3]s AND id_%[2]s=?) `,
		titleField, parentTable,
		tableProps, tableValue)

}

func getSQLFromNodeID(key, parentTable string) string {
	var tableProps, titleField string

	tableValue := strings.TrimPrefix(key, "nodeid_")

	var ns db.FieldsTable
	ns.GetColumnsProp(tableValue)

	var fields FieldsTable

	fields.PutDataFrom(ns)

	for _, field := range fields.Rows {
		if (field.COLUMN_NAME != "id_"+parentTable) && strings.HasPrefix(field.COLUMN_NAME, "id_") {
			tableProps = field.COLUMN_NAME[3:]
			titleField = field.getForeignFields(tableProps)
			break
		}
	}

	if titleField == "" {
		return ""
	}

	return fmt.Sprintf(`SELECT p.id, %s, id_%s
	FROM %s p LEFT JOIN %s v ON (p.id=v.id_%[3]s AND id_%[2]s=?) `,
		titleField, parentTable,
		tableProps, tableValue)

}

func (field *FieldStructure) getOptionsNODEID(ns *FieldsTable, key string) {
	var sqlCommand string

	if strings.HasPrefix(key, "setid_") {
		sqlCommand = field.getSQLFromSETID(key, ns.Name)
	} else if strings.HasPrefix(key, "nodeid_") {

		sqlCommand = getSQLFromNodeID(key, ns.Name)

	}

	if sqlCommand == "" {
		field.Html += "<option disabled>Нет значений связанной таблицы!</option>"
		return
	}

	where := field.whereFromSet(ns)

	rows, err := db.DoSelect(sqlCommand+where, ns.ID)
	if err != nil {
		logs.ErrorLog(err, sqlCommand)
		return
	}
	defer rows.Close()
	idx := 0

	field.Html = ""
	for rows.Next() {
		var id string
		var title, selected string
		var idParent sql.NullInt64

		if err := rows.Scan(&id, &title, &idParent); err != nil {
			logs.ErrorLog(err)
			continue
		}
		if idParent.Valid {
			selected = "selected"
		}
		idx++

		field.Html += renderOption(id, title, selected)
	}

}

func (field *FieldStructure) getMultiSelect(ns *FieldsTable, key string) {
	var sqlCommand string

	if strings.HasPrefix(key, "setid_") {
		sqlCommand = field.getSQLFromSETID(key, ns.Name)
	} else if strings.HasPrefix(key, "nodeid_") {

		sqlCommand = getSQLFromNodeID(key, ns.Name)

	}

	if sqlCommand == "" {
		field.Html += "не получается собрать запрос для поля " + key
		return
	}

	where := field.whereFromSet(ns)

	rows, err := db.DoSelect(sqlCommand+where, ns.ID)
	if err != nil {
		logs.ErrorLog(err, sqlCommand)
		return
	}
	defer rows.Close()
	idx := 0

	field.Html = ""
	for rows.Next() {
		var id string
		var title, checked string
		var idParent sql.NullInt64

		if err := rows.Scan(&id, &title, &idParent); err != nil {
			logs.ErrorLog(err)
			continue
		}
		if idParent.Valid {
			checked = "checked"
		}
		idx++

		field.Html += "<li role='presentation'>" + RenderCheckBox(key+"[]", id, title, idx, checked, "", "", "") + "</li>"
	}

}

func (field *FieldStructure) getForeignFields(tableName string) string {

	if field.ForeignFields > "" {
		return field.ForeignFields
	} else {
		return db.GetParentFieldName(tableName)
	}
}

func (field *FieldStructure) GetOptions(tableName, val string) {

	var where string
	ForeignFields := field.getForeignFields(tableName)

	if ForeignFields == "" {
		field.Html += "<option disabled>Нет значений связанной таблицы!</option>"
		return
	}

	if field.Where > "" {
		where = " WHERE "
		enumVal := field.Where
		i, j := strings.Index(enumVal, ":"), 0

		for i > 0 {
			param := enumVal[i+1:]
			if j = strings.IndexAny(param, ", )"); j > 0 {
				param = param[:j]
			}
			// мы добавим условие со значением поля текущей записи, если это поле найдено и в нем установлено значение
			if param == "currentUser" {
				where += enumVal[:i] + fmt.Sprintf("%s", "users.IsLogin(nil)")
			} else if paramField := field.Table.FindField(param); (paramField != nil) && (paramField.Value != "") {
				where += enumVal[:i] + fmt.Sprintf("%s", paramField.Value)
			}
			/// попозже перепроверить
			enumVal = enumVal[i+len(param)+1:]
			i = strings.Index(enumVal, ":")
		}

		where += enumVal

		logs.DebugLog("where=", where)
	}

	sqlCommand := "select id, " + ForeignFields + " from " + tableName + where
	rows, err := db.DoSelect(sqlCommand)
	if err != nil {
		logs.ErrorLog(err, sqlCommand)
		return
	}
	defer rows.Close()
	idx := 0
	//valueID, _ := strconv.Atoi(val)

	field.Html = ""

	for rows.Next() {

		var id, title, selected string

		if err := rows.Scan(&id, &title); err != nil {
			logs.ErrorLog(err)
			continue
		}
		if val == id {
			selected = "selected"
		}
		idx++

		field.Html += renderOption(id, title, selected)
	}
}

func (field *FieldStructure) RenderSet(key, val, required, events, dataJson string) (result string) {
	fields := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)

	for idx, title := range fields {
		enumVal := title[len(title)-1]
		checked := ""
		if strings.Contains(val, enumVal) {
			checked = "checked"
		}
		result += RenderCheckBox(key+"[]", enumVal, enumVal, idx, checked, required, events, dataJson)
	}

	return result
}

func (field *FieldStructure) RenderEnum(key, val, required, events, dataJson string) (result string) {

	fields := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)
	// TODO: придумать параметр, который будет определять элемент тега ывне зависемости от количества
	isRenderSelect := (len(fields) > 2) || (field.InputType == "select")

	// если не имеет значение, покажем placeholder как первый option в списке
	if isRenderSelect && (val == "") {
		if field.Placeholder > "" {

			result = "<option disabled>" + field.Placeholder + "</option>"
		} else {

			result = "<option disabled>Выберите значение</option>"
		}
	}

	for idx, title := range fields {
		enumVal := title[len(title)-1]
		checked, selected := "", ""
		if val == enumVal {
			checked, selected = "checked", "selected"
		}
		if isRenderSelect {
			result += renderOption(enumVal, enumVal, selected)
		} else {
			result += renderRadioBox(key, enumVal, enumVal, idx, checked, required, events, dataJson)
		}
	}
	if isRenderSelect {
		return renderSelect(key, result, required, events, dataJson)
	}

	return result
}

func cutPartFromTitle(title, pattern, defaultStr string) (titleFull, titlePart string) {
	titleFull = title
	if title == "" {
		return "", ""
	}
	posPattern := strings.Index(titleFull, pattern)
	if posPattern > 0 {
		titlePart = titleFull[posPattern+len(pattern):]
		titleFull = titleFull[:posPattern]
	} else {
		titlePart = defaultStr
	}

	return
}

func (fieldStrc *FieldStructure) GetColumnTitles() (titleFull, titleLabel, placeholder, pattern, dataJson string) {

	counter := 1
	comma := ""
	for key, val := range fieldStrc.DataJSOM {

		dataJson += comma + fmt.Sprintf(`"%s": "%s"`, key, val)
		counter++
		comma = ","
	}
	return fieldStrc.COLUMN_COMMENT, fieldStrc.COLUMN_COMMENT, fieldStrc.Placeholder, fieldStrc.Pattern, dataJson
}

func getPattern(name string) string {
	rows, err := db.DoSelect("select pattern from patterns_list where name=?", name)
	if err != nil {
		logs.ErrorLog(err)
		return ""
	}
	for rows.Next() {
		var pattern sql.NullString
		if err := rows.Scan(&pattern); err != nil {
			logs.ErrorLog(err)
			return ""
		}
		if pattern.Valid {
			return pattern.String
		} else {
			return ""
		}

	}

	return ""
}

func (fieldStrc *FieldStructure) parseWhere(field db.FieldStructure, whereJSON interface{}) {
	switch whereJSON.(type) {
	case map[string]interface{}:

		comma := ""
		fieldStrc.Where = ""
		for key, value := range whereJSON.(map[string]interface{}) {
			enumVal := value.(string)
			// отбираем параметры типы :имя_поля
			if i := strings.Index(enumVal, ":"); i > -1 {
				param := enumVal[i+1:]
				// считаем, что окончанием параметра могут быть символы ", )"
				if j := strings.IndexAny(param, ", )"); j > 0 {
					param = param[:j]
				}
				// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
				if paramField := fieldStrc.Table.FindField(param); paramField == nil {
					continue
				}
			}
			fieldStrc.Where += comma + key + enumVal
			comma = " OR "
			logs.DebugLog("fieldStrc.Where", fieldStrc.Where)

		}
	default:
		logs.ErrorLog(errors.New("not correct type WhereJSON !"), whereJSON)
	}

}

func convertDatePattern(strDate string) string {
	switch strDate {
	case "today":
		return time.Now().Format("2006.01.02")
	case "tomorrow":
		return time.Now().Format("2006.01.02")
	case "yestoday":
		return time.Now().Format("2006.01.02")
	}
	return strDate
}

func (fieldStrc *FieldStructure) GetTitle(field db.FieldStructure) string {

	if !field.COLUMN_COMMENT.Valid {
		return ""
	}
	titleFull := field.COLUMN_COMMENT.String
	titleFull, fieldStrc.Pattern = cutPartFromTitle(titleFull, "//", "")
	if posPattern := strings.Index(titleFull, "{"); posPattern > 0 {

		dataJson := titleFull[posPattern:]

		var properMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataJson), &properMap); err != nil {
			logs.ErrorLog(err, "dataJson=", dataJson)
		} else {
			for key, val := range properMap {

				//buff, err := val.MarshalJSON()
				if err != nil {
					logs.ErrorLog(err)
					continue
				}
				switch key {
				case "figure":
					fieldStrc.Figure = val.(string)
				case "classCSS":
					fieldStrc.CSSClass = val.(string)
				case "placeholder":
					fieldStrc.Placeholder = val.(string)
				case "pattern":
					fieldStrc.Pattern = getPattern(val.(string))
				case "foreingKeys":
					fieldStrc.ForeignFields = val.(string)
				case "inputType":
					fieldStrc.InputType = val.(string)
				case "isHidden":
					fieldStrc.IsHidden = val.(bool)
				case "linkTD":
					fieldStrc.LinkTD = val.(string)
				case "where":
					fieldStrc.parseWhere(field, val)
				case "maxDate":
					fieldStrc.MaxDate = convertDatePattern(val.(string))
				case "minDate":
					fieldStrc.MinDate = convertDatePattern(val.(string))
				case "events":
					fieldStrc.Events = make(map[string]string, 0)
					for name, event := range val.(map[string]interface{}) {
						fieldStrc.Events[name] = event.(string)
					}
				default:
					fieldStrc.DataJSOM[key] = val
				}
			}
		}

		fieldStrc.COLUMN_COMMENT = titleFull[:posPattern]
	} else {
		fieldStrc.COLUMN_COMMENT = titleFull
	}

	return fieldStrc.COLUMN_COMMENT
}

// заполняет структуру для формы данными, взятыми из структуры БД
func (fields *FieldsTable) PutDataFrom(ns db.FieldsTable) {

	for _, field := range ns.Rows {
		fieldStrc := &FieldStructure{
			COLUMN_NAME: field.COLUMN_NAME,
			DATA_TYPE:   field.DATA_TYPE,
			IS_NULLABLE: field.IS_NULLABLE,
			COLUMN_TYPE: field.COLUMN_TYPE,
			Events:      make(map[string]string, 0),
			DataJSOM:    make(map[string]interface{}, 0),
			Table:       fields,
			IsHidden:    false,
		}
		if field.CHARACTER_SET_NAME.Valid {
			fieldStrc.CHARACTER_SET_NAME = field.CHARACTER_SET_NAME.String
		}
		fieldStrc.GetTitle(field)

		if field.CHARACTER_MAXIMUM_LENGTH.Valid {
			fieldStrc.CHARACTER_MAXIMUM_LENGTH = int(field.CHARACTER_MAXIMUM_LENGTH.Int64)
		}
		if field.COLUMN_DEFAULT.Valid {
			fieldStrc.COLUMN_DEFAULT = field.COLUMN_DEFAULT.String
		}

		fields.Rows = append(fields.Rows, *fieldStrc)
	}

	if fields.Name == "" {
		return
	}
	var tableOpt db.TableOptions
	tableOpt.GetTableProp(fields.Name)

	fields.SaveFormEvents = make(map[string]string, 0)

	if pos := strings.Index(tableOpt.TABLE_COMMENT, "onload:"); pos > 0 {
		fields.Comment = tableOpt.TABLE_COMMENT[:pos]
		fields.DataJSOM = make(map[string]interface{}, 0)

		fields.DataJSOM["onload"] = tableOpt.TABLE_COMMENT[pos+len("onload:"):]
	} else {
		fields.Comment = tableOpt.TABLE_COMMENT
	}
}

//AppendNewFieldRows - Добаляет в fields поля из других таблиц
//@version 1.00 2017-05-13
//@author Serg Litvinov
func (fields *FieldsTable) AppendNewFieldRows(fields1 FieldsTable, args ...interface{}) {
	for _, row := range fields1.Rows {
		for _, arg := range args {
			if row.COLUMN_NAME == arg {
				fields.Rows = append(fields.Rows, row)
			}
		}
	}
}

func (field *FieldStructure) GetListJSON(key, val, required, events, dataJson string) {

	field.Html = ""

	fields := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)

	for idx, title := range fields {
		enumVal := title[len(title)-1]
		checked := ""
		if strings.Contains(val, enumVal) {
			checked = "checked"
		}
		field.Html += RenderCheckBox(key+"[]", enumVal, enumVal, idx, checked, required, events, dataJson)
	}

}

func (field *FieldStructure) GetOptionsJson(tableName string) {

	var where string
	ForeignFields := field.getForeignFields(tableName)

	if ForeignFields == "" {
		field.Html += "\"0\": \"Нет значений связанной таблицы!\""
		return
	}

	if field.Where > "" {
		where = " WHERE "
		enumVal := field.Where
		i, j := strings.Index(enumVal, ":"), 0

		for i > 0 {
			param := enumVal[i+1:]
			if j = strings.IndexAny(param, ", )"); j > 0 {
				param = param[:j]
			}
			// мы добавим условие со значением поля текущей записи, если это поле найдено и в нем установлено значение
			if param == "currentUser" {
				where += enumVal[:i] + fmt.Sprintf("%s", "users.IsLogin(nil)")
			} else if paramField := field.Table.FindField(param); (paramField != nil) && (paramField.Value != "") {
				where += enumVal[:i] + fmt.Sprintf("%s", paramField.Value)
			}
			///TODO попозже перепроверить
			enumVal = enumVal[i+len(param)+1:]
			i = strings.Index(enumVal, ":")
		}

		where += enumVal

		logs.DebugLog("where=", where)
	}

	sqlCommand := "select id, " + ForeignFields + " from " + tableName + where
	rows, err := db.DoSelect(sqlCommand)
	if err != nil {
		logs.ErrorLog(err, sqlCommand)
		return
	}
	defer rows.Close()
	idx := 0

	field.Html = ""
	comma := ""
	for rows.Next() {

		var id, title string

		if err := rows.Scan(&id, &title); err != nil {
			logs.ErrorLog(err)
			continue
		}

		idx++
		title = strings.Replace(title, "\"", "'", -1)
		field.Html += comma + "\"" + id + "\"" + ": \"" + title + "\""
		comma = ","
	}
}

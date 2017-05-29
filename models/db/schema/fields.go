// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

//todo необходима вариативность вывода в input select значений из енумов и справочников,пример - есть енум из 9-ти позиций а вывести нужно только 1,5,6(соответственно и юзер может что либо делать только с ними а не со всем списком 1-9)



import (
	"strings"
	"regexp"
	"fmt"
	"log"
	"database/sql"
	"encoding/json"
	_ "strconv"
	"time"
)
var (
	enumValidator = regexp.MustCompile(`(?:'([^,]+)',?)`)

)
//Сия структура нужна для подготовки к отображению поля на форме (возможо, в таблице и еще других компонентах веб-старницы)
//На данный момент создается на лету, в будущем
//TODO: перенести в сервис отдачи структур и сделать независемым от реализации СУБД
type FieldStructure struct {
	Table 		*FieldsTable
	COLUMN_NAME   	string
	DATA_TYPE 	string
	COLUMN_DEFAULT 	string
	IS_NULLABLE 	string
	CHARACTER_SET_NAME       string
	COLUMN_COMMENT           string
	COLUMN_TYPE              string
	CHARACTER_MAXIMUM_LENGTH int
	Value                    string
	IsHidden                 bool
	InputType                string
	CSSClass                 string
	CSSStyle                string
	TableName               string
	Events                  map[string] string
	Where                   string
	Figure                  string
	Placeholder             string
	Pattern                 string
	MinDate                 string
	MaxDate                 string
	BeforeHtml              string
	Html                    string
	AfterHtml               string
	ForeignFields           string
	LinkTD                  string
	DataJSOM                map[string] interface{}
	EnumValues              []string
	SQLforFORMList          string `отдаем в списках полей для формы`
	SQLforDATAList          string `отдаем в составе данных`
	SETID, NODEID, TABLEID  bool
	IdForeign		bool
	SelectValues            map[int] string
	TableProps, TableValues string
	ChildrenFields		FieldsTable
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
func (field *FieldStructure) TypeInput() string{
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
	if strings.HasPrefix(field.COLUMN_NAME, "setid_") || strings.HasPrefix(field.COLUMN_NAME, "nodeid_"){
		return "set"
	}
	if strings.HasPrefix(field.COLUMN_NAME, "tableid_") {
		return "table"
	}
	if field.InputType == "" {
		switch (field.DATA_TYPE) {
		case "varchar":
			field.InputType = "text"
		case "set":
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
		case "timestamp", "datetime":
			field.InputType = "datetime"
		case "text":
			field.InputType = "textarea"
		case "blob":
			field.InputType = "file"
		}
	}

	return field.InputType

}
//старый метод, обсолете, буду избавляться
func StyleInput(dataType string) string{
	switch (dataType) {
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


// todo: проверить работу
// create where for  query from SETID_ / NODEID_ / TABLEID_ fields
// условия вынимаем из определения поля типа SET
// и все условия оборачиваем в скобки для того, что бы потом можно было навесить еще усло
func (field *FieldStructure) WhereFromSet(fields *FieldsTable) (result string) {

	defer func() {
		result := recover()
		switch err := result.(type) {
		case ErrNotFoundTable:
			log.Println(err)
		case nil:
		case error:
			panic(err)
		default:
			log.Println(err)
		}
	}()
	enumValues := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)
	comma  := ""
	for _, title := range enumValues {
		enumVal := title[len(title) - 1]
		if i := strings.Index(enumVal, ":"); i > 0 {
			param := ""
			// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
			if paramField := fields.FindField(enumVal[i+1:]); (paramField != nil) && (paramField.Value != "") {
				param = paramField.Value
				enumVal = enumVal[:i] + fmt.Sprintf("%s", param)
			} else {
				continue
			}
		}
		result += comma + enumVal
		comma = " OR "
	}

	if (result > "") && (result != "1") {

		return " WHERE (" + result + ")"
	}

	return ""
}
func (field *FieldStructure) GetSQLFromSETID(key, parentTable string) string{
	tableProps := strings.TrimPrefix(key, "setid_")
	tableValue := parentTable + "_" + tableProps + "_has"

	titleField := field.GetForeignFields()
	if titleField == "" {
		return ""
	}
	// LEFT JOIN for get all propertyes values
	return fmt.Sprintf( `SELECT p.id, %s, id_%s
	FROM %s p LEFT JOIN %s v ON (p.id=v.id_%[3]s AND id_%[2]s=?) `,
		titleField, parentTable,
		tableProps, tableValue)

}
// возвращает поле в связанной таблице, которое будет отдано пользователю
//например, для вторичных ключей отдает не idзаписи, а name || title || какой-либо складное поле
func (field *FieldStructure) GetForeignFields()  string {


	if field.ForeignFields > "" {
		return field.ForeignFields
	} else {
		return field.GetParentFieldName()
	}
}

func (field *FieldStructure) GetParentFieldName() (name string) {

	// получаем имя связанной таблицы
	var tableName string
	if field.SETID {
		tableName = field.TableProps
	} else if field.NODEID {
		tableName = field.TableValues
	} else if field.TABLEID {
		tableName = strings.TrimPrefix(field.COLUMN_NAME, "tableid_")
	} else if strings.HasPrefix(field.COLUMN_NAME, "id_") {
		tableName = strings.TrimPrefix(field.COLUMN_NAME, "id_")
	}

	defer func() {
		err := recover()
		switch err.(type) {
		case ErrNotFoundTable:
			name = ""
		case nil:
		default:
			panic(err)
		}
	}()

	fields := GetFieldsTable(tableName)

	for _, list := range fields.Rows {
			switch list.COLUMN_NAME {
			case "name":
				return "name"
			case "title":
				return "title"
			case "fullname":
				return "fullname"
			}
	}

	return name

}

func cutPartFromTitle(title, pattern, defaultStr string) (titleFull, titlePart string)  {
	titleFull = title
	if title == "" {
		return "", ""
	}
	posPattern := strings.Index(titleFull, pattern)
	if posPattern > 0 {
		titlePart = titleFull[posPattern + len(pattern):]
		titleFull = titleFull[:posPattern]
	} else {
		titlePart = defaultStr
	}

	return titleFull, titlePart
}
func (fieldStrc *FieldStructure) GetColumnTitles() (titleFull, titleLabel, placeholder, pattern, dataJson string)  {

	counter := 1
	comma := ""
	for key, val := range fieldStrc.DataJSOM {

		dataJson += comma + fmt.Sprintf( `"%s": "%s"`, key, val)
		counter++
		comma = ","
	}
	return fieldStrc.COLUMN_COMMENT, fieldStrc.COLUMN_COMMENT, fieldStrc.Placeholder, fieldStrc.Pattern, dataJson
}
func (fieldStrc *FieldStructure) parseWhere (whereJSON interface{}) {
	switch mapWhere := whereJSON.(type) {
	case map[string] interface{}:

		comma := ""
		fieldStrc.Where = ""

		for key, value := range mapWhere {
			enumVal := value.(string)
			// отбираем параметры типы :имя_поля
			if i := strings.Index(enumVal, ":"); i > -1 {
				param := enumVal[i+1:]
				// считаем, что окончанием параметра могут быть символы ", )"
				if j := strings.IndexAny(param, ", )"); j > 0 {
					param = param[:j]
				}
				// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
				//if paramField := fieldStrc.Table.FindField(param); paramField == nil {
				//	continue
				//}
			}
			fieldStrc.Where += comma + key + enumVal
			comma = " OR "

		}
	default:
		log.Println("not correct type WhereJSON !", whereJSON)
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
func (fieldStrc *FieldStructure) ParseComment(COLUMN_COMMENT string) string{

	titleFull := COLUMN_COMMENT
	titleFull, fieldStrc.Pattern = cutPartFromTitle(titleFull, "//", "")
	if posPattern := strings.Index(COLUMN_COMMENT, "{"); posPattern > 0 {

		dataJson := COLUMN_COMMENT[posPattern:]

		var properMap map[string] interface{}
		if err := json. Unmarshal([]byte(dataJson), &properMap); err != nil {
			log.Println(err)
			log.Println(dataJson)
		} else {
			for key, val := range properMap {

				//buff, err := val.MarshalJSON()
				if err != nil {
					log.Println(err)
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
					//fieldStrc.Pattern = getPattern( val.(string) )
				case "foreingKeys":
					fieldStrc.ForeignFields = val.(string)
				case "inputType":
					fieldStrc.InputType = val.(string)
				case "isHidden":
					fieldStrc.IsHidden = val.(bool)
				case "linkTD":
					fieldStrc.LinkTD   = val.(string)
				case "where":
					fieldStrc.parseWhere(val)
				case "maxDate":
					fieldStrc.MaxDate = convertDatePattern(val.(string))
				case "minDate":
					fieldStrc.MinDate = convertDatePattern(val.(string))
				case "events":
					fieldStrc.Events = make(map[string] string, 0)
					for name, event := range val.(map[string] interface{}) {
						fieldStrc.Events[name] = event.(string)
					}
				default:
					fieldStrc.DataJSOM[key] = val
				}
			}
		}

		fieldStrc.COLUMN_COMMENT = COLUMN_COMMENT[:posPattern]
	} else {
		fieldStrc.COLUMN_COMMENT = COLUMN_COMMENT
	}

	return fieldStrc.COLUMN_COMMENT
}
func (fieldStrc *FieldStructure) writeSQLbySETID() error {

	where := fieldStrc.WhereFromSet(fieldStrc.Table)

	fieldStrc.SQLforFORMList = fmt.Sprintf(`SELECT p.id, %s
		FROM %s p LEFT JOIN %s v
		ON (p.id = v.id_%[2]s) ` + where, fieldStrc.GetForeignFields(),
		fieldStrc.TableProps, fieldStrc.TableValues, fieldStrc.Table.Name)

	fieldStrc.SQLforDATAList = fmt.Sprintf(`SELECT p.id, %s
		FROM %s p JOIN %s v
		ON (p.id = v.id_%[2]s AND v.id_%[4]s=?)` + where, fieldStrc.GetForeignFields(),
		fieldStrc.TableProps, fieldStrc.TableValues, fieldStrc.Table.Name)
	return nil
}
//getSQLFromNodeID(field *schema.FieldStructure) string
func (fieldStrc *FieldStructure) writeSQLByNodeID() (err error){
	var titleField string


	defer func() {
		err := recover()
		switch err.(type) {
		case ErrNotFoundTable:
			err = ErrNotFoundTable{Table:fieldStrc.TableProps}
		case nil:
		default:
			panic(err)
		}
	}()
	if fieldStrc.TableProps == "" {
		fieldsValues := GetFieldsTable(fieldStrc.TableValues)

		for _, field := range fieldsValues.Rows {
			if strings.HasPrefix(field.COLUMN_NAME, "id_") && (field.COLUMN_NAME != "id_"+fieldStrc.Table.Name) {
				fieldStrc.TableProps = field.COLUMN_NAME[3:]
				titleField = field.GetForeignFields()
				break
			}
		}
	}

	where := fieldStrc.WhereFromSet(fieldStrc.Table)

	fieldStrc.SQLforFORMList =  fmt.Sprintf(`SELECT p.id, %s
		FROM %s p LEFT JOIN %s v
		ON (p.id = v.id_%[2]s) ` + where,
		titleField,  fieldStrc.TableProps, fieldStrc.TableValues)

	fieldStrc.SQLforDATAList =  fmt.Sprintf(`SELECT p.id, %s
		FROM %s p JOIN %s v
		ON (p.id = v.id_%[2]s AND v.id_%[4]s=?) ` + where,
		titleField, fieldStrc.TableProps, fieldStrc.TableValues, fieldStrc.Table.Name)

	return nil
}

func (fieldStrc *FieldStructure) writeSQLByTableID() error {


	where := fieldStrc.WhereFromSet(fieldStrc.Table)
	if where > "" {
		where += " OR (id_%s=?)"
	} else {
		where = " WHERE (id_%s=?)"
	}

	fieldStrc.SQLforFORMList =  fmt.Sprintf( `SELECT * FROM %s p ` + where, fieldStrc.TableProps, fieldStrc.Table.Name )

	return nil
}


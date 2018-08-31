// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

//todo необходима вариативность вывода в input select значений из енумов и справочников,пример - есть енум из 9-ти позиций а вывести нужно только 1,5,6(соответственно и юзер может что либо делать только с ними а не со всем списком 1-9)

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/logs"
	"regexp"
	"strings"
	"time"
)

var (
	enumValidator = regexp.MustCompile(`(?:'([^,]+)',?)`)
)

//FieldStructure нужна для подготовки к отображению поля на форме (возможно, в таблице и еще других компонентах веб-старницы)
//На данный момент создается на лету, в будущем
//TODO: hfpltkbnm yf cnfnbxtcre. b lbyfvbxtcre. xfcnb
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
	PrimaryKey               bool
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
	SETID, NODEID, TABLEID   bool
	IdForeign                bool
	TableProps, TableValues  string
	ChildrenFields           *FieldsTable
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

// TypeInput стиль показа для разных типов полей
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

// GetSQLFromSETID return sql-query string fro field {key}
func (field *FieldStructure) GetSQLFromSETID(key, parentTable string) string {
	tableProps := strings.TrimPrefix(key, "setid_")
	tableValue := parentTable + "_" + tableProps + "_has"

	titleField := field.GetForeignFields()
	if titleField == "" {
		return ""
	}
	// LEFT JOIN for get all propertyes values
	return fmt.Sprintf(`SELECT p.id, %s, id_%s
	FROM %s p LEFT JOIN %s v ON (p.id=v.id_%[3]s AND id_%[2]s=?) `,
		titleField, parentTable,
		tableProps, tableValue)

}

// GetForeignFields возвращает поле в связанной таблице, которое будет отдано пользователю
// например, для вторичных ключей отдает не idзаписи, а name || title || какой-либо складное поле
func (field *FieldStructure) GetForeignFields() string {

	if field.ForeignFields > "" {
		return field.ForeignFields
	}

	return field.GetParentFieldName()
}

// GetParentFieldName fro field by her name
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

	return "id"

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

	return titleFull, titlePart
}

// GetColumnTitles split title field for form render
func (field *FieldStructure) GetColumnTitles() (titleFull, titleLabel, placeholder, pattern, dataJson string) {

	counter := 1
	comma := ""
	for key, val := range field.DataJSOM {

		dataJson += comma + fmt.Sprintf(`"%s": "%s"`, key, val)
		counter++
		comma = ","
	}
	return field.COLUMN_COMMENT, field.COLUMN_COMMENT, field.Placeholder, field.Pattern, dataJson
}
func (field *FieldStructure) parseWhere(whereJSON interface{}) {
	switch mapWhere := whereJSON.(type) {
	case map[string]interface{}:

		comma := ""
		field.Where = ""

		for key, value := range mapWhere {
			enumVal := value.(string)
			// отбираем параметры типы :имя_поля
			//if i := strings.Index(enumVal, ":"); i > -1 {
			//	param := enumVal[i+1:]
			//	// считаем, что окончанием параметра могут быть символы ", )"
			//	if j := strings.IndexAny(param, ", )"); j > 0 {
			//		param = param[:j]
			//	}
			//	// мы добавим условие созначением пол текущей записи, если это поле найдено и в нем установлено значение
			//	//if paramField := field.Table.FindField(param); paramField == nil {
			//	//	continue
			//	//}
			//}
			field.Where += comma + key + " " + enumVal
			comma = " OR "

		}
	default:
		logs.ErrorLog(errors.New("not correct type WhereJSON !"), field.Table.Name, field.COLUMN_NAME, whereJSON)
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

// ParseComment извлекает из комментария к полю значимые токены
func (field *FieldStructure) ParseComment(COLUMN_COMMENT string) string {

	COLUMN_COMMENT, field.Pattern = cutPartFromTitle(COLUMN_COMMENT, "//", "")
	if posPattern := strings.Index(COLUMN_COMMENT, "{"); posPattern > 0 {

		dataJson := COLUMN_COMMENT[posPattern:]

		var properMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataJson), &properMap); err != nil {
			logs.ErrorLog(err.(error), dataJson)
		} else {
			for key, val := range properMap {

				switch key {
				case "figure":
					field.Figure = val.(string)
				case "classCSS":
					field.CSSClass = val.(string)
				case "placeholder":
					field.Placeholder = val.(string)
				case "pattern":
					//field.Pattern = getPattern( val.(string) )
				case "foreingKeys":
					field.ForeignFields = val.(string)
				case "inputType":
					field.InputType = val.(string)
				case "isHidden":
					field.IsHidden = val.(bool)
				case "linkTD":
					field.LinkTD = val.(string)
				case "where":
					field.parseWhere(val)
				case "maxDate":
					field.MaxDate = convertDatePattern(val.(string))
				case "minDate":
					field.MinDate = convertDatePattern(val.(string))
				case "events":
					field.Events = make(map[string]string, 0)
					for name, event := range val.(map[string]interface{}) {
						field.Events[name] = event.(string)
					}
				default:
					field.DataJSOM[key] = val
				}
			}
		}

		field.COLUMN_COMMENT = COLUMN_COMMENT[:posPattern]
	} else {
		field.COLUMN_COMMENT = COLUMN_COMMENT
	}

	return field.COLUMN_COMMENT
}

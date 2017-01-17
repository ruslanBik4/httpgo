package forms

import (
	"strings"
	"github.com/ruslanBik4/httpgo/models/db"
	"regexp"
	"fmt"
	"log"
	//"strconv"
	"database/sql"
)
var (
	enumValidator = regexp.MustCompile(`(?:'([^,]+)',?)`)

)
type FieldStructure struct {
	COLUMN_NAME   	string
	DATA_TYPE 	string
	COLUMN_DEFAULT 	string
	IS_NULLABLE 	string
	CHARACTER_SET_NAME string
	COLUMN_COMMENT 	string
	COLUMN_TYPE 	string
	CHARACTER_MAXIMUM_LENGTH int
	Value 		string
	IsHidden bool
	CSSClass  	string
	TableName 	string
	Events 		map[string] string
	Html		string
}
type FieldsTable struct {
	Name string
	IsDadata bool
	Rows [] FieldStructure
	Hiddens map[string] string
}

func (ns *FieldsTable) findField(name string) *FieldStructure {
	for _, field := range ns.Rows {
		if field.COLUMN_NAME == name {
			return &field
		}
	}

	return nil
}
func (field *FieldStructure) whereFromSet(ns *FieldsTable) (result string) {
	fields := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)
	comma  := " WHERE "
	for _, title := range fields {
		enumVal := title[len(title) - 1]
		if i := strings.Index(enumVal, ":"); i > 0 {
			param := ""
			if paramField := ns.findField(enumVal[i+1:]); paramField != nil {
				param = paramField.Value
			}
			enumVal = enumVal[:i] + fmt.Sprintf("'%s'", param)
		}
		result += comma + enumVal
		comma = " OR "
	}

	return result
}
func (field *FieldStructure) getMultiSelect(ns *FieldsTable){

	key := field.COLUMN_NAME
	tableProps := strings.TrimLeft(key, "setid_")
	tableValue := ns.Name + "_" + tableProps + "_has"

	titleField := getParentFieldName(tableProps)
	if titleField == "" {
		return
	}
	rows, err := db.DoSelect( fmt.Sprintf( "select %s, id_%s from %s p left join %s v ON p.id=v.id_%s %s",
					titleField, ns.Name,
		tableProps, tableValue,  tableProps, field.whereFromSet(ns) ) )
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	idx := 0

	field.Html = ""
	for rows.Next() {
		var title, checked string
		var idRooms sql.NullInt64

		if err := rows.Scan(&title, &idRooms); err != nil {
				log.Println(err)
				continue
		}
		if idRooms.Valid {
			checked = "checked"
		}
		idx++


		field.Html += "<li role='presentation'>" + renderCheckBox(key, title, idx, checked, "events", "") + "</li>"
	}

}
func (field *FieldStructure) renderSet(key, required, events, dataJson string) (result string) {
	fields := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)

	for idx, title := range fields {
		enumVal := title[len(title)-1]
		checked := ""
		if (field.Value > "") && (strings.Index(field.Value, enumVal) > -1) || (field.Value == "") && (enumVal == field.COLUMN_DEFAULT) {
			checked = "checked"
		}
		result += renderCheckBox(key + "[]", enumVal, idx, checked, events, dataJson)
	}

	return result
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
func (field *FieldStructure) GetColumnTitles() (titleFull, titleLabel, placeholder, pattern, dataJson string)  {
	titleFull = field.COLUMN_COMMENT
	if titleFull=="" {
		titleLabel = field.COLUMN_NAME
	} else if strings.Index(titleFull, ".") > 0 {
		titleLabel = titleFull[:strings.Index(titleFull, ".")]
	} else {
		titleLabel = titleFull
	}
	titleFull, pattern = cutPartFromTitle(titleFull, "//", "")
	titleFull, dataJson = cutPartFromTitle(titleFull, "{", "")
	titleFull, placeholder = cutPartFromTitle(titleFull, "#", titleFull)

	return titleFull, titleLabel, placeholder, pattern, dataJson
}
// заполняет структуру для формы данными, взятыми из структуры БД
func (fields *FieldsTable) PutDataFrom(ns db.FieldsTable) {

	for _, field := range ns.Rows {
		fieldStrc := &FieldStructure{
			COLUMN_NAME: field.COLUMN_NAME,
			DATA_TYPE  : field.DATA_TYPE,
			IS_NULLABLE: field.IS_NULLABLE,
			COLUMN_TYPE: field.COLUMN_TYPE,
			Events     : make(map[string] string, 0),
		}
		if field.CHARACTER_SET_NAME.Valid {
			fieldStrc.CHARACTER_SET_NAME = field.CHARACTER_SET_NAME.String
		}
		if field.COLUMN_COMMENT.Valid {
			fieldStrc.COLUMN_COMMENT = field.COLUMN_COMMENT.String
		}
		if field.CHARACTER_MAXIMUM_LENGTH.Valid {
			fieldStrc.CHARACTER_MAXIMUM_LENGTH = int(field.CHARACTER_MAXIMUM_LENGTH.Int64)
		}
		if field.COLUMN_DEFAULT.Valid {
			fieldStrc.COLUMN_DEFAULT = field.COLUMN_DEFAULT.String
		}
		fieldStrc.IsHidden = false
		fields.Rows = append(fields.Rows,*fieldStrc)
	}
}
func getParentFieldName(tableName string) (name string) {
	var listNs db.FieldsTable

	if err := listNs.GetColumnsProp(tableName); err != nil {
		return ""
	}
	for _, list := range listNs.Rows {
		switch list.COLUMN_NAME {
		case "name":
			name = "name"
		case "title":
			name = "title"
		case "fullname":
			name = "fullname"
		}
	}

	return name

}


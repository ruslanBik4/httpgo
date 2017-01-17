package forms

import (
	"strings"
	"github.com/ruslanBik4/httpgo/models/db"
	"regexp"
)
var (
	enumValidator = regexp.MustCompile(`(?:'([^,]+)',?)`)

)
type FieldStructure struct {
	COLUMN_NAME   string
	DATA_TYPE string
	COLUMN_DEFAULT string
	IS_NULLABLE string
	CHARACTER_SET_NAME string
	COLUMN_COMMENT string
	COLUMN_TYPE string
	CHARACTER_MAXIMUM_LENGTH int
	Value string
	IsHidden bool
	CSSClass  string
	TableName string
	Events map[string] string
}
type FieldsTable struct {
	Name string
	IsDadata bool
	Rows [] FieldStructure
	Hiddens map[string] string
}
func (field *FieldStructure) whereFromSet() (result string) {
	fields := enumValidator.FindAllStringSubmatch(field.COLUMN_TYPE, -1)
	comma  := ""
	for _, title := range fields {
		enumVal := title[len(title) - 1]
		result += comma + enumVal
		comma = " OR "
	}

	return result
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


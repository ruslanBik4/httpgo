package tables

import (
	"database/sql"
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"regexp"
)

var enumValidator = regexp.MustCompile(`(?:'([^,]+)',?)`)
// QueryStruct has property for form record in table view
type QueryStruct struct {
	HrefEdit   string
	Href       string
	row        []interface{}
	columns    []string
	fields     []*forms.FieldStructure
	Rows       *sql.Rows
	Tables     []*forms.FieldsTable
	widthTable int
	Order      string
	PostFields []*forms.FieldStructure
}

func (query *QueryStruct) findField(fieldName string) *forms.FieldStructure {
	for _, fields := range query.Tables {
		if field := fields.FindField(fieldName); field != nil {
			field.Table = fields
			return field
		}
	}

	return nil

}
func (query *QueryStruct) beforeRender() (err error) {

	query.columns, err = query.Rows.Columns()
	if err != nil {
		logs.ErrorLog(err)
		return err
	}

	// mfields может не соответствовать набору столбцов, потому завязываем на имеющиеся, прочие - игнорируем
	for _, fieldName := range query.columns {
		var field *forms.FieldStructure
		if field = query.findField(fieldName); field == nil {
			field.COLUMN_NAME = fieldName
			field.COLUMN_COMMENT = fieldName
			//field.Table =
		}
		query.row = append(query.row, field)
		query.fields = append(query.fields, field)
	}

	return nil
}

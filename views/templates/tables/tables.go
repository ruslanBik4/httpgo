package tables

import (
	"log"
	"database/sql"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
)

type QueryStruct struct {
	HrefEdit string
	Href     string
	row [] interface {}
	columns [] string
	fields  [] *forms.FieldStructure
	Rows 	*sql.Rows
	Tables [] *forms.FieldsTable
}

func (query *QueryStruct) findField(fieldName string) *forms.FieldStructure {
	for _, fields := range query.Tables {
		if field := fields.FindField(fieldName); field != nil {
			field.Table  = fields
			query.fields = append(query.fields, field)
			return field
		}
	}

	return nil

}
func (query *QueryStruct) beforeRender() (err error) {

	query.columns, err = query.Rows.Columns()
	if (err != nil) {
		log.Println(err)
		return err
	}


	// mfields может не соответствовать набору столбцов, потому завязываем на имеющиеся, прочие - игнорируем
	for _, fieldName := range query.columns {
		if field := query.findField(fieldName); field == nil  {
			query.row = append(query.row, new(sql.NullString) )
			query.fields = append(query.fields, &forms.FieldStructure{COLUMN_NAME: fieldName, COLUMN_COMMENT: fieldName})
		} else {
			query.row = append(query.row, field)
		}
	}

	return nil
}
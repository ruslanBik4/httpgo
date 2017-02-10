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

func (query *QueryStruct) beforeRender() (err error) {

	query.columns, err = query.Rows.Columns()
	if (err != nil) {
		log.Println(err)
		return err
	}


	// mfields может не соответствовать набору столбцов, потому завязываем на имеющиеся, прочие - игнорируем
	for _, fieldName := range query.columns {
		var field interface {}

		for _, fields := range query.Tables {
			if field := fields.FindField(fieldName); field != nil {
				field.Table  = fields
				query.fields = append(query.fields, field)
				break
			}
		}
		if field == nil  {
			field = new(sql.NullString)
		}
		query.row = append( query.row, field )
	}

	return nil
}
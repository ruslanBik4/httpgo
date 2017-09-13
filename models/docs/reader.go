package docs

import (
	"github.com/nguyenthenguyen/docx"
	"github.com/ruslanBik4/httpgo/models/logs"
	"net/http"
)
// ReplaceDocx replaces exist output an input in map
func ReplaceDocx(input, output string, replaces map[string]string) bool {
	r, err := docx.ReadDocxFile(input);
	if err != nil {
		logs.ErrorLog(err)
		return false
	}

	docx1 := r.Editable()
	defer r.Close()

	for search, replace := range replaces {
		docx1.Replace(search, replace, -1)
	}

	if err := docx1.WriteToFile(output); err != nil {
		logs.ErrorLog(err)
		return false
	}

	return true
}
// RenderReplaesDoc read file with name {templatesName}, replace string from map & write to {w}
func RenderReplaesDoc(w http.ResponseWriter, templatesName string, replaces map[string]string) error {
	r, err := docx.ReadDocxFile(templatesName);
	if err != nil {
		logs.ErrorLog(err)
		return err
	}

	template := r.Editable()
	defer r.Close()
	for search, replace := range replaces {
		template.Replace(search, replace, -1)
	}
	template.Write(w)


	return nil
}

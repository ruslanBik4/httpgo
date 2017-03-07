package docx

import (
	"github.com/nguyenthenguyen/docx"
	"log"
	"net/http"
)

 func ReplaceDocx(input, output string, replaces map[string] string) bool {
	if r, err := docx.ReadDocxFile(input); err != nil {
		log.Println(err)
		return false
	} else {
		docx1 := r.Editable()
		defer r.Close()

		for search, replace := range replaces {
			docx1.Replace(search, replace, -1)
		}

		if err := docx1.WriteToFile(output); err != nil {
			log.Println(err)
			return false
		}


		return true
	}
}

func RenderReplaesDoc(w http.ResponseWriter, templatesName string, replaces map[string] string)  error {
	if r, err := docx.ReadDocxFile(templatesName); err != nil {
		log.Println(err)
		return err
	} else {
		template := r.Editable()
		defer r.Close()
		for search, replace := range replaces {
			template.Replace(search, replace, -1)
		}
		template.Write(w)
	}

	return nil
}
package docx

import (
	"github.com/nguyenthenguyen/docx"
	"log"
)

func ReadDocx(fileName string) string {
	if r, err := docx.ReadDocxFile(fileName); err != nil {
		log.Println(err)
		return ""
	} else {
		docx1 := r.Editable()
		docx1.Replace("«00» _____ 2016 г.", "«21» лютого 2017 г.", -1)
		docx1.Replace("Иванова Ивана Ивановича,", "Бикчентаева Руслана Ильдаровича", -1)
		docx1.WriteToFile("./new.docx")

		r.Close()

		return "./new.docx"
	}
}
package docx

import (
	"github.com/nguyenthenguyen/docx"
	"log"
	"flag"
)

var (
	f_web    = flag.String("web","/home/travel/thetravel/web/","path to web files")

)
func ReplaceDocx(input, output string, replaces map[string] string) bool {
	if r, err := docx.ReadDocxFile(input); err != nil {
		log.Println(err)
		return false
	} else {
		docx1 := r.Editable()

		for search, replace := range replaces {
			docx1.Replace(search, replace, -1)
		}

		if err := docx1.WriteToFile(output); err != nil {
			return false

		}

		r.Close()

		return true
	}
}
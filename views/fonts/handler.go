package fonts

import (
	"net/http"
	"strings"
	"io/ioutil"
	"log"
)
var 	PathWeb string
func GetPath(path string) {
	PathWeb = path
}
func contains(array [] string, str string) bool {
	for _, value := range array {
		if strings.Contains(value, str) {
			return true
		}
	}

	return false
}
func HandleGetFont(w http.ResponseWriter, r *http.Request) {

	//w.Header().Set("Content-Type", "mime/type; ttf")

	//log.Println(PathWeb + r.URL.Path)
	ext := ".ttf"
	if browser:= r.Header["User-Agent"]; contains(browser, "Safari") {
		ext = ".woff"
	} else {
		//http.ServeFile(w, r, PathWeb+r.URL.Path+ext)
		log.Println(browser)
	}

	w.Header().Set("Content-Type", "mime/type: font/x-woff")
	if data, err := ioutil.ReadFile(PathWeb+r.URL.Path+ext); err != nil {
		log.Println(err)
	} else {
		w.Write(data)
	}
}

package fonts

import (
	"net/http"
	"strings"
	"io/ioutil"
	"github.com/ruslanBik4/httpgo/models/logs"
)
var 	PathWeb string
func GetPath(path *string) {
	PathWeb = *path
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

	//PathWeb = "/home/travel/thetravel/web"
	ext := ".ttf"
	if browser:= r.Header["User-Agent"]; contains(browser, "Safari") {
		ext = ".woff"
		w.Header().Set("Content-Type", "mime/type: font/x-woff")
	} else {
		w.Header().Set("Content-Type", "mime/type: font/font-sfnt")
		//http.ServeFile(w, r, PathWeb+r.URL.Path+ext)
		logs.DebugLog("browser=",browser)
	}

	if data, err := ioutil.ReadFile(PathWeb+r.URL.Path+ext); err != nil {
		logs.ErrorLog(err)
	} else {
		w.Write(data)
	}
}

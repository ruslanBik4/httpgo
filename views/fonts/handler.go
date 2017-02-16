package fonts

import (
	"net/http"
	"strings"
)
var 	PathWeb string
func GetPath(path string) {
	PathWeb = path
}
func HandleGetFont(w http.ResponseWriter, r *http.Request) {

	//w.Header().Set("Content-Type", "mime/type; ttf")

	http.ServeFile(w, r, PathWeb + r.URL.Path)
}

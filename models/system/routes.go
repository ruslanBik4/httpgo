package system

import (
	"net/http"
	"fmt"
)
// you may create your routes and handler for custom web-site
var (
	CustomRoutes = map[string] func(w http.ResponseWriter, r *http.Request) {
		"/custom/": handlerCustom,
}
)

func handlerCustom(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "this is custom page %v", r )

}

package system

import (
	"net/http"
	"github.com/ruslanBik4/httpgo/models/users"
	"log"
	"fmt"
	"github.com/ruslanBik4/httpgo/views"
)

func Catch(w http.ResponseWriter, r *http.Request) {
	err := recover()

	switch err.(type) {
	case users.ErrNotLogin:
		fmt.Fprintf(w, "<title>%s</title>", "Для начала работы необходимо авторизоваться!" )
		views.RenderSignForm(w, r, "")
	case nil:
	default:
		log.Print("panic runtime! ", err)
		fmt.Fprintf(w, "Error during executing %v", err)
	}
}

func WrapCatchHandler(fnc http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Catch(w,r)
		fnc(w,r)
	})
}


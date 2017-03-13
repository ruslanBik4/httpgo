package system

import (
	"net/http"
	"github.com/ruslanBik4/httpgo/models/users"
	"log"
	"fmt"
)

func Catch(w http.ResponseWriter, r *http.Request) {
	err := recover()

	switch err.(type) {
	case users.ErrNotLogin:
		http.Redirect(w,r, "/show/forms/?name=signin", http.StatusSeeOther)
	case nil:
	default:
		log.Print("panic runtime! ", err)
		fmt.Fprint(w, "Error during executing %v", err)
	}
}

func WrapCatchHandler(fnc http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Catch(w,r)
		fnc(w,r)
	})
}


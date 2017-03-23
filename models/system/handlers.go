package system

import (
	"net/http"
	"log"
	"fmt"
)

type ErrNotLogin struct {
	Message string
}
func (err *ErrNotLogin) Error() string{
	return err.Message
}

func Catch(w http.ResponseWriter, r *http.Request) {
	err := recover()

	switch err.(type) {
	case ErrNotLogin:
		fmt.Fprintf(w, "<title>%s</title>", "Для начала работы необходимо авторизоваться!" )
		http.Redirect(w, r, "/show/forms/?name=signin", http.StatusSeeOther)
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


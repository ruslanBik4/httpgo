package system

import (
	"net/http"
	"log"
	"fmt"
	"github.com/ruslanBik4/httpgo/views"
)

type ErrNotLogin struct {
	Message string
}
func (err *ErrNotLogin) Error() string{
	return err.Message
}
//Структура для ошибок базы данных
type ErrDb struct {
	Message string
}
//Функция для обработк структуры ошибок базы данных
func (err *ErrDb) Error() string{
	return err.Message
}

type ErrNotPermission struct {
	Message string
}
func (err *ErrNotPermission) Error() string{
	return err.Message
}

func Catch(w http.ResponseWriter, r *http.Request) {
	err := recover()

	switch err.(type) {
	case ErrNotLogin:
		fmt.Fprintf(w, "<title>%s</title>", "Для начала работы необходимо авторизоваться!" )
		views.RenderSignForm(w, r, "")
	case ErrNotPermission:
		fmt.Fprintf(w, "<title>%s</title>", "Доступ закрыт" )
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


package views

import (
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"net/http"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/httpgo/views/templates/json"
	_ "github.com/ruslanBik4/httpgo/views/templates/system"
	"log"
//	"views/templates/layouts/common"
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"runtime"
)

//noinspection GoInvalidConstType
var HEADERS = map[string] string {
	"Content-Type": "text/html; charset=utf-8",
	"author":	"uStudio",
}
func WriteHeaders(w http.ResponseWriter) {
	// выдаем стандартные заголовки страницы
	for key, value := range HEADERS {
		w.Header().Set(key, value)
	}
}
func IsAJAXRequest(r *http.Request) bool {
	return len(r.Header["X-Requested-With"]) > 0
}
func RenderAnyPage(w http.ResponseWriter, r *http.Request, strContent string) {
	if IsAJAXRequest(r) {
		w.Write( []byte( strContent ) )
	} else {
		p := &pages.IndexPageBody{ Content: strContent, Route: r.URL.Path }
		RenderTemplate(w, r, "index", p)
	}
}
func RenderSignForm(w http.ResponseWriter, r *http.Request, email string )  {

	RenderAnyPage(w, r, forms.SigninForm(email, "Введите пароль, полученный по почте") )
}
func RenderSignUpForm(w http.ResponseWriter, r *http.Request, placeholder string )  {

	RenderAnyPage(w, r, forms.SignUpForm(placeholder) )
}
func RenderAnotherSignUpForm(w http.ResponseWriter, r *http.Request, placeholder string )  {

	RenderAnyPage(w, r, forms.AnotherSignUpForm(placeholder) )
}
func RenderNoPermissionPage(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
}
// render errors
func RenderBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
func RenderInternalError(w http.ResponseWriter, err error) {
	_, fn, line, _ := runtime.Caller(0)
	log.Printf("[error] %s:  in line %d. Error -  %v", fn, line, err)
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
}
func RenderUnAuthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
}
func RenderNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}
// render from template
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, Content interface{} ) error {

	WriteHeaders(w)

	headPage  := &layouts.HeadHTMLPage{
		Charset: "charset=utf-8",
		Language: "ru",
		Title: "Заголовок новой страницы",
	}

	switch tmplName {
	case "index":
		var p *pages.IndexPageBody = Content.(*pages.IndexPageBody)

		if p.Content == "" {

			//p.Title   = "Авторизация"
			headPage.Title = "Страничка управления миром - бета-версия"
			p.Route = "/"
		}

		w.Write( []byte( headPage.HeadHTML() + p.IndexHTML() ) )
	case "signinForm":
		RenderSignForm(w, r, "Введите пароль, полученный по почте")
	case "signupForm":
		RenderSignUpForm(w, r, "Введите ФАМИЛИЮ ИМЯ ОТЧЕСТВО")
	case "anothersignupForm":
		RenderAnotherSignUpForm(w, r, "Введите ФАМИЛИЮ ИМЯ ОТЧЕСТВО")

	case "adminPage":
		var p *pages.AdminPageBody = Content.(*pages.AdminPageBody)

		w.Write( []byte(p.ShowAdminPage("")) )
	default:
		w.Write( []byte( "no rendering with page " + tmplName ) )
	}
	return nil
}

func RenderAnyForm(w http.ResponseWriter, r *http.Request, Title string, fields forms.FieldsTable,
			Inputs map[string] []string, head, foot string ) error  {

	WriteHeaders(w)

	if Inputs != nil {
		head += layouts.DadataHead()
		foot += layouts.DadataScript(Inputs)
	}
	RenderAnyPage(w, r, head + layouts.PutHeadForm() + fields.ShowAnyForm("/admin/exec/", Title) + layouts.PutEndForm() + foot )

	return nil

}
// render JSON from any data type
var jsonHEADERS = map[string] string {
	"Content-Type": "application/json; charset=utf-8",
}
func WriteJSONHeaders(w http.ResponseWriter) {
	// выдаем стандартные заголовки страницы
	for key, value := range jsonHEADERS {
		w.Header().Set(key, value)
	}
}
func RenderAnyJSON(w http.ResponseWriter, arrJSON map[string] interface {}) {

	WriteJSONHeaders(w)
	w.Write( []byte( json.WriteAnyJSON(arrJSON) ) )
}
func RenderAnySlice(w http.ResponseWriter, arrJSON []interface{}) {

	WriteJSONHeaders(w)
	w.Write( []byte( json.WriteArrJSON(arrJSON) ) )
}
func RenderStringSliceJSON(w http.ResponseWriter, arrJSON []string) {

	WriteJSONHeaders(w)
	w.Write( []byte( json.WriteStringDimension(arrJSON) ) )
}

func RenderArrayJSON(w http.ResponseWriter, arrJSON [] map[string] interface {}) {

	WriteJSONHeaders(w)
	w.Write( []byte( json.WriteSliceJSON(arrJSON) ) )
}
// render JSON for form by fields map
func RenderJSONAnyForm(w http.ResponseWriter, fields schema.FieldsTable, form *json.FormStructure,
	AddJson map[string] string) {

	WriteJSONHeaders(w)
	w.Write( []byte(form.JSONAnyForm(fields, AddJson)) )
}
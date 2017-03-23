package views

//go:generate /Users/rus/go/bin/qtc -dir=views/templates

import (
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"fmt"
	"net/http"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/httpgo/views/templates/json"
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
		fmt.Fprint(w, strContent)
	} else {
		p := &pages.IndexPageBody{ Content: strContent, Route: r.URL.Path }
		RenderTemplate(w, r, "index", p)
	}
}
func RenderSignForm(w http.ResponseWriter, r *http.Request, email string )  {

	RenderAnyPage(w, r, forms.SigninForm(email, "Введите пароль, полученный по почте") )

}
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

		fmt.Fprint(w, headPage.HeadHTML())
		fmt.Fprint(w, p.IndexHTML())
	case "signinForm":
		var p *pages.IndexPageBody = Content.(*pages.IndexPageBody)
		RenderAnyPage(w, r, forms.SigninForm(p.Title, "Введите пароль, полученный по почте") )

	case "adminPage":
		var p *pages.AdminPageBody = Content.(*pages.AdminPageBody)

		fmt.Fprint(w, headPage.HeadHTML())
		fmt.Fprint(w, p.ShowAdminPage(""))
	default:
		fmt.Fprint(w, "no rendering with page %s with data %v", tmplName, Content)
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
func RenderAnyJSON(w http.ResponseWriter, arrJSON map[string] interface {}) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, json.WriteAnyJSON(arrJSON) )
}

func RenderArrayJSON(w http.ResponseWriter, arrJSON [] map[string] interface {}) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, json.WriteSliceJSON(arrJSON))
}

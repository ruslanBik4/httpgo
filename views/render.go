package views

//go:generate /Users/rus/go/bin/qtc -dir=views/templates

import (
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"fmt"
	"net/http"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/httpgo/models/users"
)

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, Content interface{} ) error {

	var menu db.MenuItems

	headPage  := &layouts.HeadHTMLPage{
		Charset: "charset=utf-8",
		Language: "ru",
		Title: "rcgthbvtyn gj ds;bdfyb. dhfujd jn yfc",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	switch tmplName {
	case "index":
		var p *pages.IndexPageBody = Content.(*pages.IndexPageBody)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		headPage.Title = "введите Ваше имя отчество фамилию"

		fmt.Fprint(w, headPage.HeadHTML())

		p.TopMenu = make( map[string] *pages.ItemMenu, 0)

		menu.GetMenu("indexTop")

		for _, item := range menu.Items {
			p.TopMenu[item.Title] = &pages.ItemMenu{ Link: "/menu/" + item.Name + "/"}

		}
		if p.Content == "" {

			p.Title   = "Авторизация"
			if userID, ok := users.IsLogin(r); ok {
				p.Content = fmt.Sprintf("<script>afterLogin({login:'%d',sex:'0'})</script>", userID)
			} else {
				p.Content = forms.SigninForm("", "Введите пароль") + forms.ShowForm("введите фамилию имя отчество")
			}
			p.Route = "/"
		}
		fmt.Fprint(w, p.IndexHTML())
	case "signinForm":
		var p *pages.IndexPageBody = Content.(*pages.IndexPageBody)
		fmt.Fprint(w, forms.SigninForm(p.Title, "Введите пароль, полученный по почте"))

	case "adminPage":
		var p *pages.AdminPageBody = Content.(*pages.AdminPageBody)

		p.TopMenu = make( map[string] *pages.ItemMenu, 0)
		menu.GetMenu("indexTop")

		for _, item := range menu.Items {
			p.TopMenu[item.Title] = &pages.ItemMenu{ Link: "/menu/" + item.Name + "/"}

		}
		fmt.Fprint(w, headPage.HeadHTML())
		fmt.Fprint(w, p.ShowAdminPage(""))

	default:
		fmt.Fprint(w, "no rendering with page %s with data %v", tmplName, Content)
	}
	return nil
}

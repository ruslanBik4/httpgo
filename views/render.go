// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package views подготовка вывода данных в поток возврата
package views

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/json"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
)

// HEADERS - list standard header for html page - noinspection GoInvalidConstType
var HEADERS = map[string]string{
	"Content-Type":     "text/html; charset=utf-8",
	"author":           "ruslanBik4",
	"Server":           "HTTPGO/0.9 (CentOS) Go 1.14",
	"Content-Language": "en, ru",
	"Age":              fmt.Sprintf("%f", time.Since(server.GetServerConfig().StartTime).Seconds()),
}

// WriteHeaders выдаем стандартные заголовки страницы
func WriteHeaders(ctx *fasthttp.RequestCtx) {
	for key, value := range HEADERS {
		ctx.Response.Header.Set(key, value)
	}
}

func WriteHeadersHTML(ctx *fasthttp.RequestCtx) {
	for key, value := range HEADERS {
		ctx.Response.Header.Set(key, value)
	}
}

// IsAJAXRequest - is this AJAX-request
func IsAJAXRequest(r *fasthttp.Request) bool {
	return len(r.Header.Peek("X-Requested-With")) > 0
}

// RenderOutput render for output script execute
func RenderHTMLPage(ctx *fasthttp.RequestCtx, fncWrite func(w io.Writer)) {

	WriteHeadersHTML(ctx)

	fncWrite(ctx)
}

// RenderAnyPage (deprecate)
//TODO: replace string output by streaming
func RenderAnyPage(ctx *fasthttp.RequestCtx, strContent string) error {
	if IsAJAXRequest(&ctx.Request) {
		_, err := ctx.Write([]byte(strContent))
		return err
	}
	p := &pages.IndexPageBody{Content: strContent, Route: string(ctx.Path()), Buff: ctx}

	return RenderTemplate(ctx, "index", p)
}

// RenderSignForm show form for authorization user
func RenderSignForm(ctx *fasthttp.RequestCtx, email string) {

	signForm := &forms.SignForm{Email: email, Password: "Введите пароль, полученный по почте"}
	RenderHTMLPage(ctx, signForm.WriteSigninForm)
}

// RenderSignUpForm show form registration user
func RenderSignUpForm(ctx *fasthttp.RequestCtx, placeholder string) {

	RenderAnyPage(ctx, forms.SignUpForm(placeholder))
}

// RenderAnotherSignUpForm  - new form for registration
func RenderAnotherSignUpForm(ctx *fasthttp.RequestCtx, placeholder string) {

	RenderAnyPage(ctx, forms.AnotherSignUpForm(placeholder))
}

// ParamNotCorrect - map bad parameters on this request
type ParamNotCorrect map[string]string

// RenderTemplate render from template tmplName
func RenderTemplate(ctx *fasthttp.RequestCtx, tmplName string, Content interface{}) error {

	WriteHeaders(ctx)

	headPage := &layouts.HeadHTMLPage{
		Charset:  "charset=utf-8",
		Language: "ru",
		Title:    "Заголовок новой страницы",
	}

	switch tmplName {
	case "index":
		var p *pages.IndexPageBody = Content.(*pages.IndexPageBody)

		if p.Content == "" {

			//p.Title   = "Авторизация"
			headPage.Title = "Страничка управления миром - бета-версия"
			p.Route = "/"
		}
		if p.Buff == nil {
			p.Buff = ctx
		}

		headPage.WriteHeadHTML(ctx)
		p.WriteIndexHTML(ctx)
	case "signinForm":
		RenderSignForm(ctx, "Введите пароль, полученный по почте")
	case "signupForm":
		RenderSignUpForm(ctx, "Введите ФАМИЛИЮ ИМЯ ОТЧЕСТВО")
	case "anothersignupForm":
		RenderAnotherSignUpForm(ctx, "Введите ФАМИЛИЮ ИМЯ ОТЧЕСТВО")

	case "adminPage":
		var p *pages.AdminPageBody = Content.(*pages.AdminPageBody)

		p.WriteShowAdminPage(ctx, "")
	default:
		ctx.Write([]byte("no rendering with page " + tmplName))
	}
	return nil
}

// RenderAnyForm show form for list fields
func RenderAnyForm(ctx *fasthttp.RequestCtx, Title string, fields forms.FieldsTable,
	Inputs map[string][]string, head, foot string) error {

	WriteHeaders(ctx)

	if Inputs != nil {
		head += layouts.DadataHead()
		foot += layouts.DadataScript(Inputs)
	}
	//TODO: replace on stream buffer function
	return RenderAnyPage(ctx, head+layouts.PutHeadForm()+fields.ShowAnyForm("/admin/exec/", Title)+layouts.PutEndForm()+foot)
}

// render JSON from any data type
var jsonHEADERS = map[string]string{
	"Content-Type": "application/json; charset=utf-8",
}

// WriteJSONHeaders return standart headers for JSON
func WriteJSONHeaders(ctx *fasthttp.RequestCtx) {
	// выдаем стандартные заголовки страницы
	for key, value := range jsonHEADERS {
		ctx.Response.Header.Set(key, value)
	}
}

// RenderAnyJSON marshal JSON from arrJSON
func RenderAnyJSON(w *fasthttp.RequestCtx, arrJSON map[string]interface{}) {

	WriteJSONHeaders(w)
	json.WriteAnyJSON(w, arrJSON)
}

// RenderAnySlice marshal JSON from slice
func RenderAnySlice(w *fasthttp.RequestCtx, arrJSON []interface{}) {

	WriteJSONHeaders(w)
	json.WriteArrJSON(w, arrJSON)
}

// RenderStringSliceJSON marshal JSON from slice strings
func RenderStringSliceJSON(w *fasthttp.RequestCtx, arrJSON []string) {

	WriteJSONHeaders(w)
	json.WriteStringDimension(w, arrJSON)
}

// RenderArrayJSON marshal JSON from arrJSON
func RenderArrayJSON(ctx *fasthttp.RequestCtx, arrJSON []map[string]interface{}) {

	WriteJSONHeaders(ctx)
	json.WriteSliceJSON(ctx, arrJSON)
}

// RenderJSONAnyForm render JSON for form by fields map
func RenderJSONAnyForm(w *fasthttp.RequestCtx, fields qb.QBTable, form *json.FormStructure,
	AddJson json.MultiDimension) {

	WriteJSONHeaders(w)
	form.WriteJSONAnyForm(w, fields, AddJson)
}

// RenderOutput render for output script execute
func RenderOutput(w *fasthttp.RequestCtx, stdoutStderr []byte) {

	WriteHeaders(w)
	w.Write([]byte("<pre>"))
	w.Write(bytes.Replace(stdoutStderr, []byte("\n"), []byte("<br>"), 0))
	w.Write([]byte("</pre>"))
}

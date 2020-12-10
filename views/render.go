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

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
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

	_ = RenderAnyPage(ctx, forms.SignUpForm(placeholder))
}

// RenderAnotherSignUpForm  - new form for registration
func RenderAnotherSignUpForm(ctx *fasthttp.RequestCtx, placeholder string) {

	_ = RenderAnyPage(ctx, forms.AnotherSignUpForm(placeholder))
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
		_, err := ctx.Write([]byte("no rendering with page " + tmplName))
		if err != nil {
			return errors.Wrap(err, tmplName)
		}
	}
	return nil
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

// RenderOutput render for output script execute
func RenderOutput(ctx *fasthttp.RequestCtx, stdoutStderr []byte, err error) error {

	if err != nil {
		return err
	} else {
		WriteHeaders(ctx)
	}

	return RenderOutWithWrapLine(ctx, stdoutStderr)
}

var (
	startPre = bytes.NewBufferString("<pre>")
	endPre   = bytes.NewBufferString("</pre>")
	wrapLine = []byte("<br>")
)

func RenderOutWithWrapLine(ctx *fasthttp.RequestCtx, out []byte) error {
	_, err := startPre.WriteTo(ctx)
	if err == nil {
		_, err = ctx.Write(bytes.Replace(out, []byte("\n"), wrapLine, -1))
		if err == nil {
			_, err = endPre.WriteTo(ctx)
		}
	}

	return errors.Wrap(err, "RenderOutWithWrapLine")
}

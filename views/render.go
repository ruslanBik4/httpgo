/*
 * Copyright (c) 2022-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Package views подготовка вывода данных в поток возврата
package views

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/json"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/logs"
)

// HEADERS - list standard header for html page - noinspection GoInvalidConstType
var HEADERS = map[string]string{
	"author":           "ruslanBik4",
	"Server":           "%v HTTPGO/%v (CentOS) Go 1.21",
	"Content-Language": "en,uk",
}

// WriteHeaders выдаем стандартные заголовки страницы
func WriteHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentEncoding("utf-8")
	age, ok := ctx.UserValue(AgeOfServer).(float64)
	if ok {
		ctx.Response.Header.Set("Age", fmt.Sprintf("%f", age))
	}
	ctx.Response.Header.SetLastModified(time.Now().Add(-(time.Second * time.Duration(age))))
	for key, value := range HEADERS {
		if key == "Server" {
			value = fmt.Sprintf(value, ctx.UserValue("name of server httpgo"), ctx.UserValue("ACC_VERSION"))
		}
		if len(ctx.Response.Header.Peek(key)) == 0 {
			ctx.Response.Header.Set(key, value)
		}
	}
}

func WriteHeadersHTML(ctx *fasthttp.RequestCtx) {
	WriteHeaders(ctx)
	ctx.Response.Header.SetContentType("text/html; charset=utf-8")
}

// IsAJAXRequest - is this AJAX-request
func IsAJAXRequest(r *fasthttp.Request) bool {
	return len(r.Header.Peek("X-Requested-With")) > 0
}

func WriteDownloadHeaders(ctx *fasthttp.RequestCtx, lastModify time.Time, fileName string, length int) {
	ctx.Response.Header.Set("Content-Description", "File Transfer")
	ctx.Response.Header.Set("Content-Transfer-Encoding", "binary")
	ctx.Response.Header.Set("Cache-Control", "must-revalidate")
	ctx.Response.Header.SetLastModified(lastModify)
	if length > 0 {
		ctx.Response.Header.SetContentLength(length)
	}

	if ext := path.Ext(fileName); ext > "" {
		ctx.Response.Header.SetContentType(mime.TypeByExtension(ext))
	} else {
		ct := http.DetectContentType(ctx.Response.Body())
		if ext, err := mime.ExtensionsByType(ct); err != nil {
			logs.ErrorLog(err)
		} else if len(ext) > 0 {
			fileName += ext[0]
		}

		ctx.Response.Header.SetContentType(ct)
	}

	ctx.Response.Header.Set("Content-Disposition", "attachment; filename="+fileName)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// RenderHTMLPage render for output script execute
func RenderHTMLPage(ctx *fasthttp.RequestCtx, fncWrite func(w io.Writer)) {

	WriteHeadersHTML(ctx)

	fncWrite(ctx)
}

// RenderAnyPage (deprecate)
// TODO: replace string output by streaming
func RenderAnyPage(ctx *fasthttp.RequestCtx, strContent string) error {
	content := gotools.StringToBytes(strContent)
	if IsAJAXRequest(&ctx.Request) {
		_, err := ctx.Write(content)
		return err
	}
	p := &pages.IndexPageBody{Content: content, Route: string(ctx.Path()), Buff: ctx}

	return RenderTemplate(ctx, "index", p)
}

// RenderSignForm show form for authorization user
func RenderSignForm(ctx *fasthttp.RequestCtx, email string) {

	signForm := &forms.SignForm{Email: email, Password: "Enter password that was sending on email"}
	RenderHTMLPage(ctx, signForm.WriteSigningForm)
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

	WriteHeadersHTML(ctx)

	headPage := &layouts.HeadHTMLPage{
		Charset:  "charset=utf-8",
		Language: "eu",
		Title:    "Title of new page",
	}

	switch tmplName {
	case "index":
		var p *pages.IndexPageBody = Content.(*pages.IndexPageBody)

		if len(p.Content) == 0 {

			//p.Title   = "Авторизация"
			headPage.Title = "Страничка управления миром - бета-версия"
			p.Route = "/"
		}
		if p.Buff == nil {
			p.Buff = ctx
		}

		if p.HeadHTML == nil {
			headPage.WriteHeadHTML(ctx)
		}
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

// render JSON from any data type
const jsonHEADERSContentType = "application/json; charset=utf-8"

// WriteJSONHeaders return standart headers for JSON
func WriteJSONHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(jsonHEADERSContentType)
	WriteHeaders(ctx)
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
		_, err = ctx.Write(ReplaceWrapLines(out))
		if err == nil {
			_, err = endPre.WriteTo(ctx)
		}
	}

	return errors.Wrap(err, "RenderOutWithWrapLine")
}

func ReplaceWrapLines(out []byte) []byte {
	return bytes.ReplaceAll(out, []byte("\n"), wrapLine)
}

const AgeOfServer = "AGE"

// WriteResponse to ctx body according to type of resp
func WriteResponse(ctx *fasthttp.RequestCtx, resp any) error {
	switch resp := resp.(type) {
	case nil:
	case []byte:
		ctx.Response.SetBodyRaw(resp)
	case string:
		ctx.Response.SetBodyString(resp)
	case int, int16, int32, int64, bool, float32, float64:
		_, err := fmt.Fprintf(ctx, "%v", resp)
		return err
	//case crud.DtoFileField:
	//	for _, header := range resp {
	//		return Download(ctx, header)
	//	}
	case *multipart.FileHeader:
		return Download(ctx, resp)
	case []*multipart.FileHeader:
		for _, header := range resp {
			return Download(ctx, header)
		}
	default:
		return WriteJSON(ctx, resp)
	}

	return nil
}

// WriteJSON write JSON to response
func WriteJSON(ctx *fasthttp.RequestCtx, r any) (err error) {

	defer func() {
		if err == nil {
			errR := recover()
			if errR != nil {
				logs.ErrorStack(err, "WriteJSON")
				err = errors.Wrap(errR.(error), "marshal json")
			}
		}
	}()

	json.WriteElement(ctx, r)
	WriteJSONHeaders(ctx)

	return nil
}

func Download(ctx *fasthttp.RequestCtx, fHeader *multipart.FileHeader) error {
	f, err := fHeader.Open()
	if err != nil {
		logs.DebugLog(err, fHeader)
		return errors.Wrap(err, fHeader.Filename)
	}

	size := int(fHeader.Size)
	ctx.Response.SetBodyStream(f, size)
	WriteDownloadHeaders(ctx, time.Now(), fHeader.Filename, size)

	return nil
}

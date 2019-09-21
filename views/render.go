// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package views подготовка вывода данных в поток возврата
package views

import (
	"bytes"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/json"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"io"
	"net/http"
	"strings"
	"time"
)

// HEADERS - list standard header for html page - noinspection GoInvalidConstType
var HEADERS = map[string]string{
	"Content-Type":     "text/html; charset=utf-8",
	"author":           "ruslanBik4",
	"Server":           "HTTPGO/0.9 (CentOS) Go 1.12",
	"Content-Language": "en, ru",
	"Age":              fmt.Sprintf("%f", time.Since(server.GetServerConfig().StartTime).Seconds()),
}

// WriteHeaders выдаем стандартные заголовки страницы
func WriteHeaders(w http.ResponseWriter) {
	for key, value := range HEADERS {
		w.Header().Set(key, value)
	}
}

// IsAJAXRequest - is this AJAX-request
func IsAJAXRequest(r *http.Request) bool {
	return len(r.Header["X-Requested-With"]) > 0
}

// RenderContentFromAJAXRequest NEW! эта функция позволяет определить - пришел ли запрос как AJAX
// и, если нет, добавить в вывод текст основной страницы
// получает на вход функцию qtpl, которая пишет сразу в буфер вывода
func RenderContentFromAJAXRequest(w http.ResponseWriter, r *http.Request, fncWrite func(w io.Writer)) {
	if IsAJAXRequest(r) {
		fncWrite(w)
	} else {
		p := &pages.IndexPageBody{ContentWrite: fncWrite, Route: r.URL.Path, Buff: w}
		RenderTemplate(w, r, "index", p)
	}

}

// RenderAnyPage (deprecate)
//TODO: replace string output by streaming
func RenderAnyPage(w http.ResponseWriter, r *http.Request, strContent string) {
	if IsAJAXRequest(r) {
		w.Write([]byte(strContent))
	} else {
		p := &pages.IndexPageBody{Content: strContent, Route: r.URL.Path, Buff: w}
		RenderTemplate(w, r, "index", p)
	}
}

// RenderSignForm show form for authorization user
func RenderSignForm(w http.ResponseWriter, r *http.Request, email string) {

	signForm := &forms.SignForm{Email: email, Password: "Введите пароль, полученный по почте"}
	RenderContentFromAJAXRequest(w, r, signForm.WriteSigninForm)
}

// RenderSignUpForm show form registration user
func RenderSignUpForm(w http.ResponseWriter, r *http.Request, placeholder string) {

	RenderAnyPage(w, r, forms.SignUpForm(placeholder))
}

// RenderAnotherSignUpForm  - new form for registration
func RenderAnotherSignUpForm(w http.ResponseWriter, r *http.Request, placeholder string) {

	RenderAnyPage(w, r, forms.AnotherSignUpForm(placeholder))
}

// ParamNotCorrect - map bad parameters on this request
type ParamNotCorrect map[string]string

// render errors

// RenderNotParamsInPOST get list params thoese not found in request
func RenderNotParamsInPOST(w http.ResponseWriter, params ...string) {
	http.Error(w, strings.Join(params, ",")+": not found", http.StatusBadRequest)

}

// RenderBadRequest return header "BADREQUEST" & descriptors bad params
func RenderBadRequest(w http.ResponseWriter, params ...ParamNotCorrect) {

	description, comma := "", ""
	for _, param := range params {
		description += comma
		for key, value := range param {
			description += key + "=" + value
		}

		comma = "; "
	}

	http.Error(w, description, http.StatusBadRequest)
}

// RenderHandlerError для отдачи и записи в лог паники системы при работе хендлеров
func RenderHandlerError(w http.ResponseWriter, err error, args ...interface{}) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	logs.ErrorLogHandler(err, args)
}

// RenderInternalError для отдачи и записи в лог ошибок системы при работе хендлеров
func RenderInternalError(w http.ResponseWriter, err error, args ...interface{}) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	logs.ErrorLog(err, args)
}

// RenderUnAuthorized - returs error code
func RenderUnAuthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
}

// RenderNotFound - returs error code
func RenderNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

// RenderNoPermissionPage - returs error code
func RenderNoPermissionPage(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
}

// RenderTemplate render from template tmplName
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, Content interface{}) error {

	WriteHeaders(w)

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
			p.Buff = w
		}

		headPage.WriteHeadHTML(w)
		p.WriteIndexHTML(w)
	case "signinForm":
		RenderSignForm(w, r, "Введите пароль, полученный по почте")
	case "signupForm":
		RenderSignUpForm(w, r, "Введите ФАМИЛИЮ ИМЯ ОТЧЕСТВО")
	case "anothersignupForm":
		RenderAnotherSignUpForm(w, r, "Введите ФАМИЛИЮ ИМЯ ОТЧЕСТВО")

	case "adminPage":
		var p *pages.AdminPageBody = Content.(*pages.AdminPageBody)

		p.WriteShowAdminPage(w, "")
	default:
		w.Write([]byte("no rendering with page " + tmplName))
	}
	return nil
}

// RenderAnyForm show form for list fields
func RenderAnyForm(w http.ResponseWriter, r *http.Request, Title string, fields forms.FieldsTable,
	Inputs map[string][]string, head, foot string) error {

	WriteHeaders(w)

	if Inputs != nil {
		head += layouts.DadataHead()
		foot += layouts.DadataScript(Inputs)
	}
	//TODO: replace on stream buffer function
	RenderAnyPage(w, r, head+layouts.PutHeadForm()+fields.ShowAnyForm("/admin/exec/", Title)+layouts.PutEndForm()+foot)

	return nil

}

// render JSON from any data type
var jsonHEADERS = map[string]string{
	"Content-Type": "application/json; charset=utf-8",
}

// WriteJSONHeaders return standart headers for JSON
func WriteJSONHeaders(w http.ResponseWriter) {
	// выдаем стандартные заголовки страницы
	for key, value := range jsonHEADERS {
		w.Header().Set(key, value)
	}
}

// RenderAnyJSON marshal JSON from arrJSON
func RenderAnyJSON(w http.ResponseWriter, arrJSON map[string]interface{}) {

	WriteJSONHeaders(w)
	json.WriteAnyJSON(w, arrJSON)
}

// RenderAnySlice marshal JSON from slice
func RenderAnySlice(w http.ResponseWriter, arrJSON []interface{}) {

	WriteJSONHeaders(w)
	json.WriteArrJSON(w, arrJSON)
}

// RenderStringSliceJSON marshal JSON from slice strings
func RenderStringSliceJSON(w http.ResponseWriter, arrJSON []string) {

	WriteJSONHeaders(w)
	json.WriteStringDimension(w, arrJSON)
}

// RenderArrayJSON marshal JSON from arrJSON
func RenderArrayJSON(w http.ResponseWriter, arrJSON []map[string]interface{}) {

	WriteJSONHeaders(w)
	json.WriteSliceJSON(w, arrJSON)
}

// RenderJSONAnyForm render JSON for form by fields map
func RenderJSONAnyForm(w http.ResponseWriter, fields qb.QBTable, form *json.FormStructure,
	AddJson json.MultiDimension) {

	WriteJSONHeaders(w)
	form.WriteJSONAnyForm(w, fields, AddJson)
}

// RenderOutput render for output script execute
func RenderOutput(w http.ResponseWriter, stdoutStderr []byte) {

	WriteHeaders(w)
	w.Write([]byte("<pre>"))
	w.Write(bytes.Replace(stdoutStderr, []byte("\n"), []byte("<br>"), 0))
	w.Write([]byte("</pre>"))
}

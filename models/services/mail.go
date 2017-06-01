// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Creating 01.06.17

package services

import (
	"gopkg.in/gomail.v2"
	netMail "net/mail"
	"github.com/ruslanBik4/httpgo/models/logs"
)

type mailService struct {
	name   string
	status string
	email, password, body string
}

var (
	mail *mailService = &mailService{name: "mail"}
)

func (mail *mailService) Init() error {
	schema.status = "ready"
	return nil
}

//TODO: нужно методы ниже имплементировать сюда
func (mail *mailService) Send(messages ...interface{}) error {
	return nil

}
//TODO: настройки отправки надо вынести в конфигфайл
func SendMail(email, password, body string)  {

	m := gomail.NewMessage()
	m.SetHeader("From", "ruslan-bik@yandex.ru")
	m.SetHeader("To", email )
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Регистрация на travel.com.ua!")
	m.SetBody("text/html", body)
	m.Attach("/home/travel/bootstrap/ico/favicon.png")

	d := gomail.NewDialer("smtp.yandex.ru", 587, "ruslan-bik", "FalconSwallow")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		logs.ErrorLog(err, m)
	}
	logs.DebugLog(email, password)
}
func VerifyMail(email, password string) {

	if _, err := netMail.ParseAddress(email); err != nil {
		logs.ErrorLog(err, email)
		logs.DebugLog( "Что-то неверное с вашей почтой, не смогу отослать письмо! %v", err)
		return
	}
}

func (mail *mailService) Get(messages ... interface{}) (responce interface{}, err error) {
	return nil, nil

}
func (mail *mailService) Connect(in <-chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (mail *mailService) Close(out chan<- interface{}) error {

	return nil
}
func (mail *mailService) Status() string {

	return ""
}

func init() {
	AddService(mail.name, mail)
}


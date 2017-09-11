// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Creating 01.06.17

package services

import (
	"github.com/ruslanBik4/httpgo/models/logs"
	"gopkg.in/gomail.v2"
	netMail "net/mail"
	"os"
	"errors"
	"path/filepath"
	"gopkg.in/yaml.v2"
)

type mailService struct {
	name   string
	status string
	mConfig struct {
		Server   string `yaml:"smtp-server"`
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
		Port     int `yaml:"port"`
	}
}
type Mail struct {
	From         string
	To           string
	Subject      string
	Content_type string
	Body         string
	Attachments  []string
}

var (
	mailServ *mailService = &mailService{name: "mail"}
)

const PATH_FLAG = "-path"
const STATUS_PREPARING = "preparing data"
const STATUS_ERROR = "error"
const STATUS_READY = "ready"
const TYPE_PLAIN_TEXT = "text/plain"
const TYPE_HTML = "text/html"

func init() {
	AddService(mailServ.name, mailServ)
}
func Examlpe_SendEmail() {

	mail := Mail{
		From:         "ruslan@gmail.com",
		To:           "sizov.mykhailo@gmail.com",
		Body:         "you shall no pass!",
		Subject:      "massage subject text",
		Content_type: TYPE_PLAIN_TEXT, // TYPE_PLAIN_TEXT || TYPE_HTML
	}
	if err := Send("mail", mail); err != nil {
		logs.ErrorLog(err, mail)
	}
}

func (mailServ *mailService) Init() error {

	mailServ.status = STATUS_PREPARING
	f, err := os.Open(filepath.Join(mailServ.getStaticFilePath(), "config/mail.yml"))
	if err != nil {
		mailServ.status = STATUS_ERROR
		return err
	}
	fileInfo, _ := f.Stat()
	b := make([]byte, fileInfo.Size())
	if n, err := f.Read(b); err != nil {
		mailServ.status = STATUS_ERROR
		logs.ErrorLog(err, "n=", n)
		return err
	}
	if err := yaml.Unmarshal(b, &mailServ.mConfig); err != nil {
		mailServ.status = STATUS_ERROR
		return err
	}
	mailServ.status = STATUS_READY
	return nil
}

func (mailServ *mailService) Send(messages ...interface{}) error {

	currentMail, err := mailServ.getMailStruct(messages...)
	if err != nil {
		return err
	}
	if err := mailServ.SendMail(currentMail); err != nil {
		return err
	}
	return nil
}

func (mailServ *mailService) SendMail(mail *Mail) error {

	m := gomail.NewMessage()
	from := mailServ.mConfig.Email
	if mail.From != "" {
		from = mail.From
	}
	m.SetHeader("From", from)
	m.SetHeader("To", mail.To)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", mail.Subject)
	m.SetBody(mail.Content_type, mail.Body)
	for _, file := range mail.Attachments {
		m.Attach(file)
	}
	d := mailServ.getDialer()
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		logs.ErrorLog(err)
		return err
	}
	return nil
}

func (mailServ *mailService) Get(messages ...interface{}) (response interface{}, err error) {
	return nil, nil

}
func (mailServ *mailService) Connect(in <-chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (mailServ *mailService) Close(out chan<- interface{}) error {

	return nil
}
func (mailServ *mailService) Status() string {

	return mailServ.status
}
func (mailServ *mailService) getStaticFilePath() string {

	for key, val := range os.Args {
		if val == PATH_FLAG {
			return os.Args[key+1]
		}
	}
	return ""
}
func (mailServ *mailService) getDialer() *gomail.Dialer {
	return gomail.NewDialer(mailServ.mConfig.Server, mailServ.mConfig.Port, mailServ.mConfig.Email, mailServ.mConfig.Password)
}
func (mailServ *mailService) getMailStruct(messages ...interface{}) (*Mail, error) {

	currentMail := new(Mail)
	for _, message := range messages {
		switch mess := message.(type) {
		case Mail:
			currentMail = &mess
		default:
			return nil, &ErrServiceNotCorrectParamType{
				Name: mailServ.name,
			}

		}
	}
	if err := currentMail.validate(); err != nil {
		return nil, err
	}
	return currentMail, nil
}
func (mail *Mail) validate() error {

	if mail.To == "" {
		return &ErrServiceNotEnoughParameter{
			Name: "mail From",
		}
	}
	if _, err := netMail.ParseAddress(mail.To); err != nil {
		return err
	}

	if mail.Subject == "" {
		return &ErrServiceNotEnoughParameter{
			Name: "mail Subject",
		}
	}
	if len(mail.Subject) > 255 {
		return errors.New("Massage Subject is too long")
	}
	if mail.Content_type != TYPE_HTML && mail.Content_type != TYPE_PLAIN_TEXT {
		return errors.New("Content type is illegal")
	}
	return nil
}

func VerifyMail(email, password string) {

	if _, err := netMail.ParseAddress(email); err != nil {
		logs.ErrorLog(err, email)
		logs.DebugLog("Что-то неверное с вашей почтой, не смогу отослать письмо! %v", err)
		return
	}
}

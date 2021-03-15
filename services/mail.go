// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Creating 01.06.17

package services

import (
	"io/ioutil"
	netMail "net/mail"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"gopkg.in/gomail.v2"
	"gopkg.in/yaml.v2"

	"github.com/ruslanBik4/logs"
)

type mailService struct {
	name    string
	status  string
	mConfig struct {
		Server   string `yaml:"smtp-server"`
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
		Port     int    `yaml:"port"`
	}
}

// Mail - тип данных для письма
type Mail struct {
	From        string
	To          string
	Subject     string
	ContentType string
	Body        string
	Attachments []string
}

var (
	mailServ *mailService = &mailService{name: "mail"}
)

// PATH_FLAG - сиситемая константа - флаг пути к файлам конфигурации.
const PATH_FLAG = "-path"

// TYPE_PLAIN_TEXT - тип отправляемого сообщения - "Текст"
const TYPE_PLAIN_TEXT = "text/plain"

// TYPE_HTML - тип отправляемого сообщения - "HTML"
const TYPE_HTML = "text/html"

// ExamlpeSendEmail - пример отправки сообщения
func ExampLeSendEmail() {

	mail := Mail{
		From:        "ruslan@gmail.com",
		To:          "sizov.mykhailo@gmail.com",
		Body:        "you shall no pass!",
		Subject:     "massage subject text",
		ContentType: TYPE_PLAIN_TEXT, // TYPE_PLAIN_TEXT || TYPE_HTML
	}
	if err := Send(context.TODO(), "mail", mail); err != nil {
		logs.ErrorLog(err, mail)
	}
}

func (mailServ *mailService) Init(ctx context.Context) error {

	mailServ.status = STATUS_PREPARING
	cfg, ok := ctx.Value("cfg_path").(string)
	if !ok {
		cfg = "config/mail.yml"
	}

	fileName := filepath.Join(mailServ.getStaticFilePath(), cfg)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		mailServ.status = STATUS_ERROR

		return errors.Wrap(err, "open file "+fileName)
	}

	if err := yaml.Unmarshal(b, &mailServ.mConfig); err != nil {
		mailServ.status = STATUS_ERROR
		return errors.Wrap(err, "Unmarshal "+string(b))
	}

	mailServ.status = STATUS_READY

	return nil
}

func (mailServ *mailService) Send(ctx context.Context, messages ...interface{}) error {

	currentMail, err := mailServ.getMailStruct(messages...)
	if err != nil {
		return err
	}
	if err := mailServ.SendMail(ctx, currentMail); err != nil {
		return err
	}
	return nil
}

func (mailServ *mailService) SendMail(ctx context.Context, mail *Mail) error {

	m := gomail.NewMessage()
	from := mailServ.mConfig.Email
	if mail.From != "" {
		from = mail.From
	}
	m.SetHeader("From", from)
	m.SetHeader("To", mail.To)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", mail.Subject)
	m.SetBody(mail.ContentType, mail.Body)
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

func (mailServ *mailService) Get(ctx context.Context, messages ...interface{}) (response interface{}, err error) {
	logs.DebugLog(messages)
	return nil, nil

}
func (mailServ *mailService) Connect(in <-chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (mailServ *mailService) Close(out chan<- interface{}) error {

	close(out)

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
			Name: "email To",
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
	if mail.ContentType != TYPE_HTML && mail.ContentType != TYPE_PLAIN_TEXT {
		return errors.New("Content type is illegal")
	}
	return nil
}

// VerifyMail - проверка на валидность email - адреса
func VerifyMail(email, password string) {

	if _, err := netMail.ParseAddress(email); err != nil {
		logs.ErrorLog(err, email)
		logs.DebugLog("Что-то неверное с вашей почтой, не смогу отослать письмо! %v", err)
		return
	}
}

func init() {
	AddService(mailServ.name, mailServ)
}

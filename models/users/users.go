// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package users - авторизация, регистрация юзеров и проверка прав для разделов сайта
package users

import (
	"crypto/rand"
	// "gopkg.in/gomail.v2"
	"database/sql"
	"encoding/base64"
	"flag"
	"hash/crc32"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/gomail.v2"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/views/templates/mails"
)

const nameSession = "PHPSESSID"

// NOT_AUTHORIZE message for response
const NOT_AUTHORIZE = "Нет данных об авторизации!"

var (
	fTest             = flag.Bool("test", false, "test mode")
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "",
		ClientID:     os.Getenv("googlekey"),
		ClientSecret: os.Getenv("googlesecret"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
	oauthStateString = "random"
	Store            = sessions.NewFilesystemStore("/var/lib/php/session", []byte("travel.com.ua"))
)

// SetSessionPath set path for session storage
func SetSessionPath(fSession string) {
	Store = sessions.NewFilesystemStore(fSession, []byte("travel.com.ua"))
}

// UserRecord for user data
type UserRecord struct {
	Id   int
	Name string
	Sex  int
}

var greetings = []string{"господин", "госпожа"}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// GeneratePassword run password by email
func GeneratePassword(email string) (string, error) {
	logs.DebugLog("email", email)
	return GenerateRandomString(16)

}

// HashPassword create hash from {password} & return checksumm
func HashPassword(password string) interface{} {
	// crypto password
	crc32q := crc32.MakeTable(0xD5828281)
	return crc32.Checksum([]byte(password), crc32q)
}

const _2K = (1 << 10) * 2

// SendMail create mail with new {password} & send to {email}
func SendMail(email, password string) {

	m := gomail.NewMessage()
	m.SetHeader("From", "ruslan-bik@yandex.ru")
	m.SetHeader("To", email)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Регистрация на travel.com.ua!")
	m.SetBody("text/html", mails.InviteEmail(email, password))
	//m.Attach("/favicon.ico")

	d := gomail.NewDialer("smtp.yandex.ru", 587, "ruslan-bik", "FalconSwallow")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		logs.ErrorLog(err)
	}

	logs.DebugLog("email-", email, ", password=", password)
}

// CheckUserCredentials check user data & return id + name
func CheckUserCredentials(login string, password string) (row UserRecord, err error) {

	var rows *sql.Rows
	rows, err = db.DoSelect("select id, fullname, sex from users where login=? and hash=?", login, HashPassword(password))
	if err != nil {
		logs.ErrorLog(err)
		return
	}
	defer rows.Close()

	for rows.Next() {

		err = rows.Scan(&row.Id, &row.Name, &row.Sex)

		if err != nil {
			logs.ErrorLog(err)
			return
		}

		return
	}

	err = &system.ErrNotLogin{Message: "Wrong email or password"}

	return
}

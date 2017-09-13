// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package users - авторизация, регистрация юзеров и проверка прав для разделов сайта
package users

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	//"gopkg.in/gomail.v2"
	"github.com/gorilla/sessions"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/mails"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/gomail.v2"
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

//func HandlerQauth2(w http.ResponseWriter, r *http.Request) {
//
//
//	googleOauthConfig.RedirectURL = r.Host +  "/user/GoogleCallback/"
//	url := googleOauthConfig.AuthCodeURL(oauthStateString)
//	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
//	//var ctx context.Context = appengine.NewContext(r)
//	//client := &http.Client{
//	//	Transport: &oauth2.Transport{
//	//		Source: google.AppEngineTokenSource(ctx, "scope"),
//	//		Base:   &urlfetch.Transport{Context: ctx},
//	//	},
//	//}
//	//resp, _ := client.Get("...")
//	//w.Write(resp.Body)
//}

// HandleGoogleCallback Эти callback было бы неплохо регистрировать в одну общую библиотеку для авторизации
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {

		logs.ErrorLog(err, "Code exchange failed with")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		logs.ErrorLog(err, "access_token")
		return
	}
	defer response.Body.Close()
	if contents, err := ioutil.ReadAll(response.Body); err != nil {
		logs.ErrorLog(err, "read_token")
	} else {
		fmt.Fprintf(w, "Content: %s\n", contents)
	}
}
// UserRecord for user data
type UserRecord struct {
	Id   int
	Name string
	Sex  int
}

var greetings = []string{"господин", "госпожа"}
// GetSession return session for cutrrent user (create is needing)
func GetSession(r *http.Request, name string) *sessions.Session {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := Store.Get(r, name)
	if err != nil {
		logs.ErrorLog(err)
		return nil
	}
	return session
}
// IsLogin return ID current user or panic()
func IsLogin(r *http.Request) string {
	if *fTest {
		return "8"
	}
	session := GetSession(r, nameSession)
	if session == nil {
		panic(http.ErrNotSupported)
	}
	userID, ok := session.Values["id"]
	if !ok {
		panic(system.ErrNotLogin{Message: "not login user!"})
	}

	return strconv.Itoa(userID.(int))
}
func deleteCurrentUser(w http.ResponseWriter, r *http.Request) error {
	session := GetSession(r, nameSession)
	delete(session.Values, "id")
	delete(session.Values, "email")
	return session.Save(r, w)

}
// HandlerProfile show data on profile current user
func HandlerProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	session := GetSession(r, nameSession)
	email, ok := session.Values["email"]
	if !ok {
		http.Redirect(w, r, "/show/forms/?name=signin", http.StatusSeeOther)
		return
	}
	rows := db.DoQuery("select id, fullname, sex from users where login=?", email)

	var row UserRecord

	defer rows.Close()
	for rows.Next() {

		err := rows.Scan(&row.Id, &row.Name, &row.Sex)

		if err != nil {
			logs.ErrorLog(err)
			continue
		}
	}

	p := &layouts.MenuOwnerBody{Title: greetings[row.Sex] + " " + row.Name, TopMenu: make(map[string]*layouts.ItemMenu, 0)}

	var menu db.MenuItems

	menu.GetMenu("menuOwner")

	for _, item := range menu.Items {
		p.TopMenu[item.Title] = &layouts.ItemMenu{Link: "/menu/" + item.Name + "/"}

	}
	fmt.Fprint(w, p.MenuOwner())
}
// HandlerSignIn run user authorization
func HandlerSignIn(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	email := r.FormValue("login")
	password := r.FormValue("password")

	if (email == "") || (password == "") {
		panic(&system.ErrNotLogin{Message: "Not enoug login parameters!"})
	}

	err, userId, userName := CheckUserCredentials(email, password)

	if err != nil {
		panic(&system.ErrNotLogin{Message: "Wrong email or password"})
	}

	// session save BEFORE write page
	SaveSession(w, r, userId, email)

	p := &forms.PersonData{Id: userId, Login: userName, Email: email}
	fmt.Fprint(w, p.JSON())
}
// HandlerSignOut sign out current user & show authorization form
func HandlerSignOut(w http.ResponseWriter, r *http.Request) {

	if err := deleteCurrentUser(w, r); err != nil {
		logs.ErrorLog(err)
	}
	fmt.Fprintf(w, "<title>%s</title>", "Для начала работы необходимо авторизоваться!")
	views.RenderSignForm(w, r, "")
}
// SaveSession save in session some data from user
func SaveSession(w http.ResponseWriter, r *http.Request, id int, email string) {
	session := sessions.NewSession(Store, nameSession)
	session.Options = &sessions.Options{Path: "/", HttpOnly: true, MaxAge: int(3600)}
	session.Values["id"] = id
	session.Values["email"] = email
	if err := session.Save(r, w); err != nil {
		logs.ErrorLog(err)
	}
}

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

// HandlerSignUp регистрация нового пользователя с генерацией пароля
// и отсылка письмо об регистрации
// пароль отсылаем в письме, у себя храним только кеш
// @/user/signup/
func HandlerSignUp(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(_2K)

	var args []interface{}

	JSON := make(map[string]interface{}, 4)

	sql, comma, values := "insert into users (", "", ") values ("

	for key, val := range r.Form {
		args = append(args, val[0])
		sql += comma + key
		values += comma + "?"
		comma = ","

		JSON[key] = val[0]
	}

	email := JSON["login"].(string)
	if JSON["form-radio"] == "radio-male" {
		JSON["sex"] = "господин"
	} else {
		JSON["sex"] = "госпожа"
	}
	password, err := GeneratePassword(email)
	if err == nil {
		sql += comma + "hash"
		values += comma + "?"
		// получаем кеш
		args = append(args, HashPassword(password))
		JSON["id"], err = db.DoInsert(sql+values+")", args...)
		if err == nil {
			// проверка корректности email
			if _, err := mail.ParseAddress(email); err == nil {
				go SendMail(email, password)
				views.RenderAnyJSON(w, JSON)
			}
		}
	}

	if err != nil {
		views.RenderInternalError(w, err)
		return
	}
}
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
// HandlerActivateUser - obsolete
func HandlerActivateUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.FormValue("email") == "" {

		logs.ErrorLog(errors.New("activate user not has email"))
	}
	if result, err := db.DoUpdate("update users set active=1 where login=?", r.Form["email"][0]); err != nil {
		logs.ErrorLog(err)
	} else {
		fmt.Fprint(w, result)

	}
}
// CheckUserCredentials check user data & return id + name
func CheckUserCredentials(login string, password string) (error, int, string) {

	rows, err := db.DoSelect("select id, fullname, sex from users where login=? and hash=?", login, HashPassword(password))
	if err != nil {
		logs.ErrorLog(err)
		return err, 0, ""
	}
	defer rows.Close()
	var row UserRecord

	for rows.Next() {

		err := rows.Scan(&row.Id, &row.Name, &row.Sex)

		if err != nil {
			logs.ErrorLog(err)
			continue
		}

		return nil, row.Id, row.Name
	}

	return &system.ErrNotLogin{Message: "Wrong email or password"}, 0, ""
}

package users

import (
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/mails"
	"github.com/ruslanBik4/httpgo/models/db"
	"net/http"
	"fmt"
	"strconv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"os"
	"log"
	"io/ioutil"
	"gopkg.in/gomail.v2"
	"net/mail"
	"crypto/rand"
	"encoding/base64"
	"hash/crc32"
	"github.com/gorilla/sessions"
)
const nameSession = "PHPSESSID"
const NOT_AUTHORIZE = "Нет данных об авторизации!"
var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "",
		ClientID:     os.Getenv("googlekey"),
		ClientSecret: os.Getenv("googlesecret"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	oauthStateString = "random"
	store = sessions.NewFilesystemStore("/var/lib/php/session",[]byte("travel.com.ua"))

)
func HandlerQauth2(w http.ResponseWriter, r *http.Request) {


	googleOauthConfig.RedirectURL = r.Host +  "/user/GoogleCallback/"
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	//var ctx context.Context = appengine.NewContext(r)
	//client := &http.Client{
	//	Transport: &oauth2.Transport{
	//		Source: google.AppEngineTokenSource(ctx, "scope"),
	//		Base:   &urlfetch.Transport{Context: ctx},
	//	},
	//}
	//resp, _ := client.Get("...")
	//w.Write(resp.Body)
}
//Эти callback было бы неплохо регистрировать в одну общую библиотеку для авторизации
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
		log.Printf("Code exchange failed with '%v'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	fmt.Fprintf(w, "Content: %s\n", contents)
}
type UserRecord struct {
	Id int
	Name string
	Sex int
}
var greetings = [] string {"господин", "госпожа"}

func GetSession(r *http.Request, name string) *sessions.Session {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := store.Get(r, name)
	if err != nil {
		log.Println(err)
		return nil
	}
	return session
}
func IsLogin(r *http.Request) (string, bool) {
	session := GetSession(r, nameSession)
	if session == nil {
		return "", false
	}
	if userID, ok := session.Values["id"]; ok {

		return strconv.Itoa(userID.(int)), ok
	} else {
		return "", ok
	}

}
func deleteCurrentUser(w http.ResponseWriter, r *http.Request) error {
	session := GetSession(r, nameSession )
	delete(session.Values, "id")
	delete(session.Values, "email")
	return session.Save(r, w)

}
func HandlerProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	session := GetSession(r, nameSession )
	email, ok := session.Values["email"]
	if !ok {
		http.Redirect(w,r, "/show/forms/?name=signin", http.StatusSeeOther)
		return
	}
	rows := db.DoQuery("select id, fullname, sex from users where login=?", email )

	var row UserRecord

	defer rows.Close()
	for rows.Next() {

		err := rows.Scan(&row.Id, &row.Name, &row.Sex)

		if err != nil {
			log.Println(err)
			continue
		}
	}

	p := &layouts.MenuOwnerBody{ Title: greetings[row.Sex] + " " + row.Name, TopMenu: make(map[string] *layouts.ItemMenu, 0)}

	var menu db.MenuItems

	menu.GetMenu("menuOwner")

	for _, item := range menu.Items {
		p.TopMenu[item.Title] = &layouts.ItemMenu{ Link: "/menu/" + item.Name + "/"  }

	}
	fmt.Fprint(w, p.MenuOwner() )
}
func HandlerSignIn(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	email := r.Form["login"][0]
	password := r.Form["password"][0]

	rows := db.DoQuery("select id, fullname, sex from users where login=? and hash=?", email, hashPassword(password) )

	defer rows.Close()
	var row UserRecord

	for rows.Next() {

		err := rows.Scan(&row.Id, &row.Name, &row.Sex)

		if err != nil {
			log.Println(err)
			continue
		}

		// session save BEFORE write page
		session := sessions.NewSession(store, nameSession)
		session.Options = &sessions.Options{Path: "/", HttpOnly: true, MaxAge: int(3600)}
		session.Values["id"] = row.Id
		session.Values["email"] = email
		if err := session.Save(r, w); err != nil {
			log.Println(err)
		}

		p := &forms.PersonData{ Id: row.Id, Login: row.Name, Email: email }
		fmt.Fprint(w, p.JSON())
	}
}
func HandlerSignOut(w http.ResponseWriter, r *http.Request) {

	if err := deleteCurrentUser(w, r); err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/show/forms/?name=signin", http.StatusContinue)
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
func generatePassword(email string) (string, error) {

	log.Println(email)
	return GenerateRandomString(16)

}
func hashPassword(password string) interface{} {
	// crypto password
	crc32q := crc32.MakeTable(0xD5828281)
return 	crc32.Checksum([]byte(password), crc32q)
}
func HandlerSignUp(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32000)

	var args [] interface{}
	sql, comma, values := "insert into users (", "", ") values ("

	for key, val := range r.MultipartForm.Value {
		args = append(args, val[0])
		sql += comma + key
		values += comma + "?"
		comma = ","
	}
	email := r.MultipartForm.Value["login"][0]
	password, err := generatePassword(email)
	if err != nil {
		log.Println(err)
	}
	sql += comma + "hash"
	values += comma + "?"

	args = append(args, hashPassword(password) )
	lastInsertId, err := db.DoInsert(sql + values + ")", args... )
	if err != nil {

		fmt.Fprintf(w, "%v", err)
		return
	}
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	mRow := forms.MarshalRow{Msg: "Append row", N: lastInsertId}
	sex, _ := strconv.Atoi(r.MultipartForm.Value["sex"][0])

	if _, err := mail.ParseAddress(email); err !=nil {
		log.Println(err)
		fmt.Fprintf(w, "Что-то неверное с вашей почтой, не смогу отослать письмо! %v", err)
		return
	}
	p := &forms.PersonData{ Id: lastInsertId, Login: r.MultipartForm.Value["fullname"][0], Sex: sex,
		Rows: []forms.MarshalRow{mRow}, Email: email }
	fmt.Fprint(w, p.JSON())

	go sendMail(email, password)
}
func sendMail(email, password string)  {

	m := gomail.NewMessage()
	m.SetHeader("From", "ruslan-bik@yandex.ru")
	m.SetHeader("To", email )
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Регистрация на travel.com.ua!")
	m.SetBody("text/html", mails.InviteEmail(email, password) )
	m.Attach("/home/travel/bootstrap/ico/favicon.png")

	d := gomail.NewDialer("smtp.yandex.ru", 587, "ruslan-bik", "FalconF99")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
	}
}
func HandlerActivateUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm();

	//var args [] interface{}

	//args = append(args, r.Form["email"][0])
	result, _ := db.DoUpdate("update users set active=1 where login=?", r.Form["email"][0])
	fmt.Fprint(w, result)
}



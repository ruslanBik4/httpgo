package admin
import (
	"encoding/base64"
	"net/http"
	"strings"
	"log"
	"fmt"
	"database/sql"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views"
)

const ccApiKey = "SVwaLLaJCUSUV5XPsjmdmiV5WBakh23a7ehCFdrR68pXlT8XBTvh25OO_mUU4_vuWbxsQSW_Ww8zqPG5-w6kCA"
const userDir  = "../store/nav/"
var username = []byte("admin")
var password = []byte("password")

func correctURL(url string) string {

	if strings.HasPrefix(url, "//") {
		return "https:" + url
	}

	return url
}
//func handlerAddPage(w http.ResponseWriter, r *http.Request) {
//
//	err := r.ParseForm()
//	if err != nil {
//		fmt.Fprintf(w, "Error create process! %q", err )
//		return
//
//	}
//	if r.Form["step"][0] != "finished" {
//		log.Println("error status:", r.Form["step"][0])
//		return
//
//	}
//
//	url := correctURL( r.Form["url"][0] )
//
//	resp, err := http.Get( url )
//	if err != nil {
//		fmt.Fprintf(w, "\n Error download file - %s! %v", url, err )
//		log.Println("error:", err)
//		return
//	}
//
//	var response map[string] interface{}
//	err = json.NewDecoder(resp.Body).Decode(&response)
//
//	if err != nil {
//		log.Println("error decode %q from %q", err)
//		log.Println( resp.Body)
//		return
//	}
//
//	var output cloudconvert.StatusOutput
//
//
//	switch typeOutput := response["output"].(type) {
//	case cloudconvert.StatusOutput:
//		output = response["output"].(cloudconvert.StatusOutput)
//	case map[string] interface{}:
//		for key, value := range response["output"].(map[string] interface{}) {
//			switch key {
//			case "url":
//				output.URL = value.(string)
//			case "filename":
//				output.FileName = value.(string)
//			}
//		}
//
//	default:
//		log.Println("Error type output structure %v : %v", typeOutput, response["output"] )
//		return
//	}
//	filename := userDir + output.FileName
//	url = correctURL( output.URL )
//
//	resp, err = http.Get( url )
//	if err != nil {
//		log.Println("Error download file :", err, url)
//		return
//	}
//
//	f, err := os.Create(filename)
//	if err != nil {
//		log.Println("Error create file :", err, filename)
//		return
//	}
//	defer f.Close()
//	if _, err = io.Copy(f, resp.Body); err != nil {
//		log.Println("Error write to file :", err, filename, resp.Body)
//		return
//	}
//}
func basicAuth(w http.ResponseWriter, r *http.Request, user, pass []byte) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	return pair[0] == string(user) && pair[1] == string(pass)
}
func HandlerAdminLists(w http.ResponseWriter, r *http.Request) {

	p := &layouts.MenuOwnerBody{ Title: "Menu admina", TopMenu: make(map[string] *layouts.ItemMenu, 0)}
	var ns db.RecordsTables
	ns.Rows = make([] db.TableOptions, 0)
	ns.GetSelectTablesProp("TABLE_SCHEMA='travel' AND TABLE_NAME like '%_list'" )
	for _, value := range ns.Rows {
		p.TopMenu[value.TABLE_COMMENT] = &layouts.ItemMenu{ Link: "/admin/table/" + value.TABLE_NAME + "/"}

	}
	fmt.Fprint(w, p.MenuOwner() )

}
func HandlerAdmin(w http.ResponseWriter, r *http.Request) {

	// pass from global variables
	if basicAuth(w, r, username, password) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		p := &pages.AdminPageBody{ Name: username, Pass : password, Content : "", Catalog: make(map[string] *pages.ItemMenu) }
		var menu db.MenuItems

		if menu.GetMenu("admin") > 0 {


			for _, item := range menu.Items {
				p.Catalog[item.Title] = &pages.ItemMenu{ Link: "/menu/" + item.Name + "/" }

			}
		}
		if menu.Self.Link > ""  {
			p.Content = fmt.Sprintf("<div class='autoload' data-href='%s'></div>", menu.Self.Link )
		}

		if err := views.RenderTemplate(w, r, "adminPage", p); err != nil {
			log.Println(err)
			return
		}
		return
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="Beware! Protected REALM! "`)
	w.WriteHeader(401)
	w.Write([]byte("401 Unauthorized\n"))
}
func HandlerAdminTable (w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tableName := r.URL.Path[ len("/admin/table/") : len(r.URL.Path)-1]
	fmt.Fprint(w, tableName )

	p := &layouts.MenuOwnerBody{ Title: tableName, TopMenu: make(map[string] *layouts.ItemMenu, 0)}

	p.TopMenu["Добавить"] = &layouts.ItemMenu{ Link: "/admin/row/new/" + tableName + "/" }

	fmt.Fprint(w, p.MenuOwner() )

	fields := getFields(tableName)
	rows := db.DoQuery("select * from " + tableName)

	defer rows.Close()
	fmt.Fprint(w, pages.ShowTable(tableName, fields, rows) )


}
func getFields(tableName string) (fields forms.FieldsTable){

	var ns db.FieldsTable
	ns.Rows = make([] db.FieldStructure, 0)
	ns.GetColumnsProp(tableName)

	fields.Rows = make([] forms.FieldStructure, 0)
	fields.Name = tableName

	fields.PutDataFrom(ns)

	return fields

}
func HandlerNewRecord(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tableName := r.URL.Path[ len("/admin/row/new/") : len(r.URL.Path)-1]

	fields := getFields(tableName)
	fmt.Fprint(w, fields.ShowAnyForm("/admin/row/add/", "Новая запись в таблицу " + tableName) )
}
func HandlerEditRecord(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tableName := r.FormValue("table")
	id        := r.FormValue("id")

	fields := getFields(tableName)
	rows := db.DoQuery("select * from " + tableName + " where id=?", id)

	defer rows.Close()
	var row [] interface {}

	columns, err := rows.Columns()
	if (err != nil) {
		log.Println(err)
	}
	rowField := make( [] *sql.NullString, len(columns) )
	for idx, _ := range columns {

		rowField[idx] = new(sql.NullString)
		row = append( row, rowField[idx] )
	}

	for rows.Next() {


		if err := rows.Scan(row...); err != nil {
			log.Println(err)
		}
		for idx, field := range rowField {
			if field.Valid {
				fields.Rows[idx].Value = field.String
			}
		}

	}

	fmt.Fprint(w, fields.ShowAnyForm("/admin/row/update/", "Меняем запись №" + id + " в таблице " + tableName) )

}
func HandlerAddRecord(w http.ResponseWriter, r *http.Request)  {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if result, err := db.DoInsertFromForm(r); err != nil {
		log.Println(err)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
	} else {
		fmt.Fprintf(w, "{\"result\":\"%d\", \"contentURL\":\"/admin/table/%s/\"}", result, r.FormValue("table"))
	}
}
func HandlerUpdateRecord(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if result, err := db.DoUpdateFromForm(r); err != nil {
		log.Println(err)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
	} else {
		fmt.Fprintf(w, "{\"result\":\"%d\", \"contentURL\":\"/admin/table/%s/\"}", result, r.FormValue("table"))
	}

}
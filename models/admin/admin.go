package admin
import (
	"encoding/base64"
	"net/http"
	"strings"
	"log"
	"fmt"
	"database/sql"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"strconv"
	"github.com/ruslanBik4/httpgo/models/users"
	"github.com/ruslanBik4/httpgo/views/templates/tables"
	_ "github.com/ruslanBik4/httpgo/models/system"
)

const ccApiKey = "SVwaLLaJCUSUV5XPsjmdmiV5WBakh23a7ehCFdrR68pXlT8XBTvh25OO_mUU4_vuWbxsQSW_Ww8zqPG5-w6kCA"
const nameSession = "PHPSESSID"
var store = users.Store

type UserRecord struct {
	Id int
	Name string
	Sex int
}

func correctURL(url string) string {

	if strings.HasPrefix(url, "//") {
		return "https:" + url
	}

	return url
}
func HandlerUMUTables(w http.ResponseWriter, r *http.Request) {

	p := &layouts.MenuOwnerBody{ Title: "Menu admina", TopMenu: make(map[string] *layouts.ItemMenu, 0)}
	var ns db.RecordsTables
	ns.Rows = make([] db.TableOptions, 0)
	ns.GetSelectTablesProp("TABLE_SCHEMA='travel' AND TABLE_NAME in ('users', 'object', 'business') " )
	for _, value := range ns.Rows {
		p.TopMenu[value.TABLE_COMMENT] = &layouts.ItemMenu{ Link: "/admin/table/" + value.TABLE_NAME + "/"}

	}
	if views.IsAJAXRequest(r) {
		fmt.Fprint(w, p.MenuOwner() )
	} else {
		HandlerAdmin(w, r)
	}
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
func basicAuth(w http.ResponseWriter, r *http.Request) (bool, []byte, []byte) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false, nil, nil
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false, nil, nil
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false, nil, nil
	}

	err, userId, userName := users.CheckUserCredentials(pair[0], pair[1])

	if err != nil {
		return false, nil, nil
	}

	// session save BEFORE write page
	users.SaveSession(w, r, userId, pair[0])

	return true, []byte( userName), []byte (pair[1])
}

func HandlerAdminLists(w http.ResponseWriter, r *http.Request) {

	p := &layouts.MenuOwnerBody{ Title: "Menu admina", TopMenu: make(map[string] *layouts.ItemMenu, 0)}
	var ns db.RecordsTables
	ns.Rows = make([] db.TableOptions, 0)
	ns.GetSelectTablesProp("TABLE_SCHEMA='travel' AND TABLE_NAME like '%_list'" )
	for _, value := range ns.Rows {
		p.TopMenu[value.TABLE_COMMENT] = &layouts.ItemMenu{ Link: "/admin/table/" + value.TABLE_NAME + "/"}

	}
	if views.IsAJAXRequest(r) {
		fmt.Fprint(w, p.MenuOwner() )
	} else {
		HandlerAdmin(w, r)
	}
}
func HandlerAdmin(w http.ResponseWriter, r *http.Request) {

	// pass from global variables
	result, username, password := basicAuth(w, r)
	if  result {
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

		p.TopMenu = make( map[string] string, 0)
		menu.GetMenu("indexTop")

		for _, item := range menu.Items {
			p.TopMenu[item.Title] = "/menu/" + item.Name + "/"

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

	tableName := strings.Trim( r.URL.Path[ len("/admin/table/") : ], "/" )

	var menu db.MenuItems

	p := &layouts.MenuOwnerBody{ Title: tableName, TopMenu: make(map[string] *layouts.ItemMenu, 0)}
	if menu.GetMenu(tableName) > 0 {

		for _, item := range menu.Items {
			p.TopMenu[item.Title] = &layouts.ItemMenu{ Link: "/menu/" + item.Name + "/" }

		}

		// return into parent menu if he occurent
		if menu.Self.ParentID > 0 {
			p.TopMenu["< на уровень выше"] = &layouts.ItemMenu{ Link: fmt.Sprintf("/menu/%d/", menu.Self.ParentID ) }
		}
	} else {

		p.TopMenu["Добавить"] = &layouts.ItemMenu{Link: "/admin/row/new/" + tableName + "/" }
	}

	fmt.Fprint(w, p.MenuOwner() )

	var tableOpt db.TableOptions
	tableOpt.GetTableProp(tableName)

	fields := GetFields(tableName)

	fmt.Fprint(w, fields.Comment )

	sqlCommand := "select * from " + tableName

	var query tables.QueryStruct

	query.Order = r.FormValue("order")
	if query.Order > "" {
		sqlCommand += " order by " + query.Order
	}
	rows, err := db.DoSelect(sqlCommand)
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Error during run query %s", sqlCommand)
		return
	}

	defer rows.Close()

	query.Rows = rows
	query.Href = "/admin/table/" + tableName
	query.HrefEdit = "/admin/row/edit/?table=" + tableName + "&id="
	query.Tables = append(query.Tables, &fields)

	fmt.Fprintf(w, `<script src="/%s.js"></script>`, tableName )
	fmt.Fprint(w, query.RenderTable() )


}
func HandlerSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var fields forms.FieldsTable

	tableName := r.FormValue("table")
	id        := r.FormValue("id")
	if tableName > "" {

		fields = GetFields(tableName)
		//fmt.Fprint(w, fields.ShowAnyForm("/admin/row/update/", "Меняем запись №" + id + " в таблице " + tableName) )
	} else {
		r.ParseForm()
		for key, val := range r.Form {
			if len(val) > 1 {
				tableName = key[: strings.Index(key, "[")]

				var ns db.FieldsTable
				ns.GetColumnsProp(tableName)

				//fields.Name = tableName
				for _, name := range val {
					for _, field := range ns.Rows {

						if field.COLUMN_NAME == name {

							fieldStrc := &forms.FieldStructure{
								COLUMN_NAME: field.COLUMN_NAME,
								DATA_TYPE:   field.DATA_TYPE,
								IS_NULLABLE: field.IS_NULLABLE,
								COLUMN_TYPE: field.COLUMN_TYPE,
								TableName:   tableName,
							}
							if field.CHARACTER_SET_NAME.Valid {
								fieldStrc.CHARACTER_SET_NAME = field.CHARACTER_SET_NAME.String
							}
							if field.COLUMN_COMMENT.Valid {
								fieldStrc.COLUMN_COMMENT = field.COLUMN_COMMENT.String
							}
							if field.CHARACTER_MAXIMUM_LENGTH.Valid {
								fieldStrc.CHARACTER_MAXIMUM_LENGTH = int(field.CHARACTER_MAXIMUM_LENGTH.Int64)
							}
							if field.COLUMN_DEFAULT.Valid {
								fieldStrc.COLUMN_DEFAULT = field.COLUMN_DEFAULT.String
							}
							//fieldStrc.Value = value
							fields.Rows = append(fields.Rows, *fieldStrc)
							break
						}
					}

				}
			}
		}
	}
	fmt.Fprint(w, fields.ShowAnyForm("/admin/exec/", "Меняем запись №" + id + " в таблице " + tableName) )

}
func GetFields(tableName string) (fields forms.FieldsTable){

	var ns db.FieldsTable
	ns.Options.GetTableProp(tableName)
	//ns.Rows = make([] db.FieldStructure, 0)
	ns.GetColumnsProp(tableName)

	//fields.Rows = make([] forms.FieldStructure, 0)
	fields.Name = tableName
	//fields.Comment = ns.Options.TABLE_COMMENT

	fields.PutDataFrom(ns)

	return fields

}
func HandlerNewRecord(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tableName := r.URL.Path[ len("/admin/row/new/") : len(r.URL.Path)-1]

	fields := GetFields(tableName)
	fmt.Fprint(w, fields.ShowAnyForm("/admin/row/add/", "Новая запись в таблицу " + tableName) )
}
func GetRecord(tableName, id string) (fields forms.FieldsTable, err error) {


	fields = GetFields(tableName)
	rows, err := db.DoSelect("select * from "+tableName+" where id=?", id)
	if (err != nil) {
		log.Println(err)
		return fields, err
	}

	defer rows.Close()
	var row [] interface{}

	columns, err := rows.Columns()
	if (err != nil) {
		log.Println(err)
	}
	rowField := make([] *sql.NullString, len(columns))
	for idx, _ := range columns {

		rowField[idx] = new(sql.NullString)
		row = append(row, rowField[idx])
	}

	for rows.Next() {

		if err := rows.Scan(row...); err != nil {
			log.Println(err)
		}
		for idx, field := range rowField {
			if field.Valid {
				fields.Rows[idx].Value = field.String
				if fields.Rows[idx].COLUMN_NAME == "id" {
					fields.ID, _ = strconv.Atoi(id)
				}
			}
		}

	}

	return fields, nil
}
// удаление записи - помечаем специальное поле isDel
func HandlerDeleteRecord(w http.ResponseWriter, r *http.Request) {

	tableName := r.FormValue("table")
	id	  := r.FormValue("id")

	if tableName == "" {
		fmt.Fprint(w, "Not table name!")
		return
	}
	if _, err := db.DoUpdate("update " + tableName + " set isDel=1 where id=?", id); err != nil {
		log.Println(err)
	} else {
		fmt.Fprint(w, "Успешно удалили запись с номером " + id)
	}
}
func HandlerShowRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if fields, err := GetRecord(tableName, id); err != nil {
		fmt.Fprint(w, "Error during reading record with id=%s", id)

	} else {
		fmt.Fprint(w, fields.ShowRecord( "Меняем запись №"+id+" в таблице " + fields.Comment ) )
	}


}
func HandlerEditRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if fields, err := GetRecord(tableName, id); err != nil {
		fmt.Fprint(w, "Error during reading record with id=%s", id)

	} else {
		fmt.Fprintf(w, `<script src="/%s.js"></script>`, tableName )
		views.RenderAnyForm(w, r, "Меняем запись №"+id+" в таблице " + fields.Comment, fields, nil, "", "")
	}

}
func checkUserLogin(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID := users.IsLogin(r)

	return userID, true

}
func HandlerRecord(w http.ResponseWriter, r *http.Request, operation string)  {

	var arrJSON map[string] interface {}
	arrJSON = make( map[string] interface {}, 0)

	userID, ok := checkUserLogin(w, r)
	if !ok {
		arrJSON["error"] = "true"
		arrJSON["message"] = fmt.Sprintf("%s", users.NOT_AUTHORIZE)
	} else {
		var err error
		var id int

		if operation == "id" {
			id, err = db.DoInsertFromForm(r, userID)
		} else {
			id, err = db.DoUpdateFromForm(r, userID)
		}

		tableName := r.FormValue("table")
		if err != nil {
			log.Println(err)
			arrJSON["error"] = "true"
			arrJSON["message"] = fmt.Sprintf("Error %v during uodate table '%s' ", err, tableName)
		} else {
			arrJSON[operation] = id
			arrJSON["contentURL"] = fmt.Sprintf("/admin/table/%s/", tableName )
		}
	}
	views.RenderAnyJSON(w, arrJSON)
}
func HandlerAddRecord(w http.ResponseWriter, r *http.Request) {

	HandlerRecord(w, r, "id" )
}
func HandlerUpdateRecord(w http.ResponseWriter, r *http.Request)  {

	HandlerRecord(w, r, "rowAffected" )

}
func HandlerExec(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.FormValue("table") > "" {
		if r.FormValue("id") > "" {
			HandlerUpdateRecord(w, r)
		} else {
			HandlerAddRecord(w, r)
		}
	} else {
		var params db.MultiQuery
		var arrJSON map[string] interface {}

		arrJSON = make( map[string] interface {}, 0)

		params.Queryes = make(map[string] *db.ArgsQuery, 0)
		for key, val := range r.Form {

			indSeparator := strings.Index(key, ":")
			if indSeparator < 1 {
				log.Printf("Error in name field %s^ don't write to DB!", key)
				arrJSON["error"]   = "true"
				arrJSON["message"] = fmt.Sprintf("Error in name field %s^ don't write to DB!", key)
				continue
			}

			tableName := key[: indSeparator ]

			query, ok := params.Queryes[tableName]
			if !ok {
				query =  &db.ArgsQuery{
					Comma: "",
					SQLCommand: "insert into " + tableName + "(",
					Values: "values (",
				}
			}
			fieldName := key[ strings.Index(key, ":")+1 : ]

			if strings.Contains(fieldName, "[]") {
				query.SQLCommand += query.Comma + "`" + strings.TrimRight(fieldName, "[]") + "`"
				str, comma := "", ""
				for _, value := range val {
					str += comma + value
					comma = ","
				}
				query.Args = append(query.Args, str)
			} else if strings.Contains(fieldName, "[")  {
				log.Println(fieldName)
				pos := strings.Index(fieldName, "[")
				//number := fieldName[ pos+1 : strings.Index(fieldName, "]") ]
				fieldName = "`" + fieldName[ :pos] + "`"

				// пока беда в том, что количество должно точно соответствовать!
				//если первый  - то создаем новый список параметров для вставки
				if strings.HasPrefix(query.SQLCommand, "insert into " + tableName + "(" + fieldName) {
					query.Comma = "), ("
					//args, ok := query.args[0][fieldName]
				} else if !strings.Contains(query.SQLCommand, fieldName )  {
					query.SQLCommand += query.Comma + fieldName
					//query.args = append(query.args, make(map[string] string, 0))
				}

				log.Println(fieldName)
				query.Args = append(query.Args, val[0])

			} else {
				query.SQLCommand += query.Comma + "`" + fieldName + "`"
				query.Args = append( query.Args, val[0] )
			}
			query.Values += query.Comma + "?"
			query.Comma = ", "
			params.Queryes[tableName] = query
		}

		primaryTable := ""
		primaryID := 0

		for key, query := range params.Queryes {
			if primaryTable == "" {
				primaryTable = key
			} else {
				query.SQLCommand += query.Comma + "`id_" + primaryTable + "`"
				query.Args = append( query.Args, primaryID )
				query.Values += query.Comma + "?"
			}

			id, err := db.DoInsert(query.SQLCommand + ") " + query.Values + ")", query.Args ... )
			if err != nil {
				log.Println(err)
				arrJSON["error"] = "true"
				arrJSON["message"] = fmt.Sprintf("Error during insert into %s ", key)
				break
			} else {
				arrJSON["message"] = fmt.Sprintf("insert into %s record #%d", key, id)
				arrJSON["id"]      = id

				 if primaryID == 0 {
					 primaryID = id
				 }
			}
		}
		arrJSON["contentURL"] = fmt.Sprintf("/admin/table/%s/", primaryTable)
		views.RenderAnyJSON(w, arrJSON)
	}
}
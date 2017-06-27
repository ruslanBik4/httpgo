package admin

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/logs"
	_ "github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/models/users"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/httpgo/views/templates/tables"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
)

const ccApiKey = "SVwaLLaJCUSUV5XPsjmdmiV5WBakh23a7ehCFdrR68pXlT8XBTvh25OO_mUU4_vuWbxsQSW_Ww8zqPG5-w6kCA"
const nameSession = "PHPSESSID"

var store = users.Store

type UserRecord struct {
	Id   int
	Name string
	Sex  int
}

func correctURL(url string) string {

	if strings.HasPrefix(url, "//") {
		return "https:" + url
	}

	return url
}
func HandlerUMUTables(w http.ResponseWriter, r *http.Request) {

	/*userID  := users.IsLogin(r)
	resultId,_ := strconv.Atoi(userID)
	if resultId > 0 {
		if !GetUserPermissionForPageByUserId(resultId, r.URL.Path, "View") {
			views.RenderNoPermissionPage(w, r)
		}
	} else {
		views.RenderNoPermissionPage(w, r)
	}*/

	p := &layouts.MenuOwnerBody{Title: "Menu admina", TopMenu: make(map[string]*layouts.ItemMenu, 0)}
	var ns db.RecordsTables
	ns.Rows = make([]db.TableOptions, 0)
	ns.GetSelectTablesProp("TABLE_SCHEMA=? AND TABLE_NAME in (?, ?, ?) ", "travel", "users", "object", "business" )
	for _, value := range ns.Rows {
		p.TopMenu[value.TABLE_COMMENT] = &layouts.ItemMenu{Link: "/admin/table/" + value.TABLE_NAME + "/"}

	}
	if views.IsAJAXRequest(r) {
		fmt.Fprint(w, p.MenuOwner())
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
func basicAuth(w http.ResponseWriter, r *http.Request) (bool, []byte, []byte, int) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false, nil, nil, 0
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false, nil, nil, 0
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false, nil, nil, 0
	}

	err, userId, userName := users.CheckUserCredentials(pair[0], pair[1])

	if err != nil {
		return false, nil, nil, 0
	}

	// session save BEFORE write page
	users.SaveSession(w, r, userId, pair[0])

	return true, []byte(userName), []byte(pair[1]), userId
}

func HandlerAdminLists(w http.ResponseWriter, r *http.Request) {

	userID := users.IsLogin(r)
	resultId, err := strconv.Atoi(userID)
	if err != nil || !CheckAdminPermissions(resultId) {
		views.RenderNoPermissionPage(w)
		return
	}
	p := &layouts.MenuOwnerBody{Title: "Menu admina", TopMenu: make(map[string]*layouts.ItemMenu, 0)}
	var ns db.RecordsTables
	ns.Rows = make([]db.TableOptions, 0)
	ns.GetSelectTablesProp("TABLE_SCHEMA=? AND (RIGHT(table_name, 5) = ?)", "travel", "%_list")
	for _, value := range ns.Rows {
		p.TopMenu[value.TABLE_COMMENT] = &layouts.ItemMenu{Link: "/admin/table/" + value.TABLE_NAME + "/"}

	}
	if views.IsAJAXRequest(r) {
		fmt.Fprint(w, p.MenuOwner())
	} else {
		HandlerAdmin(w, r)
	}
}
func HandlerAdmin(w http.ResponseWriter, r *http.Request) {

	// pass from global variables
	result, username, password, userId := basicAuth(w, r)
	if result {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		headPage := &layouts.HeadHTMLPage{
			Charset:  "charset=utf-8",
			Language: "ru",
			Title:    "Заголовок новой страницы",
		}
		p := &pages.AdminPageBody{Name: username, Pass: password, Content: "", Catalog: make(map[string]*pages.ItemMenu), Head: headPage}
		var menu db.MenuItems

		if menu.GetMenuByUserId(userId) > 0 {

			for _, item := range menu.Items {
				p.Catalog[item.Title] = &pages.ItemMenu{Link: "/menu/" + item.Name + "/"}

			}
		}
		if menu.Self.Link > "" {
			p.Content = fmt.Sprintf("<div class='autoload' data-href='%s'></div>", menu.Self.Link)
		}

		p.TopMenu = make(map[string]string, 0)
		menu.GetMenu("indexTop")

		for _, item := range menu.Items {
			p.TopMenu[item.Title] = "/menu/" + item.Name + "/"

		}
		if err := views.RenderTemplate(w, r, "adminPage", p); err != nil {
			logs.ErrorLog(err)
			return
		}
		return
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="Beware! Protected REALM! "`)
	w.WriteHeader(401)
	w.Write([]byte("401 Unauthorized\n"))
}
func HandlerAdminTable(w http.ResponseWriter, r *http.Request) {

	/*userID  := users.IsLogin(r)
	resultId,_ := strconv.Atoi(userID)
	if resultId > 0 {
		if !GetUserPermissionForPageByUserId(resultId, r.URL.Path, "View") {
			views.RenderNoPermissionPage(w, r)
		}
	} else {
		views.RenderNoPermissionPage(w, r)
	}*/

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tableName := strings.Trim(r.URL.Path[len("/admin/table/"):], "/")

	var menu db.MenuItems

	p := &layouts.MenuOwnerBody{Title: tableName, TopMenu: make(map[string]*layouts.ItemMenu, 0)}
	if menu.GetMenu(tableName) > 0 {

		for _, item := range menu.Items {
			p.TopMenu[item.Title] = &layouts.ItemMenu{Link: "/menu/" + item.Name + "/"}

		}

		// return into parent menu if he occurent
		if menu.Self.ParentID > 0 {
			p.TopMenu["< на уровень выше"] = &layouts.ItemMenu{Link: fmt.Sprintf("/menu/%d/", menu.Self.ParentID)}
		}
	} else {

		p.TopMenu["Добавить"] = &layouts.ItemMenu{Link: "/admin/row/new/" + tableName + "/"}
	}

	w.Write([]byte(p.MenuOwner()))

	var tableOpt db.TableOptions
	tableOpt.GetTableProp(tableName)

	fields := GetFields(tableName)

	w.Write([]byte(fields.Comment))

	sqlCommand := "select * from " + tableName

	var query tables.QueryStruct

	query.Order = r.FormValue("order")
	if query.Order > "" {
		sqlCommand += " order by " + query.Order
	}
	rows, err := db.DoSelect(sqlCommand)
	if err != nil {
		logs.ErrorLog(err)
		fmt.Fprintf(w, "Error during run query %s", sqlCommand)
		return
	}

	defer rows.Close()

	query.Rows = rows
	query.Href = "/admin/table/" + tableName
	query.HrefEdit = "/admin/row/edit/?table=" + tableName + "&id="
	query.Tables = append(query.Tables, &fields)

	fmt.Fprintf(w, `<script src="/%s.js"></script>`, tableName)
	w.Write([]byte(query.RenderTable()))

}
func HandlerSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var fields forms.FieldsTable

	tableName := r.FormValue("table")
	id := r.FormValue("id")
	if tableName > "" {

		fields = GetFields(tableName)
		//fmt.Fprint(w, fields.ShowAnyForm("/admin/row/update/", "Меняем запись №" + id + " в таблице " + tableName) )
	} else {
		r.ParseForm()
		for key, val := range r.Form {
			if len(val) > 1 {
				tableName = key[:strings.Index(key, "[")]

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
	fmt.Fprint(w, fields.ShowAnyForm("/admin/exec/", "Меняем запись №"+id+" в таблице "+tableName))

}
func GetFields(tableName string) (fields forms.FieldsTable) {

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

	tableName := r.URL.Path[len("/admin/row/new/") : len(r.URL.Path)-1]

	fields := GetFields(tableName)
	fmt.Fprint(w, fields.ShowAnyForm("/admin/row/add/", "Новая запись в таблицу "+tableName))
}
func GetRecord(tableName, id string) (fields forms.FieldsTable, err error) {

	//TODO научить данную функцию получать значения из полей типа tableid_ setid_ nodeid_
	fields = GetFields(tableName)
	rows, err := db.DoSelect("select * from "+tableName+" where id=?", id)
	if err != nil {

		logs.ErrorLog(err, "select * from ", tableName, " where id=?", id)
		return fields, err
	}

	defer rows.Close()
	var row []interface{}

	columns, err := rows.Columns()
	if err != nil {
		logs.ErrorLog(err)
	}
	rowField := make([]*sql.NullString, len(columns))
	for idx, _ := range columns {

		rowField[idx] = new(sql.NullString)
		row = append(row, rowField[idx])
	}

	for rows.Next() {

		if err := rows.Scan(row...); err != nil {
			logs.ErrorLog(err)
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
	id := r.FormValue("id")

	if tableName == "" {
		fmt.Fprint(w, "Not table name!")
		return
	}
	if _, err := db.DoUpdate("update "+tableName+" set isDel=1 where id=?", id); err != nil {
		logs.ErrorLog(err)
	} else {
		fmt.Fprint(w, "Успешно удалили запись с номером "+id)
	}
}
func HandlerShowRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if fields, err := GetRecord(tableName, id); err != nil {
		fmt.Fprint(w, "Error during reading record with id=%s", id)

	} else {
		fmt.Fprint(w, fields.ShowRecord("Меняем запись №"+id+" в таблице "+fields.Comment))
	}

}
func HandlerEditRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if fields, err := GetRecord(tableName, id); err != nil {
		fmt.Fprint(w, "Error during reading record with id=%s", id)

	} else {
		fmt.Fprintf(w, `<script src="/%s.js"></script>`, tableName)
		views.RenderAnyForm(w, r, "Меняем запись №"+id+" в таблице "+fields.Comment, fields, nil, "", "")
	}

}
func checkUserLogin(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID := users.IsLogin(r)

	return userID, true

}
func HandlerRecord(w http.ResponseWriter, r *http.Request, operation string) {

	var arrJSON map[string]interface{}
	arrJSON = make(map[string]interface{}, 0)

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
			logs.ErrorLog(err)
			arrJSON["error"] = "true"
			arrJSON["message"] = fmt.Sprintf("Error %v during uodate table '%s' ", err, tableName)
		} else {
			arrJSON[operation] = id
			arrJSON["contentURL"] = fmt.Sprintf("/admin/table/%s/", tableName)
		}
	}
	views.RenderAnyJSON(w, arrJSON)
}
func HandlerAddRecord(w http.ResponseWriter, r *http.Request) {
	HandlerRecord(w, r, "id")
}
func HandlerUpdateRecord(w http.ResponseWriter, r *http.Request) {
	HandlerRecord(w, r, "rowAffected")

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
		var arrJSON map[string]interface{}

		arrJSON = make(map[string]interface{}, 0)

		params.Queryes = make(map[string]*db.ArgsQuery, 0)
		for key, val := range r.Form {

			indSeparator := strings.Index(key, ":")
			if indSeparator < 1 {
				message := "Error in name   don't write to DB. Field="
				logs.ErrorLog(errors.New(message), key)
				arrJSON["error"] = "true"
				arrJSON["message"] = fmt.Sprintf(message+"%s^", key)
				continue
			}

			tableName := key[:indSeparator]

			query, ok := params.Queryes[tableName]
			if !ok {
				query = &db.ArgsQuery{
					Comma:     "",
					FieldList: "insert into " + tableName + "(",
					Values:    "values (",
				}
			}
			fieldName := key[strings.Index(key, ":")+1:]

			if strings.Contains(fieldName, "[]") {
				query.FieldList += query.Comma + "`" + strings.TrimRight(fieldName, "[]") + "`"
				str, comma := "", ""
				for _, value := range val {
					str += comma + value
					comma = ","
				}
				query.Args = append(query.Args, str)
			} else if strings.Contains(fieldName, "[") {
				logs.DebugLog("fieldName=", fieldName)
				pos := strings.Index(fieldName, "[")
				//number := fieldName[ pos+1 : strings.Index(fieldName, "]") ]
				fieldName = "`" + fieldName[:pos] + "`"

				// пока беда в том, что количество должно точно соответствовать!
				//если первый  - то создаем новый список параметров для вставки
				if strings.HasPrefix(query.FieldList, "insert into "+tableName+"("+fieldName) {
					query.Comma = "), ("
					//args, ok := query.args[0][fieldName]
				} else if !strings.Contains(query.FieldList, fieldName) {
					query.FieldList += query.Comma + fieldName
					//query.args = append(query.args, make(map[string] string, 0))
				}

				logs.DebugLog("fieldName=", fieldName)
				query.Args = append(query.Args, val[0])

			} else {
				query.FieldList += query.Comma + "`" + fieldName + "`"
				query.Args = append(query.Args, val[0])
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
				query.FieldList += query.Comma + "`id_" + primaryTable + "`"
				query.Args = append(query.Args, primaryID)
				query.Values += query.Comma + "?"
			}

			id, err := db.DoInsert(query.FieldList+") "+query.Values+")", query.Args...)
			if err != nil {
				logs.ErrorLog(err)
				arrJSON["error"] = "true"
				arrJSON["message"] = fmt.Sprintf("Error during insert into %s ", key)
				break
			} else {
				arrJSON["message"] = fmt.Sprintf("insert into %s record #%d", key, id)
				arrJSON["id"] = id

				if primaryID == 0 {
					primaryID = id
				}
			}
		}
		arrJSON["contentURL"] = fmt.Sprintf("/admin/table/%s/", primaryTable)
		views.RenderAnyJSON(w, arrJSON)
	}
}

//проверка прав пользователя на доступ по url с учётом проверки на права администратора (доступны все области)
func GetUserPermissionForPageByUserId(userId int, url, action string) bool {

	var rows *sql.Rows
	var err error
	switch action {
	case "Create":
		rows, err = db.DoSelect("SELECT menu_items.`id` "+
			"FROM users_roles_list_has "+
			"LEFT JOIN roles_permission_list ON `roles_permission_list`.`id_roles_list`=users_roles_list_has.id_roles_list "+
			"INNER JOIN roles_list ON users_roles_list_has.`id_roles_list`=`roles_list`.id "+
			"INNER JOIN `menu_items` ON `roles_permission_list`.`id_menu_items` = menu_items.`id` "+
			"WHERE users_roles_list_has.id_users=? AND (menu_items.`name`=? OR menu_items.`link`=? OR roles_list.`is_general`=1) AND roles_permission_list.`allow_create`=1", userId, getMenuNameFromUrl(url), url)

	case "Edit":
		rows, err = db.DoSelect("SELECT menu_items.`id` "+
			"FROM users_roles_list_has "+
			"LEFT JOIN roles_permission_list ON `roles_permission_list`.`id_roles_list`=users_roles_list_has.id_roles_list "+
			"INNER JOIN roles_list ON users_roles_list_has.`id_roles_list`=`roles_list`.id "+
			"INNER JOIN `menu_items` ON `roles_permission_list`.`id_menu_items` = menu_items.`id` "+
			"WHERE users_roles_list_has.id_users=? AND (menu_items.`name`=? OR menu_items.`link`=? OR roles_list.`is_general`=1) AND roles_permission_list.`allow_edit`=1", userId, getMenuNameFromUrl(url), url)

	case "Delete":
		rows, err = db.DoSelect("SELECT menu_items.`id` "+
			"FROM users_roles_list_has "+
			"LEFT JOIN roles_permission_list ON `roles_permission_list`.`id_roles_list`=users_roles_list_has.id_roles_list "+
			"INNER JOIN roles_list ON users_roles_list_has.`id_roles_list`=`roles_list`.id "+
			"INNER JOIN `menu_items` ON `roles_permission_list`.`id_menu_items` = menu_items.`id` "+
			"WHERE users_roles_list_has.id_users=? AND (menu_items.`name`=? OR menu_items.`link`=? OR roles_list.`is_general`=1) AND roles_permission_list.`allow_delete`=1", userId, getMenuNameFromUrl(url), url)

	default:
		rows, err = db.DoSelect("SELECT menu_items.`id` "+
			"FROM users_roles_list_has "+
			"LEFT JOIN roles_permission_list ON `roles_permission_list`.`id_roles_list`=users_roles_list_has.id_roles_list "+
			"INNER JOIN roles_list ON users_roles_list_has.`id_roles_list`=`roles_list`.id "+
			"INNER JOIN `menu_items` ON `roles_permission_list`.`id_menu_items` = menu_items.`id` "+
			"WHERE users_roles_list_has.id_users=? AND (menu_items.`name`=? OR menu_items.`link`=? OR roles_list.`is_general`=1)", userId, getMenuNameFromUrl(url), url)

	}

	if err != nil {
		logs.ErrorLog(err)
		return false
	}

	defer rows.Close()
	for rows.Next() {
		return true
	}

	return false
}

//проверка пользователя на права администратора в екстранете
func CheckAdminPermissions(userId int) bool {

	rows, err := db.DoSelect("SELECT users_roles_list_has.`id_users` "+
		"FROM users_roles_list_has "+
		"LEFT JOIN roles_list ON `roles_list`.`id`=users_roles_list_has.id_roles_list "+
		"WHERE users_roles_list_has.id_users=? AND roles_list.is_general=1", userId)

	if err != nil {
		logs.ErrorLog(err)
		return false
	}

	defer rows.Close()
	for rows.Next() {
		return true
	}

	return false
}

//получить действие по урл (для таблиц)
func getActionByUrl(url string) string {
	urlParts := strings.Split(url, ",")

	var result string
	switch urlParts[2] {
	case "new":
		result = "Create"
	case "edit":
		result = "Edit"
	case "del":
		result = "Delete"
	default:
		result = "Undefined action"
	}
	return result
}

//получить имя пункта меню исходя из урл
func getMenuNameFromUrl(url string) string {
	urlParts := strings.Split(url, "/")

	return urlParts[2]
}

func HandlerSignUpAnotherUser(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32000)

	var args []interface{}
	sql, comma, values := "insert into users (", "", ") values ("

	for key, val := range r.MultipartForm.Value {
		args = append(args, val[0])
		sql += comma + key
		values += comma + "?"
		comma = ","
	}
	email := r.MultipartForm.Value["login"][0]
	password, err := users.GeneratePassword(email)
	if err != nil {
		logs.ErrorLog(err)
	}
	sql += comma + "hash"
	values += comma + "?"

	args = append(args, users.HashPassword(password))
	lastInsertId, err := db.DoInsert(sql+values+")", args...)
	if err != nil {

		fmt.Fprintf(w, "%v", err)
		return
	}
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	mRow := forms.MarshalRow{Msg: "Append row", N: lastInsertId}
	sex, _ := strconv.Atoi(r.MultipartForm.Value["sex"][0])

	if _, err := mail.ParseAddress(email); err != nil {
		logs.ErrorLog(err)
		fmt.Fprintf(w, "Что-то неверное с вашей почтой, не смогу отослать письмо! %v", err)
		return
	}
	p := &forms.PersonData{Id: lastInsertId, Login: r.MultipartForm.Value["fullname"][0], Sex: sex,
		Rows: []forms.MarshalRow{mRow}, Email: email}
	fmt.Fprint(w, p.JSON())

	go users.SendMail(email, password)
}

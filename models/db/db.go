package db

import (
	"net/http"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"strconv"
)
type
	argsRAW  [] interface {}

func GetParentFieldName(tableName string) (name string) {
	var listNs FieldsTable

	if err := listNs.GetColumnsProp(tableName); err != nil {
		return ""
	}
	for _, list := range listNs.Rows {
		switch list.COLUMN_NAME {
		case "name":
			name = "name"
		case "title":
			name = "title"
		case "fullname":
			name = "fullname"
		}
	}

	return name

}

//TODO: добавить запись для мультиполей (setid_)
func insertMultiSet(tableName, key string, values []string, id int) {

	tableProps := strings.TrimLeft(key, "setid_")


	sqlCommand := fmt.Sprintf("insert IGNORE into %s_%s_has (id_%[1]s, id_%[2]s) values (%d, ?)",
		tableName, tableProps, id)
	smtp, err := prepareQuery(sqlCommand)
	if err != nil {
		log.Println(err)
		return
	}
	for _, value := range values {
		if resultSQL, err := smtp.Exec(value); err != nil {
			log.Println(err)
			log.Println(sqlCommand)
		} else {
			log.Println(resultSQL.LastInsertId())
			log.Println(sqlCommand)
		}
	}

}
func DoInsertFromForm( r *http.Request ) (lastInsertId int, err error) {

	r.ParseForm()

	if r.FormValue("table") == "" {
		log.Println("not table name")
		return -1, http.ErrNotSupported
	}

	tableName := r.FormValue("table")

	var row argsRAW

	comma, sqlCommand, values := "", "insert into " + tableName + "(", "values ("

	for key, val := range r.Form {

		if key == "table" {
			continue
		} else if strings.HasPrefix(key, "setid_"){
			defer func(tableName, key string, values []string) {
				insertMultiSet(tableName, key, values, lastInsertId)
			}(tableName, strings.TrimRight(key, "[]"), val)
		}
		if strings.Contains(key, "[]") {
			sqlCommand += comma + "`" + strings.TrimRight(key, "[]") + "`"
			str, comma := "", ""
			for _, value := range val {
				str += comma + value
				comma = ","
			}
			row = append(row, str)
		} else {
			sqlCommand += comma + "`" + key + "`"
			row = append( row, val[0] )
		}
		values += comma + "?"
		comma = ", "

	}

	lastInsertId, err = DoInsert(sqlCommand + ") " + values + ")", row ... )

	return lastInsertId, err

}
func DoUpdateFromForm( r *http.Request ) (id int, err error) {

	r.ParseForm()

	if r.FormValue("table") == "" {
		log.Println("not table name")
		return -1, http.ErrNotSupported
	}

	tableName := r.FormValue("table")
	var row argsRAW

	comma, sqlCommand, where := "", "update " + tableName + " set ", " where id="

	for key, val := range r.Form {

		if key == "table" {
			continue
		} else if key == "id" {
			where += val[0]
			id, _ = strconv.Atoi(val[0])
			continue
		} else if strings.HasPrefix(key, "setid_"){
			defer func(tableName, key string, values []string) {
				insertMultiSet(tableName, key, values, id)
			}(tableName, strings.TrimRight(key, "[]"), val)
		}

		if strings.Contains(key, "[]") {
			sqlCommand += comma + "`" + strings.TrimRight(key, "[]") + "`=?"
			str, comma := "", ""
			for _, value := range val {
				str += comma + value
				comma = ","
			}
			row = append(row, str)
		} else {
			sqlCommand += comma + "`" + key + "`=?"
			row = append( row, val[0] )
		}
		comma = ", "

	}
	return DoUpdate(sqlCommand + where, row ... )

}
func createCommand( sqlCommand string, r *http.Request, typeQuery string ) (row argsRAW, sqlQuery string) {

	comma := ""
	where := ""

	for key, val := range r.Form {


		switch key {
		case "call":
		case "sql":
		case "select":
		case "insert":
		case "update":
		case "where":
			where = " where " + val[0]
		default:
			row = append( row, val[0] )
			switch (typeQuery) {
			case "select":
				if (comma == "") {
					sqlCommand += " where " + key + "=?"
				} else {
					sqlCommand += " AND " + key + "=?"
				}
			case "update":
				sqlCommand += comma + key + "=?"
				comma = ","
			case "insert":
				sqlCommand += comma + key + ", ?"
			}
			comma = ", "
		}
	}

	return row, sqlCommand + where
}

func HandlerDBQuery(w http.ResponseWriter, r *http.Request) {

	var rows *sql.Rows

	_ = r.ParseForm()
	var row [] interface {}
	var sqlCommand string

	if command, isSelect := r.Form["sql"]; isSelect {
		row, sqlCommand = createCommand( command[0], r, "select")
	} else if command, isUpdate := r.Form["update"]; isUpdate {
		row, sqlCommand = createCommand( "update " + command[0] + " set ", r, "update")
	} else if command, isCall:= r.Form["call"]; isCall {
		row, sqlCommand = createCommand( "call " + command[0], r, "call")
	} else {
		var command, isInsert = r.Form["insert"]
		if (isInsert) {
			row, sqlCommand = createCommand( command[0], r, "insert")
		}
	}

	if (sqlCommand > "") {

		//defer main.Catch(w)
		switch len(row) {
		case 0:
			rows = DoQuery( sqlCommand )
		default:
			rows = DoQuery( sqlCommand, row... )

		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if rows == nil {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Что-то пошло не так" + sqlCommand))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write( GetResultToJSON(rows) )
	} else {
		fmt.Fprintf(w, "%q", row)
	}

}
type menuItem struct {
	Id int32
	Name string
	ParentID int32
	Title string
	SQL []byte
	Link string
}
type MenuItems struct {
	Self menuItem
	Items [] *menuItem
}
// find submenu (init menu first) return count submenu items
func (menu *MenuItems) GetMenu(id string) int {


	rows := DoQuery("select * from menu_items where parent_id=?", menu.Init(id))

	if rows == nil {
		log.Println("nil row")
		return 0
	}
	defer rows.Close()
	for rows.Next() {

		item := &menuItem{}
		if err := rows.Scan(&item.Id, &item.Name, &item.ParentID, &item.Title, &item.SQL, &item.Link); err != nil {
			log.Println(err)
			continue
		}
		menu.Items = append(menu.Items, item)
	}

	return len(menu.Items)
}
// -1 означает, что нет нужного нам пункта в меню
func (menu *MenuItems) Init(id string) int32 {

	var sqlQuery string

	if _, err := strconv.Atoi(id); err == nil {
		sqlQuery = "select * from menu_items where id=?"
	} else {
		sqlQuery = "select * from menu_items where name=?"
	}

	rows := DoQuery(sqlQuery, id)
	if rows == nil {
		log.Println("Not find menu wich id = ", id)
		return -1
	}

	defer rows.Close()
	for rows.Next() {

		if err := rows.Scan(&menu.Self.Id, &menu.Self.Name, &menu.Self.ParentID, &menu.Self.Title,
			&menu.Self.SQL, &menu.Self.Link); err != nil {
			log.Println(err)
			continue
		}
	}
	menu.Items = make( [] *menuItem, 0)

	return menu.Self.Id
}

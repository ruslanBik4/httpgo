package db

import (
	"net/http"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"strconv"
	"regexp"
)
type
	argsRAW  [] interface {}

type ArgsQuery struct {
	Comma, SQLCommand, Values string
	Args                      [] interface {}
}
type MultiQuery struct {
	Queryes map[string] *ArgsQuery
}

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

var (
	DigitsValidator = regexp.MustCompile(`^\d+$`)
)
func addNewItem(tableProps, value, userID string) (int, error) {

	if newId, err := DoInsert("insert into " + tableProps + "(title, id_users) values (?, ?)", value, userID); err != nil {
		return -1, err
	} else {
		return newId, nil
	}

}
//TODO: добавить запись для мультиполей (setid_)
func  insertMultiSet(tableName, tableProps, tableValues, userID string, values []string, id int) (err error) {

	sqlCommand := fmt.Sprintf("insert IGNORE into %s (id_%s, id_%s) values (%d, ?)",
		tableValues, tableName, tableProps, id)
	smtp, err := prepareQuery(sqlCommand)
	if err != nil {
		log.Println(err)
		return err
	}
	var params, comma string
	var valParams argsRAW

	for _, value := range values {

		// если не числовое значение - стало быть, это новое свойство и его добавим в таблицу свойств
		if !DigitsValidator.MatchString(value) {
			newId, err := addNewItem(tableProps, value, userID)
			if err != nil {
				log.Println(err)
				continue
			}
			value =  strconv.Itoa(newId)
		}
		if resultSQL, err := smtp.Exec(value); err != nil {
			log.Println(err)
			log.Println(sqlCommand)
		} else {
			log.Println(resultSQL.LastInsertId())
		}
		params += comma + "?"
		valParams = append(valParams, value )
		comma = ","
	}
	sqlCommand = fmt.Sprintf("delete from %s where id_%s = %d AND id_%s not in (%s)",
		tableValues, tableName, id, tableProps, params)

	if smtp, err = prepareQuery(sqlCommand); err != nil {
		log.Println(err)
		return err
	}

	if resultSQL, err := smtp.Exec(valParams ...); err != nil {
		log.Println(err)
		log.Println(sqlCommand)
		return err
	}else {
		log.Println(resultSQL.RowsAffected())
	}

	return err


}
func (tableIDQueryes *MultiQuery) addNewParam(key string, indSeparator int, val []string) {
	tableName := key[: indSeparator ]
	query, ok := tableIDQueryes.Queryes[tableName]
	if !ok {
		query = &ArgsQuery{
			Comma:      "",
			SQLCommand: "",
			Values:     "",
		}
	}
	fieldName := key[ strings.Index(key, ":") + 1: ]
	pos := strings.Index(fieldName, "[")
	fieldName = "`" + fieldName[ :pos] + "`"

	// пока беда в том, что количество должно точно соответствовать!
	//если первый  - то создаем новый список параметров для вставки
	if strings.HasPrefix(query.SQLCommand, fieldName) {
		query.Comma = "), ("
	} else if !strings.Contains(query.SQLCommand, fieldName )  {
		query.SQLCommand += query.Comma + fieldName
	}

	query.Values += query.Comma + "?"
	query.Comma = ", "
	query.Args = append(query.Args, val)
	for i, v := range val {
		log.Println(i,v)
	}
	tableIDQueryes.Queryes[tableName] = query

}
func (tableIDQueryes *MultiQuery) runQueryes(tableName string, lastInsertId int, Queryes map[string] *ArgsQuery) (err error){
	for childTableName, query := range Queryes {
		//query.Args = append(query.Args, lastInsertId)
		query.SQLCommand += query.Comma + "id_" + tableName
		query.Values += query.Comma + "?"
		fullCommand := fmt.Sprintf("replace into %s (%s) values (%s)", childTableName, query.SQLCommand, query.Values)

		var args [] interface{}

		for i := range query.Args[0].([]string) {
			if i > 0 {
				fullCommand += ",(" + query.Values + ")"
			}
			for _, valArr := range query.Args {
				switch valArr.(type) {
				case [] string:
					args = append(args, valArr.([]string)[i])
				default:
					args = append(args, valArr)

				}
			}
			// последним добавляем вторичный ключ
			args = append(args, lastInsertId)
		}
		if id, err := DoInsert(fullCommand,  args ...); err != nil {
			log.Println(err)
		} else {
			log.Println(fullCommand, id)
		}
	}

	return err
}
func GetNameTableProps(tableValue, parentTable string) string{
	var ns FieldsTable
	ns.GetColumnsProp(tableValue)

	for _, field := range ns.Rows {
		if (field.COLUMN_NAME != "id_" + parentTable) && strings.HasPrefix(field.COLUMN_NAME, "id_") {
			return field.COLUMN_NAME[3:]
		}
	}

	return ""
}
func DoInsertFromForm( r *http.Request, userID string ) (lastInsertId int, err error) {

	r.ParseForm()

	if r.FormValue("table") == "" {
		log.Println("not table name")
		return -1, http.ErrNotSupported
	}

	tableName := r.FormValue("table")

	var row argsRAW
	var tableIDQueryes MultiQuery
	tableIDQueryes.Queryes = make(map[string] *ArgsQuery, 0)

	comma, sqlCommand, values := "", "insert into " + tableName + "(", "values ("

	for key, val := range r.Form {

		indSeparator := strings.Index(key, ":")

		if key == "table" {
			continue
		} else if strings.HasPrefix(key, "setid_"){
			defer func(tableName, tableProps string, values []string) {
				if err != nil {
					err = insertMultiSet(tableName,  tableProps,
						tableName + "_" + tableProps + "_has", userID, values, lastInsertId)
				}
			} (tableName, key[len("setid_"):len(key)-2], val)
			continue
		} else if strings.HasPrefix(key, "nodeid_"){
			defer func(tableName, tableValues string, values []string) {
				if err != nil {
					err = insertMultiSet(tableName, GetNameTableProps(tableValues, tableName),
						tableValues, userID, values, lastInsertId)
				}
			} (tableName, key[len("nodeid_"):len(key)-2], val)
			continue
		} else if key == "id_users" {

			sqlCommand += comma + "`" + key + "`"
			row = append( row, userID )

		} else if strings.Contains(key, "[]") {
			sqlCommand += comma + "`" + strings.TrimRight(key, "[]") + "`"
			str, comma := "", ""
			for _, value := range val {
				str += comma + value
				comma = ","
			}
			row = append(row, str)
		} else if (indSeparator > 1) && strings.Contains(key, "[")  {
			tableIDQueryes.addNewParam(key, indSeparator, val)
			continue
		} else {
			sqlCommand += comma + "`" + key + "`"
			row = append( row, val[0] )
		}
		values += comma + "?"
		comma = ", "

	}

	// если будут дополнительные запросы
	if len(tableIDQueryes.Queryes) > 0 {
		// исполнить по завершению функции, чтобы получить lastInsertId
		defer func() {
			if err != nil {
				err = tableIDQueryes.runQueryes(tableName, lastInsertId, tableIDQueryes.Queryes)
			}
		} ()

	}
	return DoInsert(sqlCommand + ") " + values + ")", row ... )

}
func DoUpdateFromForm( r *http.Request, userID string ) (RowsAffected int, err error) {

	r.ParseForm()

	if r.FormValue("table") == "" {
		log.Println("not table name")
		return -1, http.ErrNotSupported
	}

	tableName := r.FormValue("table")
	var row argsRAW
	var id int
	var tableIDQueryes MultiQuery
	tableIDQueryes.Queryes = make(map[string] *ArgsQuery, 0)

	comma, sqlCommand, where := "", "update " + tableName + " set ", " where id="

	for key, val := range r.Form {

		indSeparator := strings.Index(key, ":")
		if key == "table" {
			continue
		} else if key == "id" {
			where += val[0]
			id, _ = strconv.Atoi(val[0])
			continue
		} else if strings.HasPrefix(key, "setid_"){
			defer func(tableProps string, values []string) {
				if err != nil {
					err = insertMultiSet(tableName,  tableProps,
						tableName + "_" + tableProps + "_has", userID, values, id)
				}
			} (key[len("setid_"):len(key)-2], val)
			continue
		} else if strings.HasPrefix(key, "nodeid_"){
			defer func(tableValues string, values []string) {
				if err != nil {
					err = insertMultiSet(tableName, GetNameTableProps(tableValues, tableName),
						tableValues, userID, values, id)
				}
			} (key[len("nodeid_"):len(key)-2], val)
			continue
		} else if key == "id_users" {

			sqlCommand += comma + "`" + key + "`=?"
			row = append( row, userID )

		} else if strings.Contains(key, "[]") {
			sqlCommand += comma + "`" + strings.TrimRight(key, "[]") + "`=?"
			str, comma := "", ""
			for _, value := range val {
				str += comma + value
				comma = ","
			}
			row = append(row, str)
		} else if (indSeparator > 1) && strings.Contains(key, "[")  {
			tableIDQueryes.addNewParam(key, indSeparator, val)
			continue
		} else {
			sqlCommand += comma + "`" + key + "`=?"
			row = append( row, val[0] )
		}
		comma = ", "

	}
	// если будут дополнительные запросы
	if len(tableIDQueryes.Queryes) > 0 {
		// исполнить по завершению функции, чтобы получить lastInsertId
		defer func() {
			if err != nil {
				err = tableIDQueryes.runQueryes(tableName, id, tableIDQueryes.Queryes)
			}
		} ()

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

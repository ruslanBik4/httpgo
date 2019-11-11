// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// some function for request БД
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/models/db/multiquery"
	"github.com/ruslanBik4/httpgo/models/db/schema"
)

type argsRAW []interface{}

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

func GetNameTableProps(tableValue, parentTable string) string {
	var ns FieldsTable
	ns.GetColumnsProp(tableValue)

	for _, field := range ns.Rows {
		if (field.COLUMN_NAME != "id_"+parentTable) && strings.HasPrefix(field.COLUMN_NAME, "id_") {
			return field.COLUMN_NAME[3:]
		}
	}

	return ""
}

var isArrayPostParam = regexp.MustCompile(`/\[\S*\]/`)

//получает имя таблицы свойств для суррогатных полей типа
func getTableProps(key, typeField string) string {
	if strings.HasSuffix(key, "[]") {
		return key[len(typeField) : len(key)-2]
	} else if isArrayPostParam.MatchString(key) {
		return isArrayPostParam.ReplaceAllString(key, "")[len(typeField):]

	}
	return key[len(typeField):]
}

const _2K = (1 << 10) * 2

func checkPOSTParams(r *http.Request) string {

	r.ParseMultipartForm(_2K)

	return r.FormValue("table")

}

// check field in schema
func checkField(key string, table *schema.FieldsTable) (field *schema.FieldStructure, indSeparator int, err error) {

	indSeparator = strings.Index(key, ":")
	fieldName := key
	if pos := strings.Index(fieldName, "["); pos > 0 {
		fieldName = fieldName[:pos]
		logs.DebugLog(fieldName)
	}
	if indSeparator > 0 {
		tableChild := schema.GetFieldsTable(fieldName[:indSeparator])
		fieldName = fieldName[indSeparator+1:]
		field = tableChild.FindField(fieldName)
	} else {
		field = table.FindField(fieldName)
	}

	if field == nil {
		return nil, -1, ErrBadParam{Name: fieldName, BadName: " field not in table" + table.Name, FuncName: "DoInsertFromForm"}
	}

	return field, indSeparator, nil

}

//выполняет запрос согласно переданным данным в POST,
//для суррогатных полей готовит запросы для изменения связанных полей
//возвращает id новой записи
func DoInsertFromForm(r *http.Request, userID string, txConn ...*TxConnect) (lastInsertId int, err error) {

	tableName := checkPOSTParams(r)
	if len(r.Form) < 3 {
		return -1, errors.New("Count params is less 3!")
	}
	if tableName == "" {
		logs.ErrorLog(errors.New("not table name"))
		return -1, ErrParamNotFound{Name: "table", FuncName: "DoInsertFromForm"}
	}

	if r.FormValue("id") > "" {
		return -1, ErrBadParam{Name: "id", BadName: "found uncorrect params", FuncName: "DoInsertFromForm"}
	}
	table := schema.GetFieldsTable(tableName)
	// удаляем определяющие параметры
	r.Form.Del("table")

	tableIDQueryes := multiquery.Create(tableName)
	var tx *TxConnect
	// проверяем, что мы в контексте транзакции
	if len(txConn) > 0 {
		tx = txConn[0]
	}

	var args argsRAW

	comma, sqlCommand, values := "", "insert into "+tableName+"(", "values ("

	for key, val := range r.Form {

		field, indSeparator, err := checkField(key, table)
		if err != nil {
			return -1, err
		}
		if strings.HasPrefix(key, "setid_") {
			if tx == nil {
				tx, err = startTXforQuery()
				defer func() {
					finishTX(tx, err)
				}()
			}
			defer func(tableProps string, values []string) {
				if err == nil {
					err = tx.insertMultiSet(tableName, tableProps,
						tableName+"_"+tableProps+"_has", userID, values, lastInsertId)
				}
			}(getTableProps(key, "setid_"), val)
			continue
		} else if strings.HasPrefix(key, "nodeid_") {
			if tx == nil {
				tx, err = startTXforQuery()
				defer func() {
					finishTX(tx, err)
				}()
			}

			defer func(tableValues string, values []string) {
				if err == nil {
					tableProps := GetNameTableProps(tableValues, tableName)
					if tableProps == "" {
						logs.DebugLog("Empty tableProps! ", tableValues)
					}
					err = tx.insertMultiSet(tableName, tableProps, tableValues, userID, values, lastInsertId)
				}
			}(getTableProps(key, "nodeid_"), val)
			continue
		} else if key == "id_users" {

			sqlCommand += comma + "`" + key + "`"
			args = append(args, userID)

		} else if strings.Contains(key, "[]") {
			sqlCommand += comma + "`" + strings.TrimRight(key, "[]") + "`"
			str, comma := "", ""
			for _, value := range val {
				str += comma + value
				comma = ","
			}
			args = append(args, str)
		} else if (indSeparator > 1) && strings.Contains(key, "[") {
			tableIDQueryes.AddNewParam(key, indSeparator, val, field)
			continue
		} else {
			sqlCommand += comma + "`" + key + "`"
			args = append(args, val[0])
		}
		values += comma + "?"
		comma = ", "

	}

	if len(tableIDQueryes.Queryes) > 0 {
		if tx == nil {
			tx, err = startTXforQuery()
			defer func() {
				finishTX(tx, err)
			}()
		}
		defer func() {
			if err == nil {
				err = runMultiQuery(tableIDQueryes, lastInsertId, tx)
			}
		}()
	}

	if tx == nil {
		return DoInsert(sqlCommand+") "+values+")", args...)
	} else {
		return tx.DoInsert(sqlCommand+") "+values+")", args...)
	}

}

//выполняет запрос согласно переданным данным в POST,
//для суррогатных полей готовит запросы для изменения связанных полей
//возвращает количество измененных записей
//TODO: сменить проверку параметров в цикле на предпроверку и добавить связку с схемой БД
func DoUpdateFromForm(r *http.Request, userID string, txConn ...*TxConnect) (RowsAffected int, err error) {

	tableName := checkPOSTParams(r)

	if len(r.Form) < 3 {
		return -1, errors.New("Count params is less 3!")
	}
	idText := r.FormValue("id")
	if (tableName == "") && (idText == "") {
		logs.ErrorLog(errors.New("not table name"))
		return -1, ErrParamNotFound{Name: "table", FuncName: "DoUpdateFromForm"}
	}

	id, err := strconv.Atoi(idText)
	if (err != nil) || (id < 1) {
		return -1, ErrBadParam{Name: "id", BadName: idText, FuncName: "DoUpdateFromForm"}
	}
	table := schema.GetFieldsTable(tableName)
	// удаляем определяющие параметры
	r.Form.Del("table")
	r.Form.Del("id")

	tableIDQueryes := multiquery.Create(tableName)
	var tx *TxConnect
	// проверяем, что мы в контексте транзакции
	if len(txConn) > 0 {
		tx = txConn[0]
	}
	var args argsRAW

	comma, sqlCommand, where := "", "update "+tableName+" set ", " where id=?"

	for key, val := range r.Form {

		switch key {
		case "id_users":
			sqlCommand += comma + "`" + key + "`=?"
			args = append(args, userID)
		default:
			field, indSeparator, err := checkField(key, table)
			if err != nil {
				return -1, err
			}
			if strings.HasPrefix(key, "setid_") {
				if tx == nil {
					tx, err = startTXforQuery()
					defer func() {
						finishTX(tx, err)
					}()
				}

				defer func(tableProps string, values []string) {
					if err == nil {
						err = tx.insertMultiSet(tableName, tableProps,
							tableName+"_"+tableProps+"_has", userID, values, id)
					}
				}(getTableProps(key, "setid_"), val)
				continue
			} else if strings.HasPrefix(key, "nodeid_") {
				if tx == nil {
					tx, err = startTXforQuery()
					defer func() {
						finishTX(tx, err)
					}()
				}

				defer func(tableValues string, values []string) {
					if err == nil {
						tableProps := GetNameTableProps(tableValues, tableName)
						if tableProps == "" {
							logs.DebugLog("Empty tableProps! ", tableValues)
						}
						err = tx.insertMultiSet(tableName, tableProps, tableValues, userID, values, id)
					}
				}(getTableProps(key, "nodeid_"), val)
				continue

			} else if strings.Contains(key, "[]") {
				// fields type SET | ENUM
				sqlCommand += comma + "`" + strings.TrimRight(key, "[]") + "`=?"
				str := strings.Join(val, ",")
				args = append(args, str)
			} else if (indSeparator > 1) && strings.Contains(key, "[") {
				tableIDQueryes.AddNewParam(key, indSeparator, val, field)
				continue
			} else {
				sqlCommand += comma + "`" + key + "`=?"
				args = append(args, val[0])
			}
		}
		comma = ", "

	}

	args = append(args, id)
	// если будут дополнительные запросы
	if len(tableIDQueryes.Queryes) > 0 {
		if tx == nil {
			tx, err = startTXforQuery()
			defer func() {
				finishTX(tx, err)
			}()
		}
		defer func() {
			if err == nil {
				err = runMultiQuery(tableIDQueryes, id, tx)
			}
		}()
	}

	if tx == nil {
		return DoUpdate(sqlCommand+where, args...)
	} else {
		return tx.DoUpdate(sqlCommand+where, args...)
	}

}

// запускаем транзакцию, если в этом есть необходимость
func startTXforQuery() (*TxConnect, error) {
	return StartTransaction()
}

// и закрываем потом (в зависемости от результатов запросов)
func finishTX(tx *TxConnect, err error) {
	if err != nil {
		tx.RollbackTransaction()
	} else {
		tx.CommitTransaction()
	}
}
func runMultiQuery(tableIDQueryes *multiquery.MultiQuery, parentId int, tx *TxConnect) error {
	for _, query := range tableIDQueryes.Queryes {

		sql, args := query.GetUpdateSQL(parentId)
		idQuery, err := tx.DoUpdate(sql, args...)
		if err != nil {
			return err
		}
		logs.DebugLog("Insert new child", idQuery)
	}
	return nil
}
func createCommand(sqlCommand string, r *http.Request, typeQuery string) (row argsRAW, sqlQuery string) {

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
			row = append(row, val[0])
			switch typeQuery {
			case "select":
				if comma == "" {
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
	var row []interface{}
	var sqlCommand string

	if command, isSelect := r.Form["sql"]; isSelect {
		row, sqlCommand = createCommand(command[0], r, "select")
	} else if command, isUpdate := r.Form["update"]; isUpdate {
		row, sqlCommand = createCommand("update "+command[0]+" set ", r, "update")
	} else if command, isCall := r.Form["call"]; isCall {
		row, sqlCommand = createCommand("call "+command[0], r, "call")
	} else {
		var command, isInsert = r.Form["insert"]
		if isInsert {
			row, sqlCommand = createCommand(command[0], r, "insert")
		}
	}

	if sqlCommand > "" {

		//defer main.Catch(w)
		switch len(row) {
		case 0:
			rows = DoQuery(sqlCommand)
		default:
			rows = DoQuery(sqlCommand, row...)

		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if rows == nil {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Что-то пошло не так" + sqlCommand))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(GetResultToJSON(rows))
	} else {
		fmt.Fprintf(w, "%q", row)
	}

}

type menuItem struct {
	Id        int32
	Name      string
	ParentID  int32
	Title     string
	SQL       []byte
	Link      string
	SortOrder int32
	Elements  string
}
type MenuItems struct {
	Self  menuItem
	Items []*menuItem
}

var IDvalidator = regexp.MustCompile(`^\d+$`)

// find submenu (init menu first) return count submenu items
func (menu *MenuItems) GetMenu(id string) int {

	rows, err := DoSelect("select * from menu_items where parent_id=?", menu.Init(id))

	if err != nil {
		logs.ErrorLog(err)
		return 0
	}

	defer rows.Close()
	for rows.Next() {

		item := &menuItem{}
		if err := item.Scan(rows); err != nil {
			logs.ErrorLog(err)
			continue
		}
		menu.Items = append(menu.Items, item)
	}

	return len(menu.Items)
}

//получаем пункты мену по id_user (если пользователь является администратором то показываем меню администратора)
func (menu *MenuItems) GetMenuByUserId(user_id int) int {

	isAdmin, err := DoSelect("SELECT is_general FROM roles_list "+
		"INNER JOIN users_roles_list_has UR ON UR.id_roles_list=roles_list.id "+
		"WHERE UR.id_users=?", user_id)

	if err != nil {
		logs.ErrorLog(err)
		return 0
	}

	defer isAdmin.Close()
	for isAdmin.Next() {
		is_admin := 0
		if err := isAdmin.Scan(&is_admin); err != nil {
			logs.ErrorLog(err)
			continue
		}
		if is_admin > 0 {
			return menu.GetMenu("admin")
		}

	}

	extranetMenuId := "admin"

	rows, err := DoSelect("SELECT menu_items.`id`, menu_items.`name`, menu_items.`parent_id`, menu_items.`title`, menu_items.`sql`, menu_items.`link` "+
		"FROM users_roles_list_has "+
		"LEFT JOIN roles_permission_list ON `roles_permission_list`.`id_roles_list`=users_roles_list_has.id_roles_list "+
		"INNER JOIN `menu_items` ON `roles_permission_list`.`id_menu_items` = menu_items.`id` "+
		"WHERE users_roles_list_has.id_users=? AND menu_items.parent_id=?", user_id, menu.Init(extranetMenuId))

	if err != nil {
		logs.ErrorLog(err)
		return 0
	}

	defer rows.Close()
	for rows.Next() {

		item := &menuItem{}
		if err := item.Scan(rows); err != nil {
			logs.ErrorLog(err)
			continue
		}
		menu.Items = append(menu.Items, item)
	}

	return len(menu.Items)
}

func (menu *menuItem) Scan(rows *sql.Rows) error {
	return rows.Scan(&menu.Id, &menu.Name, &menu.ParentID, &menu.Title,
		&menu.SQL, &menu.Link, &menu.SortOrder, &menu.Elements)
}

// -1 означает, что нет нужного нам пункта в меню
func (menu *MenuItems) Init(id string) int32 {

	sqlQuery := "select * from menu_items where "

	if IDvalidator.MatchString(id) {
		sqlQuery += "id=?"
	} else {
		sqlQuery += "name=?"
	}

	rows, err := DoSelect(sqlQuery, id)
	if err != nil {
		logs.ErrorLog(err)
		return -1
	}

	defer rows.Close()
	// если нет записей
	if !rows.Next() {
		logs.DebugLog("Not find menu which id = ", id)
		return -1

	}

	if err := menu.Self.Scan(rows); err != nil {
		logs.ErrorLog(err)
		return -1
	}

	menu.Items = make([]*menuItem, 0)

	return menu.Self.Id
}

// функция возвращает стандартный sql для вставки данных в таблицу
// пример работы https://play.golang.org/p/4KeGhkskh5
func GetSimpleInsertSQLString(table string, args ...string) string {

	var comma string
	gravis := "`"
	placeholders := ""
	sqlString := "INSERT INTO " + gravis + table + gravis + " ("
	for _, val := range args {
		sqlString += comma + gravis + val + gravis
		placeholders += comma + "?"
		comma = ","
	}
	sqlString += ") VALUES (" + placeholders + ")"

	return sqlString
}

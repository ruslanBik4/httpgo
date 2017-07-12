// Copyright 2017 Author: Yurii Kravchuk. All rights reserved.

package services

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"sync"
)

type linkPermission struct {
	link string
	allow_create int
	allow_delete int
	allow_edit int
	id_users int
}

type cpService struct {
	name string
	region string
	status string
	Rows map[string] rowsRoles
	roles map[int][]interface{}
}

var crm_permission *cpService = &cpService{name:"crm_permission"}
var cacheMu sync.RWMutex

//реализация обязательных методов интерейса
func (crm_permission *cpService) Init() error{

	rows, err := db.DoSelect("SELECT `menu_items`.`link`, `roles_permission_list`.`allow_create`, " +
		"`roles_permission_list`.`allow_delete`, `roles_permission_list`.`allow_edit`, `users_roles_list_has`.`id_users` " +
		"FROM users_roles_list_has " +
		"LEFT JOIN roles_permission_list ON `roles_permission_list`.`id_roles_list`=users_roles_list_has.id_roles_list " +
		"INNER JOIN roles_list ON users_roles_list_has.`id_roles_list`=`roles_list`.id " +
		"INNER JOIN `menu_items` ON `roles_permission_list`.`id_menu_items` = menu_items.`id` " +
		"ORDER BY users_roles_list_has.`id_users` ASC")

	if err != nil {
		return err
	}

	roles := make(map[int][]interface{}, 0)
	for rows.Next() {
		var link string
		var allow_create, allow_delete, allow_edit, id_users int
		if err := rows.Scan(&link, &allow_create, &allow_delete, &allow_edit, &id_users); err != nil {
			continue
		}

		newRow := make(map[string]interface{}, 0)
		newRow["link"] = link
		newRow["allow_create"] = allow_create
		newRow["allow_delete"] = allow_delete
		newRow["allow_edit"] = allow_edit

		roles[id_users] = append(roles[id_users], newRow)
	}

	crm_permission.roles = roles
	crm_permission.status = "ready"
	return nil
}

// args: 0 => admin part, 1 => user id, 2 => url what test on permiss, 3 => set/delete action with permiss
// 4 => is allow create for this url, 5 => is allow delete for this url, 6 => is allow edit for this url
// 4,5,6 (for set permiss only)
func (crm_permission *cpService) Send(args ...interface{}) error {

	if crm_permission.status != "ready" {
		return ErrBrokenConnection{Name: crm_permission.name, Param: args}
	}

	if len(args) < 4 {
		return ErrServiceNotEnoughParameter{Name: crm_permission.name, Param: args}
	}
	if _,ok := args[1].(int); !ok {
		return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[1], Number: 2}
	}
	if _,ok := args[2].(string); !ok {
		return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[2], Number: 3}
	}

	switch permission_type := args[0].(type) {
	case string:
		if permission_type == "crm" {
			if args[3].(string) == "set" {
				return crm_permission.setPermissForUser(args[1].(int), args[2].(string), args[4].(bool), args[5].(bool), args[6].(bool));
			} else if args[3].(string) == "delete" {
				return crm_permission.deletePermissForUser(args[1].(int), args[2].(string));
			}

		}
	default:
		return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: permission_type, Number: 1}
	}

	return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}

}

// args: 0 => admin part, 1 => user id, 2 => url what test on permiss, 3 => action for test access (Create/Delete/Edit/View)
func (crm_permission *cpService) Get(args ... interface{}) ( interface{}, error) {

	if len(args) < 4 {
		return nil, ErrServiceNotEnoughParameter{Name: crm_permission.name, Param: args}
	}
	if _,ok := args[1].(int); !ok {
		return nil, ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[1], Number: 2}
	}

	connection_status := Status("crm_permission")

	if connection_status != "ready" {
		return nil, ErrBrokenConnection{Name: crm_permission.name, Param: args}
	}

	switch permission_type := args[0].(type) {
	case string:
		if permission_type == "crm" {
			return crm_permission.getCRMPermissions(args[1].(int), args[2].(string), args[3].(string)), nil;
		}
	default:
		return nil, ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: permission_type, Number: 1}
	}

	return nil, ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
}
func (crm_permission *cpService) Connect(in <- chan interface{}) (out chan interface{}, err error) {
	out = make(chan interface{})

	go func() {
		out<-"open"
		for {
			select {
			case v := <-in:
				if v.(string) == "close" {
					crm_permission.Close(out)
				} else {
					out <- v
				}
			}
		}
	}()
	return out, nil
}
func (crm_permission *cpService) Close(out chan <- interface{}) error {
	close(out)
	return nil

}
func (crm_permission *cpService) Status() string {
	return crm_permission.status
}

func (crm_permission *cpService) getCRMPermissions(user_id int, url, action string) bool {

	if crm_permission.roles[user_id] == nil || len(crm_permission.roles[user_id]) == 0 {
		return false
	}

	for _,permission := range crm_permission.roles[user_id] {
		resRow := permission.(map[string]interface{})
		if resRow["link"].(string) == url {
			return checkAction(resRow, action)
		}
	}
	return false
}

func (crm_permission *cpService) deletePermissForUser(user_id int, url string) error {

	cacheMu.Lock()

	if crm_permission.roles[user_id] == nil || len(crm_permission.roles[user_id]) == 0 {
		return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
	}

	for key,permission := range crm_permission.roles[user_id] {
		resRow := permission.(map[string]interface{})
		if resRow["link"].(string) == url {
			crm_permission.roles[user_id] = append(crm_permission.roles[user_id][:key], crm_permission.roles[user_id][key+1:]...)
			return nil
		}
	}

	cacheMu.Unlock()

	return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
}

func (crm_permission *cpService) setPermissForUser(user_id int, link string, allow_create, allow_delete, allow_edit bool) error {

	cacheMu.Lock()

	newRow := make(map[string]interface{}, 0)
	newRow["link"] = link

	if allow_create {
		newRow["allow_create"] = 1
	} else {
		newRow["allow_create"] = 0
	}

	if allow_delete {
		newRow["allow_delete"] = 1
	} else {
		newRow["allow_delete"] = 0
	}

	if allow_edit {
		newRow["allow_edit"] = 1
	} else {
		newRow["allow_edit"] = 0
	}

	crm_permission.roles[user_id] = append(crm_permission.roles[user_id], newRow)

	cacheMu.Unlock()
	return nil
}

func init() {
	AddService(crm_permission.name, crm_permission)
}

func checkAction(permiss interface{}, action string) bool {
	convert := permiss.(map[string]interface{})
	switch action {
	case "Create":
		if convert["allow_create"].(int) == 1 {
			return true
		} else {
			return false
		}
	case "Delete":
		if convert["allow_delete"].(int) == 1 {
			return true
		} else {
			return false
		}
	case "Edit":
		if convert["allow_edit"].(int) == 1 {
			return true
		} else {
			return false
		}
	case "View":
		return true
	default:
		return false
	}
}
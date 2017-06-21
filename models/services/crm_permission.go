// Copyright 2017 Author: Yurii Kravchuk. All rights reserved.

package services

import (
	"github.com/ruslanBik4/httpgo/models/db"
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
func (crm_permission *cpService) Send(messages ...interface{}) error {



	return nil

}
func (crm_permission *cpService) Get(args ... interface{}) ( interface{}, error) {

	if len(args) < 4 {
		return nil, ErrServiceNotEnougnParameter{Name: crm_permission.name, Param: args}
	}
	if _,ok := args[1].(int); !ok {
		return nil, ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[1], Number: 2}
	}
	if _,ok := args[2].(string); !ok {
		return nil, ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[2], Number: 3}
	}
	if _,ok := args[3].(string); !ok {
		return nil, ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[3], Number: 4}
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

// Copyright 2017 Author: Yurii Kravchuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//Реализует работу с правами пользователя для доступа в CRM/Extranet
package crm_permission

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/services"
	"sync"
)

//структура прав для пользователя по ссылке (CRM)
type linkPermission struct {
	link string
	allow_create int
	allow_delete int
	allow_edit int
	id_users int
}

//структура прав роли по ссылке (Extranet)
type roles struct {
	link string
	allow_create int
	allow_delete int
	allow_edit int
}

type cpService struct {
	name string
	region string
	status string
	crm_permissions_roles map[int][]interface{}
	extranet_roles map[int][]roles
	extranet_permissions map[int]map[int]int
}

const CREATE_ACTION = "Create"
const DELETE_ACTION = "Delete"
const EDIT_ACTION = "Edit"
const VIEW_ACTION = "View"
const SET_PERMISS = "Set"
const DROP_PERMISS = "Drop"
const CRM_PART = "crm"
const EXTRANET_PART = "extranet"

var crm_permission *cpService = &cpService{name:"crm_permission", status: "create"}
var cacheMu sync.RWMutex

//реализация обязательных методов интерейса
func (crm_permission *cpService) Init() error{

	crm_permission.status = "init"

	err := crm_permission.setUserPermissionForCRM()

	if err != nil {
		return err
	}

	roles_err := crm_permission.setExtranetRoles()

	if roles_err != nil {
		return roles_err
	}

	user_roles_err := crm_permission.setExtranetUserRoles()

	if user_roles_err != nil {
		return user_roles_err
	}

	crm_permission.status = "ready"
	return nil
}

// args: 0 => admin part, 1 => user id, 2 => url what test on permiss, 3 => set/delete action with permiss
// 4 => is allow create for this url, 5 => is allow delete for this url, 6 => is allow edit for this url
// 4,5,6 (for set permiss only)
//for Extranet 7 => id_hotels for set permiss
//for Extranet 5 => id_role for set permiss
//for Extranet 4 => id_hotels for drop permiss
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
		if permission_type == CRM_PART {
			if args[3].(string) == SET_PERMISS {
				return crm_permission.setPermissForUser(args[1].(int), args[2].(string), args[4].(bool), args[5].(bool), args[6].(bool));
			} else if args[3].(string) == DROP_PERMISS {
				return crm_permission.deletePermissForUser(args[1].(int), args[2].(string));
			}
		} else if permission_type == EXTRANET_PART {
			if args[3].(string) == SET_PERMISS {
				return crm_permission.setPermissForUserExtranet(args[1].(int), args[4].(int), args[5].(int));
			} else if args[3].(string) == DROP_PERMISS {
				return crm_permission.deletePermissForUserExtranet(args[1].(int), args[4].(int));
			}
		}
	default:
		return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: permission_type, Number: 1}
	}

	return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}

}

// args: 0 => admin part, 1 => user id, 2 => url what test on permiss, 3 => action for test access (Create/Delete/Edit/View)
//for Extranet 4 => id_hotels
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
		if permission_type == CRM_PART {
			return crm_permission.getCRMPermissions(args[1].(int), args[2].(string), args[3].(string)), nil;
		} else if permission_type == EXTRANET_PART {

			if _,ok := args[4].(int); !ok {
				return nil, ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[4], Number: 2}
			}

			return crm_permission.getExtranetPermissions(args[1].(int), args[2].(string),
				args[3].(string), args[4].(int)), nil;
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

//получение статуса сервиса
func (crm_permission *cpService) Status() string {
	return crm_permission.status
}

//получение права пользователя в CRM на выполнения конкретного действия по конкретному url
func (crm_permission *cpService) getCRMPermissions(user_id int, url, action string) bool {

	if crm_permission.crm_permissions_roles[user_id] == nil || len(crm_permission.crm_permissions_roles[user_id]) == 0 {
		return false
	}

	for _,permission := range crm_permission.crm_permissions_roles[user_id] {
		resRow := permission.(map[string]interface{})
		if resRow["link"].(string) == url {
			return checkAction(resRow, action)
		}
	}
	return false
}

//получение прав пользователя в Extranet для конкретного отеля
func (crm_permission *cpService) getExtranetPermissions(user_id int, url, action string, id_hotels int) bool {

	role_id := crm_permission.getUserRole(user_id, id_hotels)

	if role_id == 0 || crm_permission.extranet_roles[role_id] == nil {
		return false
	}

	for _,permission := range crm_permission.extranet_roles[role_id] {
		if permission.link == url {
			return checkActionExtranet(permission, action)
		}
	}

	return false

}

//получение роли пользователя для Extranet для конкретного отеля
func (crm_permission *cpService) getUserRole(user_id, id_hotels int) int {

	for hotel,role := range crm_permission.extranet_permissions[user_id] {
		if hotel == id_hotels {
			return role
		}
	}

	return 0
}
//удаление роли для пользователя в CRM
func (crm_permission *cpService) deletePermissForUser(user_id int, url string) error {

	cacheMu.Lock()
	// TODO: Unlock must to allow there as defer func

	if crm_permission.crm_permissions_roles[user_id] == nil || len(crm_permission.crm_permissions_roles[user_id]) == 0 {
		return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
	}

	for key,permission := range crm_permission.crm_permissions_roles[user_id] {
		resRow := permission.(map[string]interface{})
		if resRow["link"].(string) == url {
			crm_permission.crm_permissions_roles[user_id] = append(crm_permission.crm_permissions_roles[user_id][:key],
				crm_permission.crm_permissions_roles[user_id][key+1:]...)
			return nil
		}
	}

	cacheMu.Unlock()

	return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
}

//выставление роли для пользователя в CRM
func (crm_permission *cpService) setPermissForUser(user_id int, link string, allow_create, allow_delete, allow_edit bool) error {

	cacheMu.Lock()
	// TODO: Unlock must to allow there as defer func

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

	crm_permission.crm_permissions_roles[user_id] = append(crm_permission.crm_permissions_roles[user_id], newRow)

	cacheMu.Unlock()
	return nil
}

//удаление роли для пользователя в Extranet для конкретного отеля
func (crm_permission *cpService) deletePermissForUserExtranet(user_id int, id_hotels int) error {

	cacheMu.Lock()
	// TODO: Unlock must to allow there as defer func

	if crm_permission.extranet_permissions[user_id] == nil || len(crm_permission.extranet_permissions[user_id]) == 0 {
		return ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
	}

	crm_permission.extranet_permissions[user_id][id_hotels] = 0

	cacheMu.Unlock()

	return nil
}

//выставление роли для пользователя в Extranet для конкретного отеля
func (crm_permission *cpService) setPermissForUserExtranet(user_id, id_hotels, id_role int) error {

	cacheMu.Lock()
	// TODO: Unlock must to allow there as defer func

	crm_permission.extranet_permissions[user_id][id_hotels] = id_role

	cacheMu.Unlock()
	return nil
}

//заполнение масива прав для CRM
func (crm_permission *cpService) setUserPermissionForCRM() error {

	rows, err := db.DoSelect("SELECT `menu_items`.`link`, `roles_permission_list`.`allow_create`, " +
		"`roles_permission_list`.`allow_delete`, `roles_permission_list`.`allow_edit`, `users_roles_list_has`.`id_users` " +
		"FROM users_roles_list_has " +
		"LEFT JOIN roles_permission_list ON `roles_permission_list`.`id_roles_list`=users_roles_list_has.id_roles_list " +
		"INNER JOIN roles_list ON users_roles_list_has.`id_roles_list`=`roles_list`.id " +
		"INNER JOIN `menu_items` ON `roles_permission_list`.`id_menu_items` = menu_items.`id` " +
		"WHERE roles_list.is_extranet = 0 " +
		"ORDER BY users_roles_list_has.`id_users` ASC")

	if err != nil {
		return err
	}

	crm_permissions_roles := make(map[int][]interface{}, 0)
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

		crm_permissions_roles[id_users] = append(crm_permissions_roles[id_users], newRow)
	}

	crm_permission.crm_permissions_roles = crm_permissions_roles

	return nil
}

//заполнение масива прав для ролей Extranet
func (crm_permission *cpService) setExtranetRoles() error {
	roles_rows, roles_err := db.DoSelect("SELECT `roles_list`.id AS id_role, " +
		"`menu_items`.`link`, `roles_permission_list`.`allow_create`, " +
		"`roles_permission_list`.`allow_delete`, `roles_permission_list`.`allow_edit` " +
		"FROM `roles_list` " +
		"LEFT JOIN roles_permission_list ON `roles_list`.`id`=roles_permission_list.id_roles_list " +
		"INNER JOIN `menu_items` ON `roles_permission_list`.`id_menu_items` = menu_items.`id` " +
		"WHERE roles_list.is_extranet = 1")

	if roles_err != nil {
		return roles_err
	}

	extranet_roles := make(map[int][]roles, 0)

	for roles_rows.Next() {
		var extranet_role roles

		var link string
		var allow_create, allow_delete, allow_edit, id_role int
		if err := roles_rows.Scan(&id_role, &link, &allow_create, &allow_delete, &allow_edit); err != nil {
			continue
		}

		extranet_role.link = link
		extranet_role.allow_create = allow_create
		extranet_role.allow_delete = allow_delete
		extranet_role.allow_edit = allow_edit

		extranet_roles[id_role] = append(extranet_roles[id_role], extranet_role)
	}

	crm_permission.extranet_roles = extranet_roles

	return nil
}

//заполнение масива ролей для Extranet
func (crm_permission *cpService) setExtranetUserRoles() error {
	extranet_user_roles, extranet_user_roles_err := db.DoSelect("SELECT id_users, id_roles_list, id_hotels FROM " +
		"users_roles_list_has_extranet")

	if extranet_user_roles_err != nil {
		return extranet_user_roles_err
	}

	extranet_user_permiss := make(map[int]map[int]int, 0)

	for extranet_user_roles.Next() {


		extranet_user_permiss_info := make(map[int]int, 0)

		var id_users, id_roles_list, id_hotels int
		if err := extranet_user_roles.Scan(&id_users, &id_roles_list, &id_hotels); err != nil {
			continue
		}

		extranet_user_permiss_info[id_hotels] = id_roles_list

		extranet_user_permiss[id_users] = extranet_user_permiss_info
	}


	crm_permission.extranet_permissions = extranet_user_permiss

	return nil
}

//проверка на конкретную операцию CRM
func checkAction(permiss interface{}, action string) bool {
	convert := permiss.(map[string]interface{})
	switch action {
	case CREATE_ACTION:
		if convert["allow_create"].(int) == 1 {
			return true
		} else {
			return false
		}
	case DELETE_ACTION:
		if convert["allow_delete"].(int) == 1 {
			return true
		} else {
			return false
		}
	case EDIT_ACTION:
		if convert["allow_edit"].(int) == 1 {
			return true
		} else {
			return false
		}
	case VIEW_ACTION:
		return true
	default:
		return false
	}
}

//проверка на конкретную операцию Extranet
func checkActionExtranet(permiss roles, action string) bool {

	switch action {
	case CREATE_ACTION:
		if permiss.allow_create == 1 {
			return true
		} else {
			return false
		}
	case DELETE_ACTION:
		if permiss.allow_delete == 1 {
			return true
		} else {
			return false
		}
	case EDIT_ACTION:
		if permiss.allow_edit == 1 {
			return true
		} else {
			return false
		}
	case VIEW_ACTION:
		return true
	default:
		return false
	}
}

func init() {
	services.AddService(crm_permission.name, crm_permission)
}
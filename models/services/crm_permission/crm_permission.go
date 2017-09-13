// Copyright 2017 Author: Yurii Kravchuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//Package crm_permission Реализует работу с правами пользователя для доступа в CRM/Extranet
package crm_permission

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/services"
	"sync"
)

//структура прав для пользователя по ссылке (CRM)
type linkPermission struct {
	link string
	allowCreate int
	allowDelete int
	allowEdit int
	idUsers int
}

//структура прав роли по ссылке (Extranet)
type roles struct {
	link string
	allowCreate int
	allowDelete int
	allowEdit int
}

type cpService struct {
	name string
	region string
	status string
	crmPermissionsRoles map[int][]interface{}
	extranetRoles map[int][]roles
	extranetPermissions map[int]map[int]int
}

//константы для работы с сервисом
const (
	//Создание
	CREATE_ACTION = "Create"
	//Удаление
	DELETE_ACTION = "Delete"
	//Редактирование
	EDIT_ACTION = "Edit"
	//Просмотр
	VIEW_ACTION = "View"
	//Устрановление прав
	SET_PERMISS = "Set"
	//Удаление прав
	DROP_PERMISS = "Drop"
	//Часть CRM
	CRM_PART = "crm"
	//Часть Extranet
	EXTRANET_PART = "extranet"
)

var crm_permission *cpService = &cpService{name:"crm_permission", status: "create"}
var cacheMu sync.RWMutex

//реализация обязательных методов интерейса
func (crm_permission *cpService) Init() error{

	crm_permission.status = "init"

	err := crm_permission.setUserPermissionForCRM()

	if err != nil {
		return err
	}

	rolesErr := crm_permission.setExtranetRoles()

	if rolesErr != nil {
		return rolesErr
	}

	userRolesErr := crm_permission.setExtranetUserRoles()

	if userRolesErr != nil {
		return userRolesErr
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
		return services.ErrBrokenConnection{Name: crm_permission.name, Param: args}
	}

	if len(args) < 4 {
		return services.ErrServiceNotEnoughParameter{Name: crm_permission.name, Param: args}
	}
	if _,ok := args[1].(int); !ok {
		return services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[1], Number: 2}
	}
	if _,ok := args[2].(string); !ok {
		return services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[2], Number: 3}
	}

	switch permissionType := args[0].(type) {
	case string:
		if permissionType == CRM_PART {
			if args[3].(string) == SET_PERMISS {
				return crm_permission.setPermissForUser(args[1].(int), args[2].(string), args[4].(bool), args[5].(bool), args[6].(bool));
			} else if args[3].(string) == DROP_PERMISS {
				return crm_permission.deletePermissForUser(args[1].(int), args[2].(string));
			}
		} else if permissionType == EXTRANET_PART {
			if args[3].(string) == SET_PERMISS {
				return crm_permission.setPermissForUserExtranet(args[1].(int), args[4].(int), args[5].(int));
			} else if args[3].(string) == DROP_PERMISS {
				return crm_permission.deletePermissForUserExtranet(args[1].(int), args[4].(int));
			}
		}
	default:
		return services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: permissionType, Number: 1}
	}

	return services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}

}

// args: 0 => admin part, 1 => user id, 2 => url what test on permiss, 3 => action for test access (Create/Delete/Edit/View)
//for Extranet 4 => id_hotels
func (crm_permission *cpService) Get(args ... interface{}) ( interface{}, error) {

	if len(args) < 4 {
		return nil, services.ErrServiceNotEnoughParameter{Name: crm_permission.name, Param: args}
	}
	if _,ok := args[1].(int); !ok {
		return nil, services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[1], Number: 2}
	}

	connection_status := services.Status("crm_permission")

	if connection_status != "ready" {
		return nil, services.ErrBrokenConnection{Name: crm_permission.name, Param: args}
	}

	switch permissionType := args[0].(type) {
	case string:
		if permissionType == CRM_PART {
			return crm_permission.getCRMPermissions(args[1].(int), args[2].(string), args[3].(string)), nil;
		} else if permissionType == EXTRANET_PART {

			if _,ok := args[4].(int); !ok {
				return nil, services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: args[4], Number: 2}
			}

			return crm_permission.getExtranetPermissions(args[1].(int), args[2].(string),
				args[3].(string), args[4].(int)), nil;
		}
	default:
		return nil, services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: permissionType, Number: 1}
	}

	return nil, services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
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

	if crm_permission.crmPermissionsRoles[user_id] == nil || len(crm_permission.crmPermissionsRoles[user_id]) == 0 {
		return false
	}

	for _,permission := range crm_permission.crmPermissionsRoles[user_id] {
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

	if role_id == 0 || crm_permission.extranetRoles[role_id] == nil {
		return false
	}

	for _,permission := range crm_permission.extranetRoles[role_id] {
		if permission.link == url {
			return checkActionExtranet(permission, action)
		}
	}

	return false

}

//получение роли пользователя для Extranet для конкретного отеля
func (crm_permission *cpService) getUserRole(user_id, id_hotels int) int {

	for hotel,role := range crm_permission.extranetPermissions[user_id] {
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

	if crm_permission.crmPermissionsRoles[user_id] == nil || len(crm_permission.crmPermissionsRoles[user_id]) == 0 {
		return services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
	}

	for key,permission := range crm_permission.crmPermissionsRoles[user_id] {
		resRow := permission.(map[string]interface{})
		if resRow["link"].(string) == url {
			crm_permission.crmPermissionsRoles[user_id] = append(crm_permission.crmPermissionsRoles[user_id][:key],
				crm_permission.crmPermissionsRoles[user_id][key+1:]...)
			return nil
		}
	}

	cacheMu.Unlock()

	return services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
}

//выставление роли для пользователя в CRM
func (crm_permission *cpService) setPermissForUser(user_id int, link string, allowCreate, allowDelete, allowEdit bool) error {

	cacheMu.Lock()
	// TODO: Unlock must to allow there as defer func

	newRow := make(map[string]interface{}, 0)
	newRow["link"] = link

	if allowCreate {
		newRow["allow_create"] = 1
	} else {
		newRow["allow_create"] = 0
	}

	if allowDelete {
		newRow["allow_delete"] = 1
	} else {
		newRow["allow_delete"] = 0
	}

	if allowEdit {
		newRow["allow_edit"] = 1
	} else {
		newRow["allow_edit"] = 0
	}

	crm_permission.crmPermissionsRoles[user_id] = append(crm_permission.crmPermissionsRoles[user_id], newRow)

	cacheMu.Unlock()
	return nil
}

//удаление роли для пользователя в Extranet для конкретного отеля
func (crm_permission *cpService) deletePermissForUserExtranet(user_id int, id_hotels int) error {

	cacheMu.Lock()
	// TODO: Unlock must to allow there as defer func

	if crm_permission.extranetPermissions[user_id] == nil || len(crm_permission.extranetPermissions[user_id]) == 0 {
		return services.ErrServiceNotCorrectParamType{Name: crm_permission.name, Param: "", Number: 1}
	}

	crm_permission.extranetPermissions[user_id][id_hotels] = 0

	cacheMu.Unlock()

	return nil
}

//выставление роли для пользователя в Extranet для конкретного отеля
func (crm_permission *cpService) setPermissForUserExtranet(user_id, id_hotels, id_role int) error {

	cacheMu.Lock()
	// TODO: Unlock must to allow there as defer func

	crm_permission.extranetPermissions[user_id][id_hotels] = id_role

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

	crmPermissionsRoles := make(map[int][]interface{}, 0)
	for rows.Next() {
		var link string
		var allowCreate, allowDelete, allowEdit, idUsers int
		if err := rows.Scan(&link, &allowCreate, &allowDelete, &allowEdit, &idUsers); err != nil {
			continue
		}

		newRow := make(map[string]interface{}, 0)
		newRow["link"] = link
		newRow["allow_create"] = allowCreate
		newRow["allow_delete"] = allowDelete
		newRow["allow_edit"] = allowEdit

		crmPermissionsRoles[idUsers] = append(crmPermissionsRoles[idUsers], newRow)
	}

	crm_permission.crmPermissionsRoles = crmPermissionsRoles

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

	extranetRoles := make(map[int][]roles, 0)

	for roles_rows.Next() {
		var extranetRole roles

		var link string
		var allowCreate, allowDelete, allowEdit, id_role int
		if err := roles_rows.Scan(&id_role, &link, &allowCreate, &allowDelete, &allowEdit); err != nil {
			continue
		}

		extranetRole.link = link
		extranetRole.allowCreate = allowCreate
		extranetRole.allowDelete = allowDelete
		extranetRole.allowEdit = allowEdit

		extranetRoles[id_role] = append(extranetRoles[id_role], extranetRole)
	}

	crm_permission.extranetRoles = extranetRoles

	return nil
}

//заполнение масива ролей для Extranet
func (crm_permission *cpService) setExtranetUserRoles() error {
	extranetUserRoles, extranetUserRolesErr := db.DoSelect("SELECT id_users, id_roles_list, id_hotels FROM " +
		"users_roles_list_has_extranet")

	if extranetUserRolesErr != nil {
		return extranetUserRolesErr
	}

	extranetUserPermiss := make(map[int]map[int]int, 0)

	for extranetUserRoles.Next() {


		extranetUserPermissInfo := make(map[int]int, 0)

		var idUsers, idRolesList, idHotels int
		if err := extranetUserRoles.Scan(&idUsers, &idRolesList, &idHotels); err != nil {
			continue
		}

		extranetUserPermissInfo[idHotels] = idRolesList

		extranetUserPermiss[idUsers] = extranetUserPermissInfo
	}


	crm_permission.extranetPermissions = extranetUserPermiss

	return nil
}

//проверка на конкретную операцию CRM
func checkAction(permiss interface{}, action string) bool {
	convert := permiss.(map[string]interface{})
	switch action {
	case CREATE_ACTION:
		if convert["allow_create"].(int) == 1 {
			return true
		}
		return false
	case DELETE_ACTION:
		if convert["allow_delete"].(int) == 1 {
			return true
		}
		return false
	case EDIT_ACTION:
		if convert["allow_edit"].(int) == 1 {
			return true
		}
		return false

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
		if permiss.allowCreate == 1 {
			return true
		}
		return false
	case DELETE_ACTION:
		if permiss.allowDelete == 1 {
			return true
		}
		return false
	case EDIT_ACTION:
		if permiss.allowEdit	 == 1 {
			return true
		}
		return false
	case VIEW_ACTION:
		return true
	default:
		return false
	}
}

func init() {
	services.AddService(crm_permission.name, crm_permission)
}
// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"log"
)

type pRoles struct {
	idUser int
	idRoles string
}
type rowsRoles [] *pRoles
type pService struct {
	name string
	region string
	status string
	Rows map[string] rowsRoles
	roles [] string
}
var permissions *pService = &pService{name:"permission"}

//реализация обязательных методов интерейса
func (permissions *pService) Init() error{

	permissions.Rows = make(map[string] rowsRoles, 0)
	rows, err := db.DoSelect("SELECT p.id_users, g.title FROM permissions p JOIN permission_group_list g ")
	if err != nil {
		return err
	}

	roles := make(rowsRoles, 0)
	for rows.Next() {
		var row *pRoles
		if err := rows.Scan(&row); err != nil {
			continue
		}

		roles = append(roles, row)
	}

	permissions.Rows["system"] = roles

	permissions.status = "ready"
	return nil
}
func (permissions *pService) Send(messages ...interface{}) error {

	for _, message := range messages {
		switch mess := message.(type) {
		case string:
			log.Println(mess)
		default:
			log.Println(mess)
		}
	}

	return nil

}
func (permissions *pService) Get(messages ... interface{}) ( interface{}, error) {

	responce := make(map[string] bool)
	for _, message := range messages {
		for _, mess1 := range message.([] interface{}) {
			switch mess := mess1.(type) {
			case map[string]string:
				for key, val := range mess {
					pRole, ok := permissions.Rows[key]
					responce[key] = ok
					log.Println(val, pRole)
				}
			case []interface{}:
				log.Println(mess)
				continue
			case string:
				responce[mess] = true
			default:
				log.Println(mess)
				responce["Unknow type"] = false

			}
		}
	}

	return responce, nil
}
func (permissions *pService) Connect(in <- chan interface{}) (out chan interface{}, err error) {
	out = make(chan interface{})

	go func() {
		out<-"open"
		for {
			select {
			case v := <-in:
				if v.(string) == "close" {
					permissions.Close(out)
				} else {
					out <- v
				}
			}
		}
	}()
	return out, nil
}
func (permissions *pService) Close(out chan <- interface{}) error {
	close(out)
	return nil

}
func (permissions *pService) Status() string {
	return permissions.status
}

//func (permissions *pService) RegPermission(name, tableName string) error {
//	_, ok := permissions.Rows[name]
//	if !ok {
//		temp := make(rowsRoles,0)
//		rows, err := db.DoSelect("SELECT * FROM " + tableName)
//		if err != nil {
//			return err
//		}
//		for rows.Next() {
//			var User int, Roles string
//
//			rows.Scan(&User, &Roles)
//			temp = append(temp, &pRoles{idUser:User, idRoles:Roles})
//
//		}
//		permissions.Rows[name] = temp
//	}
//
//	return nil
//}
func (permissions *pService) GetPermission(name string, idUser int) error {
	perm, ok := permissions.Rows[name]
	if ok {
		for i, val := range perm {
log.Println(i,val)
		}

	}

	return nil
}
func init() {
	AddService(permissions.name, permissions)
}
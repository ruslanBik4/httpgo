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
	idRoles int
}
type rowsRoles [] *pRoles
type pService struct {
	name string
	region string
	Rows map[string] rowsRoles
	roles [] string
}
var permissions *pService

func (*pService) init() error{

	permissions.Rows = make(map[string] rowsRoles, 0)
	rows, err := db.DoSelect("SELECT * FROM permissions")
	if err != nil {
		return err
	}

	for rows.Next() {

	}

	return nil
}
func (*pService) RegPermission(name, tableName string) error {
	_, ok := permissions.Rows[name]
	if !ok {
		temp := make(rowsRoles,0)
		rows, err := db.DoSelect("SELECT * FROM " + tableName)
		if err != nil {
			return err
		}
		for rows.Next() {
			var User, Roles int

			rows.Scan(&User, &Roles)
			temp = append(temp, &pRoles{idUser:User, idRoles:Roles})

		}
		permissions.Rows[name] = temp
	}

	return nil
}
func (*pService) GetPermission(name string, idUser int) error {
	perm, ok := permissions.Rows[name]
	if ok {
		for i, val := range perm {
log.Println(i,val)
		}

	}

	return nil
}
func init() {
	AddService("permission", permissions)
}
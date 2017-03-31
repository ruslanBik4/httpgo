// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import "github.com/ruslanBik4/httpgo/models/db"

type pService struct {
	name string
	region string
	roles [] string
}
var permission *pService

func (*pService) init() error{

	rows, err := db.DoSelect("select * from permission")
	if err != nil {
		return err
	}

	for rows.Next() {

	}

	return nil
}
func init() {
	AddService("permission", permission)
}
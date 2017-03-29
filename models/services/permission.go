// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import "github.com/ruslanBik4/httpgo/models/db"

type permission struct {
	name string
	region string
}
type pService struct {

}

func (* pService) init() {

	rows, _ := db.DoSelect("select * from permission")

	for rows.Next() {

	}
}
// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Creating %{DATE}

//Обслуживает кеш для справочников БД

package services

import (
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/db/cache"
)

type DBlistsService struct {
	name   string
	status string
}

var (
	DBlists *DBlistsService = &DBlistsService{name: "DBlists"}
)

func (DBlists *DBlistsService) Init() error {
	DBlists.status = "starting"
	db.InitLists()
	DBlists.status = "ready"

	return nil
}
func (DBlists *DBlistsService) Send(messages ...interface{}) error {
	return nil

}
func (DBlists *DBlistsService) Get(messages ...interface{}) (responce interface{}, err error) {
	switch tableName := messages[0].(type) {
	case string:
		return cache.GetListRecord(tableName), nil
	default:
		return nil, &ErrServiceNotCorrectParamType{Name: schema.name, Param: messages[0]}
	}

	return nil, nil

}
func (DBlists *DBlistsService) Connect(in <-chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (DBlists *DBlistsService) Close(out chan<- interface{}) error {

	return nil
}
func (DBlists *DBlistsService) Status() string {

	return ""
}

func init() {
	AddService(DBlists.name, DBlists)
}

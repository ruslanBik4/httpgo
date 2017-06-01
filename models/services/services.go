// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"github.com/ruslanBik4/httpgo/models/logs"
)

type IService interface {
	Init() error
	Send(messages ...interface{}) error
	Get(messages ... interface{}) (responce interface{}, err error)
	Connect(in <- chan interface{}) (out chan interface{}, err error)
	Close(out chan <- interface{}) error
	Status() string
}

type rootServices struct {
  services map[string]IService
}

var sServices = &rootServices{ services : make(map[string]IService, 0) }

func InitServices() *rootServices {
	for name, service := range sServices.services {
		go startService(name, service)
	}

	return sServices
}
//получение сервиса по имени
func getService(name string) (pServ IService) {
	pServ, ok := sServices.services[name]
	if !ok {
		return nil
	}
	return pServ
}
func startService(name string, pService IService) {

	defer catch(name)
	if err := pService.Init(); err != nil {
		logs.ErrorLog(err, name)
	} else {
		logs.StatusLog(name,   " starting ",", Status - ", pService.Status())
	}
}
func catch(name string) {
	err := recover()

	switch err.(type) {
	case ErrServiceNotFound:
		logs.ErrorLog(err.(error), name)
	case nil:
	default:
		logs.ErrorLog(err.(error), name)
	}
}
func AddService(name string, pService IService) {
	sServices.services[name] = pService
}
func Send(name string, messages ...interface{}) (err error) {

	pService := getService(name)
	if pService == nil {
		return 	&ErrServiceNotFound{Name: name}
	}

	return pService.Send(messages...)
}
func Get(name string, messages ... interface{}) (responce interface{}, err error) {

	pService := getService(name)
	if pService == nil {
		return 	nil, &ErrServiceNotFound{Name: name}
	}

	return pService.Get(messages...)
}
func Connect(name string, in <- chan interface{}) (out chan  interface{}, err error) {
	pService := getService(name)
	if pService == nil {
		return 	nil, &ErrServiceNotFound{Name: name}
	}

	return pService.Connect(in)
}
func Close(name string, out chan <- interface{}) error {
	pService := getService(name)
	if pService == nil {
		return 	&ErrServiceNotFound{Name: name}
	}

	return pService.Close(out)
}
func Status(name string ) string {
	pService := getService(name)
	if pService == nil {
		return name + MessServNotFound
	}

	return pService.Status()
}

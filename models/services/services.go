// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"log"
)

type IService interface {
	init() error
}

type rootServices struct {
  services map[string]IService
}

var sServices = &rootServices{ services : make(map[string]IService, 0) }

func InitServices() *rootServices {
	for name, service := range sServices.services {
		go startService(name, service)
		log.Println(name)
	}

	return sServices
}
func GetService(name string) (pServ IService) {
	pServ, ok := sServices.services[name]
	if !ok {
		return nil
	}
	return pServ
}
func startService(name string, pService IService) {

	defer Catch(name)
	if err := pService.init(); err != nil {
		log.Println(err, name)
	}
}
func Catch(name string) {
	err := recover()

	switch err.(type) {
	case nil:
	default:
		log.Println(err, name)
	}
}
func AddService(name string, pService IService) {
	sServices.services[name] = pService
}
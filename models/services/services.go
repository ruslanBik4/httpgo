// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import "log"

type iService interface {
	init() error
}

type rootServices struct {
  services map[string] iService
}

var sServices = &rootServices{ services : make(map[string] iService, 0) }

func InitServices() *rootServices {
	for name, service := range sServices.services {
		service.init()
		log.Println(name)
	}

	return sServices
}
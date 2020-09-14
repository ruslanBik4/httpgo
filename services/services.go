// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package services для подключения сервисов и устроения взаимодействия сервисов с главным потоком
package services

import (
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/logs"
)

// IService root service interface
type IService interface {
	Init(ctx context.Context) error
	Send(ctx context.Context, messages ...interface{}) error
	Get(ctx context.Context, messages ...interface{}) (response interface{}, err error)
	Connect(in <-chan interface{}) (out chan interface{}, err error)
	Close(out chan<- interface{}) error
	Status() string
}

// IChildService interface of service with parent dependencies
type IChildService interface {
	Dependencies() []IService
	IsReadyToStart() bool
}

type rootServices struct {
	services map[string]IService
}

var sServices = &rootServices{services: make(map[string]IService, 0)}

// InitServices started all services from sServices.services in some goroutins
func InitServices(ctx context.Context, list ...string) *rootServices {
	if len(list) > 0 {
		for _, name := range list {
			if service, ok := sServices.services[name]; ok {
				go startService(ctx, name, service)
			} else {
				logs.ErrorLog(ErrServiceNotFound{Name: name})
			}
		}
	} else {
		for name, service := range sServices.services {
			go startService(ctx, name, service)
		}
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

func startService(ctx context.Context, name string, pService IService) {

	defer catch(name)
	if iChild, ok := pService.(IChildService); ok && !iChild.IsReadyToStart() {
		logs.ErrorLog(errors.New("not ready parent services"), iChild)
	} else if pService.Status() == "" {
		logs.DebugLog("attempt to restart the service %s", name)
	} else if err := pService.Init(ctx); err != nil {
		logs.ErrorLog(err, name)
		logs.StatusLog("[[%s]]; not starting, Status - %s", name, pService.Status())
	} else {
		logs.StatusLog("[[%s]]; starting, Status - %s", name, pService.Status())
	}
}

func catch(name string) {
	result := recover()

	switch err := result.(type) {
	case ErrServiceNotFound:
		logs.ErrorLogHandler(err, name)
	case nil:
	case error:
		logs.ErrorLogHandler(err, name)
	}
}

// AddService adding new service with {name} to services list
func AddService(name string, pService IService) {
	sServices.services[name] = pService
}

// Send messages to service {name}
func Send(ctx context.Context, name string, messages ...interface{}) (err error) {

	pService := getService(name)
	if pService == nil {
		return &ErrServiceNotFound{Name: name}
	}

	if pService.Status() != "ready" {
		return &ErrServiceNotReady{Name: name}
	}

	for pService.Status() == "starting" {
		if status := pService.Status(); strings.HasPrefix(status, "failed") {
			return &ErrServiceNotReady{Name: name, Status: status}
		}
	}

	return pService.Send(ctx, messages...)
}

// Get messages to service {name} & return result
func Get(ctx context.Context, name string, messages ...interface{}) (response interface{}, err error) {

	pService := getService(name)
	if pService == nil {
		return nil, &ErrServiceNotFound{Name: name}
	}

	return pService.Get(ctx, messages...)
}

// Connect to service {name} from channel in & return channel service
func Connect(name string, in <-chan interface{}) (out chan interface{}, err error) {
	pService := getService(name)
	if pService == nil {
		return nil, &ErrServiceNotFound{Name: name}
	}

	return pService.Connect(in)
}

// Close service {name}
func Close(name string, out chan<- interface{}) error {
	pService := getService(name)
	if pService == nil {
		return &ErrServiceNotFound{Name: name}
	}

	return pService.Close(out)
}

// Status service {name} return
func Status(name string) string {
	pService := getService(name)
	if pService == nil {
		return name + messServNotFound
	}

	return pService.Status()
}

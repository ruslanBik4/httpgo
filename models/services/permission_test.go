// Copyright 2017 Author: Yurii Kravchuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"flag"
	"testing"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/models/services/crmPermission"
)

const permissName = "crmPermission"

var (
	fStatic  = flag.String("path", "/opt/lampp/htdocs/go_src/src/github.com/ruslanBik4/httpgo", "path to static files")
	fWeb     = flag.String("web", "/opt/lampp/htdocs/travel/web", "path to web files")
	fSession = flag.String("sessionPath", "/opt/lampp/htdocs/go_sessions", "path to store sessions data")
)

func TestPermissSend(t *testing.T) {

	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(fStatic, fWeb, fSession); err != nil {
		t.Error(err)
	}

	service := getService(permissName)
	startService(permissName, service)

	var result interface{}
	result = Send(permissName, "crm", 8, "/admin/business/", "set", false, false, false)
	if result == nil {
		t.Skipped()
		logs.DebugLog("result", result)
		return
	}

	switch err := result.(type) {

	case ErrServiceNotCorrectParamType:
		t.Errorf("Error - %s, parameter #%d - %v", err.Error(), err.Number, err.Param)
	case error:
		t.Error("Not correct error type - " + err.Error())
	default:
		t.Error("Not correct error type - ")
	}
}

func TestPermissSendExtranet(t *testing.T) {

	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(fStatic, fWeb, fSession); err != nil {
		t.Error(err)
	}

	service := getService(permissName)
	startService(permissName, service)

	var result interface{}
	result = Send(permissName, crmPermission.EXTRANET_PART, 8, "/admin/business/", crmPermission.DROP_PERMISS, 52)
	if result == nil {
		t.Skipped()
		logs.DebugLog("result", result)
		return
	}

	switch err := result.(type) {

	case ErrServiceNotCorrectParamType:
		t.Errorf("Error - %s, parameter #%d - %v", err.Error(), err.Number, err.Param)
	case error:
		t.Error("Not correct error type - " + err.Error())
	default:
		t.Error("Not correct error type - ")
	}
}

func TestPermissSendWrongParam(t *testing.T) {

	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(fStatic, fWeb, fSession); err != nil {
		t.Error(err)
	}

	service := getService(permissName)
	startService(permissName, service)

	var result interface{}
	result = Send(permissName, "crm", "8", "/admin/business/", "set", false, false, false)
	if result == nil {
		t.Skipped()
		logs.DebugLog("result", result)
		return
	}

	switch err := result.(type) {

	case ErrServiceNotCorrectParamType:
		t.Skipped()
		logs.DebugLog("result", result)
		return
	case error:
		t.Error("Not correct error type - " + err.Error())
	default:
		t.Error("Not correct error type - ")
	}
}

func TestPermissGet(t *testing.T) {

	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(fStatic, fWeb, fSession); err != nil {
		t.Error(err)
	}

	service := getService(permissName)
	startService(permissName, service)

	result, sErr := Get(permissName, "crm", 8, "/admin/business/", "Create")

	switch err := sErr.(type) {

	case ErrServiceNotEnoughParameter:
		t.Skipped()
	case ErrServiceNotCorrectParamType:
		t.Errorf("Error - %s, parameter #%d - %v", err.Error(), err.Number, err.Param)
	case nil:
		t.Skipped()
		logs.DebugLog("result", result)
		return
	default:
		t.Error("Not correct error type - ")
	}
}

func TestPermissGetExtranet(t *testing.T) {

	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(fStatic, fWeb, fSession); err != nil {
		t.Error(err)
	}

	service := getService(permissName)
	startService(permissName, service)

	result, sErr := Get(permissName, "extranet", 8, "/payment_rules", "Delete", 52)

	switch err := sErr.(type) {

	case ErrServiceNotEnoughParameter:
		t.Skipped()
	case ErrServiceNotCorrectParamType:
		t.Errorf("Error - %s, parameter #%d - %v", err.Error(), err.Number, err.Param)
	case nil:
		t.Skipped()
		logs.DebugLog("result", result)
		return
	default:
		t.Error("Not correct error type - ")
	}
}

func TestPermissGetWrongParam(t *testing.T) {

	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(fStatic, fWeb, fSession); err != nil {
		t.Error(err)
	}

	service := getService(permissName)
	startService(permissName, service)

	result, sErr := Get(permissName, "crm", "8", "/admin/business/", "Create")

	switch err := sErr.(type) {

	case ErrServiceNotEnoughParameter:
		t.Errorf("Error - %s, parameter %s = %v", err.Error(), err.Name, err.Param)
	case ErrServiceNotCorrectParamType:
		t.Skipped()
		logs.DebugLog("result", result)
		return
	case nil:
		t.Errorf("No error on wrong params count")
	default:
		t.Error("Not correct error type - ")
	}
}

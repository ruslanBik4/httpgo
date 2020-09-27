// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"log"
	"testing"

	"golang.org/x/net/context"

	"github.com/ruslanBik4/logs"
)

const serviceName = "photos"

func TestPhotosSend(t *testing.T) {

	var result interface{}
	result = Send(context.TODO(), serviceName, "open", 3)
	if result == nil {
		t.Error("Not error by operation with error name")
	}
	switch err := result.(type) {

	case ErrServiceNotCorrectOperation:
		t.Log(err.Error())
		t.Skipped()
	case error:
		t.Error("Not correct error type - " + err.Error())
	default:
		t.Error("Not correct error type - ")
	}
}

func TestPhotosGetList(t *testing.T) {

	//var iErr interface{}
	result, iErr := Get(context.TODO(), serviceName, "rooms", "1", 1)

	switch err := iErr.(type) {

	case ErrServiceNotCorrectOperation:
		t.Skipped()
	case ErrServiceNotCorrectParamType:
		t.Errorf("Error - %s, parameter #%d - %v", err.Error(), err.Number, err.Param)
	case ErrServiceWrongIndex:
		t.Errorf("Wrong index %d", err.Index)
	case nil:
		t.Skipped()
		logs.DebugLog("result", result)
		return

	default:
		t.Error("Not correct error type - ")
		log.Println(err.(ErrServiceNotCorrectParamType).Param)
	}
}

// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"testing"
)

const serviceName = "photos"

func TestPhotosSend(t *testing.T) {

	var result interface{}
	result = Send(serviceName, "open", 3)
	if result == nil {
		t.Error("Not error by operation with error name")
	}
	switch err := result.(type){

	case ErrServiceNotCorrectOperation:
		t.Skipped()
	case error:
		t.Error("Not correct error type - " + err.Error())
	default:
		t.Error("Not correct error type - " )
	}
}


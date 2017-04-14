// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"testing"
)

func TestConnect(t *testing.T) {

	nameServ := "permission"

	t.Log("before connect")
	in := make(chan interface{})
	out, err := Connect(nameServ, in)

	if err != nil {
		t.Errorf("error initialization: filename=%s, error=%q", nameServ, err)
		return
	}

	for {
		select {
		case v := <-out:
			switch v.(string) {
			case "open":
				in <- "first"
			case "first":
				in <- "second"
			case "second":
				in <- "close"
				//close(in)
				return
			default:
				t.Log(v)
			}
			t.Log(v)
		//default:
		//	t.Log("-")
		}
	}

	t.Skipped()

}
func TestSend(t *testing.T) {
	nameServ := "permission"
	err := Send(nameServ, "open")
	if err != nil {
		t.Error(err)
	}
	t.Skipped()
}
func TestGet(t *testing.T) {
	nameServ := "permission"
	messages := make(map[string] string, 0)
	messages["system"] = "10"
	responce, err := Get(nameServ, messages, "stress")
	if err != nil {
		t.Error(err)
	} else if responce == nil {
		t.Errorf("Get return nil !")
	} else {
		t.Log(responce)

	}
	t.Skipped()
}
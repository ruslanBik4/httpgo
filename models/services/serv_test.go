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
t.Log(out)
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
		default:
			t.Log("-")
		}
	}

	t.Skipped()

}
func TestSend(t *testing.T) {
	nameServ := "permission"
	err :=Send(nameServ, "close")
	if err != nil {
		t.Error(err)
	}
	t.Skipped()
}
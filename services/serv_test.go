// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
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
			if !assert.IsType(t, "", v) {
				t.FailNow()
			}
			switch v.(string) {
			case "open":
				in <- "first"
			case "first":
				in <- "second"
			case "second":
				in <- "close"
				//close(in)
				t.Skipped()
				break
			default:
				t.Log(v)
			}
			t.Log(v)
			//default:
			//	t.Log("-")
		}
	}

}
func TestSend(t *testing.T) {
	nameServ := "permission"
	err := Send(context.TODO(), nameServ, "open")
	if err != nil {
		t.Error(err)
	}
	t.Skipped()
}
func TestGet(t *testing.T) {
	nameServ := "permission"
	messages := make(map[string]string, 0)
	messages["system"] = "10"
	response, err := Get(context.TODO(), nameServ, messages, "stress")
	if err != nil {
		t.Error(err)
	} else if response == nil {
		t.Errorf("Get return nil !")
	} else {
		t.Log(response)

	}
	t.Skipped()
}

func TestModConnect(t *testing.T) {

	nameServ := "moderation"

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
				t.Skipped()
				break
			default:
				t.Log(v)
			}
			t.Log(v)
			//default:
			//	t.Log("-")
		}
	}

}

func TestModSendInsert(t *testing.T) {
	var config = make(map[string]string, 0)
	config["table"] = "test2"
	config["key"] = "3333"
	config["action"] = "insert"
	var a []url.Values
	result := make(map[string][]string, 0)
	result["key"] = []string{
		"11",
	}
	a = append(a, result)

	//result = append(result, config)
	//result = append(result, a)

	defer func() {
		err := recover()
		if err != nil {
			t.Error(err)
		}
	}()
	err := Send(context.TODO(), "moderation", config, a)
	if err != nil {
		t.Error(err)
	}

	t.Skipped()
}

func TestModSendDelete(t *testing.T) {
	var config = make(map[string]string, 0)
	config["table"] = "test2"
	config["key"] = "72"
	config["action"] = "delete"

	result := make([]interface{}, 0)

	result = append(result, config)

	Send(context.TODO(), "moderation", result)
	t.Skipped()
}

func TestModGet(t *testing.T) {
	var config = make(map[string]string, 0)
	config["table"] = "test2"
	config["key"] = "72"

	response, err := Get(context.TODO(), "moderation", config)

	if err != nil {
		t.Error(err)
	} else if response == nil {
		t.Errorf("Get return nil !")
	} else {
		t.Log(response)
		t.Skipped()

	}
}

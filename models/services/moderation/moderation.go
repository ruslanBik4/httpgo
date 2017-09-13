// Copyright 2017 Author: Yurii Kravchuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//Package moderation Реализует работу с записями для последуйщей модерации
package moderation

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/services"
	"github.com/ruslanBik4/httpgo/models/services/mongod"
	mongo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/url"
)

var (
	moderation *mService = &mService{name: "moderation"}
)

//Record являет собой структуру по получению данных для модерации
type Record struct {
	Config map[string]string
	Data   []url.Values
}

//Struct вляет собой структуру для записи модерации данных
type Struct struct {
	Key  string
	Data string
}

type mService struct {
	name    string
	connect *mongo.Session
	status  string
}

func (moderation *mService) Init() error {

	session, err := mongo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}

	session.SetMode(mongo.Monotonic, true)

	moderation.connect = session
	moderation.status = "ready"
	return nil
}

func (moderation *mService) Connect(in <-chan interface{}) (out chan interface{}, err error) {
	out = make(chan interface{})

	go func() {
		out <- "open"
		for {
			select {
			case v := <-in:
				if v.(string) == "close" {
					moderation.Close(out)
				} else {
					out <- v
				}
			}
		}
	}()
	return out, nil
}

func (moderation *mService) Close(out chan<- interface{}) error {
	close(out)
	return nil
}

func (moderation *mService) Status() string {
	return moderation.status
}

//отправка данных на хранение для последуйщей модерации
func (moderation *mService) Send(messages ...interface{}) error {

	setData := Record{
		Config: make(map[string]string, 0),
		Data:   make([]url.Values, 0),
	}

	for _, message := range messages {
		switch mess := message.(type) {
		case map[string]string:
			setData.Config["table"] = mess["table"]
			setData.Config["key"] = mess["key"]
			setData.Config["action"] = mess["action"]
		case []url.Values:
			setData.Data = mess
		default:

			return &services.ErrServiceNotCorrectParamType{
				Name: moderation.name,
			}

		}
	}

	if setData.Config["table"] == "" || setData.Config["key"] == "" ||
		(setData.Config["action"] != "insert" && setData.Config["action"] != "delete") {

		return &services.ErrServiceNotCorrectParamType{
			Name: moderation.name,
		}
	}

	cConnect := mongod.GetMongoCollectionConnect(setData.Config["table"])

	if setData.Config["action"] == "delete" {
		//err := cConnect.Remove(bson.M{"key": setData.Config["key"]})
		err := services.Send("mongod", setData.Config["table"], "Remove", bson.M{"key": setData.Config["key"]})
		if err != nil {
			return err
		}

		return nil
	}

	checkRow := Struct{}
	err := cConnect.Find(bson.M{"key": setData.Config["key"]}).One(&checkRow)

	if checkRow.Data != "" {
		return &services.ErrServiceNotCorrectParamType{
			Name: moderation.name,
		}
	}

	data := toGoB64(setData.Data)

	err = cConnect.Insert(&Struct{setData.Config["key"], data})

	if err != nil {
		return err
	}

	return nil
}

//получение данных для модерации
func (moderation *mService) Get(messages ...interface{}) (interface{}, error) {

	getData := Record{
		Config: make(map[string]string),
		Data:   make([]url.Values, 0),
	}

	for _, message := range messages {
		switch mess := message.(type) {
		case map[string]string:
			getData.Config["table"] = mess["table"]
			getData.Config["key"] = mess["key"]
		}
	}

	//cConnect := moderation.connect.DB("newDB").C(getData.Config["table"])
	cConnect := mongod.GetMongoCollectionConnect(getData.Config["table"])

	response := Struct{}

	err := cConnect.Find(bson.M{"key": getData.Config["key"]}).One(&response)

	if err != nil {
		return nil, err
	}

	data := fromGoB64(response.Data)

	return data, nil
}

//GetMongoConnection для получение связи к монго
func GetMongoConnection() *mongo.Session {
	return moderation.connect
}

// go binary encoder
func toGoB64(m []url.Values) string {

	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(m); err != nil {
		fmt.Println(`failed gob Encode`, err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

// go binary decoder
func fromGoB64(str string) []url.Values {

	var m []url.Values
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println(`failed base64 Decode`, err)
	}
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	if err := d.Decode(&m); err != nil {
		fmt.Println(`failed gob Decode`, err)
	}
	return m
}

func init() {
	services.AddService(moderation.name, moderation)
}

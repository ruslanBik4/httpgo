// Copyright 2017 Author: Yurii Kravchuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//Package mongod Реализует работу с базой данных mongodb
package mongod

import (
	mongo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/models/services"
)

var (
	mongod *mdService = &mdService{name: "mongod"}
)

type mdService struct {
	name    string
	connect *mongo.Session
	status  string
}

//Запускает связь с mongodb на дефолтный порт. (localhost:27017)
//TODO: перенести порт для связи в config для его настройки
func (mongod *mdService) Init() error {

	session, err := mongo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}

	session.SetMode(mongo.Monotonic, true)

	mongod.connect = session
	mongod.status = "ready"
	return nil
}

func (mongod *mdService) Connect(in <-chan interface{}) (out chan interface{}, err error) {
	out = make(chan interface{})

	go func() {
		out <- "open"
		for {
			select {
			case v := <-in:
				if v.(string) == "close" {
					mongod.Close(out)
				} else {
					out <- v
				}
			}
		}
	}()
	return out, nil
}

func (mongod *mdService) Close(out chan<- interface{}) error {
	close(out)
	return nil
}

//получаем текущий статус сервиса
func (mongod *mdService) Status() string {
	return mongod.status
}

func (mongod *mdService) Get(args ...interface{}) (interface{}, error) {

	if len(args) < 2 {
		return nil, services.ErrServiceNotEnoughParameter{Name: mongod.name, Param: args}
	}

	connectionStatus := services.Status("mongod")

	if connectionStatus != "ready" {
		return nil, services.ErrBrokenConnection{Name: mongod.name, Param: args}
	}

	var collection string
	switch col := args[0].(type) {
	case string:
		collection = col
	default:
		return nil, services.ErrServiceNotCorrectParamType{Name: mongod.name, Param: col, Number: 1}
	}

	connetion := mongod.connect
	cConnect := connetion.DB(server.GetMongodConfig().MongoDBName()).C(collection)

	switch option := args[1].(type) {
	case string:
		switch option {
		case "Find":
			return findRecord(cConnect, args)

		default:
			return nil, services.ErrServiceNotCorrectParamType{Name: mongod.name, Param: args, Number: 2}
		}
	}

	return nil, services.ErrServiceNotEnoughParameter{Name: mongod.name, Param: args}
}

func (mongod *mdService) Send(args ...interface{}) error {

	if len(args) < 2 {
		return services.ErrServiceNotEnoughParameter{Name: mongod.name, Param: args}
	}

	connectionStatus := services.Status("mongod")

	if connectionStatus != "ready" {
		return services.ErrBrokenConnection{Name: mongod.name, Param: args}
	}

	var collection string
	switch col := args[0].(type) {
	case string:
		collection = col
	default:
		return services.ErrServiceNotCorrectParamType{Name: mongod.name, Param: col, Number: 1}
	}

	connetion := mongod.connect
	cConnect := connetion.DB(server.GetMongodConfig().MongoDBName()).C(collection)

	switch option := args[1].(type) {
	case string:
		switch option {
		case "Insert":
			return insertRecord(cConnect, args)

		case "Update":
			return updateRecord(cConnect, args)

		case "Remove":
			return removeRecord(cConnect, args)

		default:
			return services.ErrServiceNotCorrectParamType{Name: mongod.name, Param: args, Number: 2}
		}
	}

	return services.ErrServiceNotEnoughParameter{Name: mongod.name, Param: args}
}

//поиск записей в mongodb
func findRecord(cConnect *mongo.Collection, args []interface{}) (interface{}, error) {

	if len(args) < 4 {
		return nil, services.ErrServiceNotEnoughParameter{Name: mongod.name, Param: args}
	}
	switch oType := args[2].(type) {
	case string:
		switch oType {
		case "All":
			switch args[3].(type) {
			case bson.M:
				response := make([]interface{}, 0)
				err := cConnect.Find(args[3]).All(&response)

				if err != nil {
					return nil, err
				}

				return response, nil

			default:
				return nil, services.ErrServiceNotCorrectParamType{Name: mongod.name, Param: args, Number: 4}
			}
		case "One":
			switch args[3].(type) {
			case bson.M:
				var response interface{}
				err := cConnect.Find(args[3]).One(&response)

				if err != nil {
					return nil, err
				}

				return response, nil

			default:
				return nil, services.ErrServiceNotCorrectParamType{Name: mongod.name, Param: args, Number: 4}
			}
		default:
			return nil, services.ErrServiceNotCorrectParamType{Name: mongod.name, Param: args, Number: 3}
		}
	}

	return nil, nil
}

//создание новой записи в mongodb
func insertRecord(cConnect *mongo.Collection, args []interface{}) error {

	if len(args) < 3 {
		return services.ErrServiceNotEnoughParameter{Name: mongod.name, Param: args}
	}
	err := cConnect.Insert(args[2])
	if err != nil {
		return err
	}

	return nil
}

//обновление записи в mongodb
func updateRecord(cConnect *mongo.Collection, args []interface{}) error {

	if len(args) < 4 {
		return services.ErrServiceNotEnoughParameter{Name: mongod.name, Param: args}
	}
	_, err := cConnect.UpdateAll(args[2], args[3])
	if err != nil {
		return err
	}

	return nil
}

//удаление записи в mongodb
func removeRecord(cConnect *mongo.Collection, args []interface{}) error {

	if len(args) < 3 {
		return services.ErrServiceNotEnoughParameter{Name: mongod.name, Param: args}
	}
	err := cConnect.Remove(args[2])

	if err != nil {
		return err
	}

	return nil
}

//GetMongoCollectionConnect функция для получения соединения к колекции по названию
func GetMongoCollectionConnect(collection string) *mongo.Collection {
	return mongod.connect.DB(server.GetMongodConfig().MongoDBName()).C(collection)
}

func init() {
	services.AddService(mongod.name, mongod)
}

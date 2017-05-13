// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	DBschema "github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/db"
)

type schemaService struct {
	name string
}

var (schema *schemaService = &schemaService{name:"schema"})

func (schema *schemaService) Init() error{

	db.InitSchema()
	return nil
}
func (schema *schemaService) Send(messages ...interface{}) error{
	return nil

}
func (schema *schemaService) Get(messages ... interface{}) (responce interface{}, err error) {

	for _, message := range messages {
		switch args := message.(type) {
		case [] interface{}:
			switch tableName := args[0].(type) {
			case string:
					return DBschema.GetFieldsTable(tableName), nil
			default:
				return nil, &ErrServiceNotCorrectParamType{Name: schema.name, Param: args[0]}
			}
		}
	}

	return nil, nil

}
func (schema *schemaService) Connect(in <- chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (schema *schemaService) Close(out chan <- interface{}) error {

	return nil
}
func (schema *schemaService) Status() string {

	return ""
}

func init() {
	AddService(schema.name, schema)
}


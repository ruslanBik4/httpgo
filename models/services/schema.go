// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"github.com/ruslanBik4/httpgo/models/db"
	DBschema "github.com/ruslanBik4/httpgo/models/db/schema"
)

type schemaService struct {
	name   string
	status string
}

var (
	schema *schemaService = &schemaService{name: "schema", status: "create"}
)

func (schema *schemaService) Init() error {

	schema.status = "starting"
	db.InitSchema()
	schema.status = "ready"

	return nil
}
func (schema *schemaService) Send(messages ...interface{}) error {
	return nil

}
func (schema *schemaService) Get(messages ...interface{}) (responce interface{}, err error) {

	switch tableName := messages[0].(type) {
	case string:
		return DBschema.GetFieldsTable(tableName), nil
	default:
		return nil, &ErrServiceNotCorrectParamType{Name: schema.name, Param: messages[0]}
	}

	return nil, nil

}
func (schema *schemaService) Connect(in <-chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (schema *schemaService) Close(out chan<- interface{}) error {

	return nil
}
func (schema *schemaService) Status() string {

	return schema.status
}

func init() {
	AddService(schema.name, schema)
}

// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"golang.org/x/net/context"

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

func (schema *schemaService) Init(ctx context.Context) error {

	schema.status = "starting"
	db.InitSchema()
	schema.status = "ready"

	return nil
}
func (schema *schemaService) Send(ctx context.Context, messages ...interface{}) error {
	return nil

}
func (schema *schemaService) Get(ctx context.Context, messages ...interface{}) (response interface{}, err error) {

	if tableName, ok := messages[0].(string); ok {
		return DBschema.GetFieldsTable(tableName), nil
	}

	return nil, &ErrServiceNotCorrectParamType{Name: schema.name, Param: messages[0]}
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

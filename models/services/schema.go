// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

type schemaService struct {
	name string
}

var (schema *schemaService = &schemaService{name:"schema"})

func (schema *schemaService) Init() error{
	return nil
}
func (schema *schemaService) Send(messages ...interface{}) error{
	return nil

}
func (schema *schemaService) Get(messages ... interface{}) (responce interface{}, err error) {
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


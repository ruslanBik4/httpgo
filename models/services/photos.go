// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"path/filepath"
	"os"
	"log"
	"github.com/ruslanBik4/httpgo/models/server"
	"io"
)

type photosService struct {
	name string
	path string
	fileName string
}

var (
	photos *photosService = &photosService{name:"photos"}
)

func (photos *photosService) Init() error{
	ServerConfig := server.GetServerConfig()

	photos.path = filepath.Join( ServerConfig.WWWPath(), "files/photos" )
	photos.fileName = "test.txt"
	return nil
}
// выполняет операцию по записи/чтении файла
// это зависит от первого параметра - "save" или "read"
// третий параметр - имя файла
func (photos *photosService) Send(args ...interface{}) error{

	var oper string
			if len(args) < 2 {
				return &ErrServiceNotEnougnParameter{Name: photos.name, Param: args}
			}
			switch message := args[0].(type) {
			case string:
				oper = message
				log.Println(oper)
			default:
				return &ErrServiceNotCorrectParamType{Name: photos.name, Param: message}
			}

			if oper == "save" {
				switch message := args[1].(type) {
				case string:
					photos.fileName = message
				default:
					return &ErrServiceNotCorrectParamType{Name: photos.name, Param: message}
				}
				if len(args) < 3 {
					return &ErrServiceNotEnougnParameter{Name: photos.name, Param: args}
				}
				photos.saveFile(args[2].(io.Reader))
			} else if oper == "read" {
				log.Println(oper)
				return &ErrServiceNotCorrectOperation{Name: photos.name, OperName: oper}

			} else {
				return &ErrServiceNotCorrectOperation{Name: photos.name, OperName: oper}

			}

	return nil

}
func (photos *photosService) Get(messages ... interface{}) (responce interface{}, err error) {
	return nil, nil

}
func (photos *photosService) Connect(in <- chan interface{}) (out chan interface{}, err error) {

	return nil, nil
}
func (photos *photosService) Close(out chan <- interface{}) error {

	return nil
}
func (photos *photosService) Status() string {

	return ""
}
func (photos *photosService) saveFile(inFile io.Reader) error {

	fullName := filepath.Join(photos.path, photos.fileName)
	if outFile, err := os.Create(fullName); err != nil {
		log.Println(err)
		return err
	} else {
		defer outFile.Close()
		_, err := io.Copy(outFile, inFile )
		if err != nil {
			log.Println("Error saving file: "+err.Error())
			return err
		}
	}
	return nil

}
func init() {
	AddService(photos.name, photos)
}


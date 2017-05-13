// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

const MessServNotFound = " service not found in list services"
type ErrServiceNotFound struct {
	Name string
}
func (err *ErrServiceNotFound) Error() string{
	return err.Name + MessServNotFound
}

const photosNotCorrectOperation = " operation name is incorrect or not string type - "

type ErrServiceNotCorrectOperation struct {
	Name string
	OperName string
	Message string
}
func (err *ErrServiceNotCorrectOperation) Error() string{
	err.Message = err.Name + photosNotCorrectOperation + err.OperName
	return err.Message
}

const photosNotCorrectParameterType = " operation name is not string type - "

type ErrServiceNotCorrectParamType struct {
	Name string
	Param interface{}
}
func (err *ErrServiceNotCorrectParamType) Error() string{
	return err.Name + photosNotCorrectParameterType
}

const photosNotEnougnParameter = " not enougn parameters "

type ErrServiceNotEnougnParameter struct {
	Name string
	Param interface{}
}
func (err *ErrServiceNotEnougnParameter) Error() string{
	return err.Name + photosNotEnougnParameter
}

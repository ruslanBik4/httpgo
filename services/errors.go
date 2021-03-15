// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

const messServNotFound = " service not found in list services"

// ErrServiceNotFound for errors when current service not found
type ErrServiceNotFound struct {
	Name string
}

func (err ErrServiceNotFound) Error() string {
	return err.Name + messServNotFound
}

const messServNotReady = " service not ready to operation"

// ErrServiceNotReady for errors if service not ready
type ErrServiceNotReady struct {
	Name   string
	Status string
}

func (err ErrServiceNotReady) Error() string {
	return err.Name + messServNotReady
}

const photosNotCorrectOperation = " operation name is incorrect or not string type - "

// ErrServiceNotCorrectOperation for errors if input operation is not valid
type ErrServiceNotCorrectOperation struct {
	Name     string
	OperName string
	Message  string
}

func (err ErrServiceNotCorrectOperation) Error() string {
	err.Message = err.Name + photosNotCorrectOperation + err.OperName
	return err.Message
}

const photosNotCorrectParameterType = " Wrong params type "

// ErrServiceNotCorrectParamType for errors if parameter is not valid
type ErrServiceNotCorrectParamType struct {
	Name   string
	Number int
	Param  interface{}
}

func (err ErrServiceNotCorrectParamType) Error() string {
	return err.Name + photosNotCorrectParameterType
}

const photosNotEnoughParameter = " not enough parameters: "

// ErrServiceNotEnoughParameter for errors if not found required parameter
type ErrServiceNotEnoughParameter struct {
	Name  string
	Param interface{}
}

func (err ErrServiceNotEnoughParameter) Error() string {
	return photosNotEnoughParameter + err.Name
}

const brokenStatus = " broken status "

// ErrBrokenConnection for errors broken connection
type ErrBrokenConnection struct {
	Name  string
	Param interface{}
}

func (err ErrBrokenConnection) Error() string {
	return err.Name + brokenStatus
}

// ErrServiceWrongIndex for errors wrong index in array range
// TODO: wrote correct comment for this type
type ErrServiceWrongIndex struct {
	Name  string
	Index int
}

func (err ErrServiceWrongIndex) Error() string {
	return err.Name + photosNotEnoughParameter
}

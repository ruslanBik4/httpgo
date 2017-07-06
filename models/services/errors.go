// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

const MessServNotFound = " service not found in list services"

type ErrServiceNotFound struct {
	Name string
}

func (err ErrServiceNotFound) Error() string {
	return err.Name + MessServNotFound
}

const photosNotCorrectOperation = " operation name is incorrect or not string type - "

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

type ErrServiceNotCorrectParamType struct {
	Name   string
	Number int
	Param  interface{}
}

func (err ErrServiceNotCorrectParamType) Error() string {
	return err.Name + photosNotCorrectParameterType
}

const photosNotEnoughParameter = " not enough parameters "

type ErrServiceNotEnoughParameter struct {
	Name  string
	Param interface{}
}

func (err ErrServiceNotEnoughParameter) Error() string {
	return err.Name + photosNotEnoughParameter
}

const brokenStatus = " broken status "

type ErrBrokenConnection struct {
	Name  string
	Param interface{}
}

func (err ErrBrokenConnection) Error() string {
	return err.Name + brokenStatus
}

type ErrServiceWrongIndex struct {
	Name  string
	Index int
}

func (err ErrServiceWrongIndex) Error() string {
	return err.Name + photosNotEnoughParameter
}

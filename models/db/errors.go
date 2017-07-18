// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// error types for package

package db

const MessParamNotFound = "Param not found:"

type ErrParamNotFound struct {
	Name string
	FuncName string
}

func (err ErrParamNotFound) Error() string {
	return  MessParamNotFound + err.Name
}

const MessBadParam = "Bad value in params:"

type ErrBadParam struct {
	Name string
	BadName string
	FuncName string
}

func (err ErrBadParam) Error() string {
	return MessBadParam + err.Name + ", " + err.BadName
}

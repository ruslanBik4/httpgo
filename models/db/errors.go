// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// error types for package

package db

const messParamNotFound = "Param not found:"

// ErrParamNotFound if not found requiared parameter
type ErrParamNotFound struct {
	Name     string
	FuncName string
}

func (err ErrParamNotFound) Error() string {
	return messParamNotFound + err.Name
}

const messBadParam = "Bad value in params:"

// ErrBadParam is parameter not valid
type ErrBadParam struct {
	Name     string
	BadName  string
	FuncName string
}

func (err ErrBadParam) Error() string {
	return messBadParam + err.Name + ", " + err.BadName
}

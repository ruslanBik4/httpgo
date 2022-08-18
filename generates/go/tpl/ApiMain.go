/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package tpl

//go:generate stringer -type TypeAuth
type TypeAuth uint8

const (
	Basic TypeAuth = iota
	JWT
	OAuth2
)

type ApiMain struct {
	Name string
	auth TypeAuth
}

func NewApiMain(name string, a TypeAuth) *ApiMain {

	//if a == OAuth2 {
	//	t := reflect.TypeOf(auth.OAuth2{})
	//	logs.StatusLog("%s", t.)
	//}
	return &ApiMain{
		name,
		a,
	}
}

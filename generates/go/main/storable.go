/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package main

import (
	"os"
	"path"

	"github.com/pkg/errors"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/logs"
)

// MakeSrcDir create with destination directory 'dst'
func MakeSrcDir(dst string) error {
	err := os.Mkdir(dst, os.ModePerm)

	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "mkDirAll")
	}

	return nil
}

// MakeSrcFile create table interface with Columns operations
func MakeSrcFile(dst string, table dbEngine.Table) (*os.File, error) {
	logs.SetDebug(true)
	//name := strcase.ToCamel(table.Name())
	f, err := os.Create(path.Join(dst, table.Name()) + ".go")
	if err != nil && !os.IsExist(err) {
		// err.(*os.PathError).Err
		return nil, errors.Wrap(err, "creator")
	}

	return f, nil
}

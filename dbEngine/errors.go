// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbEngine

import (
	"fmt"
	"strings"

	"github.com/ruslanBik4/httpgo/logs"
)

// ErrNotFoundTable if not found table by name {Table}
type ErrNotFoundTable struct {
	Table string
}

func (err ErrNotFoundTable) Error() string {

	return fmt.Sprintf("Not table `%s` in schema ", err.Table)
}

// ErrNotFoundField if not found in table {Table} field by name {FieldName}
type ErrNotFoundField struct {
	Table     string
	FieldName string
}

func (err ErrNotFoundField) Error() string {

	return fmt.Sprintf("Not field `%s` for table `%s` in schema ", err.FieldName, err.Table)

}

func isErrorAlreadyExists(err error) bool {
	ignoreErrors := []string{
		"already exists",
	}

	for _, val := range ignoreErrors {
		if strings.Contains(err.Error(), val) {
			return true
		}
	}

	return false
}

func isErrorForReplace(err error) bool {
	ignoreErrors := []string{
		"cannot change return type of existing function",
		"cannot change name of input parameter",
	}
	for _, val := range ignoreErrors {
		if strings.Contains(err.Error(), val) {
			return true
		}

	}

	logs.DebugLog(" %+v", err)
	return false
}

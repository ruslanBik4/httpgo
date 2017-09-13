// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"github.com/ruslanBik4/httpgo/models/db/schema"
	"github.com/ruslanBik4/httpgo/models/logs"
)

// ErrNotFoundParam for errors not found requared parameters
type ErrNotFoundParam struct {
	Param string
}

func (err ErrNotFoundParam) Error() string {
	return err.Param
}
func schemaError() {
	result := recover()
	switch err := result.(type) {
	case schema.ErrNotFoundTable:
		logs.ErrorLogHandler(err, err.Table)
		panic(err)
	case nil:
	case error:
		panic(err)
	}
}

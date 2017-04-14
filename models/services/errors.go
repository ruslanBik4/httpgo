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


// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multiquery


func Create(tableName string) *MultiQuery{
	return &MultiQuery{	Queryes: make(map[string]*ArgsQuery, 0), parentName: tableName}
}

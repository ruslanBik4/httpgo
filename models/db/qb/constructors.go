// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

// constructors
func Create(where, groupBy, orderBy string) *QueryBuilder{

	qb := &QueryBuilder{Where: where, OrderBy: orderBy, GroupBy: groupBy}
	return qb
}
func CreateEmpty() *QueryBuilder{

	qb := &QueryBuilder{}
	return qb
}
func CreateFromSQL(sqlCommand string) *QueryBuilder {
	qb := &QueryBuilder{sqlCommand: sqlCommand}
	return qb
}


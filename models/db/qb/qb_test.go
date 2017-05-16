// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"testing"
)

func TestQBCreate(t *testing.T) {
	qb := &QueryBuilder{OrderBy:"name"}

	table := &qbTables{Name:"rooms"}
	fields := make(map[string] *qbFields, 2)
	fields["name"] = &qbFields{Name: "title" }
	fields["num"]  = &qbFields{Name: "id"}

	qb.Tables = make( map[string] *qbTables, 2)
	qb.Tables["a"] = table

	v, err := qb.SelectToMultidimension()

	if err != nil {
		t.Error(err)
	} else {
		t.Log(v)
		t.Skipped()
	}
}

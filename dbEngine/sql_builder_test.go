// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbEngine

import (
	"reflect"
	"testing"
	"time"
)

func TestArgsForSelect(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name string
		args args
		want BuildSqlOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ArgsForSelect(tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArgsForSelect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnsForSelect(t *testing.T) {
	type args struct {
		columns []string
	}
	tests := []struct {
		name string
		args args
		want BuildSqlOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ColumnsForSelect(tt.args.columns...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ColumnsForSelect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLBuilder_InsertSql(t *testing.T) {
	type fields struct {
		Args          []interface{}
		columns       []string
		filter        []string
		posFilter     int
		Table         Table
		SelectColumns []Column
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"simple insert",
			fields{
				[]interface{}{time.Now()},
				[]string{"last_login"},
				nil,
				0,
				tableString{},
				nil,
			},
			"INSERT INTO StringTable(last_login) VALUES ($1)",
			false,
		},
		{
			"two columns insert",
			fields{
				[]interface{}{1, "ruslan"},
				[]string{"last_login", "name"},
				nil,
				0,
				tableString{},
				nil,
			},
			"INSERT INTO StringTable(last_login,name) VALUES ($1,$2)",
			false,
		},
		{
			"two columns insert according two filter columns",
			fields{
				[]interface{}{"ruslan", time.Now()},
				[]string{"last_login", "name"},
				nil,
				0,
				tableString{},
				nil,
			},
			"INSERT INTO StringTable(last_login,name) VALUES ($1,$2)",
			false,
		},
		{
			"two columns insert according two filter columns & wrong args",
			fields{
				[]interface{}{1, "ruslan", time.Now()},
				[]string{"last_login", "name"},
				[]string{"id", "id_roles"},
				0,
				tableString{},
				nil,
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := SQLBuilder{
				Args:          tt.fields.Args,
				columns:       tt.fields.columns,
				filter:        tt.fields.filter,
				posFilter:     tt.fields.posFilter,
				Table:         tt.fields.Table,
				SelectColumns: tt.fields.SelectColumns,
			}
			got, err := b.InsertSql()
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertSql() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("InsertSql() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLBuilder_Select(t *testing.T) {
	type fields struct {
		Args          []interface{}
		columns       []string
		filter        []string
		posFilter     int
		Table         Table
		SelectColumns []Column
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			"simple insert",
			fields{
				[]interface{}{1, time.Now()},
				[]string{"last_login"},
				[]string{"id"},
				0,
				tableString{},
				nil,
			},
			"last_login",
		},
		{
			"two columns update",
			fields{
				[]interface{}{1, "ruslan", time.Now()},
				[]string{"last_login", "name"},
				[]string{"id"},
				0,
				tableString{},
				nil,
			},
			"last_login,name",
		},
		{
			"two columns update according two filter columns",
			fields{
				[]interface{}{1, 2, "ruslan", time.Now()},
				[]string{"last_login", "name"},
				[]string{"id", "id_roles"},
				0,
				tableString{},
				nil,
			},
			"last_login,name",
		},
		{
			"two columns update according four filter columns",
			fields{
				[]interface{}{1, "ruslan", time.Now()},
				[]string{"last_login", "name", "id", "id_roles"},
				nil,
				0,
				tableString{},
				nil,
			},
			"last_login,name,id,id_roles",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &SQLBuilder{
				Args:          tt.fields.Args,
				columns:       tt.fields.columns,
				filter:        tt.fields.filter,
				posFilter:     tt.fields.posFilter,
				Table:         tt.fields.Table,
				SelectColumns: tt.fields.SelectColumns,
			}
			if got := b.Select(); got != tt.want {
				t.Errorf("Select() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLBuilder_SelectSql(t *testing.T) {
	type fields struct {
		Args          []interface{}
		columns       []string
		filter        []string
		posFilter     int
		Table         Table
		SelectColumns []Column
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"simple select",
			fields{
				[]interface{}{1},
				[]string{"last_login"},
				[]string{"id"},
				0,
				tableString{},
				nil,
			},
			"SELECT last_login FROM StringTable WHERE  id=$1",
			false,
		},
		{
			"two columns select",
			fields{
				[]interface{}{1},
				[]string{"last_login", "name"},
				[]string{"id"},
				0,
				tableString{},
				nil,
			},
			"SELECT last_login,name FROM StringTable WHERE  id=$1",
			false,
		},
		{
			"two columns select according two filter columns",
			fields{
				[]interface{}{1, 2},
				[]string{"last_login", "name"},
				[]string{"id", "id_roles"},
				0,
				tableString{},
				nil,
			},
			"SELECT last_login,name FROM StringTable WHERE  id=$1 AND  id_roles=$2",
			false,
		},
		{
			"two columns select according two filter columns & wrong args",
			fields{
				[]interface{}{1},
				[]string{"last_login", "name"},
				[]string{"id", "id_roles"},
				0,
				tableString{},
				nil,
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := SQLBuilder{
				Args:          tt.fields.Args,
				columns:       tt.fields.columns,
				filter:        tt.fields.filter,
				posFilter:     tt.fields.posFilter,
				Table:         tt.fields.Table,
				SelectColumns: tt.fields.SelectColumns,
			}
			got, err := b.SelectSql()
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectSql() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SelectSql() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLBuilder_Set(t *testing.T) {
	type fields struct {
		Args          []interface{}
		columns       []string
		filter        []string
		posFilter     int
		Table         Table
		SelectColumns []Column
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &SQLBuilder{
				Args:          tt.fields.Args,
				columns:       tt.fields.columns,
				filter:        tt.fields.filter,
				posFilter:     tt.fields.posFilter,
				Table:         tt.fields.Table,
				SelectColumns: tt.fields.SelectColumns,
			}
			if got := b.Set(); got != tt.want {
				t.Errorf("Set() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLBuilder_UpdateSql(t *testing.T) {
	type fields struct {
		Args          []interface{}
		columns       []string
		filter        []string
		posFilter     int
		Table         Table
		SelectColumns []Column
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"simple update",
			fields{
				[]interface{}{1, time.Now()},
				[]string{"last_login"},
				[]string{"id"},
				0,
				tableString{},
				nil,
			},
			"UPDATE StringTable SET  last_login=$1 WHERE  id=$2",
			false,
		},
		{
			"two columns update",
			fields{
				[]interface{}{1, "ruslan", time.Now()},
				[]string{"last_login", "name"},
				[]string{"id"},
				0,
				tableString{},
				nil,
			},
			"UPDATE StringTable SET  last_login=$1, name=$2 WHERE  id=$3",
			false,
		},
		{
			"two columns update according two filter columns",
			fields{
				[]interface{}{1, 2, "ruslan", time.Now()},
				[]string{"last_login", "name"},
				[]string{"id", "id_roles"},
				0,
				tableString{},
				nil,
			},
			"UPDATE StringTable SET  last_login=$1, name=$2 WHERE  id=$3 AND  id_roles=$4",
			false,
		},
		{
			"two columns update according two filter columns & wrong args",
			fields{
				[]interface{}{1, "ruslan", time.Now()},
				[]string{"last_login", "name"},
				[]string{"id", "id_roles"},
				0,
				tableString{},
				nil,
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := SQLBuilder{
				Args:          tt.fields.Args,
				columns:       tt.fields.columns,
				filter:        tt.fields.filter,
				posFilter:     tt.fields.posFilter,
				Table:         tt.fields.Table,
				SelectColumns: tt.fields.SelectColumns,
			}
			got, err := b.UpdateSql()
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateSql() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UpdateSql() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLBuilder_Where(t *testing.T) {
	type fields struct {
		Args          []interface{}
		columns       []string
		filter        []string
		posFilter     int
		Table         Table
		SelectColumns []Column
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &SQLBuilder{
				Args:          tt.fields.Args,
				columns:       tt.fields.columns,
				filter:        tt.fields.filter,
				posFilter:     tt.fields.posFilter,
				Table:         tt.fields.Table,
				SelectColumns: tt.fields.SelectColumns,
			}
			if got := b.Where(); got != tt.want {
				t.Errorf("Where() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLBuilder_values(t *testing.T) {
	type fields struct {
		Args          []interface{}
		columns       []string
		filter        []string
		posFilter     int
		Table         Table
		SelectColumns []Column
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &SQLBuilder{
				Args:          tt.fields.Args,
				columns:       tt.fields.columns,
				filter:        tt.fields.filter,
				posFilter:     tt.fields.posFilter,
				Table:         tt.fields.Table,
				SelectColumns: tt.fields.SelectColumns,
			}
			if got := b.values(); got != tt.want {
				t.Errorf("values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhereForSelect(t *testing.T) {
	type args struct {
		columns []string
	}
	tests := []struct {
		name string
		args args
		want BuildSqlOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WhereForSelect(tt.args.columns...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WhereForSelect() = %v, want %v", got, tt.want)
			}
		})
	}
}

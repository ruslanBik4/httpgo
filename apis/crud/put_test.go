/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/httpgo/apis"
)

func TestAddColumnAndValue(t *testing.T) {
	type args struct {
		name      string
		table     dbEngine.Table
		arg       interface{}
		buf       *bytes.Buffer
		badParams map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			"simple",
			args{
				name: "name",
				table: dbEngine.NewTableString("test",
					"",
					[]dbEngine.Column{dbEngine.NewStringColumn("name", "", true)},
					nil,
					nil,
				),
				arg:       nil,
				buf:       bytes.NewBufferString(""),
				badParams: map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, arg := AddColumnAndValue(tt.args.name, tt.args.table, tt.args.arg, tt.args.buf, tt.args.badParams)
			t.Log(name, arg, tt.args.badParams)
		})
	}
}

func TestReadByteA(t *testing.T) {
	type args struct {
		fHeaders []*multipart.FileHeader
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want1   [][]byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ReadByteA(tt.args.fHeaders)
			if !tt.wantErr(t, err, fmt.Sprintf("ReadByteA(%v)", tt.args.fHeaders)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ReadByteA(%v)", tt.args.fHeaders)
			assert.Equalf(t, tt.want1, got1, "ReadByteA(%v)", tt.args.fHeaders)
		})
	}
}

func TestTableInsert(t *testing.T) {
	type args struct {
		preRoute string
		table    dbEngine.Table
		params   []string
	}
	tests := []struct {
		name string
		args args
		want apis.ApiRouteHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, TableInsert(tt.args.preRoute, tt.args.table, tt.args.params), "TableInsert(%v, %v, %v)", tt.args.preRoute, tt.args.table, tt.args.params)
		})
	}
}

func TestTableUpdate(t *testing.T) {
	type args struct {
		preRoute   string
		table      dbEngine.Table
		columns    []string
		priColumns []string
	}
	tests := []struct {
		name string
		args args
		want apis.ApiRouteHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, TableUpdate(tt.args.preRoute, tt.args.table, tt.args.columns, tt.args.priColumns), "TableUpdate(%v, %v, %v, %v)", tt.args.preRoute, tt.args.table, tt.args.columns, tt.args.priColumns)
		})
	}
}

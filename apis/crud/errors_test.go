// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crud

import (
	"reflect"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestCreateErrResult(t *testing.T) {
	const (
		testMobileMsg = `duplicate key value violates unique constraint "candidates_mobile_uindex"`
		testMobileKey = `Key (phone)=(+380) already exists.`
		testLinkedin  = `Key (linkedin)=(https://www.linkedin.com/in/vladislav-yena/) already exists.`
	)

	testMobileRes := map[string]string{"candidates_mobile_uindex": "duplicate key value violates unique constraint"}
	testMobileKeyRes := map[string]string{"phone": "`+380` already exists"}
	testLinedinRes := map[string]string{
		"linkedin": "`https://www.linkedin.com/in/vladislav-yena/` already exists",
	}
	tests := []struct {
		name    string
		err     error
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"wrap mobile",
			errors.Wrap(errors.New(testMobileMsg), "wrap"),
			testMobileRes,
			true,
		},
		{
			"pgError mobile duplicate error msg",
			errors.Wrap(&pgconn.PgError{Detail: testMobileMsg}, "wrap"),
			testMobileRes,
			true,
		},
		{
			"pgError mobile key msg",
			errors.Wrap(&pgconn.PgError{Detail: testMobileKey}, "wrap"),
			testMobileKeyRes,
			true,
		},
		{
			"pgError linkedin ",
			errors.Wrap(&pgconn.PgError{Detail: testLinkedin}, "wrap"),
			testLinedinRes,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateErrResult(tt.err)
			assert.Equal(t, err != nil, tt.wantErr)
			if tt.wantErr {
				assert.Equal(t, apis.ErrWrongParamsList, err, tt.name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_createResult(t *testing.T) {
	type args struct {
		ctx    *fasthttp.RequestCtx
		id     int64
		msg    string
		colSel []string
		url    string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderCreatedResult(tt.args.ctx, tt.args.id, tt.args.msg, tt.args.colSel, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderCreatedResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RenderCreatedResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}

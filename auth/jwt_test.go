// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/apis"
)

func TestAuthBearer_AddToken(t *testing.T) {
	type fields struct {
		tokens Tokens
	}
	type args struct {
		hash int64
		id   int
		ctx  map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AuthBearer{
				tokens: tt.fields.tokens,
			}
			if got := a.AddToken(tt.args.hash, tt.args.id, tt.args.ctx); got != tt.want {
				t.Errorf("AddToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBearer_AdminAuth(t *testing.T) {
	type fields struct {
		tokens Tokens
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AuthBearer{
				tokens: tt.fields.tokens,
			}
			if got := a.AdminAuth(tt.args.ctx); got != tt.want {
				t.Errorf("AdminAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBearer_Auth(t *testing.T) {
	type fields struct {
		tokens Tokens
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AuthBearer{
				tokens: tt.fields.tokens,
			}
			if got := a.Auth(tt.args.ctx); got != tt.want {
				t.Errorf("Auth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBearer_GetToken(t *testing.T) {
	type fields struct {
		tokens Tokens
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AuthBearer{
				tokens: tt.fields.tokens,
			}
			if got := a.GetToken(tt.args.ctx); got != tt.want {
				t.Errorf("GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBearer_String(t *testing.T) {
	type fields struct {
		tokens Tokens
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
			a := AuthBearer{
				tokens: tt.fields.tokens,
			}
			if got := a.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBearer_getBearer(t *testing.T) {
	type fields struct {
		tokens Tokens
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AuthBearer{
				tokens: tt.fields.tokens,
			}
			if got := a.getBearer(tt.args.ctx); got != tt.want {
				t.Errorf("getBearer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAuthBearer(t *testing.T) {
	a := NewAuthBearer(nil)
	assert.Implements(t, (*apis.FncAuth)(nil), a)
}

func Test_getStringOfFnc(t *testing.T) {
	type args struct {
		pc uintptr
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStringOfFnc(tt.args.pc); got != tt.want {
				t.Errorf("getStringOfFnc() = %v, want %v", got, tt.want)
			}
		})
	}
}

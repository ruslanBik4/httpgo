// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

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
			a := &AuthBearer{
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
			a := &AuthBearer{
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
		want   TokenData
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthBearer{
				tokens: tt.fields.tokens,
			}
			if got := a.GetToken(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBearer_NewToken(t *testing.T) {
	type fields struct {
		tokens Tokens
	}
	tests := []struct {
		name   string
		fields fields
		args   *testTokenData
		ctx    *fasthttp.RequestCtx
	}{
		// TODO: Add test cases.
		{
			"1",
			fields{
				&MapTokens{
					expiresIn: time.Hour,
					tokens:    map[string]*mapToken{},
				},
			},
			&testTokenData{
				id:      1,
				isAdmin: false,
			},
			&fasthttp.RequestCtx{},
		},
		{
			"2",
			fields{
				&MapTokens{
					expiresIn: time.Hour,
					tokens:    map[string]*mapToken{},
				},
			},
			&testTokenData{
				id:      1,
				isAdmin: true,
			},
			&fasthttp.RequestCtx{},
		},
		{
			"3",
			fields{
				&MapTokens{
					expiresIn: time.Hour,
					tokens:    map[string]*mapToken{},
				},
			},
			&testTokenData{
				id:      10000,
				isAdmin: false,
			},
			&fasthttp.RequestCtx{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthBearer{
				tokens: tt.fields.tokens,
			}

			s := make([]string, 0)
			for i := 0; i < 100; i++ {
				got, err := a.NewToken(tt.args)
				assert.Nil(t, err)
				for _, str := range s {
					if !assert.NotEqual(t, got, str, "not random value of token") {
						break
					}
				}
				s = append(s, got)
			}

			token := a.tokens.GetToken(s[0])
			tt.ctx.Request.Header.Set("Authorization", "  Bearer  "+s[0])

			if assert.NotNil(t, token, "not found token") &&
				assert.True(t, a.Auth(tt.ctx), "unAuthorization") {

				assert.Equal(t, tt.args.id, token.GetUserID())
				assert.Equal(t, tt.args.isAdmin, token.IsAdmin())
				assert.Equal(t, tt.args.isAdmin, a.AdminAuth(tt.ctx))
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
			a := &AuthBearer{
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
			a := &AuthBearer{
				tokens: tt.fields.tokens,
			}
			if got := a.getBearer(tt.args.ctx); got != tt.want {
				t.Errorf("getBearer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAuthBearer(t *testing.T) {
	type args struct {
		tokens Tokens
	}
	tests := []struct {
		name string
		args args
		want *AuthBearer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAuthBearer(tt.args.tokens); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuthBearer() = %v, want %v", got, tt.want)
			}
		})
	}
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

// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/ruslanBik4/logs"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestAuthBasic_AdminAuth(t *testing.T) {
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
			a := &AuthBasic{
				tokens: tt.fields.tokens,
			}
			if got := a.AdminAuth(tt.args.ctx); got != tt.want {
				t.Errorf("AdminAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBasic_Auth(t *testing.T) {
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
			a := &AuthBasic{
				tokens: tt.fields.tokens,
			}
			if got := a.Auth(tt.args.ctx); got != tt.want {
				t.Errorf("Auth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBasic_String(t *testing.T) {
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
			a := &AuthBasic{
				tokens: tt.fields.tokens,
			}
			if got := a.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBasic_getBasic(t *testing.T) {
	type fields struct {
		tokens Tokens
	}
	type args struct {
		header string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		p      []byte
		u      []byte
		ok     bool
	}{
		// TODO: Add test cases.
		{
			"dchervakov@ukr.net",
			fields{nil},
			args{"dchervakov@ukr.net:YTk_gJ5R0kFK8cmfgvn0eQ=="},
			[]byte("dchervakov@ukr.net"),
			[]byte("YTk_gJ5R0kFK8cmfgvn0eQ=="),
			true,
		},
		{
			"savtym@gmail.com",
			fields{nil},
			args{"savtym@gmail.com:PqqSpSmTfqVlf9WO6LXJAw=="},
			[]byte("savtym@gmail.com"),
			[]byte("PqqSpSmTfqVlf9WO6LXJAw=="),
			true,
		},
		{
			"ni@gamayun.sk",
			fields{nil},
			args{"ni@gamayun.sk:3Gwz9a2ode-kbzUi-07M_A=="},
			[]byte("ni@gamayun.sk"),
			[]byte("3Gwz9a2ode-kbzUi-07M_A=="),
			true,
		},
		{
			"bik4ruslan@gmail.com",
			fields{nil},
			args{"bik4ruslan@gmail.com:QLHxis2LpzpddPJgOZCCDg=="},
			[]byte("bik4ruslan@gmail.com"),
			[]byte("QLHxis2LpzpddPJgOZCCDg=="),
			true,
		},
		{
			"zero@null.com",
			fields{nil},
			args{"zero@null.com:"},
			[]byte("zero@null.com"),
			[]byte(""),
			true,
		},
		// negative
		{
			"null",
			fields{nil},
			args{""},
			nil,
			nil,
			false,
		},
		{
			"not :",
			fields{nil},
			args{"l"},
			nil,
			nil,
			false,
		},
		{
			"noDecode",
			fields{nil},
			args{"noDecode"},
			nil,
			nil,
			false,
		},
	}
	logs.SetDebug(true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthBasic{
				tokens: tt.fields.tokens,
			}
			ctx := &fasthttp.RequestCtx{}
			enc := base64.StdEncoding

			if tt.args.header == "noDecode" {
				ctx.Request.Header.Set("Authorization", "Basic "+tt.args.header)
			} else {
				ctx.Request.Header.Set("Authorization", "Basic "+enc.EncodeToString([]byte(tt.args.header)))
			}
			P, U, ok := a.getUserPass(a.getBasic(ctx))

			assert.Equal(t, tt.p, P)
			assert.Equal(t, tt.u, U)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestAuthBasic_getBasic_Hash(t *testing.T) {
	b := `bmlAZ2FtYXl1bi5zazozR3d6OWEyb2RlLWtielVpLTA3TV9BPT0=`
	b = `c2F2dHltQGdtYWlsLmNvbTpQcXFTcFNtVGZxVmxmOVdPNkxYSkF3PT0=`
	b = `dm92YXRlc3Rwb2x5bWVyQGdtYWlsLmNvbTpTcnlZd2xBM1NfdTYxYnlYOTVhOXlBPT0=`
	a := &AuthBasic{}

	tt := struct {
		name string
		p    []byte
		u    []byte
		ok   bool
	}{
		"vovatestpolymer@gmail.com",
		[]byte("vovatestpolymer@gmail.com"),
		[]byte("SryYwlA3S_u61byX95a9yA=="),
		true,
	}

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Authorization", "Basic "+b)
	P, U, ok := a.getUserPass(a.getBasic(ctx))

	assert.Equal(t, tt.p, P)
	assert.Equal(t, tt.u, U)
	assert.Equal(t, tt.ok, ok)

}

func TestAuthBasic_getBasic_Alladin(t *testing.T) {
	b := `QWxhZGRpbjpPcGVuU2VzYW1l`
	a := &AuthBasic{}

	tt := struct {
		name string
		p    []byte
		u    []byte
		ok   bool
	}{
		"Aladdin",
		[]byte("Aladdin"),
		[]byte("OpenSesame"),
		true,
	}

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Authorization", "Basic "+b)
	P, U, ok := a.getUserPass(a.getBasic(ctx))
	assert.Equal(t, tt.p, P)
	assert.Equal(t, tt.u, U)
	assert.Equal(t, tt.ok, ok)

}

func TestNewAuthBasic(t *testing.T) {
	type args struct {
		tokens Tokens
	}
	tests := []struct {
		name string
		args args
		want *AuthBasic
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAuthBasic(tt.args.tokens, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuthBasic() = %v, want %v", got, tt.want)
			}
		})
	}
}

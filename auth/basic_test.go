/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package auth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/logs"
)

// NOTE on this file's limits: auth_basic.go (wherever AuthBasic/NewAuthBasic
// actually live) was never uploaded to this session - only this test file
// was. Every case below stays within what getBasic/getUserPass's own
// existing fixtures already demonstrate empirically (base64-decode the
// "Basic " header value, split once on the first ':', return both sides
// plus an ok flag). TestAuthBasic_AdminAuth/Auth/String/TestNewAuthBasic
// are left as TODO stubs rather than filled in with guessed behavior -
// Auth/AdminAuth almost certainly compare the decoded password against
// something (a stored hash? the Tokens store, like AuthBearer does with
// bearer tokens?), and NewAuthBasic's second parameter (passed as nil in
// the original skeleton) is a type this session has never seen. Guessing
// either would produce tests that assert invented behavior instead of
// real behavior - upload auth_basic.go (or whatever the source file is
// named) to fill these in for real.

func TestAuthBasic_AdminAuth(t *testing.T) {
	t.Skip("AuthBasic's source (Auth/AdminAuth's actual password-checking logic) was never uploaded to this session - see this file's header comment")
}

func TestAuthBasic_Auth(t *testing.T) {
	t.Skip("AuthBasic's source (Auth's actual password-checking logic) was never uploaded to this session - see this file's header comment")
}

func TestAuthBasic_String(t *testing.T) {
	t.Skip("AuthBasic's source (String's exact output format) was never uploaded to this session - see this file's header comment")
}

func TestNewAuthBasic(t *testing.T) {
	t.Skip("NewAuthBasic's second parameter's type/purpose (passed as nil in the original skeleton) was never uploaded to this session - see this file's header comment")
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
		// Symmetric counterpart to "zero@null.com" above (empty password,
		// non-empty login): an empty login with a non-empty password. Safe
		// to infer from the same demonstrated mechanism (split once on ':',
		// return both sides verbatim, ok=true as long as a colon is
		// present and the header decodes) - not a guess about unrelated
		// logic like AdminAuth/Auth.
		{
			"empty login, non-empty password",
			fields{nil},
			args{":onlypassword"},
			[]byte(""),
			[]byte("onlypassword"),
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

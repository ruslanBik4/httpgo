/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package auth

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

// newCtxWithAuth builds a *fasthttp.RequestCtx carrying the given raw
// Authorization header value - the common setup every test below needs, so
// each test case only has to state the header text itself.
func newCtxWithAuth(header string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	if header != "" {
		ctx.Request.Header.Set("Authorization", header)
	}
	return ctx
}

func TestAuthBearer_getBearer(t *testing.T) {
	type fields struct {
		tokens Tokens
	}
	type args struct {
		header string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"well-formed", args{"Bearer abc123"}, "abc123"},
		{"extra whitespace between scheme and token", args{"Bearer   abc123"}, "abc123"},
		{"no header at all", args{""}, ""},
		{"wrong scheme", args{"Basic abc123"}, ""},
		// regBearer is `Bearer\s+(\S+)` with no (?i) flag - the match is
		// case-sensitive, so a lowercase scheme name does not match. This
		// is worth having a test pin down explicitly: it's easy to assume
		// HTTP header schemes are matched case-insensitively (RFC 7235
		// doesn't actually require that of a *value* like this), and a
		// future edit adding (?i) - or a client that lowercases "bearer" -
		// would silently change behavior without this test catching it.
		{"lowercase scheme does not match", args{"bearer abc123"}, ""},
		{"Bearer with no token", args{"Bearer"}, ""},
		{"Bearer with only whitespace after it", args{"Bearer   "}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthBearer{tokens: nil}
			ctx := newCtxWithAuth(tt.args.header)
			assert.Equal(t, tt.want, a.getBearer(ctx))
		})
	}
}

func TestAuthBearer_GetToken(t *testing.T) {
	data := &testTokenData{id: 5, isAdmin: true}
	tokens := NewMapTokens(time.Hour)
	tokens.SetToken("known-token", data)

	tests := []struct {
		name   string
		header string
		want   TokenData
	}{
		{"known token resolves", "Bearer known-token", data},
		{"unknown token resolves to nil", "Bearer nope", nil},
		{"no Authorization header at all", "", nil},
		{"malformed header (no scheme)", "known-token", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthBearer{tokens: tokens}
			ctx := newCtxWithAuth(tt.header)
			got := a.GetToken(ctx)
			if tt.want == nil {
				assert.Nil(t, got)
			} else if assert.NotNil(t, got) {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAuthBearer_Auth(t *testing.T) {
	data := &testTokenData{id: 9, isAdmin: false}
	tokens := NewMapTokens(time.Hour)
	tokens.SetToken("valid", data)

	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"valid bearer token", "Bearer valid", true},
		{"unknown bearer token", "Bearer unknown", false},
		{"no header", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthBearer{tokens: tokens}
			ctx := newCtxWithAuth(tt.header)
			got := a.Auth(ctx)
			assert.Equal(t, tt.want, got)
			if tt.want {
				// Auth's whole side effect - CheckAndRun and every
				// NeedAuth-gated handler read the validated TokenData back
				// out of ctx this way, not by calling GetToken again.
				assert.Equal(t, data, ctx.UserValue(UserValueToken))
			} else {
				assert.Nil(t, ctx.UserValue(UserValueToken))
			}
		})
	}
}

func TestAuthBearer_AdminAuth(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	tokens.SetToken("admin-tok", &testTokenData{id: 1, isAdmin: true})
	tokens.SetToken("user-tok", &testTokenData{id: 2, isAdmin: false})

	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"admin token", "Bearer admin-tok", true},
		{"non-admin token", "Bearer user-tok", false},
		{"unknown token", "Bearer nope", false},
		{"no header", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthBearer{tokens: tokens}
			ctx := newCtxWithAuth(tt.header)
			assert.Equal(t, tt.want, a.AdminAuth(ctx))
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

// TestAuthBearer_NewToken_ReturnsError isn't reachable through *MapTokens
// (generateRandomString only fails if the system CSPRNG itself fails,
// which can't be simulated from here) - documenting that gap rather than
// silently having zero coverage of NewToken's error return.
func TestAuthBearer_NewToken_ReturnsError(t *testing.T) {
	t.Skip("AuthBearer.NewToken only errors if crypto/rand.Read fails (MapTokens.NewToken -> generateRandomString) - not something this test can force without a fake Tokens implementation the real code doesn't accept")
}

func TestAuthBearer_String(t *testing.T) {
	a := &AuthBearer{tokens: NewMapTokens(time.Hour)}
	got := a.String()

	assert.Contains(t, got, "Bearer standart")
	assert.Contains(t, got, "user:")
	assert.Contains(t, got, "admin:")
	// getStringOfFnc renders an <a href='...'> per method - both of
	// AuthBearer's own methods should show up as links, not just static text.
	assert.Contains(t, got, "<a href=")
}

func TestAuthBearer_getBearer_TableDriven(t *testing.T) {
	// Kept separate from TestAuthBearer_getBearer above (which already
	// covers the interesting cases) purely to exercise the zero-value
	// *AuthBearer{} shape the original generated test skeleton used, so a
	// nil tokens field is confirmed not to matter for a method that never
	// touches it.
	a := &AuthBearer{}
	ctx := newCtxWithAuth("Bearer zzz")
	assert.Equal(t, "zzz", a.getBearer(ctx))
}

func TestNewAuthBearer(t *testing.T) {
	t.Run("nil tokens gets a default MapTokens, never a nil Tokens", func(t *testing.T) {
		got := NewAuthBearer(nil)
		if assert.NotNil(t, got) {
			assert.NotNil(t, got.tokens, "NewAuthBearer(nil) must not leave tokens nil - every method below dereferences it unconditionally")
		}
	})

	t.Run("provided tokens instance is used as-is, not replaced", func(t *testing.T) {
		tokens := NewMapTokens(time.Minute)
		got := NewAuthBearer(tokens)
		if assert.NotNil(t, got) {
			assert.Same(t, tokens, got.tokens, "NewAuthBearer must keep the exact instance passed in, so a caller sharing one Tokens across an AuthBearer and something else actually shares state")
		}
	})
}

func Test_getStringOfFnc(t *testing.T) {
	pc := reflect.ValueOf(Test_getStringOfFnc).Pointer()
	got := getStringOfFnc(pc)

	assert.Contains(t, got, "<a href=")
	assert.Contains(t, got, "Test_getStringOfFnc")
	assert.Contains(t, got, "jwt_test.go")
	assert.True(t, strings.Contains(got, "target='_blank'"))
}

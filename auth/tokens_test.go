/*
 * Copyright (c) 2023-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
package auth

import (
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testTokenData struct {
	id      int
	isAdmin bool
}

func (t *testTokenData) IsAdmin() bool {
	return t.isAdmin
}

func (t *testTokenData) GetUserID() int {
	return t.id
}

func Test_mapTokens_GetToken(t *testing.T) {
	type fields struct {
		expiresIn time.Duration
		tokens    map[string]*mapToken
		lock      sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		bearer string
		want   *testTokenData
	}{
		{
			"1",
			fields{
				tokens: map[string]*mapToken{
					"1": {
						userData: &testTokenData{
							id:      1,
							isAdmin: false,
						},
					},
				},
			},
			"1",
			&testTokenData{
				id:      1,
				isAdmin: false,
			},
		},
		{
			"admin flag preserved",
			fields{
				tokens: map[string]*mapToken{
					"admin-token": {
						userData: &testTokenData{
							id:      42,
							isAdmin: true,
						},
					},
				},
			},
			"admin-token",
			&testTokenData{
				id:      42,
				isAdmin: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MapTokens{
				expiresIn: tt.fields.expiresIn,
				tokens:    tt.fields.tokens,
				lock:      tt.fields.lock,
			}

			got := m.GetToken(tt.bearer)
			if assert.NotNil(t, got) {
				assert.Equal(t, tt.want.id, got.GetUserID())
				assert.Equal(t, tt.want.isAdmin, got.IsAdmin())
			}
		})
	}
}

// Test_mapTokens_GetToken_NotFound covers what the table-driven test above
// can't express cleanly: looking up a key that was never stored (nil
// tokens map, and a populated map missing the requested key) must return
// nil, not panic or zero-value a *testTokenData.
func Test_mapTokens_GetToken_NotFound(t *testing.T) {
	t.Run("nil tokens map", func(t *testing.T) {
		m := &MapTokens{}
		assert.Nil(t, m.GetToken("anything"))
	})

	t.Run("populated map, missing key", func(t *testing.T) {
		m := &MapTokens{tokens: map[string]*mapToken{
			"1": {userData: &testTokenData{id: 1}},
		}}
		assert.Nil(t, m.GetToken("2"))
	})
}

func Test_mapTokens_NewToken(t *testing.T) {
	type fields struct {
		expiresIn time.Duration
	}
	type args struct {
		userData TokenData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"basic user",
			fields{expiresIn: time.Hour},
			args{&testTokenData{id: 1, isAdmin: false}},
		},
		{
			"admin user",
			fields{expiresIn: time.Hour},
			args{&testTokenData{id: 2, isAdmin: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMapTokens(tt.fields.expiresIn)

			got, err := m.NewToken(tt.args.userData)
			assert.NoError(t, err)
			assert.NotEmpty(t, got, "NewToken should return a non-empty token string")

			// The minted token must actually be retrievable, and resolve
			// back to the exact TokenData that was minted with it - that's
			// the entire point of NewToken over just generating a random
			// string nobody stored anywhere.
			stored := m.GetToken(got)
			if assert.NotNil(t, stored, "token returned by NewToken must be immediately retrievable via GetToken") {
				assert.Equal(t, tt.args.userData, stored)
			}
		})
	}
}

// Test_mapTokens_NewToken_Unique guards the property AuthBearer's own tests
// already lean on implicitly (see TestAuthBearer_NewToken): repeated calls
// must never mint the same token string, or two unrelated sessions could
// collide.
func Test_mapTokens_NewToken_Unique(t *testing.T) {
	m := NewMapTokens(time.Hour)

	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		got, err := m.NewToken(&testTokenData{id: i})
		assert.NoError(t, err)
		assert.False(t, seen[got], "NewToken produced a duplicate token: %q", got)
		seen[got] = true
	}
}

func Test_mapTokens_RemoveToken(t *testing.T) {
	type fields struct {
		expiresIn time.Duration
		tokens    map[string]*mapToken
		lock      sync.RWMutex
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"existing token removed cleanly",
			fields{
				tokens: map[string]*mapToken{
					"present": {userData: &testTokenData{id: 1}},
				},
			},
			args{"present"},
			false,
		},
		{
			"missing token errors",
			fields{
				tokens: map[string]*mapToken{},
			},
			args{"absent"},
			true,
		},
		{
			"nil tokens map errors rather than panicking",
			fields{
				tokens: nil,
			},
			args{"anything"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MapTokens{
				expiresIn: tt.fields.expiresIn,
				tokens:    tt.fields.tokens,
				lock:      tt.fields.lock,
			}
			err := m.RemoveToken(tt.args.s)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, m.GetToken(tt.args.s), "token must be gone after RemoveToken")
			}
		})
	}
}

func TestMapTokens_SetToken(t *testing.T) {
	m := NewMapTokens(time.Hour)

	data := &testTokenData{id: 7, isAdmin: true}
	m.SetToken("explicit-key", data)

	got := m.GetToken("explicit-key")
	if assert.NotNil(t, got) {
		assert.Equal(t, data, got)
	}

	// SetToken on an existing key overwrites rather than erroring or
	// appending - this is what SimplePasskeyStore's FindOrCreateByLogin
	// relies on being able to do implicitly via the same key (the login).
	overwritten := &testTokenData{id: 8, isAdmin: false}
	m.SetToken("explicit-key", overwritten)
	got = m.GetToken("explicit-key")
	if assert.NotNil(t, got) {
		assert.Equal(t, overwritten, got)
	}
}

// TestMapTokens_SetToken_Expires is the one timer-dependent test in this
// file: SetToken always arms a deletion timer for m.expiresIn (see its own
// doc comment on the Tokens interface for why that matters for anything
// stored under an explicit key rather than a fresh random one). A very
// short expiresIn plus a polling wait (not a single fixed sleep) keeps this
// from being flaky under a loaded CI runner while still proving the timer
// actually fires.
func TestMapTokens_SetToken_Expires(t *testing.T) {
	m := NewMapTokens(20 * time.Millisecond)
	m.SetToken("short-lived", &testTokenData{id: 1})

	assert.NotNil(t, m.GetToken("short-lived"), "token should be present immediately after SetToken")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if m.GetToken("short-lived") == nil {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("token was not removed after its expiresIn duration elapsed")
}

func TestNewMapTokens(t *testing.T) {
	m := NewMapTokens(time.Hour)
	assert.Equal(t, time.Hour, m.expiresIn)
	assert.NotNil(t, m.tokens, "NewMapTokens must initialize the tokens map so GetToken/SetToken never see a nil map")
	assert.Empty(t, m.tokens)
}

func TestNewSimpleTokenData(t *testing.T) {
	expiry := time.Now().Add(time.Hour)
	d := NewSimpleTokenData("bik4ruslan@gmail.com", "Ruslan", "en", 42, true, expiry)

	assert.Equal(t, "bik4ruslan@gmail.com", d.Name)
	assert.Equal(t, "Ruslan", d.Desc)
	assert.Equal(t, "en", d.Lang)
	assert.Equal(t, 42, d.Id)
	assert.True(t, d.Admin)
	assert.Equal(t, expiry, d.Expiry)
	assert.Empty(t, d.Token, "NewSimpleTokenData doesn't take a token - WithToken sets it later")
	assert.True(t, d.IsAdmin())
	assert.Equal(t, 42, d.GetUserID())
}

func TestSimpleTokenData_WithToken(t *testing.T) {
	d := NewSimpleTokenData("login", "desc", "en", 1, false, time.Time{})

	got := d.WithToken("tok-123")
	assert.Same(t, d, got, "WithToken returns the same receiver for chaining")
	assert.Equal(t, "tok-123", d.Token)

	// Nil-safe, like WithExtension - callers elsewhere (passkey.go's
	// tokenWithSetter check) rely on this never panicking even if handed a
	// nil *SimpleTokenData somehow.
	var nilData *SimpleTokenData
	assert.Nil(t, nilData.WithToken("x"))
}

func TestSimpleTokenData_WithExtension(t *testing.T) {
	d := NewSimpleTokenData("login", "desc", "en", 1, false, time.Time{})
	assert.Nil(t, d.Extensions)

	got := d.WithExtension("theme", "dark")
	assert.Same(t, d, got)
	assert.Equal(t, "dark", d.Extensions["theme"])

	// A second call on an already-initialized Extensions map must not
	// clobber the first entry.
	d.WithExtension("formActions", []string{"save"})
	assert.Equal(t, "dark", d.Extensions["theme"])
	assert.Equal(t, []string{"save"}, d.Extensions["formActions"])

	var nilData *SimpleTokenData
	assert.Nil(t, nilData.WithExtension("k", "v"))
}

func TestSimpleTokenData_IsNotExpired(t *testing.T) {
	future := &SimpleTokenData{Expiry: time.Now().Add(time.Hour)}
	assert.True(t, future.IsNotExpired())

	past := &SimpleTokenData{Expiry: time.Now().Add(-time.Hour)}
	assert.False(t, past.IsNotExpired())
}

func Test_generateRandomBytes(t *testing.T) {
	b, err := generateRandomBytes(16)
	assert.NoError(t, err)
	assert.Len(t, b, 16)

	b2, err := generateRandomBytes(16)
	assert.NoError(t, err)
	assert.NotEqual(t, b, b2, "two independent calls should not produce the same random bytes")
}

func Test_generateRandomString(t *testing.T) {
	s, err := generateRandomString(16)
	assert.NoError(t, err)
	assert.Equal(t, base64.URLEncoding.EncodedLen(16), len(s), "length must match base64.URLEncoding's own encoded length for 16 bytes")

	s2, err := generateRandomString(16)
	assert.NoError(t, err)
	assert.NotEqual(t, s, s2)
}

/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
package auth

import (
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
)

// --- NewSimplePasskeyStore -------------------------------------------------

func TestNewSimplePasskeyStore_NilTokensGetsDefault(t *testing.T) {
	s := NewSimplePasskeyStore(nil)
	if assert.NotNil(t, s) {
		assert.NotNil(t, s.tokens, "NewSimplePasskeyStore(nil) must not leave tokens nil - lookup/FindOrCreateByLogin dereference it unconditionally")
	}
}

func TestNewSimplePasskeyStore_ProvidedTokensUsedAsIs(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	s := NewSimplePasskeyStore(tokens)
	assert.Same(t, tokens, s.tokens)
}

// --- lookup -----------------------------------------------------------

func Test_SimplePasskeyStore_lookup_NotFound(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	rec, err := s.lookup("nobody@example.com")
	assert.NoError(t, err)
	assert.Nil(t, rec)
}

func Test_SimplePasskeyStore_lookup_WrongTypeStored(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	// Simulate an ordinary Bearer session colliding, in tokens, with a login
	// string used as a passkey key - lookup must report this as an error,
	// not silently type-assert-panic or treat it as "not found".
	tokens.SetToken("collides@example.com", &testTokenData{id: 1})
	s := NewSimplePasskeyStore(tokens)

	rec, err := s.lookup("collides@example.com")
	assert.Error(t, err)
	assert.Nil(t, rec)
}

func Test_SimplePasskeyStore_lookup_Found(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	s := NewSimplePasskeyStore(tokens)

	created, err := s.FindOrCreateByLogin("someone@example.com")
	if !assert.NoError(t, err) {
		return
	}

	rec, err := s.lookup("someone@example.com")
	assert.NoError(t, err)
	if assert.NotNil(t, rec) {
		assert.Same(t, created.(simplePasskeyUser).rec, rec)
	}
}

// --- FindByLogin --------------------------------------------------------

func TestSimplePasskeyStore_FindByLogin_Unknown(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	u, err := s.FindByLogin("never-registered@example.com")
	assert.Nil(t, u)
	assert.Equal(t, ErrPasskeyUnknownLogin, err)
}

func TestSimplePasskeyStore_FindByLogin_Existing(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	_, err := s.FindOrCreateByLogin("has-passkey@example.com")
	if !assert.NoError(t, err) {
		return
	}

	u, err := s.FindByLogin("has-passkey@example.com")
	assert.NoError(t, err)
	if assert.NotNil(t, u) {
		assert.Equal(t, "has-passkey@example.com", u.PasskeyLogin())
	}
}

func TestSimplePasskeyStore_FindByLogin_WrongTypeStored(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	tokens.SetToken("collides@example.com", &testTokenData{id: 1})
	s := NewSimplePasskeyStore(tokens)

	u, err := s.FindByLogin("collides@example.com")
	assert.Nil(t, u)
	assert.Error(t, err)
}

// --- FindOrCreateByLogin ------------------------------------------------

func TestSimplePasskeyStore_FindOrCreateByLogin_CreatesOnce(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	s := NewSimplePasskeyStore(tokens)

	first, err := s.FindOrCreateByLogin("new@example.com")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "new@example.com", first.PasskeyLogin())

	// A registered credential must survive a second FindOrCreateByLogin
	// call for the same login - it must resolve the EXISTING record, not
	// silently overwrite it with a fresh empty one.
	cred := webauthn.Credential{ID: []byte("cred-1")}
	assert.NoError(t, s.AddCredential(first, cred))

	second, err := s.FindOrCreateByLogin("new@example.com")
	if !assert.NoError(t, err) {
		return
	}
	assert.Same(t, first.(simplePasskeyUser).rec, second.(simplePasskeyUser).rec)
	assert.Len(t, second.PasskeyCredentials(), 1)
}

func TestSimplePasskeyStore_FindOrCreateByLogin_WrongTypeStored(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	tokens.SetToken("collides@example.com", &testTokenData{id: 1})
	s := NewSimplePasskeyStore(tokens)

	u, err := s.FindOrCreateByLogin("collides@example.com")
	assert.Nil(t, u)
	assert.Error(t, err)
}

// --- FindByToken --------------------------------------------------------

func TestSimplePasskeyStore_FindByToken_WrongType(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	u, err := s.FindByToken(&testTokenData{id: 1})
	assert.Nil(t, u)
	assert.Error(t, err, "SimplePasskeyStore only supports *SimpleTokenData")
}

func TestSimplePasskeyStore_FindByToken_FirstTimeCreatesRecordFromToken(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	s := NewSimplePasskeyStore(tokens)
	data := &SimpleTokenData{Name: "first-passkey@example.com"}

	u, err := s.FindByToken(data)
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, u) {
		// The new record must be built from the exact token handed in, not
		// a fresh zero-value one - AddCredential (called right after, at
		// register/finish) needs the caller's own SimpleTokenData fields
		// (Desc/Lang/Id/Admin) preserved, not just the login.
		assert.Same(t, data, u.(simplePasskeyUser).rec.data)
	}

	// A second lookup for the same login must resolve the SAME record
	// FindByToken just created, not build another one.
	again, err := s.FindByLogin("first-passkey@example.com")
	assert.NoError(t, err)
	if assert.NotNil(t, again) {
		assert.Same(t, u.(simplePasskeyUser).rec, again.(simplePasskeyUser).rec)
	}
}

func TestSimplePasskeyStore_FindByToken_ExistingRecordIgnoresPassedData(t *testing.T) {
	tokens := NewMapTokens(time.Hour)
	s := NewSimplePasskeyStore(tokens)

	original, err := s.FindOrCreateByLogin("existing@example.com")
	if !assert.NoError(t, err) {
		return
	}

	// Passing a different *SimpleTokenData for an already-registered login
	// must resolve the EXISTING record (keyed by data.Name), not replace it.
	u, err := s.FindByToken(&SimpleTokenData{Name: "existing@example.com", Desc: "ignored"})
	assert.NoError(t, err)
	if assert.NotNil(t, u) {
		assert.Same(t, original.(simplePasskeyUser).rec, u.(simplePasskeyUser).rec)
	}
}

// --- AddCredential / UpdateCredential -------------------------------------

func TestSimplePasskeyStore_AddCredential(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	user, err := s.FindOrCreateByLogin("adder@example.com")
	if !assert.NoError(t, err) {
		return
	}

	assert.Empty(t, user.PasskeyCredentials())

	cred := webauthn.Credential{ID: []byte("cred-1")}
	assert.NoError(t, s.AddCredential(user, cred))

	got := user.PasskeyCredentials()
	if assert.Len(t, got, 1) {
		assert.Equal(t, cred, got[0])
	}
}

func TestSimplePasskeyStore_AddCredential_WrongUserType(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	err := s.AddCredential(&fakePasskeyUser{login: "x"}, webauthn.Credential{ID: []byte("c")})
	assert.Error(t, err, "AddCredential must reject a PasskeyUser this store didn't create")
}

func TestSimplePasskeyStore_UpdateCredential(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	user, err := s.FindOrCreateByLogin("updater@example.com")
	if !assert.NoError(t, err) {
		return
	}

	// Authenticator.SignCount per go-webauthn v0.18.1's webauthn.Authenticator
	// struct (see passkey.go's own version note at the top of the file) -
	// not independently verified against the real dependency in this
	// sandbox (no network access to fetch it), but this is a stable,
	// long-standing field on that type.
	original := webauthn.Credential{ID: []byte("cred-1"), Authenticator: webauthn.Authenticator{SignCount: 1}}
	assert.NoError(t, s.AddCredential(user, original))

	updated := webauthn.Credential{ID: []byte("cred-1"), Authenticator: webauthn.Authenticator{SignCount: 2}}
	assert.NoError(t, s.UpdateCredential(user, updated))

	got := user.PasskeyCredentials()
	if assert.Len(t, got, 1) {
		assert.Equal(t, uint32(2), got[0].Authenticator.SignCount, "UpdateCredential must persist the authenticator's new signature counter")
	}
}

func TestSimplePasskeyStore_UpdateCredential_NotFound(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	user, err := s.FindOrCreateByLogin("no-such-cred@example.com")
	if !assert.NoError(t, err) {
		return
	}

	err = s.UpdateCredential(user, webauthn.Credential{ID: []byte("never-added")})
	assert.Error(t, err)
}

func TestSimplePasskeyStore_UpdateCredential_WrongUserType(t *testing.T) {
	s := NewSimplePasskeyStore(NewMapTokens(time.Hour))
	err := s.UpdateCredential(&fakePasskeyUser{login: "x"}, webauthn.Credential{ID: []byte("c")})
	assert.Error(t, err)
}

// --- simplePasskeyUser ------------------------------------------------

func Test_simplePasskeyUser_PasskeyID_PasskeyLogin(t *testing.T) {
	rec := &simplePasskeyRecord{data: &SimpleTokenData{Name: "id-and-login@example.com", Id: 42}}
	u := simplePasskeyUser{rec: rec}

	// PasskeyID is derived from data.Id's decimal string form, not the
	// login - see PasskeyID's own doc comment for why (Id is an int, not
	// directly convertible to []byte).
	assert.Equal(t, []byte("42"), u.PasskeyID())
	assert.Equal(t, "id-and-login@example.com", u.PasskeyLogin())
}

// Test_simplePasskeyUser_PasskeyID_PlaceholderCollision documents a real
// risk of keying PasskeyID off SimpleTokenData.Id: callers that leave Id at
// a shared placeholder (e.g. HandleSignIn's -1) produce the SAME PasskeyID
// for different accounts - see PasskeyID's own doc comment.
func Test_simplePasskeyUser_PasskeyID_PlaceholderCollision(t *testing.T) {
	a := simplePasskeyUser{rec: &simplePasskeyRecord{data: &SimpleTokenData{Name: "a@example.com", Id: -1}}}
	b := simplePasskeyUser{rec: &simplePasskeyRecord{data: &SimpleTokenData{Name: "b@example.com", Id: -1}}}

	assert.Equal(t, a.PasskeyID(), b.PasskeyID(), "two different accounts left at the same placeholder Id collide on PasskeyID - see this test's own doc comment")
}

func Test_simplePasskeyUser_PasskeyDisplayName(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{"real display name kept", "Ruslan Bikchentaev", "Ruslan Bikchentaev"},
		// HandleSignIn's placeholder literal - must fall back to the login
		// rather than showing "desc" in a Face ID/Touch ID prompt.
		{"placeholder 'desc' falls back to login", "desc", "login@example.com"},
		{"empty falls back to login", "", "login@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &simplePasskeyRecord{data: &SimpleTokenData{Name: "login@example.com", Desc: tt.desc}}
			u := simplePasskeyUser{rec: rec}
			assert.Equal(t, tt.want, u.PasskeyDisplayName())
		})
	}
}

func Test_simplePasskeyUser_PasskeyCredentials_ReturnsCopy(t *testing.T) {
	rec := &simplePasskeyRecord{
		data:  &SimpleTokenData{Name: "copy@example.com"},
		creds: []webauthn.Credential{{ID: []byte("original")}},
	}
	u := simplePasskeyUser{rec: rec}

	got := u.PasskeyCredentials()
	got[0].ID = []byte("mutated")

	// The record's own slice must be unaffected - webauthnUserAdapter must
	// never be able to see a concurrent request's in-progress mutation.
	assert.Equal(t, "original", string(rec.creds[0].ID))
}

func Test_simplePasskeyUser_NewToken_LoginResponse_SameObject(t *testing.T) {
	data := &SimpleTokenData{Name: "same-object@example.com"}
	u := simplePasskeyUser{rec: &simplePasskeyRecord{data: data}}

	tok, err := u.NewToken()
	assert.NoError(t, err)
	assert.Same(t, data, tok)

	// LoginResponse must return the SAME pointer NewToken did - this is what
	// lets FinishLogin's tokenWithSetter attachment (passkey.go) show up in
	// the response without any extra plumbing.
	assert.Same(t, data, u.LoginResponse())
}

// --- CAVEAT: passkey records inherit session expiry ------------------------
//
// Documented at length in SimplePasskeyStore's own doc comment - this test
// proves it's real behavior, not just a comment, so a future change to
// MapTokens.SetToken (e.g. an indefinite-expiry option) has something here
// to notice it broke this store's known limitation rather than silently
// changing behavior.

func TestSimplePasskeyStore_CAVEAT_RecordExpiresWithSession(t *testing.T) {
	tokens := NewMapTokens(20 * time.Millisecond)
	s := NewSimplePasskeyStore(tokens)

	_, err := s.FindOrCreateByLogin("short-lived@example.com")
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, tokens.GetToken("short-lived@example.com"), "record should be present immediately after creation")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if tokens.GetToken("short-lived@example.com") == nil {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("passkey record was not removed after tokens' expiresIn elapsed - if this now fails, SimplePasskeyStore's CAVEAT doc comment may be stale")
}

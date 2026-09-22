/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
package auth

// Minimal in-memory PasskeyStore/PasskeyUser implementation - a
// getting-started reference for wiring NewWebAuthnPasskey(...), NOT a
// production store. Swap this for one backed by your real user table:
// PasskeyID/PasskeyLogin/PasskeyDisplayName/PasskeyCredentials/NewToken/
// LoginResponse become columns + a credentials join table, and
// AddCredential/UpdateCredential become real writes instead of an
// in-process map that's lost on every restart and unsafe across more than
// one server instance (see sessionCache's own doc comment in passkey.go -
// same caveat applies here).

import (
	"bytes"
	"crypto/rand"
	"errors"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// memPasskeyUser is the in-memory stand-in for the app's real user/account
// record. In a real app this is likely your existing users struct/row with
// a few extra methods added (or a thin wrapper around it), not a separate
// type - PasskeyID especially should be a column on that row (see
// PasskeyUser.PasskeyID's doc comment on why it must never change).
type memPasskeyUser struct {
	id          []byte
	login       string
	displayName string
	lang        string

	// RWMutex, not Mutex: PasskeyCredentials/LoginResponse below only read
	// (RLock), while NewToken/AddCredential/UpdateCredential mutate (Lock) -
	// same read/write split MapTokens already uses for its own lock
	// (auth/tokens.go).
	mu    sync.RWMutex
	creds []webauthn.Credential

	// tokenData is set by NewToken and read back by LoginResponse - both
	// return the SAME *SimpleTokenData object, so the Bearer token
	// (*WebAuthnPasskey).FinishLogin attaches onto it after NewToken runs
	// (via its tokenWithSetter check) is still there when LoginResponse is
	// called right after. See PasskeyUser.LoginResponse's own doc
	// comment for why this replaces building a separate response map.
	tokenData *SimpleTokenData
}

func (u *memPasskeyUser) PasskeyID() []byte          { return u.id }
func (u *memPasskeyUser) PasskeyLogin() string       { return u.login }
func (u *memPasskeyUser) PasskeyDisplayName() string { return u.displayName }

func (u *memPasskeyUser) PasskeyCredentials() []webauthn.Credential {
	// RLock, not Lock - this only reads u.creds, it never mutates it.
	u.mu.RLock()
	defer u.mu.RUnlock()
	// Return a copy - the caller (webauthnUserAdapter, inside passkey.go)
	// must never see a slice that a concurrent request could mutate out
	// from under it.
	out := make([]webauthn.Credential, len(u.creds))
	copy(out, u.creds)
	return out
}

// NewToken issues this app's own TokenData for a successful ceremony.
// SimpleTokenData mirrors what HandleSignIn (routes.qtpl) builds for a
// password login, so a passkey login and a password login return the same
// shape to the browser. Stored on u.tokenData (not just returned) so
// LoginResponse can hand back this exact object once FinishLogin has
// attached the freshly-minted Bearer token to it.
func (u *memPasskeyUser) NewToken() (TokenData, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.tokenData = NewSimpleTokenData(u.login, u.displayName, u.lang, -1, true, time.Now().Add(time.Hour))
	return u.tokenData, nil
}

// LoginResponse returns the same *SimpleTokenData NewToken just
// created - by the time (*WebAuthnPasskey).FinishLogin calls this, it
// already carries the login's Bearer token too.
func (u *memPasskeyUser) LoginResponse() any {
	// RLock, not Lock - this only reads u.tokenData, it never mutates it.
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.tokenData
}

// memPasskeyStore is a process-local, single-instance PasskeyStore.
// Fine for a demo/single-node deployment; behind more than one server
// instance you need a shared store (a real database, same as your
// existing users table) so registration on one node is visible to a login
// on another.
type memPasskeyStore struct {
	*AuthBearer
	// RWMutex, not Mutex: FindByLogin/FindByToken below only read byLogin
	// (RLock), while addUser/FindOrCreateByLogin mutate it (Lock).
	mu      sync.RWMutex
	byLogin map[string]*memPasskeyUser
}

// newMemPasskeyStore builds the demo store. tokens is handed straight to
// auth.NewAuthBearer (nil is fine - it falls back to its own
// auth.NewMapTokens default) so this store's Bearer-auth capability shares
// the exact same Tokens instance the rest of the app uses - see main.go's
// own comment on why that matters.
func newMemPasskeyStore(tokens Tokens) *memPasskeyStore {
	return &memPasskeyStore{
		AuthBearer: NewAuthBearer(tokens),
		byLogin:    make(map[string]*memPasskeyUser),
	}
}

// addUser is demo-only plumbing to seed an account that can then register a
// passkey - a real app already has this step (whatever your sign-up flow
// is), it wouldn't be part of PasskeyStore itself.
func (s *memPasskeyStore) addUser(login, displayName, lang string) *memPasskeyUser {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)

	u := &memPasskeyUser{
		id:          buf,
		login:       login,
		displayName: displayName,
		lang:        lang,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.byLogin[login] = u

	return u
}

func (s *memPasskeyStore) FindByLogin(login string) (PasskeyUser, error) {
	// RLock, not Lock - this only reads s.byLogin, it never mutates it.
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byLogin[login]
	if !ok {
		// Deliberately the same error/shape as "found but no passkey yet"
		// would be - see PasskeyStore.FindByLogin's doc comment on not
		// leaking which logins exist.
		return nil, errors.New("passkey: no such account")
	}

	return u, nil
}

// FindOrCreateByLogin resolves login for a registration ceremony, creating
// a new account if none exists yet - this is what lets BeginRegistration
// (passkey.go) work for a login it's never seen before, now that
// register/begin identifies its caller by login instead of an existing
// Bearer token. A real store would still want its own actual sign-up flow
// for the non-passkey fields (email verification, ToS acceptance, ...);
// this demo just creates the row outright.
func (s *memPasskeyStore) FindOrCreateByLogin(login string) (PasskeyUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.byLogin[login]; ok {
		return u, nil
	}

	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	u := &memPasskeyUser{id: buf, login: login, displayName: login, lang: "en"}
	s.byLogin[login] = u

	return u, nil
}

func (s *memPasskeyStore) FindByToken(token TokenData) (PasskeyUser, error) {
	// auth.TokenData only guarantees IsAdmin/GetUserID - it has no String()
	// method (a previous version of this function called token.String(),
	// which doesn't compile). *SimpleTokenData does carry a login
	// (Name), so type-assert to that instead - swap this for however your
	// real store actually maps a validated TokenData back to an account if
	// it uses a different TokenData implementation (most likely a user id
	// embedded in the token data itself, not a login string).
	data, ok := token.(*SimpleTokenData)
	if !ok {
		return nil, errors.New("passkey: memPasskeyStore only supports *SimpleTokenData")
	}

	// RLock, not Lock - this only reads s.byLogin, it never mutates it.
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byLogin[data.Name]
	if !ok {
		return nil, errors.New("passkey: unknown token subject")
	}

	return u, nil
}

func (s *memPasskeyStore) AddCredential(user PasskeyUser, cred webauthn.Credential) error {
	u, ok := user.(*memPasskeyUser)
	if !ok {
		return errors.New("passkey: unexpected user type")
	}

	u.mu.Lock()
	defer u.mu.Unlock()
	u.creds = append(u.creds, cred)

	return nil
}

func (s *memPasskeyStore) UpdateCredential(user PasskeyUser, cred webauthn.Credential) error {
	u, ok := user.(*memPasskeyUser)
	if !ok {
		return errors.New("passkey: unexpected user type")
	}

	u.mu.Lock()
	defer u.mu.Unlock()
	for i := range u.creds {
		// bytes.Equal, not string(a) == string(b) - compares the two []byte
		// values directly instead of allocating two throwaway strings just
		// to compare them.
		if bytes.Equal(u.creds[i].ID, cred.ID) {
			u.creds[i] = cred
			return nil
		}
	}

	return errors.New("passkey: credential not found")
}

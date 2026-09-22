/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
package auth

import (
	"bytes"
	"errors"
	"strconv"
	"sync"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ruslanBik4/gotools"
)

// SimplePasskeyStore is the PasskeyStore NewWebAuthnPasskey falls back to
// when no store is supplied (store == nil). It keeps no registry of its
// own - every lookup and every write goes straight through the same Tokens
// store the app already has (via its embedded *AuthBearer's tokens field),
// keyed by login instead of by a random session-token string. That's a
// reasonable default for a project that hasn't wired up its own persistent
// user store yet, not a production account system - read the CAVEAT below
// before relying on it for more than getting a first end-to-end passkey
// flow running.
//
// CAVEAT - passkey records inherit session expiry: Tokens/MapTokens was
// built to track CURRENTLY ACTIVE sessions, each one deleted automatically
// after MapTokens's own expiresIn duration (MapTokens.SetToken always arms
// that timer - see its doc comment in tokens.go). Storing a passkey record
// under tokens.SetToken(login, rec) puts it on that SAME timer: a
// registered passkey will silently stop working and its credentials will
// be gone after expiresIn elapses, even though a passkey should keep
// working indefinitely - there's no session to "expire" in the first
// place, that's the whole point of being able to log back in with one.
// This is a direct consequence of doing every lookup through tokens
// instead of a separate registry, so it's called out here rather than
// hidden - if that expiry is a problem for your deployment, either give
// this store a real persistent backing store (a database row has no
// expiresIn), or configure a Tokens instance dedicated to passkey records
// with an expiresIn long enough that it never practically matters (there's
// still no way to make it truly indefinite with MapTokens as it stands -
// SetToken always arms a timer).
type SimplePasskeyStore struct {
	// *AuthBearer serves double duty here: it's what makes SimplePasskeyStore
	// satisfy PasskeyStore's embedded FncAuth/TokenIssuer requirement
	// (Auth/AdminAuth/String/NewToken, promoted straight through from
	// *WebAuthnPasskey - see PasskeyStore's own doc comment for why that
	// requirement exists), and it also gets every method below at its
	// tokens field and reuses NewAuthBearer's nil-tokens default
	// (NewMapTokens(tokenExpires)) for free.
	*AuthBearer
}

// simplePasskeyRecord is what SimplePasskeyStore actually persists per
// account, stored as the TokenData value against tokens under the login as
// key. It implements TokenData itself (IsAdmin/GetUserID, delegating to
// data) so it can be stored directly - Tokens has nowhere else to put
// anything that isn't a TokenData.
type simplePasskeyRecord struct {
	// RWMutex, not Mutex: PasskeyCredentials below only reads creds (RLock),
	// while AddCredential/UpdateCredential mutate it (Lock) - same
	// read/write split MapTokens already uses for its own lock (tokens.go).
	mu    sync.RWMutex
	data  *SimpleTokenData
	creds []webauthn.Credential
}

func (r *simplePasskeyRecord) IsAdmin() bool  { return r.data.IsAdmin() }
func (r *simplePasskeyRecord) GetUserID() int { return r.data.GetUserID() }

// NewSimplePasskeyStore builds the default PasskeyStore, backed by tokens -
// see SimplePasskeyStore's own doc comment for what every lookup/write
// below actually does with it, and the CAVEAT on record expiry. tokens may
// be nil (NewAuthBearer's own default then applies, same as everywhere
// else in this package).
func NewSimplePasskeyStore(tokens Tokens) *SimplePasskeyStore {
	return &SimplePasskeyStore{NewAuthBearer(tokens)}
}

// lookup fetches the *simplePasskeyRecord stored under login, if any.
// Returns nil, nil for "no such record" (not an error - both FindByLogin
// and FindOrCreateByLogin need to tell that apart from "found") and an
// error only if tokens has something else stored under that key (a plain
// *SimpleTokenData from an ordinary logged-in session, say, colliding with
// a login that also happens to match some other caller's random session
// token - astronomically unlikely with MapTokens's 16-random-byte tokens,
// but a login is caller-chosen text, not guaranteed disjoint from that
// keyspace in some other Tokens implementation).
func (s *SimplePasskeyStore) lookup(login string) (*simplePasskeyRecord, error) {
	td := s.tokens.GetToken(login)
	if td == nil {
		return nil, nil
	}

	rec, ok := td.(*simplePasskeyRecord)
	if !ok {
		return nil, errors.New("passkey: tokens already holds a non-passkey record under this login")
	}

	return rec, nil
}

// simplePasskeyUser adapts a simplePasskeyRecord (identity + credentials)
// to PasskeyUser. Unexported - the only way to get one is through this
// store's own FindByLogin/FindOrCreateByLogin/FindByToken, so
// AddCredential/UpdateCredential can safely assume any PasskeyUser they're
// handed back is one of these.
type simplePasskeyUser struct {
	rec *simplePasskeyRecord
}

func (u simplePasskeyUser) PasskeyID() []byte {
	// data.Id is an int, not directly convertible to []byte ([]byte(int) is
	// not a valid Go conversion) - go through its decimal string form
	// instead, via gotools.StringToByte (the zero-copy counterpart to
	// gotools.BytesToString already used elsewhere in this codebase) rather
	// than a plain []byte(string) conversion.
	//
	// CAVEAT: routes.qtpl's HandleSignIn (and possibly other callers) leaves
	// Id at a placeholder (e.g. -1) wherever a real database id hasn't been
	// wired through yet. Every account still on that placeholder collides on
	// the exact same PasskeyID - see
	// Test_simplePasskeyUser_PasskeyID_PlaceholderCollision, which pins this
	// down explicitly. WebAuthn expects the user handle to uniquely and
	// stably identify one account, so this needs a real per-account id
	// before relying on it beyond a first end-to-end passkey test.
	return gotools.StringToBytes(strconv.Itoa(u.rec.data.Id))
}

func (u simplePasskeyUser) PasskeyLogin() string { return u.rec.data.Name }

func (u simplePasskeyUser) PasskeyDisplayName() string {
	// HandleSignIn (routes.qtpl) currently always passes the literal
	// string "desc" as SimpleTokenData.Desc - not a real display name -
	// so fall back to the login whenever Desc looks like that placeholder
	// or is empty, rather than showing "desc" in a Face ID/Touch ID
	// prompt.
	if u.rec.data.Desc != "" && u.rec.data.Desc != "desc" {
		return u.rec.data.Desc
	}

	return u.rec.data.Name
}

func (u simplePasskeyUser) PasskeyCredentials() []webauthn.Credential {
	// RLock, not Lock - this only reads u.rec.creds, it never mutates it.
	u.rec.mu.RLock()
	defer u.rec.mu.RUnlock()

	// Return a copy - webauthnUserAdapter (passkey.go) must never see a
	// slice a concurrent request could mutate out from under it.
	out := make([]webauthn.Credential, len(u.rec.creds))
	copy(out, u.rec.creds)

	return out
}

// NewToken returns the SAME *SimpleTokenData this user was resolved from -
// nothing about it needs to change for a fresh passkey-issued session
// ((*WebAuthnPasskey).FinishLogin mints the actual new Bearer token string
// separately, via its embedded *AuthBearer, then attaches it back onto
// this exact object - see passkey.go's tokenWithSetter). Returning the
// same pointer LoginResponse() below also returns is what lets that
// attached token show up in the response without FinishLogin needing to
// build one itself.
func (u simplePasskeyUser) NewToken() (TokenData, error) {
	return u.rec.data, nil
}

// LoginResponse returns u.rec.data as-is - the *SimpleTokenData already
// carries every field the browser needs (name, lang, and - by the time
// FinishLogin calls this, right after minting one - token), the same shape
// a password login already returns. No separate map to keep in sync with
// SimpleTokenData's own fields.
func (u simplePasskeyUser) LoginResponse() any {
	return u.rec.data
}

// FindByLogin resolves whatever the user typed into the login form. Only
// succeeds for a login that already has a record in tokens - i.e. one that
// has already completed at least one successful AddCredential (registered
// a passkey, whether via register/begin's implicit sign-up or having
// registered a device while authenticated some other way first).
func (s *SimplePasskeyStore) FindByLogin(login string) (PasskeyUser, error) {
	rec, err := s.lookup(login)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		// Same error regardless of "no such login" vs. "login exists but
		// never registered a passkey" - see PasskeyStore.FindByLogin's own
		// doc comment on not leaking which logins exist.
		return nil, ErrPasskeyUnknownLogin
	}

	return simplePasskeyUser{rec: rec}, nil
}

// FindOrCreateByLogin resolves login for a registration ceremony, creating
// an empty-credentials record in tokens if none exists yet - this is what
// lets BeginRegistration work for a login it has never seen before (no
// separate sign-up step is needed with this store). See PasskeyID's own
// comment on this store's caveats around using the login itself as the
// account identity, and the type doc comment above on what storing this in
// tokens means for how long it lasts.
func (s *SimplePasskeyStore) FindOrCreateByLogin(login string) (PasskeyUser, error) {
	rec, err := s.lookup(login)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		rec = &simplePasskeyRecord{data: &SimpleTokenData{Name: login}}
		s.tokens.SetToken(login, rec)
	}

	return simplePasskeyUser{rec: rec}, nil
}

// FindByToken resolves an already-authenticated caller straight from the
// TokenData CheckAndRun already validated and attached to ctx - looked up
// (or created, if this is that account's first passkey) the same way as
// FindOrCreateByLogin, keyed by the token's own login. No longer called by
// BeginRegistration (see its doc comment in passkey.go); kept for any
// account-management endpoint a project adds later. Only *SimpleTokenData
// is supported; a project using a different TokenData implementation needs
// its own PasskeyStore.
func (s *SimplePasskeyStore) FindByToken(token TokenData) (PasskeyUser, error) {
	data, ok := token.(*SimpleTokenData)
	if !ok {
		return nil, errors.New("passkey: SimplePasskeyStore only supports *auth.SimpleTokenData - provide a custom PasskeyStore for a different TokenData implementation")
	}

	rec, err := s.lookup(data.Name)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		// First passkey for this account - register the record now so
		// AddCredential (called right after this, at register/finish) has
		// somewhere to append to.
		rec = &simplePasskeyRecord{data: data}
		s.tokens.SetToken(data.Name, rec)
	}

	return simplePasskeyUser{rec: rec}, nil
}

// AddCredential persists a newly-registered credential against the account
// it belongs to. Called once, at register/finish. No need to write back
// through tokens.SetToken after mutating rec - rec is the exact pointer
// tokens is already holding (GetToken/lookup never copies), so mutating it
// in place under its own mutex is visible to every future lookup without a
// second store.
func (s *SimplePasskeyStore) AddCredential(user PasskeyUser, cred webauthn.Credential) error {
	u, ok := user.(simplePasskeyUser)
	if !ok {
		return errors.New("passkey: SimplePasskeyStore was handed a PasskeyUser it didn't create")
	}

	u.rec.mu.Lock()
	defer u.rec.mu.Unlock()
	u.rec.creds = append(u.rec.creds, cred)

	return nil
}

// UpdateCredential is called after every successful login to persist the
// authenticator's new signature counter - see PasskeyStore.UpdateCredential's
// doc comment for why this matters for cloned-authenticator detection. Same
// in-place-mutation note as AddCredential applies here.
func (s *SimplePasskeyStore) UpdateCredential(user PasskeyUser, cred webauthn.Credential) error {
	u, ok := user.(simplePasskeyUser)
	if !ok {
		return errors.New("passkey: SimplePasskeyStore was handed a PasskeyUser it didn't create")
	}

	u.rec.mu.Lock()
	defer u.rec.mu.Unlock()
	for i := range u.rec.creds {
		// bytes.Equal, not string(a) == string(b) - compares the two []byte
		// values directly (bytes.Equal is the standard, allocation-free way
		// to do this) instead of converting both sides to string first,
		// which allocates two throwaway strings just to compare them.
		if bytes.Equal(u.rec.creds[i].ID, cred.ID) {
			u.rec.creds[i] = cred
			return nil
		}
	}

	return errors.New("passkey: credential not found")
}

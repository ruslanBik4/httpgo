/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
package auth

import (
	"context"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
)

// --- test doubles -----------------------------------------------------
//
// A real PasskeyUser/PasskeyStore pair, kept deliberately simpler than
// SimplePasskeyStore (which has its own dedicated test file) - these exist
// only to drive passkey.go's own logic (BeginRegistration/BeginLogin/
// FinishRegistration/FinishLogin, sessionCache, webauthnUserAdapter)
// without dragging in Tokens/MapTokens semantics that aren't this file's
// concern.

type fakePasskeyUser struct {
	id          []byte
	login       string
	displayName string
	creds       []webauthn.Credential
	tokenData   TokenData
	newTokenErr error
	loginResp   any
}

func (u *fakePasskeyUser) PasskeyID() []byte                         { return u.id }
func (u *fakePasskeyUser) PasskeyLogin() string                      { return u.login }
func (u *fakePasskeyUser) PasskeyDisplayName() string                { return u.displayName }
func (u *fakePasskeyUser) PasskeyCredentials() []webauthn.Credential { return u.creds }
func (u *fakePasskeyUser) NewToken() (TokenData, error)              { return u.tokenData, u.newTokenErr }
func (u *fakePasskeyUser) LoginResponse() any                        { return u.loginResp }

type fakePasskeyStore struct {
	// *AuthBearer is what makes fakePasskeyStore satisfy PasskeyStore's
	// embedded FncAuth/TokenIssuer requirement (see PasskeyStore's own doc
	// comment) - needed only so this test double compiles as a PasskeyStore
	// at all now, nothing in this file exercises Auth/AdminAuth/NewToken
	// through it.
	*AuthBearer
	byLogin         map[string]*fakePasskeyUser
	findByLoginErr  error
	addCredErr      error
	updateCredErr   error
	addCredCalls    []webauthn.Credential
	updateCredCalls []webauthn.Credential
}

func newFakePasskeyStore() *fakePasskeyStore {
	return &fakePasskeyStore{AuthBearer: NewAuthBearer(nil), byLogin: map[string]*fakePasskeyUser{}}
}

func (s *fakePasskeyStore) FindByLogin(login string) (PasskeyUser, error) {
	if s.findByLoginErr != nil {
		return nil, s.findByLoginErr
	}
	u, ok := s.byLogin[login]
	if !ok {
		return nil, ErrPasskeyUnknownLogin
	}
	return u, nil
}

func (s *fakePasskeyStore) FindOrCreateByLogin(login string) (PasskeyUser, error) {
	u, ok := s.byLogin[login]
	if !ok {
		u = &fakePasskeyUser{id: []byte(login), login: login, displayName: login}
		s.byLogin[login] = u
	}
	return u, nil
}

func (s *fakePasskeyStore) FindByToken(token TokenData) (PasskeyUser, error) {
	return nil, errNotImplemented
}

func (s *fakePasskeyStore) AddCredential(user PasskeyUser, cred webauthn.Credential) error {
	s.addCredCalls = append(s.addCredCalls, cred)
	return s.addCredErr
}

func (s *fakePasskeyStore) UpdateCredential(user PasskeyUser, cred webauthn.Credential) error {
	s.updateCredCalls = append(s.updateCredCalls, cred)
	return s.updateCredErr
}

var errNotImplemented = ErrPasskeyNotAuthenticated // reuse an existing exported sentinel rather than declaring a throwaway one

func newTestWebAuthnPasskey(t *testing.T, store PasskeyStore) *WebAuthnPasskey {
	t.Helper()
	p, err := NewWebAuthnPasskey("example.com", "Example App", []string{"https://example.com"}, store, NewMapTokens(time.Hour))
	if err != nil {
		t.Fatalf("NewWebAuthnPasskey: %v", err)
	}
	return p
}

// --- sessionCache -------------------------------------------------------

func Test_sessionCache_put_take(t *testing.T) {
	c := newSessionCache()
	user := &fakePasskeyUser{login: "a@b.com"}
	data := webauthn.SessionData{Challenge: "chal"}

	id := c.put(user, data, PasskeySessionTTL)
	assert.NotEmpty(t, id)

	gotData, gotUser, ok := c.take(id)
	assert.True(t, ok)
	assert.Equal(t, data, gotData)
	assert.Equal(t, user, gotUser)

	// One-shot: the same id must not resolve a second time.
	_, _, ok = c.take(id)
	assert.False(t, ok, "sessionCache.take must consume the entry - a ceremony challenge must never be replayable")
}

func Test_sessionCache_take_unknown(t *testing.T) {
	c := newSessionCache()
	_, _, ok := c.take("does-not-exist")
	assert.False(t, ok)
}

func Test_sessionCache_take_expired(t *testing.T) {
	c := newSessionCache()
	// Whitebox: PasskeySessionTTL is a fixed 5-minute package constant, so
	// the only practical way to exercise expiry without a 5-minute sleep is
	// to insert an already-expired entry directly (same package, same
	// file's own unexported fields).
	c.m["expired"] = sessionEntry{
		data:    webauthn.SessionData{Challenge: "old"},
		user:    &fakePasskeyUser{login: "gone@example.com"},
		expires: time.Now().Add(-time.Second),
	}

	_, _, ok := c.take("expired")
	assert.False(t, ok, "an expired entry must not be returned even though it's still physically present in the map")

	// take deletes on lookup regardless of expiry, so a repeat take should
	// also report not-found (rather than, say, panicking on a missing key).
	_, _, ok = c.take("expired")
	assert.False(t, ok)
}

func Test_sessionCache_put_generates_distinct_ids(t *testing.T) {
	c := newSessionCache()
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		id := c.put(&fakePasskeyUser{}, webauthn.SessionData{}, PasskeySessionTTL)
		assert.False(t, seen[id], "sessionCache.put produced a duplicate id")
		seen[id] = true
	}
}

// --- webauthnUserAdapter --------------------------------------------------

func Test_webauthnUserAdapter(t *testing.T) {
	creds := []webauthn.Credential{{ID: []byte("cred-1")}}
	u := &fakePasskeyUser{
		id:          []byte("user-id"),
		login:       "someone@example.com",
		displayName: "Someone",
		creds:       creds,
	}
	a := webauthnUserAdapter{u}

	assert.Equal(t, []byte("user-id"), a.WebAuthnID())
	assert.Equal(t, "someone@example.com", a.WebAuthnName())
	assert.Equal(t, "Someone", a.WebAuthnDisplayName())
	assert.Equal(t, creds, a.WebAuthnCredentials())
}

// --- tokenWithSetter ------------------------------------------------------

func Test_tokenWithSetter(t *testing.T) {
	var data TokenData = &SimpleTokenData{Name: "x"}
	wt, ok := data.(tokenWithSetter)
	if assert.True(t, ok, "*SimpleTokenData must satisfy tokenWithSetter via its existing WithToken method - FinishLogin depends on this to attach a freshly-minted Bearer token onto whatever PasskeyUser.NewToken() returned") {
		wt.WithToken("minted-token")
		assert.Equal(t, "minted-token", data.(*SimpleTokenData).Token)
	}

	// A TokenData implementation with no WithToken (e.g. testTokenData,
	// from tokens_test.go) must NOT satisfy tokenWithSetter - a PasskeyUser
	// backed by one has no way to get the minted Bearer token attached this
	// way, and needs its own mechanism for LoginResponse() to include it.
	var plain TokenData = &testTokenData{id: 1}
	_, ok = plain.(tokenWithSetter)
	assert.False(t, ok)
}

// --- error vars -------------------------------------------------------

func TestPasskeyErrors_AreDistinct(t *testing.T) {
	errs := []error{ErrPasskeyNotAuthenticated, ErrPasskeySessionExpired, ErrPasskeyUnknownLogin}
	for i := range errs {
		for j := range errs {
			if i == j {
				continue
			}
			assert.NotEqual(t, errs[i], errs[j], "passkey error sentinels must be distinct - callers switch on these")
		}
	}
}

func TestPasskeyLoginParam(t *testing.T) {
	assert.Equal(t, "login", PasskeyLoginParam)
}

// --- NewWebAuthnPasskey ---------------------------------------------------

func TestNewWebAuthnPasskey_DefaultStoreSharesTokens(t *testing.T) {
	p, err := NewWebAuthnPasskey("example.com", "Example", []string{"https://example.com"}, nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	store, ok := p.PasskeyStore.(*SimplePasskeyStore)
	if !assert.True(t, ok, "store == nil must fall back to *SimplePasskeyStore") {
		return
	}

	// p itself no longer holds any Tokens/AuthBearer of its own -
	// Auth/AdminAuth/NewToken are all promoted from the embedded
	// PasskeyStore (see PasskeyStore's own doc comment), so there's only
	// one Tokens instance in play now rather than two that could silently
	// disagree. Pin that down by minting a token through p.NewToken (which
	// can only resolve to store's own *AuthBearer) and confirming it comes
	// back non-empty.
	token, err := p.NewToken(&SimpleTokenData{Name: "someone@example.com"})
	if assert.NoError(t, err) {
		assert.NotEmpty(t, token)
	}
	assert.Same(t, store, p.PasskeyStore, "PasskeyStore embed must be the exact fallback store NewToken was minted against")
}

func TestNewWebAuthnPasskey_ProvidedStoreIsUsedAsIs(t *testing.T) {
	store := newFakePasskeyStore()
	p, err := NewWebAuthnPasskey("example.com", "Example", []string{"https://example.com"}, store, nil)
	if assert.NoError(t, err) {
		assert.Same(t, store, p.PasskeyStore)
	}
}

func TestNewWebAuthnPasskey_InvalidConfig(t *testing.T) {
	// go-webauthn's webauthn.New validates its Config (RPID/RPDisplayName/
	// RPOrigins are all required) before returning - an empty config should
	// surface that as an error rather than a panic or a half-built
	// *WebAuthnPasskey. If go-webauthn's validation ever gets looser about
	// this in a future version, this test failing is the signal to notice.
	_, err := NewWebAuthnPasskey("", "", nil, newFakePasskeyStore(), nil)
	assert.Error(t, err)
}

func TestWebAuthnPasskey_String(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())
	got := p.String()

	assert.Contains(t, got, "register:")
	assert.Contains(t, got, "login:")
	assert.Contains(t, got, "<a href=")
}

// --- BeginRegistration ----------------------------------------------------

func TestWebAuthnPasskey_BeginRegistration_EmptyLogin(t *testing.T) {
	store := newFakePasskeyStore()
	p := newTestWebAuthnPasskey(t, store)

	challenge, err := p.BeginRegistration(context.Background(), "")
	assert.Nil(t, challenge)
	assert.Equal(t, ErrPasskeyUnknownLogin, err)
	assert.Empty(t, store.byLogin, "an empty login must be rejected before ever reaching the store")
}

func TestWebAuthnPasskey_BeginRegistration_Success(t *testing.T) {
	store := newFakePasskeyStore()
	p := newTestWebAuthnPasskey(t, store)

	challenge, err := p.BeginRegistration(context.Background(), "new-user@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, challenge.Body)

	// FindOrCreateByLogin must have actually been used - the account now
	// exists in the store even though it never existed before this call.
	assert.Contains(t, store.byLogin, "new-user@example.com")

	// The returned session id must be a real, live entry in the sessions
	// cache - the caller's HTTP layer is the one that turns it into a
	// cookie (see the generic setCookie helper in httpServer.go), passkey.go
	// only hands back the plain string, bundled with Body into
	// PasskeyChallenge rather than as a separate positional return value.
	if assert.NotEmpty(t, challenge.SessionID, "BeginRegistration must return a ceremony session id") {
		_, user, found := p.sessions.take(challenge.SessionID)
		if assert.True(t, found) {
			assert.Equal(t, "new-user@example.com", user.PasskeyLogin())
		}
	}
}

func TestWebAuthnPasskey_BeginRegistration_StoreError(t *testing.T) {
	store := newFakePasskeyStore()
	// FindOrCreateByLogin never errors on fakePasskeyStore as written, so
	// exercise the same error path FindByLogin already covers below via
	// findByLoginErr - BeginRegistration itself doesn't call FindByLogin,
	// so this documents that BeginRegistration and BeginLogin resolve the
	// account through two different PasskeyStore methods on purpose (see
	// FindOrCreateByLogin's own doc comment on why they can't share one).
	store.findByLoginErr = ErrPasskeyUnknownLogin
	p := newTestWebAuthnPasskey(t, store)

	challenge, err := p.BeginRegistration(context.Background(), "someone@example.com")
	assert.NoError(t, err, "BeginRegistration must NOT be affected by findByLoginErr - it never calls FindByLogin")
	assert.NotNil(t, challenge.Body)
	assert.NotEmpty(t, challenge.SessionID)
}

// --- BeginLogin -------------------------------------------------------

func TestWebAuthnPasskey_BeginLogin_EmptyLogin(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())

	challenge, err := p.BeginLogin(context.Background(), "")
	assert.Nil(t, challenge)
	assert.Equal(t, ErrPasskeyUnknownLogin, err)
}

func TestWebAuthnPasskey_BeginLogin_UnknownLogin(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())

	challenge, err := p.BeginLogin(context.Background(), "never-registered@example.com")
	assert.Nil(t, challenge)
	assert.Equal(t, ErrPasskeyUnknownLogin, err)
}

func TestWebAuthnPasskey_BeginLogin_Success(t *testing.T) {
	store := newFakePasskeyStore()
	store.byLogin["has-a-passkey@example.com"] = &fakePasskeyUser{
		id:    []byte("has-a-passkey@example.com"),
		login: "has-a-passkey@example.com",
		creds: []webauthn.Credential{{ID: []byte("cred-1")}},
	}
	p := newTestWebAuthnPasskey(t, store)

	challenge, err := p.BeginLogin(context.Background(), "has-a-passkey@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, challenge.Body)
	assert.NotEmpty(t, challenge.SessionID)
}

// --- FinishRegistration / FinishLogin: session + body validation --------
//
// A full happy-path FinishRegistration/FinishLogin needs a real WebAuthn
// attestation/assertion response byte-for-byte as an actual browser and
// authenticator would produce it (a signed CBOR attestation object) - not
// something this test can fabricate. What IS fully testable without that:
// the session-id handling (shared by both) and the request-body parsing
// failure path.

func TestWebAuthnPasskey_FinishRegistration_NoSessionID(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())

	body, err := p.FinishRegistration(context.Background(), "", nil)
	assert.Nil(t, body)
	assert.Equal(t, ErrPasskeySessionExpired, err)
}

func TestWebAuthnPasskey_FinishRegistration_UnknownSession(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())

	body, err := p.FinishRegistration(context.Background(), "no-such-session", nil)
	assert.Nil(t, body)
	assert.Equal(t, ErrPasskeySessionExpired, err)
}

func TestWebAuthnPasskey_FinishRegistration_BadBody(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())
	user := &fakePasskeyUser{login: "x@example.com"}
	id := p.sessions.put(user, webauthn.SessionData{Challenge: "chal"}, PasskeySessionTTL)

	body, err := p.FinishRegistration(context.Background(), id, []byte("not valid json"))
	assert.Nil(t, body)
	assert.Error(t, err, "an unparseable credential creation response must error, not panic")
}

func TestWebAuthnPasskey_FinishLogin_NoSessionID(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())

	body, err := p.FinishLogin(context.Background(), "", nil)
	assert.Nil(t, body)
	assert.Equal(t, ErrPasskeySessionExpired, err)
}

func TestWebAuthnPasskey_FinishLogin_UnknownSession(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())

	body, err := p.FinishLogin(context.Background(), "garbage", nil)
	assert.Nil(t, body)
	assert.Equal(t, ErrPasskeySessionExpired, err)
}

func TestWebAuthnPasskey_FinishLogin_BadBody(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())
	user := &fakePasskeyUser{login: "x@example.com"}
	id := p.sessions.put(user, webauthn.SessionData{Challenge: "chal"}, PasskeySessionTTL)

	body, err := p.FinishLogin(context.Background(), id, []byte("{not json"))
	assert.Nil(t, body)
	assert.Error(t, err)
}

func TestWebAuthnPasskey_FinishLogin_HappyPath(t *testing.T) {
	t.Skip("needs a real, signed WebAuthn credential-request response (CBOR attestation) from an actual authenticator/browser to get past protocol.ParseCredentialRequestResponseBytes + WebAuthn.ValidateLogin - not fabricable in a unit test. The token-minting/attachment tail this would otherwise cover is covered directly by Test_tokenWithSetter above instead; the nil-LoginResponse -> ErrPasskeyEmptyResponse branch has no equivalent standalone coverage since it only runs after a successful ValidateLogin")
}

// --- takeSession ------------------------------------------------------

func Test_takeSession_EmptyID(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())

	_, user, err := p.takeSession("")
	assert.Nil(t, user)
	assert.Equal(t, ErrPasskeySessionExpired, err)
}

func Test_takeSession_ConsumesEntry(t *testing.T) {
	p := newTestWebAuthnPasskey(t, newFakePasskeyStore())
	user := &fakePasskeyUser{login: "once@example.com"}
	id := p.sessions.put(user, webauthn.SessionData{Challenge: "chal"}, PasskeySessionTTL)

	_, gotUser, err := p.takeSession(id)
	assert.NoError(t, err)
	assert.Equal(t, user, gotUser)

	// One-shot, same as sessionCache.take itself.
	_, _, err = p.takeSession(id)
	assert.Equal(t, ErrPasskeySessionExpired, err)
}

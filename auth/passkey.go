/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/logs"
)

// Verified against github.com/go-webauthn/webauthn v0.18.1. If this project
// pins an older version, double-check webauthn.User (WebAuthnIcon() was
// required before v0.11.0) and the Config field names (RPOrigins is plural,
// RPIcon/top-level Timeout were removed in v0.11.0) before building.

// FncPasskey is the Passkey (WebAuthn) counterpart to FncAuth: FncAuth
// checks a request's Bearer token on every call; FncPasskey runs the two
// two-step ceremonies (registration, login) that *produce* one. Every
// method has the same func(ctx *fasthttp.RequestCtx) (any, error) shape as
// the generated form handlers in views/templates/forms, so it plugs
// directly into an apis.ApiRoute{Fnc: ...} the same way - see the wiring
// example at the bottom of this file.
type FncPasskey interface {
	// BeginRegistration starts adding a passkey to the CURRENTLY
	// authenticated user - the request must already carry a valid Bearer
	// token (this is "add a device" in account settings, not sign-up).
	BeginRegistration(ctx *fasthttp.RequestCtx) (any, error)
	FinishRegistration(ctx *fasthttp.RequestCtx) (any, error)
	// BeginLogin/FinishLogin run BEFORE authentication - the request has
	// no Bearer token yet, that's the whole point.
	BeginLogin(ctx *fasthttp.RequestCtx) (any, error)
	FinishLogin(ctx *fasthttp.RequestCtx) (any, error)
	String() string
}

// PasskeyUser is the minimal contract a caller's own user/account type must
// satisfy to take part in a WebAuthn ceremony. It deliberately mirrors
// webauthn.User rather than embedding it, so this file doesn't force that
// library's types into the app's user model - PasskeyStore hands one of
// these back on lookup, and webauthnUserAdapter (below) wraps it to satisfy
// webauthn.User internally.
type PasskeyUser interface {
	// PasskeyID is a stable, opaque identifier for this account - NOT the
	// login/email. WebAuthn stores this as the credential's "user handle";
	// per the spec it should not directly expose personally-identifying
	// data and must never change for a given account. A random id
	// generated once at account creation and stored alongside the login
	// works well; the database primary key also works if it's never reused.
	PasskeyID() []byte
	// PasskeyLogin is whatever the user typed into the login field -
	// BeginLogin() is called with this value.
	PasskeyLogin() string
	// PasskeyDisplayName is shown by the platform's Face ID/Touch ID/
	// security-key prompt (e.g. "Ruslan Bikchentaev").
	PasskeyDisplayName() string
	// PasskeyCredentials lists every passkey already registered for this
	// account, so the browser only offers a matching one and a login
	// attempt can be verified against the right public key.
	PasskeyCredentials() []webauthn.Credential
	// NewToken issues this app's own auth.TokenData for a successful
	// passkey ceremony - handed straight to Tokens.NewToken(), exactly as
	// a password login would produce.
	NewToken() (TokenData, error)
	// LoginResponse supplies the user-facing fields (name, lang, theme,
	// formActions, ...) that go back to the browser alongside the new
	// token. Kept separate from TokenData, which is this app's internal
	// session representation and not necessarily what a client should see.
	// user.js's saveUser() reads name/lang/theme/formActions from this
	// response the same way it does after a password login.
	LoginResponse() map[string]any
}

// PasskeyStore is the persistence boundary between this file and the
// application's own user database - nothing here talks to a database
// directly, dbEngine or otherwise.
type PasskeyStore interface {
	// FindByLogin resolves whatever the user typed into the login form to
	// a PasskeyUser. Called once, at login/begin, before the caller is
	// authenticated at all - return an error for an unknown login (don't
	// leak which logins exist by returning something different for
	// "found but no passkey" vs. "no such account").
	FindByLogin(login string) (PasskeyUser, error)
	// FindByToken resolves the CURRENTLY authenticated caller (as already
	// validated by AuthBearer.Auth) to a PasskeyUser. Called at
	// register/begin - adding a passkey never needs a login lookup.
	FindByToken(token TokenData) (PasskeyUser, error)
	// AddCredential persists a newly-registered credential against the
	// given account. Called once, at register/finish.
	AddCredential(user PasskeyUser, cred webauthn.Credential) error
	// UpdateCredential is called after every successful login to persist
	// the authenticator's new signature counter. WebAuthn relies on this
	// to detect a cloned authenticator: if a future login reports a
	// counter that didn't increase, ValidateLogin returns an error.
	UpdateCredential(user PasskeyUser, cred webauthn.Credential) error
}

var (
	ErrPasskeyNotAuthenticated = errors.New("passkey: registering a passkey requires an existing session")
	ErrPasskeySessionExpired   = errors.New("passkey: ceremony session expired, was already used, or the browser didn't send it back")
	ErrPasskeyUnknownLogin     = errors.New("passkey: no login supplied")
)

// webauthnUserAdapter satisfies webauthn.User for whatever PasskeyUser the
// app's PasskeyStore hands back.
type webauthnUserAdapter struct{ PasskeyUser }

func (u webauthnUserAdapter) WebAuthnID() []byte          { return u.PasskeyID() }
func (u webauthnUserAdapter) WebAuthnName() string        { return u.PasskeyLogin() }
func (u webauthnUserAdapter) WebAuthnDisplayName() string { return u.PasskeyDisplayName() }
func (u webauthnUserAdapter) WebAuthnCredentials() []webauthn.Credential {
	return u.PasskeyCredentials()
}

// passkeySessionCookie correlates a ceremony's "begin" call with its
// "finish" call. AuthBearer has no server-side session to piggyback on
// (auth here is a stateless Bearer token), so the WebAuthn challenge itself
// needs its own short-lived, single-use one - a plain HttpOnly cookie is
// the simplest way to round-trip an opaque id for that, and every fetch()
// call in passkey-auth.js is same-origin, which sends/receives cookies by
// default without any extra JS.
const passkeySessionCookie = "px_session"
const passkeySessionTTL = 5 * time.Minute

type sessionEntry struct {
	data    webauthn.SessionData
	user    PasskeyUser
	expires time.Time
}

// sessionCache is a minimal in-memory, single-instance store for in-flight
// ceremony challenges. That's fine for one server; behind a load balancer
// with multiple instances, swap this for something shared (Redis, the
// existing Tokens store's backing store, etc.) keyed the same way.
type sessionCache struct {
	mu sync.Mutex
	m  map[string]sessionEntry
}

func newSessionCache() *sessionCache {
	return &sessionCache{m: make(map[string]sessionEntry)}
}

func (c *sessionCache) put(user PasskeyUser, data webauthn.SessionData) string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	id := base64.RawURLEncoding.EncodeToString(buf)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[id] = sessionEntry{data: data, user: user, expires: time.Now().Add(passkeySessionTTL)}
	return id
}

// take is one-shot: a ceremony's challenge is consumed whether it succeeds
// or fails, so it can never be replayed.
func (c *sessionCache) take(id string) (webauthn.SessionData, PasskeyUser, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[id]
	delete(c.m, id)
	if !ok || time.Now().After(e.expires) {
		return webauthn.SessionData{}, nil, false
	}
	return e.data, e.user, true
}

// WebAuthnPasskey implements FncPasskey. Build one with NewWebAuthnPasskey
// and store it wherever the app keeps its long-lived singletons (it holds
// no per-request state).
type WebAuthnPasskey struct {
	webAuthn *webauthn.WebAuthn
	store    PasskeyStore
	tokens   Tokens
	sessions *sessionCache
}

// NewWebAuthnPasskey builds the passkey ceremony handler. rpID is the site's
// domain with no scheme/port (e.g. "example.com"); rpOrigins are the exact
// origins passkeys may be used from (e.g. "https://example.com") - a
// mismatch here is the most common source of a confusing "origin not
// allowed" failure. tokens may be nil to reuse the same NewMapTokens(...)
// default AuthBearer falls back to.
func NewWebAuthnPasskey(rpID, rpDisplayName string, rpOrigins []string, store PasskeyStore, tokens Tokens) (*WebAuthnPasskey, error) {
	w, err := webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: rpDisplayName,
		RPOrigins:     rpOrigins,
	})
	if err != nil {
		return nil, err
	}

	if tokens == nil {
		tokens = NewMapTokens(tokenExpires)
	}

	return &WebAuthnPasskey{
		webAuthn: w,
		store:    store,
		tokens:   tokens,
		sessions: newSessionCache(),
	}, nil
}

func (p *WebAuthnPasskey) String() string {
	return `implement passkey (WebAuthn) auth:
	 register: ` + getStringOfFnc(reflect.ValueOf(p.BeginRegistration).Pointer()) + `
	 login: ` + getStringOfFnc(reflect.ValueOf(p.BeginLogin).Pointer())
}

func (p *WebAuthnPasskey) setSessionCookie(ctx *fasthttp.RequestCtx, id string) {
	c := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(c)
	c.SetKey(passkeySessionCookie)
	c.SetValue(id)
	c.SetPath("/webauthn/")
	c.SetHTTPOnly(true)
	c.SetSecure(true)
	c.SetSameSite(fasthttp.CookieSameSiteStrictMode)
	c.SetMaxAge(int(passkeySessionTTL.Seconds()))
	ctx.Response.Header.SetCookie(c)
}

func (p *WebAuthnPasskey) takeSession(ctx *fasthttp.RequestCtx) (webauthn.SessionData, PasskeyUser, error) {
	id := string(ctx.Request.Header.Cookie(passkeySessionCookie))
	if id == "" {
		return webauthn.SessionData{}, nil, ErrPasskeySessionExpired
	}

	session, user, ok := p.sessions.take(id)
	if !ok {
		return webauthn.SessionData{}, nil, ErrPasskeySessionExpired
	}

	return session, user, nil
}

// BeginRegistration -> POST /webauthn/register/begin (Bearer-authenticated)
func (p *WebAuthnPasskey) BeginRegistration(ctx *fasthttp.RequestCtx) (any, error) {
	token, ok := ctx.UserValue(UserValueToken).(TokenData)
	if !ok {
		return nil, ErrPasskeyNotAuthenticated
	}

	user, err := p.store.FindByToken(token)
	if err != nil {
		return nil, err
	}

	creation, session, err := p.webAuthn.BeginRegistration(webauthnUserAdapter{user})
	if err != nil {
		return nil, err
	}

	p.setSessionCookie(ctx, p.sessions.put(user, *session))

	// creation marshals as {"publicKey": {...CredentialCreationOptions}} -
	// exactly what navigator.credentials.create({publicKey}) needs on the
	// browser side once passkey-auth.js unwraps the "publicKey" key.
	return creation, nil
}

// FinishRegistration -> POST /webauthn/register/finish (Bearer-authenticated)
func (p *WebAuthnPasskey) FinishRegistration(ctx *fasthttp.RequestCtx) (any, error) {
	session, user, err := p.takeSession(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := protocol.ParseCredentialCreationResponseBytes(ctx.PostBody())
	if err != nil {
		return nil, err
	}

	cred, err := p.webAuthn.CreateCredential(webauthnUserAdapter{user}, session, parsed)
	if err != nil {
		return nil, err
	}

	if err := p.store.AddCredential(user, *cred); err != nil {
		return nil, err
	}

	return map[string]any{"message": "Passkey added."}, nil
}

// BeginLogin -> POST /webauthn/login/begin (no auth yet - {"login": "..."})
func (p *WebAuthnPasskey) BeginLogin(ctx *fasthttp.RequestCtx) (any, error) {
	var body struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil || body.Login == "" {
		return nil, ErrPasskeyUnknownLogin
	}

	user, err := p.store.FindByLogin(body.Login)
	if err != nil {
		return nil, err
	}

	assertion, session, err := p.webAuthn.BeginLogin(webauthnUserAdapter{user})
	if err != nil {
		return nil, err
	}

	p.setSessionCookie(ctx, p.sessions.put(user, *session))

	return assertion, nil
}

// FinishLogin -> POST /webauthn/login/finish (no auth yet)
func (p *WebAuthnPasskey) FinishLogin(ctx *fasthttp.RequestCtx) (any, error) {
	session, user, err := p.takeSession(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := protocol.ParseCredentialRequestResponseBytes(ctx.PostBody())
	if err != nil {
		return nil, err
	}

	cred, err := p.webAuthn.ValidateLogin(webauthnUserAdapter{user}, session, parsed)
	if err != nil {
		return nil, err
	}

	// Persist the authenticator's new signature counter - required by the
	// spec so a cloned authenticator can be detected on its next use.
	// Non-fatal: the login itself already succeeded, don't fail the
	// request over bookkeeping that only matters for the *next* login.
	if err := p.store.UpdateCredential(user, *cred); err != nil {
		logs.ErrorLog(err)
	}

	tokenData, err := user.NewToken()
	if err != nil {
		return nil, err
	}

	token, err := p.tokens.NewToken(tokenData)
	if err != nil {
		return nil, err
	}

	// DelClientCookie (not DelCookie) - this tells the *browser* to drop
	// the cookie it already has; DelCookie only removes it from this
	// response's own headers before sending, which wouldn't touch what
	// the browser is already holding.
	ctx.Response.Header.DelClientCookie(passkeySessionCookie)

	response := user.LoginResponse()
	if response == nil {
		response = map[string]any{}
	}
	// Same field user.js's saveUser() already reads (token/access_token/
	// bearer_token/auth_token) from a password login, so
	// afterLogin(userData) in passkey-auth.js can hand this straight to
	// saveUser() unchanged.
	response["token"] = token

	return response, nil
}

// --- Wiring example (goes in the app's own routes package, NOT here - auth
// stays free of any dependency on apis/views so it can't cause an import
// cycle) ---
//
// passkeys, err := auth.NewWebAuthnPasskey("example.com", "Example App",
//     []string{"https://example.com"}, myPasskeyStore, nil)
// ...
// var BeginRegistrationRoute = &apis.ApiRoute{
//     Desc:          "# begin passkey registration for the current user",
//     Fnc:           passkeys.BeginRegistration,
//     Method:        apis.POST,
//     IsAJAXRequest: true,
// }
// var FinishRegistrationRoute = &apis.ApiRoute{Desc: "# finish passkey registration", Fnc: passkeys.FinishRegistration, Method: apis.POST, IsAJAXRequest: true}
// var BeginLoginRoute = &apis.ApiRoute{Desc: "# begin passkey login", Fnc: passkeys.BeginLogin, Method: apis.POST, IsAJAXRequest: true}
// var FinishLoginRoute = &apis.ApiRoute{Desc: "# finish passkey login", Fnc: passkeys.FinishLogin, Method: apis.POST, IsAJAXRequest: true}
//
// registered at /webauthn/register/begin, /webauthn/register/finish,
// /webauthn/login/begin, /webauthn/login/finish respectively - those are
// exactly the paths passkey-auth.js already calls.

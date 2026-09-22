/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ruslanBik4/logs"
)

// Verified against github.com/go-webauthn/webauthn v0.18.1. If this project
// pins an older version, double-check webauthn.User (WebAuthnIcon() was
// required before v0.11.0) and the Config field names (RPOrigins is plural,
// RPIcon/top-level Timeout were removed in v0.11.0) before building.

// FncPasskey is the Passkey (WebAuthn) counterpart to FncAuth: FncAuth
// checks a request's Bearer token on every call; FncPasskey runs the two
// two-step ceremonies (registration, login) that *produce* one.
//
// This file has NO dependency on any HTTP framework, and no HTTP-shaped
// data at all - no request, no response, no cookie type. Every piece of
// information a method needs travels as a plain, explicitly-typed
// parameter (login string, sessionID string, body []byte) or comes back as
// one (PasskeyChallenge's Body/SessionID fields, or a bare any). ctx
// context.Context is the first parameter on every method purely for the
// standard reason any Go API takes one (cancellation/deadlines/tracing) -
// this file never calls ctx.Value on it, so a caller is free to pass
// whatever context.Context it already has to hand, including its own
// request type if that happens to satisfy the interface (e.g. fasthttp's
// *fasthttp.RequestCtx implements Deadline/Done/Err/Value, so httpServer.go
// just passes its real ctx straight through below - that's the caller's
// choice to make, not something this file relies on). Cookie handling
// (reading the incoming cookie, writing/clearing the outgoing one, choosing
// Path/HttpOnly/Secure/SameSite) is entirely the caller's job - see the
// wiring note at the bottom of this file for httpServer.go's own
// implementation. This is deliberate: this package can be driven from
// net/http, fasthttp, or anything else with zero changes here, and
// swapping the HTTP framework out from under it later never touches this
// file.
type FncPasskey interface {
	// BeginRegistration starts adding a passkey to the CURRENTLY
	// authenticated user - the request must already carry a valid Bearer
	// token (this is "add a device" in account settings, not sign-up).
	// Returns the ceremony creation options plus the new ceremony-session
	// id the caller must store in a cookie for the matching
	// FinishRegistration call to read back - see PasskeyChallenge.
	BeginRegistration(ctx context.Context, login string) (*PasskeyChallenge, error)
	// FinishRegistration takes the ceremony-session id the caller read off
	// its own session cookie (empty/unknown/expired all report
	// ErrPasskeySessionExpired) and the raw credential-creation response
	// body the browser posted.
	FinishRegistration(ctx context.Context, sessionID string, body []byte) (any, error)
	// BeginLogin/FinishLogin run BEFORE authentication - the request has
	// no Bearer token yet, that's the whole point.
	BeginLogin(ctx context.Context, login string) (*PasskeyChallenge, error)
	FinishLogin(ctx context.Context, sessionID string, body []byte) (any, error)
	String() string
}

// PasskeyChallenge is what BeginRegistration/BeginLogin return: the
// ceremony creation/assertion options to send to the browser (Body) and the
// new ceremony-session id (SessionID) the caller must correlate with the
// matching Finish call - typically by putting it in a cookie, see
// httpServer.go's setCookie call (see the wiring note at the bottom of this
// file). A named struct instead of two
// positional return values, so a caller can't accidentally transpose them;
// it carries no HTTP semantics of its own (no headers, no cookie
// attributes) - it's just the two plain values a Begin* call produces.
// Returned as a pointer (nil on error) rather than a value, so an error path
// has an unambiguous zero value to return instead of a struct that merely
// looks empty.
type PasskeyChallenge struct {
	Body      any
	SessionID string
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
	// passkey ceremony - handed straight to (*WebAuthnPasskey).NewToken()
	// (promoted from its embedded PasskeyStore, which every implementation
	// must also satisfy as a TokenIssuer - see PasskeyStore's own doc
	// comment), exactly as a password login would produce. Implementations
	// that also want LoginResponse to
	// come back with the token already attached (see LoginResponse's own
	// doc comment) should return the SAME object LoginResponse will later
	// read, e.g. simplePasskeyUser returns u.rec.data from both.
	NewToken() (TokenData, error)
	// LoginResponse supplies whatever goes back to the browser alongside a
	// successful login - the whole HTTP response body, not just extra
	// fields merged into one. Returns any (not map[string]any) so a
	// PasskeyUser can hand back its own TokenData struct directly (as
	// simplePasskeyUser does: user.js's saveUser() already reads
	// name/lang/token/... straight off a plain SimpleTokenData-shaped JSON
	// object, whether it came from a password login or here) instead of
	// FinishLogin re-assembling a map by hand. See FinishLogin's own doc
	// comment for how the freshly-minted Bearer token gets into whatever
	// this returns.
	LoginResponse() any
}

// PasskeyStore is the persistence boundary between this file and the
// application's own user database - nothing here talks to a database
// directly, dbEngine or otherwise. Also the sole source of the app-wide
// Bearer-auth capability for a *WebAuthnPasskey: it embeds FncAuth
// (Auth/AdminAuth/String) and TokenIssuer (NewToken) instead of
// WebAuthnPasskey keeping its own separate *AuthBearer, since a real
// project's PasskeyStore is already the thing that knows how a passkey
// login turns into a session - a plain in-memory store like
// SimplePasskeyStore or main.go's memPasskeyStore satisfies this cheaply by
// embedding *AuthBearer itself (see either's own doc comment), but nothing
// here requires that specific shape; any FncAuth/TokenIssuer implementation
// works.
type PasskeyStore interface {
	FncAuth
	TokenIssuer
	// FindByLogin resolves whatever the user typed into the login form to
	// a PasskeyUser. Called once, at login/begin, before the caller is
	// authenticated at all - return an error for an unknown login (don't
	// leak which logins exist by returning something different for
	// "found but no passkey" vs. "no such account"). MUST fail for a login
	// that has never registered a passkey - unlike FindOrCreateByLogin
	// below, this never creates an account.
	FindByLogin(login string) (PasskeyUser, error)
	// FindOrCreateByLogin resolves login for a REGISTRATION ceremony
	// (register/begin), creating a new account if none exists yet for that
	// login. This is what lets a first-ever passkey registration for a
	// brand new login double as sign-up - register/begin is no longer
	// gated by an existing Bearer token (see BeginRegistration's own doc
	// comment for why), so a login string is the only identity it has to
	// work with, and it has to work for a login nobody has seen before.
	FindOrCreateByLogin(login string) (PasskeyUser, error)
	// FindByToken resolves an already-authenticated caller (as validated
	// by AuthBearer.Auth) to a PasskeyUser. No longer called by
	// BeginRegistration (see its doc comment) - kept in the interface for
	// any account-management endpoint a project adds later (e.g. "list my
	// devices" while logged in).
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
	// ErrPasskeyNotAuthenticated is no longer returned by BeginRegistration
	// (see its doc comment - register/begin identifies the account by a
	// login param now, not an existing Bearer token) - left exported in
	// case a project adds its own separate "add a device while already
	// logged in" endpoint that still wants this exact error.
	ErrPasskeyNotAuthenticated = errors.New("passkey: registering a passkey requires an existing session")
	ErrPasskeySessionExpired   = errors.New("passkey: ceremony session expired, was already used, or the browser didn't send it back")
	ErrPasskeyUnknownLogin     = errors.New("passkey: no login supplied")
	// ErrPasskeyEmptyResponse is returned by FinishLogin when a PasskeyUser's
	// LoginResponse() returns nil. The login itself already succeeded by
	// that point (the credential validated, the Bearer token was minted) -
	// but the browser still needs a response body to complete the flow, and
	// FinishLogin doesn't build one of its own: LoginResponse() is expected
	// to always return something usable, so nil is treated as a PasskeyUser
	// implementation bug, not silently papered over.
	ErrPasskeyEmptyResponse = errors.New("passkey: LoginResponse returned nil")
)

// PasskeyLoginParam is the request field name both /webauthn/register/begin
// and /webauthn/login/begin declare as a route param (crud.ParamsLogin in
// httpServer.go's createAdminRoutes - "login" there too) and that
// httpServer.go extracts with apis.GetValue[string](ctx, &crud.ParamsLogin)
// before calling BeginRegistration/BeginLogin with it. Defined here, in
// auth, rather than referencing crud.ParamsLogin.Name directly: auth can't
// import apis/crud without an import cycle (crud -> apis -> auth), so
// crud.ParamsLogin.Name is set FROM this constant instead, keeping the one
// wire name in sync in both directions.
const PasskeyLoginParam = "login"

// webauthnUserAdapter satisfies webauthn.User for whatever PasskeyUser the
// app's PasskeyStore hands back.
type webauthnUserAdapter struct{ PasskeyUser }

func (u webauthnUserAdapter) WebAuthnID() []byte          { return u.PasskeyID() }
func (u webauthnUserAdapter) WebAuthnName() string        { return u.PasskeyLogin() }
func (u webauthnUserAdapter) WebAuthnDisplayName() string { return u.PasskeyDisplayName() }
func (u webauthnUserAdapter) WebAuthnCredentials() []webauthn.Credential {
	return u.PasskeyCredentials()
}

// PasskeySessionCookie is the name of the cookie that correlates a
// ceremony's "begin" call with its "finish" call. AuthBearer has no
// server-side session to piggyback on (auth here is a stateless Bearer
// token), so the WebAuthn challenge itself needs its own short-lived,
// single-use one. passkey.go never reads or writes this cookie itself -
// it only hands back the session id BeginRegistration/BeginLogin generate
// as a plain string, and only accepts one back as a plain string parameter
// to FinishRegistration/FinishLogin. The caller's HTTP layer (httpServer.go)
// is the one place that actually knows this is a cookie at all: it reads
// this exact name off the incoming request and writes it on the way out -
// see the wiring note at the bottom of this file. Exported so that HTTP
// layer can reference the one true cookie name instead of hardcoding its
// own copy of it.
const PasskeySessionCookie = "px_session"

// PasskeySessionTTL is how long a ceremony session id stays valid after
// BeginRegistration/BeginLogin issues it - also exported so the caller's
// HTTP layer can set the same lifetime on the actual cookie (e.g. its
// Max-Age) instead of duplicating this number and risking the two drifting
// apart.
const PasskeySessionTTL = 5 * time.Minute

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

// put stores data/user under a freshly-generated id, valid for ttl from now.
// ttl is an explicit parameter rather than sessionCache reaching for the
// PasskeySessionTTL package constant itself, so this type carries no
// hardcoded opinion of its own about how long a ceremony session should
// live - every current caller (BeginRegistration/BeginLogin, below) passes
// PasskeySessionTTL explicitly, which remains the one default value for the
// whole package, but a caller wanting a different lifetime for some other
// ceremony is free to pass its own.
func (c *sessionCache) put(user PasskeyUser, data webauthn.SessionData, ttl time.Duration) string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	id := base64.RawURLEncoding.EncodeToString(buf)

	// id is already computed above, before the lock is even taken - the
	// critical section below is the single map write, nothing else needs
	// protecting, so it's locked and unlocked explicitly around just that
	// line rather than deferring the unlock to function exit.
	c.mu.Lock()
	c.m[id] = sessionEntry{data: data, user: user, expires: time.Now().Add(ttl)}
	c.mu.Unlock()

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
// no per-request state). Composed of two independent capabilities, each
// promoted directly (no *p.field.Method() indirection needed for any of
// them, only for the ones this type also overrides itself):
//   - *webauthn.WebAuthn: the WebAuthn ceremony math itself
//     (BeginRegistration/CreateCredential/BeginLogin/ValidateLogin).
//   - PasskeyStore: the app's own identity/credential persistence, which
//     also supplies Bearer session handling (Auth/AdminAuth/String/NewToken)
//     by embedding FncAuth/TokenIssuer itself (see PasskeyStore's own doc
//     comment) - this is what makes *WebAuthnPasskey satisfy auth.FncAuth,
//     so it can be handed to apis.NewApis(...) as the app-wide auth manager
//     exactly like a plain *AuthBearer could.
//
// There's deliberately no separate *AuthBearer field here anymore: that
// would be a second, independent Bearer-session capability living
// alongside whatever the store already provides, with no guarantee the two
// agree on which Tokens instance backs them - PasskeyStore is the one
// source of truth for both persistence and auth on this type.
type WebAuthnPasskey struct {
	*webauthn.WebAuthn
	PasskeyStore
	sessions *sessionCache
}

// NewWebAuthnPasskey builds the passkey ceremony handler. rpID is the site's
// domain with no scheme/port (e.g. "example.com"); rpOrigins are the exact
// origins passkeys may be used from (e.g. "https://example.com") - a
// mismatch here is the most common source of a confusing "origin not
// allowed" failure. tokens is only used to build the fallback store (see
// below) - it's ignored entirely when store is provided, since at that
// point which Tokens instance backs auth is entirely the store's own
// business.
func NewWebAuthnPasskey(rpID, rpDisplayName string, rpOrigins []string, store PasskeyStore, tokens Tokens) (*WebAuthnPasskey, error) {
	w, err := webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: rpDisplayName,
		RPOrigins:     rpOrigins,
	})
	if err != nil {
		return nil, err
	}

	if store == nil {
		// NewSimplePasskeyStore accepts a nil tokens itself (it falls back
		// to NewAuthBearer's own NewMapTokens(tokenExpires) default) - no
		// need to pre-resolve that default here first.
		store = NewSimplePasskeyStore(tokens)
	}

	return &WebAuthnPasskey{
		WebAuthn:     w,
		PasskeyStore: store,
		sessions:     newSessionCache(),
	}, nil
}

func (p *WebAuthnPasskey) String() string {
	return `implement passkey (WebAuthn) auth:
	 register: ` + getStringOfFnc(reflect.ValueOf(p.BeginRegistration).Pointer()) + `
	 login: ` + getStringOfFnc(reflect.ValueOf(p.BeginLogin).Pointer())
}

// takeSession recovers the in-flight ceremony challenge a "begin" call
// created, keyed by the ceremony-session id the caller read off its own
// session cookie and passed in as a plain string. One shot, same as
// sessionCache.take itself: an empty id and an unknown/expired/already-used
// one are indistinguishable to the caller, both surface as
// ErrPasskeySessionExpired.
func (p *WebAuthnPasskey) takeSession(sessionID string) (webauthn.SessionData, PasskeyUser, error) {
	if sessionID == "" {
		return webauthn.SessionData{}, nil, ErrPasskeySessionExpired
	}

	session, user, ok := p.sessions.take(sessionID)
	if !ok {
		return webauthn.SessionData{}, nil, ErrPasskeySessionExpired
	}

	return session, user, nil
}

// BeginRegistration -> POST /webauthn/register/begin. Identifies the
// account purely by a "login" value (PasskeyLoginParam) in the request -
// deliberately NOT gated by NeedAuth/an existing Bearer token anymore.
//
// Why this changed: the route used to declare NeedAuth: true, so
// apis.ApiRoute.CheckAndRun would call FncAuth.Auth(ctx) BEFORE this
// handler ever ran and only set ctx.UserValue(UserValueToken) if THAT
// succeeded - which requires the caller to already hold a valid Bearer
// token. But a brand new account's very first passkey registration has no
// prior session to hold a token from, so that check could never pass for
// exactly the case this route exists to handle. login/begin (BeginLogin,
// below) already identifies its caller the same way this now does - by a
// login value in the request, not a token - so this brings register/begin
// in line with it. See PasskeyStore.FindOrCreateByLogin for how a
// never-seen-before login now doubles as sign-up; NeedAuth was removed
// from both /webauthn/register/begin and /webauthn/register/finish in
// httpServer.go's createAdminRoutes to match (FinishRegistration never
// read UserValueToken either - it identifies its caller via the ceremony's
// session id, see takeSession).
func (p *WebAuthnPasskey) BeginRegistration(ctx context.Context, login string) (*PasskeyChallenge, error) {
	if login == "" {
		return nil, ErrPasskeyUnknownLogin
	}

	user, err := p.FindOrCreateByLogin(login)
	if err != nil {
		return nil, err
	}

	creation, session, err := p.WebAuthn.BeginRegistration(webauthnUserAdapter{user})
	if err != nil {
		return nil, err
	}

	// creation marshals as {"publicKey": {...CredentialCreationOptions}} -
	// exactly what navigator.credentials.create({publicKey}) needs on the
	// browser side once passkey-auth.js unwraps the "publicKey" key.
	// SessionID is the new ceremony-session id - the caller's HTTP layer is
	// responsible for putting it in a cookie on the response (see the
	// wiring note at the bottom of this file).
	return &PasskeyChallenge{
		Body:      creation,
		SessionID: p.sessions.put(user, *session, PasskeySessionTTL),
	}, nil
}

// FinishRegistration -> POST /webauthn/register/finish. sessionID is
// whatever the caller read back out of its own PasskeySessionCookie cookie
// (empty if there was none); body is the raw credential-creation response
// the browser posted.
func (p *WebAuthnPasskey) FinishRegistration(ctx context.Context, sessionID string, body []byte) (any, error) {
	session, user, err := p.takeSession(sessionID)
	if err != nil {
		return nil, err
	}

	parsed, err := protocol.ParseCredentialCreationResponseBytes(body)
	if err != nil {
		return nil, err
	}

	cred, err := p.CreateCredential(webauthnUserAdapter{user}, session, parsed)
	if err != nil {
		return nil, err
	}

	if err := p.AddCredential(user, *cred); err != nil {
		return nil, err
	}

	return struct {
		Message string `json:"message"`
	}{Message: "Passkey added."}, nil
}

// BeginLogin -> POST /webauthn/login/begin (no auth yet - {"login": "..."}).
// Returns the assertion options and a new ceremony-session id, same shape
// as BeginRegistration.
func (p *WebAuthnPasskey) BeginLogin(ctx context.Context, login string) (*PasskeyChallenge, error) {
	if login == "" {
		return nil, ErrPasskeyUnknownLogin
	}

	user, err := p.FindByLogin(login)
	if err != nil {
		return nil, err
	}

	assertion, session, err := p.WebAuthn.BeginLogin(webauthnUserAdapter{user})
	if err != nil {
		return nil, err
	}

	return &PasskeyChallenge{
		Body:      assertion,
		SessionID: p.sessions.put(user, *session, PasskeySessionTTL),
	}, nil
}

// tokenWithSetter is satisfied by *SimpleTokenData, via its existing
// WithToken method. FinishLogin uses it, below, to hand the freshly-minted
// Bearer token string back to whatever concrete TokenData user.NewToken()
// returned - so that when user.LoginResponse() is called right after, the
// object it hands back (for simplePasskeyUser, the very same *SimpleTokenData)
// already carries the token. FinishLogin always returns whatever
// LoginResponse() gives back, as-is - a PasskeyUser backed by a different
// TokenData type that doesn't implement tokenWithSetter needs its own way to
// get the token into whatever LoginResponse() returns (see
// ErrPasskeyEmptyResponse if it returns nil instead).
type tokenWithSetter interface {
	WithToken(token string) *SimpleTokenData
}

// FinishLogin -> POST /webauthn/login/finish (no auth yet). sessionID/body
// have the same meaning as in FinishRegistration. The caller should clear
// its session cookie once this returns, regardless of outcome - the
// ceremony session is one-shot either way (sessionCache.take already
// deleted it the moment takeSession looked it up).
func (p *WebAuthnPasskey) FinishLogin(ctx context.Context, sessionID string, body []byte) (any, error) {
	session, user, err := p.takeSession(sessionID)
	if err != nil {
		return nil, err
	}

	parsed, err := protocol.ParseCredentialRequestResponseBytes(body)
	if err != nil {
		return nil, err
	}

	cred, err := p.ValidateLogin(webauthnUserAdapter{user}, session, parsed)
	if err != nil {
		return nil, err
	}

	// Persist the authenticator's new signature counter - required by the
	// spec so a cloned authenticator can be detected on its next use.
	// Non-fatal: the login itself already succeeded, don't fail the
	// request over bookkeeping that only matters for the *next* login.
	if err := p.UpdateCredential(user, *cred); err != nil {
		logs.ErrorLog(err)
	}

	tokenData, err := user.NewToken()
	if err != nil {
		return nil, err
	}

	// p.NewToken is promoted straight from the embedded PasskeyStore -
	// PasskeyStore itself embeds TokenIssuer (see its own doc comment), so
	// minting the Bearer session token and looking up/creating the account
	// both go through the one store, with no separate *AuthBearer needed on
	// WebAuthnPasskey itself.
	token, err := p.NewToken(tokenData)
	if err != nil {
		return nil, err
	}

	// Attach the token to tokenData itself, if it supports that, so
	// user.LoginResponse() (called next) already includes it - see
	// tokenWithSetter's own doc comment.
	if wt, ok := tokenData.(tokenWithSetter); ok {
		wt.WithToken(token)
	}

	// Returned exactly as user.LoginResponse() gives it back - FinishLogin
	// doesn't assemble or fall back to a response of its own. A nil
	// LoginResponse() is treated as a PasskeyUser implementation bug (see
	// ErrPasskeyEmptyResponse's own doc comment), not something to paper
	// over silently.
	response := user.LoginResponse()
	if response == nil {
		return nil, ErrPasskeyEmptyResponse
	}

	return response, nil
}

// --- Wiring (lives in httpServer.go's createAdminRoutes, NOT here - auth
// stays free of any dependency on apis/views, or on any HTTP framework at
// all, so it can't cause an import cycle and can't be broken by swapping
// fasthttp out later) ---
//
// Every byte of cookie handling now lives in httpServer.go: reading the
// incoming ceremony cookie off the request, choosing the outgoing cookie's
// Path/HttpOnly/Secure/SameSite/lifetime, and writing or clearing it on the
// response. passkey.go itself only ever sees/returns a plain session id
// string. handleWebAuthnRegisterBegin/handleWebAuthnRegisterFinish/
// handleWebAuthnLoginBegin/handleWebAuthnLoginFinish each resolve the
// app-wide auth manager (auth.GetAuthManager(ctx)), assert it to
// auth.FncPasskey, and do that reading/writing with three small, generic
// (not passkey-specific - any cookie a handler needs can reuse these the
// same way) helpers:
//
//     // cookieValue reads the named cookie's raw value off the incoming
//     // request.
//     func cookieValue(ctx *fasthttp.RequestCtx, name string) string {
//         return string(ctx.Request.Header.Cookie(name))
//     }
//
//     // setCookie writes name/value onto the response, scoped to path and
//     // valid for maxAge - every actual cookie attribute this project cares
//     // about is a parameter here, not passkey.go's concern.
//     func setCookie(ctx *fasthttp.RequestCtx, name, value, path string, maxAge time.Duration) {
//         c := fasthttp.AcquireCookie()
//         defer fasthttp.ReleaseCookie(c)
//         c.SetKey(name)
//         c.SetValue(value)
//         c.SetPath(path)
//         c.SetHTTPOnly(true)
//         c.SetSecure(true)
//         c.SetSameSite(fasthttp.CookieSameSiteStrictMode)
//         c.SetMaxAge(int(maxAge.Seconds()))
//         ctx.Response.Header.SetCookie(c)
//     }
//
//     // clearCookie tells the browser to drop the named cookie - called
//     // before every Finish* call, since the ceremony session is one-shot
//     // either way.
//     func clearCookie(ctx *fasthttp.RequestCtx, name string) {
//         ctx.Response.Header.DelClientCookie(name)
//     }
//
//     func handleWebAuthnRegisterBegin(ctx *fasthttp.RequestCtx) (any, error) {
//         p, err := webAuthnManager(ctx)
//         if err != nil {
//             return nil, err
//         }
//         login := GetValue[string](ctx, &crud.ParamsLogin)
//         challenge, err := p.BeginRegistration(ctx, login)
//         if err != nil {
//             return nil, err
//         }
//         setCookie(ctx, auth.PasskeySessionCookie, challenge.SessionID, "/webauthn/", auth.PasskeySessionTTL)
//         return challenge.Body, nil
//     }
//
// Note ctx is passed straight through as the context.Context parameter,
// not context.Background() - *fasthttp.RequestCtx already implements
// context.Context (Deadline/Done/Err/Value), so httpServer.go's own real
// request context is exactly what these methods receive; this file simply
// never calls anything on it besides what context.Context itself defines.
// Also note challenge above is a *PasskeyChallenge (nil on error, never a
// zero-value struct) - the `if err != nil { return nil, err }` guard right
// before challenge.SessionID/challenge.Body is what makes dereferencing it
// safe, exactly as with any other pointer-returning, error-returning Go call.
// (Finish* handlers call clearCookie(ctx, auth.PasskeySessionCookie) then
// pass ctx, cookieValue(ctx, auth.PasskeySessionCookie), and ctx.PostBody()
// straight through: p.FinishRegistration(ctx, sessionID, ctx.PostBody()).)
// registered at /webauthn/register/begin, /webauthn/register/finish,
// /webauthn/login/begin, /webauthn/login/finish respectively (no NeedAuth
// on either Begin route - see BeginRegistration's own doc comment for why)
// - those are exactly the paths passkey-auth.js already calls.

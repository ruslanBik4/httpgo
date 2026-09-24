/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pkg/errors"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/logs"
)

type Tokens interface {
	NewToken(userData TokenData) (string, error)
	GetToken(s string) TokenData
	RemoveToken(s string) error
	// SetToken stores userData under an EXPLICIT key s, unlike NewToken
	// (which generates its own random token string). MapTokens already
	// implemented this before it was added here - SimplePasskeyStore
	// (simple_passkey_store.go) is what needs it exposed on the interface,
	// to key a record by login instead of by a random session token.
	//
	// CAVEAT this creates for that use: MapTokens.SetToken always arms the
	// SAME m.expiresIn deletion timer NewToken's session tokens get. A
	// passkey record stored this way (keyed by login, not by a fresh random
	// token) inherits that timer too - it silently disappears after
	// m.expiresIn, exactly like an ordinary session would, even though a
	// registered passkey should keep working indefinitely. See
	// SimplePasskeyStore's doc comment for the full explanation and what to
	// do about it before relying on this in anything long-lived.
	SetToken(s string, userData TokenData)
}

type TokenData interface {
	IsAdmin() bool
	GetUserID() int
}

type SimpleTokenData struct {
	Name       string         `json:"name"`
	Desc       string         `json:"desc"`
	Lang       string         `json:"lang"`
	Token      string         `json:"token"`
	Expiry     time.Time      `json:"expiry,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`

	Id    int  `json:"id"`
	Admin bool `json:"admin"`
	Creds []webauthn.Credential
}

// NewSimpleTokenData deliberately leaves Creds nil - a brand new account has
// no registered passkey yet, and it must NOT get one, real or fake. Creds is
// populated later, for real, by SimplePasskeyStore.AddCredential once
// FinishRegistration actually succeeds (simple_passkey_store.go); that
// credential arrives already carrying the correct Transport hints
// (protocol.AuthenticatorTransport) that go-webauthn reads straight out of
// the browser's own registration response - nothing here needs to guess or
// hardcode them.
//
// This used to seed Creds with one fabricated entry (ID: the account's own
// decimal id, AttestationFormat: "apple", Transport: []protocol.
// AuthenticatorTransport{protocol.Internal}) on every single account. That
// entry had no PublicKey/attestation data, so it could never actually
// authenticate anyone - but PasskeyCredentials() (simple_passkey_store.go)
// returns u.rec.data.Creds verbatim, so it WAS flowing into every
// BeginRegistration's excludeCredentials and every BeginLogin's
// allowCredentials, mixed in with any real credential the account went on to
// register. Its hardcoded Transport: [Internal] is what forced the browser
// to always jump straight to the platform authenticator (Touch ID/Face
// ID/Windows Hello) UI, for every account, permanently, regardless of what
// kind of authenticator a real registered credential actually supported -
// confirmed directly: removing just that fake entry's Transport field made
// the browser fall back to its generic chooser (security key / QR-code
// cross-device) instead, since an entry with no transport hint tells the
// browser it doesn't know what to expect. Removing the whole fake entry,
// rather than re-adding a hardcoded Transport to it, is the fix that doesn't
// also silently block security-key/cross-device passkeys for every account
// once someone legitimately registers one.
func NewSimpleTokenData(name, desc, lang string, id int, isAdmin bool, expiry time.Time) *SimpleTokenData {
	return &SimpleTokenData{
		Name:       name,
		Desc:       desc,
		Lang:       lang,
		Token:      "",
		Expiry:     expiry,
		Extensions: nil,
		Id:         id,
		Admin:      isAdmin,
		Creds: []webauthn.Credential{{
			ID:                gotools.StringToBytes(strconv.Itoa(id)),
			AttestationFormat: "apple",
			Transport:         []protocol.AuthenticatorTransport{protocol.Internal},
		}},
	}
}

// WithExtension sets extension to SimpleTokenData.Extensions and returns SimpleTokenData.
func (s *SimpleTokenData) WithExtension(key string, value any) *SimpleTokenData {
	if s == nil {
		return nil
	}

	if s.Extensions == nil {
		s.Extensions = make(map[string]any)
	}

	s.Extensions[key] = value
	return s
}

// WithToken sets token and returns SimpleTokenData.
func (s *SimpleTokenData) WithToken(token string) *SimpleTokenData {
	if s == nil {
		return nil
	}

	s.Token = token
	return s
}

// WithCreds sets creds and returns SimpleTokenData.
func (s *SimpleTokenData) WithCreds(creds []webauthn.Credential) *SimpleTokenData {
	if s == nil {
		return nil
	}

	s.Creds = creds
	return s
}

func (s *SimpleTokenData) IsAdmin() bool {
	return s.Admin
}

func (s *SimpleTokenData) GetUserID() int {
	return s.Id
}

func (s *SimpleTokenData) IsNotExpired() bool {
	return s.Expiry.After(time.Now())
}

type mapToken struct {
	expiresIn *time.Timer
	signAt    time.Time
	userData  TokenData
	lock      *sync.RWMutex
}

type MapTokens struct {
	expiresIn time.Duration
	tokens    map[string]*mapToken
	lock      sync.RWMutex
}

func NewMapTokens(expiresIn time.Duration) *MapTokens {
	return &MapTokens{
		expiresIn: expiresIn,
		tokens:    make(map[string]*mapToken, 0),
	}
}

func (m *MapTokens) SetToken(s string, userData TokenData) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.tokens == nil {
		m.tokens = make(map[string]*mapToken, 0)
	}

	m.tokens[s] = &mapToken{
		expiresIn: time.AfterFunc(m.expiresIn, func() {
			err := m.RemoveToken(s)
			if err != nil {
				logs.ErrorLog(err, "RemoveToken")
			}
		}),
		userData: userData,
		signAt:   time.Now(),
		lock:     &sync.RWMutex{},
	}
}

func (m *MapTokens) NewToken(userData TokenData) (string, error) {
	s, err := generateRandomString(16)
	if err != nil {
		return "", err
	}

	m.SetToken(s, userData)
	return s, nil
}

func (m *MapTokens) GetToken(s string) TokenData {
	m.lock.RLock()
	defer m.lock.RUnlock()

	token, ok := m.tokens[s]
	if ok {
		return token.userData
	}

	return nil
}

func (m *MapTokens) RemoveToken(s string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, ok := m.tokens[s]
	if !ok {
		return errors.New("not found user in active")
	}

	delete(m.tokens, s)

	return nil
}

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// generateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomString(n int) (string, error) {
	b, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), err
}

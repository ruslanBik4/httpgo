/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/ruslanBik4/logs"
)

type Tokens interface {
	NewToken(userData TokenData) (string, error)
	GetToken(s string) TokenData
	RemoveToken(s string) error
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
}

func NewSimpleTokenData(name, desc, lang string, id int, isAdmin bool, expiry time.Time) *SimpleTokenData {
	return &SimpleTokenData{Admin: isAdmin, Name: name, Desc: desc, Lang: lang, Id: id, Expiry: expiry}
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

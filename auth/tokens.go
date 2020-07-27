package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/ruslanBik4/httpgo/logs"
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

type mapToken struct {
	expiresIn *time.Timer
	signAt    time.Time
	userData  TokenData
	lock      *sync.RWMutex
}

type mapTokens struct {
	expiresIn time.Duration
	tokens    map[string]*mapToken
	lock      sync.RWMutex
}

func (m *mapTokens) NewToken(userData TokenData) (string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.tokens == nil {
		m.tokens = make(map[string]*mapToken, 0)
	}

	s, err := generateRandomString(16)
	if err != nil {
		return "", err
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

	return s, nil
}

func (m *mapTokens) GetToken(s string) TokenData {
	m.lock.RLock()
	defer m.lock.RUnlock()

	token, ok := m.tokens[s]
	if ok {
		return token.userData
	}

	return nil
}

func (m *mapTokens) RemoveToken(s string) error {
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

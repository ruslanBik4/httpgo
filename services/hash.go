// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"crypto/rand"
	"encoding/base64"
	"hash/crc32"

	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/logs"
)

type cryptoService struct {
	name   string
	status string
}

func (c cryptoService) Init(ctx context.Context) error {
	return nil
}

func (c cryptoService) Send(ctx context.Context, messages ...interface{}) error {
	switch messages[0] {
	case "password":
		return nil
	default:
		return nil
	}
}

func (c cryptoService) Get(ctx context.Context, messages ...interface{}) (response interface{}, err error) {
	switch messages[0] {
	case "password":
		return GeneratePassword(messages[1].(string))
	default:
		return nil, nil
	}
}

func (c cryptoService) Connect(in <-chan interface{}) (out chan interface{}, err error) {
	panic("implement me")
}

func (c cryptoService) Close(out chan<- interface{}) error {
	panic("implement me")
}

func (c cryptoService) Status() string {
	return c.status
}

var cryptoServ = cryptoService{"crypto", "ready"}

// GeneratePassword run password by email
func GeneratePassword(email string) (string, error) {
	logs.DebugLog("email", email)
	return GenerateRandomString(16)

}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// HashPassword create hash from {password} & return checksumm
func HashPassword(password []byte) uint32 {
	// crypto password
	crc32q := crc32.MakeTable(0xD5828281)
	return crc32.Checksum(password, crc32q)
}

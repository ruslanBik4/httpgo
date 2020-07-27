// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testTokenData struct {
	id      int
	isAdmin bool
}

func (t *testTokenData) IsAdmin() bool {
	return t.isAdmin
}

func (t *testTokenData) GetUserID() int {
	return t.id
}

func Test_mapTokens_GetToken(t *testing.T) {
	type fields struct {
		expiresIn time.Duration
		tokens    map[string]*mapToken
		lock      sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		bearer string
		want   *testTokenData
	}{
		// TODO: Add test cases.
		{
			"1",
			fields{
				tokens: map[string]*mapToken{
					"1": &mapToken{
						userData: &testTokenData{
							id:      1,
							isAdmin: false,
						},
					},
				},
			},
			"1",
			&testTokenData{
				id:      1,
				isAdmin: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mapTokens{
				expiresIn: tt.fields.expiresIn,
				tokens:    tt.fields.tokens,
				lock:      tt.fields.lock,
			}

			got := m.GetToken(tt.bearer)
			assert.Equal(t, tt.want.id, got.GetUserID())
			assert.Equal(t, tt.want.isAdmin, got.IsAdmin())
		})
	}
}

func Test_mapTokens_NewToken(t *testing.T) {
	type fields struct {
		expiresIn time.Duration
		tokens    map[string]*mapToken
		lock      sync.RWMutex
	}
	type args struct {
		userData TokenData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mapTokens{
				expiresIn: tt.fields.expiresIn,
				tokens:    tt.fields.tokens,
				lock:      tt.fields.lock,
			}
			if got, _ := m.NewToken(tt.args.userData); got != tt.want {
				t.Errorf("NewToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapTokens_RemoveToken(t *testing.T) {
	type fields struct {
		expiresIn time.Duration
		tokens    map[string]*mapToken
		lock      sync.RWMutex
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mapTokens{
				expiresIn: tt.fields.expiresIn,
				tokens:    tt.fields.tokens,
				lock:      tt.fields.lock,
			}
			if err := m.RemoveToken(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("RemoveToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

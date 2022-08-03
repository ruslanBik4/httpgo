/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package json

import (
	"bytes"
	"database/sql"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ruslanBik4/logs"
)

func TestAnyJSON(t *testing.T) {
	type args struct {
		arrJSON map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AnyJSON(tt.args.arrJSON); got != tt.want {
				t.Errorf("AnyJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

type UsersFields struct {
	Id          int32          `json:"id"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	Isdel       bool           `json:"isdel"`
	Id_roles    int32          `json:"id_roles"`
	Last_login  time.Time      `json:"last_login"`
	Hash        int64          `json:"hash"`
	Last_page   sql.NullString `json:"last_page"`
	Address     sql.NullString `json:"address"`
	Emailpool   []string       `json:"emailpool"`
	Phones      []string       `json:"phones"`
	Languages   []string       `json:"languages"`
	IdHomepages int32          `json:"homepage"`
	CreateAt    time.Time      `json:"createAt"`
}
type FormActions struct {
	FormErrors map[string]string `json:"formErrors"`
}
type User struct {
	*UsersFields
	Form        string        `json:"form"`
	Lang        string        `json:"lang"`
	Token       string        `json:"token"`
	ContentURL  string        `json:"content_url"`
	FormActions []FormActions `json:"formActions"`
}

func TestElement(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want string
	}{
		// TODO: Add test cases.
		{
			"string with escaped symbols",
			`tralal"'"as'"'as`,
			`"tralal\"\u0027\"as\u0027\"\u0027as"`,
		},
		{
			"forms",
			User{
				UsersFields: &UsersFields{
					Id:          0,
					Name:        "ruslan",
					Email:       "trep@mail.com",
					Isdel:       false,
					Id_roles:    3,
					Last_login:  time.Date(2020, 01, 14, 12, 34, 12, 0, time.UTC),
					Hash:        131313,
					Last_page:   sql.NullString{String: "/profile/user/", Valid: true},
					Address:     sql.NullString{String: `Kyiv, Xhrechatik, 2"A"/12`, Valid: true},
					Emailpool:   []string{"ru@ru.ru", "ASFSfsfs@gmail.ru"},
					Phones:      []string{"+380(66)13e23423", "(443)343434d12"},
					Languages:   []string{"ua", "en", "ru"},
					IdHomepages: 0,
					CreateAt:    time.Date(2020, 01, 14, 12, 34, 12, 0, time.UTC),
				},
				Form:       "form",
				Lang:       "en",
				Token:      "@#%&#!^$%&^$",
				ContentURL: "ww.google.com",
				FormActions: []FormActions{
					FormActions{FormErrors: map[string]string{"id": "wrong", "password": "true"}},
				},
			},
			`{"id":0,"name":"ruslan","email":"trep@mail.com","isdel":false,"id_roles":3,"last_login":"2020-01-14T12:34:12Z","hash":131313,"last_page":"/profile/user/","address":"Kyiv, Xhrechatik, 2\"A\"/12","emailpool":["ru@ru.ru","ASFSfsfs@gmail.ru"],"phones":["+380(66)13e23423","(443)343434d12"],"languages":["ua","en","ru"],"homepage":0,"createAt":"2020-01-14T12:34:12Z","form":"form","lang":"en","token":"@#%&#!^$%&^$","content_url":"ww.google.com","formActions":[{"formErrors":{"id":"wrong","password":"true"}}]}
`,
		},
	}
	logs.SetDebug(true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !assert.Equal(t, tt.want, Element(tt.args)) {
			}
		})
	}
}
func TestSliceJSON(t *testing.T) {
	type args struct {
		mapJSON []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliceJSON(tt.args.mapJSON); got != tt.want {
				t.Errorf("SliceJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamAnyJSON(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		arrJSON  map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamArrJSON(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		arrJSON  []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamElement(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		value    interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamFloat32Dimension(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		arrJSON  []float32
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamFloat64Dimension(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		arrJSON  []float64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamInt32Dimension(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		arrJSON  []int32
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamInt64Dimension(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		arrJSON  []int64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamSimpleDimension(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		arrJSON  []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamSliceJSON(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		mapJSON  []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestStreamStringDimension(t *testing.T) {
	type args struct {
		qw422016 io.Writer
		arrJSON  []string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
func TestWriteAnyJSON(t *testing.T) {
	type args struct {
		arrJSON map[string]interface{}
	}
	tests := []struct {
		name         string
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qq422016 := &bytes.Buffer{}
			WriteAnyJSON(qq422016, tt.args.arrJSON)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteAnyJSON() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

func TestWriteElement(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name         string
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qq422016 := &bytes.Buffer{}
			WriteElement(qq422016, tt.args.value)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteElement() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}
func TestWriteSliceJSON(t *testing.T) {
	type args struct {
		mapJSON []map[string]interface{}
	}
	tests := []struct {
		name         string
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qq422016 := &bytes.Buffer{}
			WriteSliceJSON(qq422016, tt.args.mapJSON)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteSliceJSON() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

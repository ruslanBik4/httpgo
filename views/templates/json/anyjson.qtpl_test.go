/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package json

import (
	"bytes"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/quicktemplate"

	"github.com/ruslanBik4/logs"
)

func TestAnyJSON(t *testing.T) {
	type args struct {
		arrJSON map[string]any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{map[string]any{"slice": []int{1, 2, 3}}},
			`{"slice":[1,2,3]}`,
		},
		{
			"NUllString simple nil",
			args{map[string]any{"null": nil}},
			`{"null":null}`,
		},
		{
			"struct with NUllString nil",
			args{map[string]any{"name": sql.NullString{
				String: "test",
				Valid:  false,
			}}},
			`{"name":null}`,
		},
		{
			"NUllString nil",
			args{map[string]any{"test": "test"}},
			`{"test":"test"}`,
		},
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
		args any
		want string
	}{
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
					{FormErrors: map[string]string{"id": "wrong", "password": "true"}},
				},
			},
			`{"id":0,"name":"ruslan","email":"trep@mail.com","isdel":false,"id_roles":3,"last_login":"2020-01-14T12:34:12Z","hash":131313,"last_page":"/profile/user/","address":"Kyiv, Xhrechatik, 2\"A\"/12","emailpool":["ru@ru.ru","ASFSfsfs@gmail.ru"],"phones":["+380(66)13e23423","(443)343434d12"],"languages":["ua","en","ru"],"homepage":0,"createAt":"2020-01-14T12:34:12Z","form":"form","lang":"en","token":"@#%&#!^$%&^$","content_url":"ww.google.com","formActions":[{"formErrors":{"id":"wrong","password":"true"}}]}`,
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
		mapJSON []map[string]any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{[]map[string]any{{"slice": []int{1, 2, 3}}}},
			`[{"slice":[1,2,3]}]`,
		},
		{
			"NUllString simple nil",
			args{[]map[string]any{{"null": nil}}},
			`[{"null":null}]`,
		},
		{
			"struct with NUllString nil",
			args{[]map[string]any{{"name": nil}}},
			`[{"name":null}]`,
		},
		{
			"NUllString nil",
			args{[]map[string]any{{"test": "test"}}},
			`[{"test":"test"}]`,
		},
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
		arrJSON map[string]any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{arrJSON: map[string]any{"slice": []int{1, 2, 3}}},
			`{"slice":[1,2,3]}`,
		},
		{
			"NUllString simple nil",
			args{map[string]any{"null": nil}},
			`{"null":null}`,
		},
		{
			"struct with NUllString nil",
			args{map[string]any{"name": nil}},
			`{"name":null}`,
		},
		{
			"NUllString nil",
			args{map[string]any{"test": "test"}},
			`{"test":"test"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			WriteAnyJSON(buf, tt.args.arrJSON)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}

func TestStreamArrJSON(t *testing.T) {
	type args struct {
		arrJSON []any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{[]any{1, 2, 3}},
			`[1,2,3]`,
		},
		{
			"struct with NUllString simple nil",
			args{[]any{"null", sql.NullString{
				String: "test",
				Valid:  false,
			}}},
			`["null",null]`,
		},
		{
			"stringAndNil",
			args{[]any{"name", nil}},
			`["name",null]`,
		},
		{
			"NUllString nil",
			args{[]any{"test", "test"}},
			`["test","test"]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			StreamSlice(quicktemplate.AcquireWriter(buf), tt.args.arrJSON)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}

func TestStreamElement(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{map[string]any{"slice": []int{1, 2, 3}}},
			`{"slice":[1,2,3]}`,
		},
		{
			"NUllString simple nil",
			args{map[string]any{"null": nil}},
			`{"null":null}`,
		},
		{
			"struct with NUllString nil",
			args{map[string]any{"name": nil}},
			`{"name":null}`,
		},
		{
			"NUllString nil",
			args{sql.NullString{
				String: "test",
				Valid:  false,
			}},
			`null`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			WriteElement(buf, tt.args.value)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}

func TestStreamFloat32Dimension(t *testing.T) {
	type args struct {
		arrFloat []float32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{[]float32{1, 2, 3}},
			`[1,2,3]`,
		},
		{
			"overload",
			args{[]float32{1.0000, 2.0, 3.00}},
			`[1,2,3]`,
		},
		{
			"double precesion",
			//todo: 1.000012344 is writing wrong as float32
			args{[]float32{1.0000123, 1.0000124, .0000000002, 3.3300000000000000000001}},
			`[1.0000123,1.0000124,2e-10,3.33]`,
		},
		{
			"NUllFloat simple nil",
			args{[]float32{0.00, 0.1}},
			`[0,0.1]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			StreamSlice(quicktemplate.AcquireWriter(buf), tt.args.arrFloat)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}

func TestStreamFloat64Dimension(t *testing.T) {
	type args struct {
		arrFloat []float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{[]float64{1, 2, 3}},
			`[1,2,3]`,
		},
		{
			"overload",
			args{[]float64{1.0000, 2.0, 3.00}},
			`[1,2,3]`,
		},
		{
			"double precesion",
			args{[]float64{1.0000123, 1.000012344321, .0000000002, 3.33}},
			`[1.0000123,1.000012344321,0.0000000002,3.33]`,
		},
		{
			"NUllFloat simple nil",
			args{[]float64{0.00, 0.0000000}},
			`[0,0]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			StreamSlice(quicktemplate.AcquireWriter(buf), tt.args.arrFloat)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}

func TestStreamInt32Dimension(t *testing.T) {
	type args struct {
		arrInt []int32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{[]int32{1, 2, 3}},
			`[1,2,3]`,
		},
		{
			"overload",
			args{[]int32{1111111111, 2, 3}},
			`[1111111111,2,3]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			StreamSlice(quicktemplate.AcquireWriter(buf), tt.args.arrInt)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}

func TestStreamInt64Dimension(t *testing.T) {
	type args struct {
		arrInt []int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{[]int64{1, 2, 3}},
			`[1,2,3]`,
		},
		{
			"overload",
			args{[]int64{11111111111111, 2, 3}},
			`[11111111111111,2,3]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			StreamSlice(quicktemplate.AcquireWriter(buf), tt.args.arrInt)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}

func TestStreamSliceJSON(t *testing.T) {
	type args struct {
		mapJSON []map[string]any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{[]map[string]any{{"slice": []int{1, 2, 3}}}},
			`[{"slice":[1,2,3]}]`,
		},
		{
			"NUllString simple nil",
			args{[]map[string]any{{"null": nil}}},
			`[{"null":null}]`,
		},
		{
			"struct with NUllString nil",
			args{[]map[string]any{{"name": nil}}},
			`[{"name":null}]`,
		},
		{
			"NUllString nil",
			args{[]map[string]any{{"test": "test"}}},
			`[{"test":"test"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			WriteSliceJSON(buf, tt.args.mapJSON)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}

func TestStreamStringDimension(t *testing.T) {
	type args struct {
		arrJSON []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{[]string{"slice", "1", "2", "3"}},
			`["slice","1","2","3"]`,
		},
		{
			"NUllString simple nil",
			args{[]string{"null", "nil"}},
			`["null","nil"]`,
		},
		{
			"struct with NUllString nil",
			args{[]string{"name", ` "n'il"`}},
			`["name"," \"n\u0027il\""]`,
		},
		{
			"NUllString nil",
			args{[]string{"test", "test"}},
			`["test","test"]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			StreamSlice(quicktemplate.AcquireWriter(buf), tt.args.arrJSON)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
		})
	}
}
func TestWriteAnyJSON(t *testing.T) {
	type args struct {
		arrJSON map[string]any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"slice",
			args{map[string]any{"slice": []int{1, 2, 3}}},
			`{"slice":[1,2,3]}`,
		},
		{
			"NUllString simple nil",
			args{map[string]any{"null": nil}},
			`{"null":null}`,
		},
		{
			"struct with NUllString nil",
			args{map[string]any{"name": nil}},
			`{"name":null}`,
		},
		{
			"NUllString nil",
			args{map[string]any{"test": "test"}},
			`{"test":"test"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			WriteAnyJSON(buf, tt.args.arrJSON)
			assert.Equal(t, tt.want, buf.String(), "error result test '%s'", tt.name)
			qq422016 := &bytes.Buffer{}
			WriteAnyJSON(qq422016, tt.args.arrJSON)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.want {
				t.Errorf("WriteAnyJSON() = %v, want %v", gotQq422016, tt.want)
			}
		})
	}
}

func TestWriteElement(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name         string
		args         args
		wantQq422016 string
	}{
		{
			"slice",
			args{map[string]any{"slice": []int{1, 2, 3}}},
			`{"slice":[1,2,3]}`,
		},
		{
			"NUllString simple nil",
			args{map[string]any{"null": nil}},
			`{"null":null}`,
		},
		{
			"struct with NUllString nil",
			args{map[string]any{"name": nil}},
			`{"name":null}`,
		},
		{
			"NUllString nil",
			args{map[string]any{"test": "test"}},
			`{"test":"test"}`,
		},
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
		mapJSON []map[string]any
	}
	tests := []struct {
		name         string
		args         args
		wantQq422016 string
	}{
		{
			"slice",
			args{[]map[string]any{{"slice": []int{1, 2, 3}}}},
			`[{"slice":[1,2,3]}]`,
		},
		{
			"NUllString simple nil",
			args{[]map[string]any{{"null": nil}}},
			`[{"null":null}]`,
		},
		{
			"struct with NUllString nil",
			args{[]map[string]any{{"name": nil}}},
			`[{"name":null}]`,
		},
		{
			"NUllString nil",
			args{[]map[string]any{{"test": "test"}}},
			`[{"test":"test"}]`,
		},
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

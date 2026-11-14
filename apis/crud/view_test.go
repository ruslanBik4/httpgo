/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestToStandardColumnValueType(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "<nil>"},
		{"string", "abc", "abc"},
		{"bytes", []byte("raw"), "raw"},
		{"int32", int32(7), "7"},
		{"strings", []string{"a", "b"}, "[a b]"},
		{"any array", []any{int64(1), "two"}, `[1 "two"]`},
		{"null interval", pgtype.Interval{}, ""},
		{"range", pgtype.Range[pgtype.Int4]{
			Lower: pgtype.Int4{Int32: 1, Valid: true}, Upper: pgtype.Int4{Int32: 5, Valid: true},
			LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true,
		}, "[ 1, 5 )"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToStandardColumnValueType(tt.in))
		})
	}
}

func TestPhotosURL(t *testing.T) {
	assert.Equal(t, "/table/photos?id=3", PhotosURL(3))
	assert.Equal(t, "/table/goods?id=3", PhotosURL(3, "goods"))
	assert.Equal(t, "/table/goods/blob?id=3", PhotosURL(3, "goods", "blob"))
	assert.Equal(t, `<img src='/table/photos/blob?id=3' id='imgblob3' style='max-width:100%'/>`, renderImg(3, "photos", "blob"))
}

func TestTranslator(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	var none Translator
	assert.Equal(t, "Name", none.of(ctx, "Name"), "nil keeps the text")
	upper := Translator(func(_ *fasthttp.RequestCtx, s string) string { return "[" + s + "]" })
	assert.Equal(t, "[Name]", upper.of(ctx, "Name"))
}

func TestUnknownArrayToStrings_Generic(t *testing.T) {
	assert.Equal(t, []string{"1", "2"}, UnknownArrayToStrings([]int{1, 2}))
	assert.Equal(t, UnknownArrayToStrings([]any{int64(1)}), UnknownArrayToStrings([]int64{1}))
}

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

	"github.com/stretchr/testify/assert"
	"github.com/valyala/quicktemplate"
)

func TestStreamWrap(t *testing.T) {
	buf := bytes.NewBufferString("")
	w := quicktemplate.AcquireWriter(buf)

	tests := []struct {
		name  string
		value interface{}
		res   string
	}{
		{
			"slice",
			[]int{1, 2, 3},
			"[1,2,3]",
		},
		{
			"NUllString simple nil",
			sql.NullString{
				String: "test",
				Valid:  false,
			},
			"null",
		},
		{
			"struct with NUllString nil",
			struct {
				Name sql.NullString `json:"name"`
			}{
				sql.NullString{
					String: "test",
					Valid:  false,
				},
			},
			`{"name":null}`,
		},
		{
			"NUllString nil",
			sql.NullString{
				String: "test",
				Valid:  true,
			},
			`"test"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StreamElement(w, tt.value)
			assert.Equal(t, tt.res, buf.String())
			buf.Reset()
		})
	}
}

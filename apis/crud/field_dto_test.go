// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crud

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

func TestDecodeDatetimeString(t *testing.T) {
	tests := []struct {
		name string
		want error
	}{
		// TODO: Add test cases.
		{
			"2021-01-31T22:00:00.000Z",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := DateTimeString{}
			assert.NotNil(t, jsoniter.UnmarshalFromString(tt.name, &v))
		})
	}
}

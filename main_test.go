// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"testing"
)

func TestNewDefaultHandler(t *testing.T) {
	tests := []struct {
		name string
		want *DefaultHandler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultHandler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

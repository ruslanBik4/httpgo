// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	type args struct {
		password []byte
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// TODO: Add test cases.
		{
			"bytes",
			args{
				[]byte("YTk_gJ5R0kFK8cmfgvn0eQ=="),
			},
			557107631,
		},
		{
			"bytes",
			args{
				[]byte("PqqSpSmTfqVlf9WO6LXJAw=="),
			},
			2966412999,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashPassword(tt.args.password); got != tt.want {
				t.Errorf("HashPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

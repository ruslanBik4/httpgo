/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func TestImportResultJSON(t *testing.T) {
	r := NewImportResult("/table/t/browse/limit={page}", 3, "COPY 3")
	b, err := jsoniter.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"formActions":[{"type":"redirect","url":"/table/t/browse/limit={page}"}],"count":3,"message":"COPY 3"}`
	if string(b) != want {
		t.Errorf("got  %s\nwant %s", b, want)
	}
	r.Excel = map[string]int{"rows": 3}
	if b, _ = jsoniter.Marshal(r); string(b) != want[:len(want)-1]+`,"excel":{"rows":3}}` {
		t.Errorf("with excel: %s", b)
	}
}

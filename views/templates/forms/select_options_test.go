/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package forms

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

func TestSelectOptionsFncAndMultiple(t *testing.T) {
	val := fastjson.MustParse(`{"id": 1, "title": "Site", "columns": [
		{"name": "id_themes", "type": "select", "multiple": true, "title": "Themes", "options": "GetThemesOptions(ctx)"}
	]}`)
	b := NewBlockColumnsFromJSON(val, nil)
	require.NotNil(t, b)
	require.NotEqual(t, 1, len(b.Columns), "columns %d", len(b.Columns))
	col := b.Columns[0]
	if !col.Multiple || col.SelectOptionsFnc != "GetThemesOptions(ctx)" {
		t.Fatalf("%+v", col)
	}

	f := &FormField{Title: "Sales site", Action: "/project/sales", Method: "POST", Blocks: []BlockColumns{*b}}
	var code bytes.Buffer
	f.WriteCreate(&code, "github.com/x/y", "views", "sales")
	for _, want := range []string{"SelectOptions: GetThemesOptions(ctx),", "Multiple: true,"} {
		if !strings.Contains(code.String(), want) {
			t.Errorf("generated code has no %q", want)
		}
	}

	// the rendered select takes several values
	col.SelectOptions = map[string]SelectOption{"black": {Value: "1"}, "white": {Value: "2"}, "red": {Value: "3"}}
	var html bytes.Buffer
	col.WriteRenderInputs(&html, nil)
	if !strings.Contains(html.String(), `multiple size="3"`) || !strings.Contains(html.String(), `name="id_themes"`) {
		t.Errorf("select: %s", html.String())
	}
}

/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package routeTable

import (
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/views/templates/forms"
)

// col is a column of any type; the dbEngine test column has a fixed type.
type col struct {
	*dbEngine.StringColumn
	basic   types.BasicKind
	typ     string
	primary bool
}

func (c *col) BasicType() types.BasicKind     { return c.basic }
func (c *col) BasicTypeInfo() types.BasicInfo { return types.Typ[c.basic].Info() }
func (c *col) Type() string                   { return c.typ }
func (c *col) Primary() bool                  { return c.primary }

func newCol(name, typ string, basic types.BasicKind, primary bool) *col {
	return &col{StringColumn: dbEngine.NewStringColumn(name, name, false), basic: basic, typ: typ, primary: primary}
}

func newRender(onlyView bool, cols []dbEngine.Column, rows [][]any) *TableRender {
	table := dbEngine.NewTableString("users", "Users", cols, nil, nil)
	decors := make([]*forms.ColumnDecor, len(cols))
	for i, c := range cols {
		decors[i] = forms.NewColumnDecor(c, nil)
	}

	return NewTableRender(table, decors, rows, "/table/users/", onlyView)
}

func users() []dbEngine.Column {
	return []dbEngine.Column{
		newCol("id", "int4", types.Int32, true),
		newCol("name", "varchar", types.String, false),
		newCol("active", "bool", types.Bool, false),
	}
}

func TestBoolValue(t *testing.T) {
	for val, want := range map[any]bool{
		true: true, false: false, "true": true, "false": false, " True ": true, "t": true, "1": true,
		"0": false, "": false, nil: false, int32(1): false,
	} {
		assert.Equal(t, want, BoolValue(val), "%#v", val)
	}
}

func TestPkIndex(t *testing.T) {
	assert.Equal(t, 0, newRender(false, users(), nil).pkIndex())

	byName := newRender(false, []dbEngine.Column{newCol("x", "varchar", types.String, false), newCol("id", "int4", types.Int32, false)}, nil)
	assert.Equal(t, 1, byName.pkIndex(), "a column named id when none is marked")

	two := newRender(false, []dbEngine.Column{newCol("a", "int4", types.Int32, true), newCol("b", "int4", types.Int32, true)}, nil)
	assert.Equal(t, -1, two.pkIndex(), "a composite key")

	none := newRender(false, []dbEngine.Column{newCol("x", "varchar", types.String, false)}, nil)
	assert.Equal(t, -1, none.pkIndex())
	assert.False(t, none.editable())
}

func TestRenderRowsSwitches(t *testing.T) {
	rows := [][]any{
		{"1", "Alice", "true"}, // ToTableView gives a bool as a text
		{"2", "Bob", "false"},
		{int32(3), "Cher", true}, // a bool as it is
		{int32(4), "Dan", nil},   // NULL
	}

	t.Run("editable", func(t *testing.T) {
		html := newRender(false, users(), rows).RenderRows()
		assert.Equal(t, 4, strings.Count(html, `class="tbl-switch"`), "every row has a switch")
		assert.Equal(t, 2, strings.Count(html, ` checked data-saved="true"`))
		assert.Equal(t, 2, strings.Count(html, ` data-saved="false"`))
		assert.NotContains(t, html, "disabled")
		assert.Contains(t, html, `data-col="active"`)
		for _, pk := range []string{"1", "2", "3", "4"} {
			assert.Contains(t, html, `data-pk="`+pk+`"`)
		}
		assert.NotContains(t, html, ">true<", "no text true/false any more")
	})

	t.Run("read only", func(t *testing.T) {
		html := newRender(true, users(), rows).RenderRows()
		assert.Equal(t, 4, strings.Count(html, "disabled"))
	})
}

// The browse handler replaces the first column (the key) by the link to the edit
// form of the row: the key for the server is in its href.
func TestRowKeyFromEditLink(t *testing.T) {
	link := func(id string) string {
		return `<a href='/table/users/form?html=true&lang=en-GB&id=` + id + `' title="Edit record" hx-swap="none" data-fancybox>§</a>`
	}

	t.Run("rendered rows", func(t *testing.T) {
		html := newRender(false, users(), [][]any{
			{link("1"), "Alice", true},
			{link("22"), "Bob", false},
		}).RenderRows()
		assert.Contains(t, html, `data-pk="1"`)
		assert.Contains(t, html, `data-pk="22"`)
		assert.NotContains(t, html, `data-pk="<`, "never a link as a key")
		assert.NotContains(t, html, "disabled")
	})

	t.Run("a link without the key", func(t *testing.T) {
		html := newRender(false, users(), [][]any{{`<a href='/table/users/form'>§</a>`, "Alice", true}}).RenderRows()
		assert.NotContains(t, html, "data-pk")
		assert.Contains(t, html, "disabled", "a switch of a row with no key can't save")
	})

	t.Run("rowKey", func(t *testing.T) {
		re := keyRegexp("id")
		for val, want := range map[any]string{
			"7":                                "7",
			int32(8):                           "8",
			link("9"):                          "9",
			`<a href="/f?id=10&amp;x=1">§</a>`: "10",
			`<a href="/f?x=1&amp;id=11">§</a>`: "11",
			`<a href='/f?grid=5'>§</a>`:        "",
			`<a href='/f?id=a%20b'>§</a>`:      "a b",
			`<a href='/f'>§</a>`:               "",
		} {
			assert.Equal(t, want, rowKey(val, re), "%v", val)
		}
	})
}

func TestAllTableAttributes(t *testing.T) {
	t.Run("editable table knows where to save", func(t *testing.T) {
		html := newRender(false, users(), nil).AllTable()
		for _, want := range []string{
			`data-update-url="/table/users/update"`,
			`data-pk-col="id"`,
			`data-batch-url="/table/users/put/batch"`,
			`data-bool-all-url="/table/users/set_bool_all"`,
		} {
			assert.Contains(t, html, want)
		}
	})

	t.Run("read only table doesn't", func(t *testing.T) {
		html := newRender(true, users(), nil).AllTable()
		assert.NotContains(t, html, "data-update-url")
		assert.NotContains(t, html, "data-batch-url")
	})

	t.Run("filter of a boolean column is a switch of three states", func(t *testing.T) {
		html := newRender(false, users(), nil).AllTable()
		assert.Contains(t, html, `class="tbl-switch tbl-switch--filter"`)
		assert.Contains(t, html, `data-state="any"`)
		assert.Contains(t, html, `data-name="active"`)
		assert.Contains(t, html, `data-class="table-col-2"`)
		assert.Equal(t, 1, strings.Count(html, "tbl-switch--filter"), "only for the boolean column")
		assert.Contains(t, html, `<input type="text" data-name="name"`, "other filters are as they were")
	})
}

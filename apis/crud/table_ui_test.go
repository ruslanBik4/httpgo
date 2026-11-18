/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"go/types"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/auth"
)

// newUsers is a table with a boolean column, a column for administrators only,
// a foreign key and a file.
func newUsers() (*fakeTable, *fakeTable) {
	roles := newTable("roles", "Roles",
		newCol("id", "int4", asPrimary),
		newCol("name", "varchar"),
	)
	users := newTable("users", "Users",
		newCol("id", "int4", asPrimary, asAutoInc),
		newCol("name", "varchar", withComment("Full name")),
		newCol("active", "bool", withBasic(types.Bool), withComment("Active")),
		newCol("note", "varchar", withComment(`Note {"admin_only": true}`)),
		newCol("id_roles", "int4", withBasic(types.Int32), refersTo(roles, "id")),
		newCol("avatar", "bytea", withBasic(types.UnsafePointer)),
		newCol("settings", "jsonb"),
	)

	return users, roles
}

func TestHasBoolColumns(t *testing.T) {
	users, roles := newUsers()
	assert.True(t, HasBoolColumns(users))
	assert.False(t, HasBoolColumns(roles))
}

func TestSinglePrimary(t *testing.T) {
	users, _ := newUsers()
	assert.Equal(t, "id", SinglePrimary(users))

	two := newTable("pairs", "",
		newCol("a", "int4", asPrimary),
		newCol("b", "int4", asPrimary),
	)
	assert.Equal(t, "", SinglePrimary(two), "a composite key is not a single one")

	byName := newTable("plain", "", newCol("id", "int4"), newCol("x", "varchar"))
	assert.Equal(t, "id", SinglePrimary(byName), "a column named id when none is marked")

	none := newTable("none", "", newCol("x", "varchar"))
	assert.Equal(t, "", SinglePrimary(none))
}

func TestBuildSetBoolSQL(t *testing.T) {
	users, _ := newUsers()
	const cte = `WITH changed AS (%s RETURNING 1) SELECT count(*) FROM changed`
	wrap := func(q string) string { return strings.Replace(cte, "%s", q, 1) }

	tests := []struct {
		name               string
		column, value, ids string
		all                bool
		wantSQL            string
		wantArgs           []any
		wantErr            string
	}{
		{name: "toggle chosen rows", column: "active", value: "toggle", ids: "1, 2,2,,3",
			wantSQL:  wrap(`UPDATE "users" SET "active" = NOT COALESCE("active", false) WHERE "id"::text = ANY($1)`),
			wantArgs: []any{[]string{"1", "2", "3"}}},
		{name: "true for chosen rows", column: "active", value: "true", ids: "7",
			wantSQL:  wrap(`UPDATE "users" SET "active" = $1 WHERE "id"::text = ANY($2)`),
			wantArgs: []any{true, []string{"7"}}},
		{name: "false for all rows", column: "active", value: "false", all: true,
			wantSQL:  wrap(`UPDATE "users" SET "active" = $1`),
			wantArgs: []any{false}},
		{name: "toggle of all rows is not allowed", column: "active", value: "toggle", all: true,
			wantErr: "'toggle' changes only chosen rows"},
		{name: "unknown column", column: "nope", value: "true", ids: "1", wantErr: "unknown column 'nope'"},
		{name: "not a boolean column", column: "name", value: "true", ids: "1", wantErr: "column 'name' isn't boolean"},
		{name: "unknown value", column: "active", value: "maybe", ids: "1", wantErr: "value must be 'true' or 'false'"},
		{name: "no rows", column: "active", value: "true", ids: " , ", wantErr: "no rows chosen"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := BuildSetBoolSQL(users, "id", tt.column, tt.value, tt.ids, tt.all)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSQL, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}

	t.Run("unknown primary column", func(t *testing.T) {
		_, _, err := BuildSetBoolSQL(users, "key", "active", "true", "1", false)
		assert.EqualError(t, err, "unknown primary column 'key'")
	})

	t.Run("names are quoted", func(t *testing.T) {
		odd := newTable("my.table", "", newCol("id", "int4", asPrimary), newCol(`a"b`, "bool", withBasic(types.Bool)))
		sql, _, err := BuildSetBoolSQL(odd, "id", `a"b`, "true", "1", false)
		require.NoError(t, err)
		assert.Contains(t, sql, `UPDATE "my"."table" SET "a""b" = $1`)
	})

	t.Run("too many rows", func(t *testing.T) {
		ids := make([]string, maxBulkIDs+1)
		for i := range ids {
			ids[i] = strconv.Itoa(i)
		}
		_, _, err := BuildSetBoolSQL(users, "id", "active", "true", strings.Join(ids, ","), false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too many rows")
	})
}

func TestCardColumns(t *testing.T) {
	users, _ := newUsers()
	names := func(isAdmin bool) []string {
		var res []string
		for _, col := range CardColumns(users, "name", isAdmin) {
			res = append(res, col.Name())
		}
		return res
	}

	// title first; no id, no foreign key, no file, no json
	assert.Equal(t, []string{"name", "active"}, names(false))
	assert.Equal(t, []string{"name", "active", "note"}, names(true), "an administrator sees the admin_only column")
}

func TestReadCard(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	users, _ := newUsers()

	t.Run("not an administrator", func(t *testing.T) {
		users.rows = [][]any{{"Alice\nSmith", true}}
		title, fields, found, err := ReadCard(ctx, users, "id", int32(3), "en", false)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "Alice Smith", title)
		assert.Equal(t, []CardField{{Label: "Active", Value: "true"}}, fields)
		assert.Contains(t, users.sql[len(users.sql)-1], "active")
		assert.NotContains(t, users.sql[len(users.sql)-1], "note")
	})

	t.Run("administrator, empty values are skipped, long ones are cut", func(t *testing.T) {
		users.rows = [][]any{{"Bob", nil, strings.Repeat("я", maxCardValueLen+10)}}
		title, fields, found, err := ReadCard(ctx, users, "id", int32(4), "en", true)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "Bob", title)
		require.Len(t, fields, 1)
		assert.Equal(t, "Note", fields[0].Label)
		assert.Equal(t, strings.Repeat("я", maxCardValueLen)+"…", fields[0].Value)
	})

	t.Run("no such row", func(t *testing.T) {
		users.rows = nil
		_, _, found, err := ReadCard(ctx, users, "id", int32(5), "en", false)
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func TestRenderCard(t *testing.T) {
	got := RenderCard("users", int32(3), "<b>Alice</b>", []CardField{{Label: "A&B", Value: `x"y`}})
	assert.Equal(t, `<dl class="tbl-card">`+
		`<dt class="tbl-card__title"></dt><dd class="tbl-card__title">`+
		`<a href='/table/users/form?html&id=3' target="_modal">&lt;b&gt;Alice&lt;/b&gt;</a></dd>`+
		`<dt>A&amp;B</dt><dd>x&#34;y</dd></dl>`, got)

	assert.Contains(t, RenderCard("users", "a b", "", nil), `id=a+b' target="_modal">…</a>`, "the id is escaped, an empty title is shown")
}

func TestForeignCell(t *testing.T) {
	_, roles := newUsers()
	assert.Equal(t,
		`<span class="tbl-fk" tabindex="0" role="button" data-col="id_roles" data-fk-table="roles" data-fk-id="2" data-card="/table/roles/card?id=">Admin &amp; Co</span>`,
		ForeignCell(roles, "id_roles", int32(2), "Admin & Co"))
}

func TestIsAdminRequest(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	assert.False(t, IsAdminRequest(ctx), "no token")

	ctx.SetUserValue(auth.UserValueToken, auth.NewSimpleTokenData("u", "", "en", 1, false, time.Time{}))
	assert.False(t, IsAdminRequest(ctx))

	ctx.SetUserValue(auth.UserValueToken, auth.NewSimpleTokenData("a", "", "en", 2, true, time.Time{}))
	assert.True(t, IsAdminRequest(ctx))
}

func TestTableSetBoolWrongParams(t *testing.T) {
	users, _ := newUsers()
	fnc := TableSetBool(nil, users, "id", false)

	ctx := &fasthttp.RequestCtx{}
	ctx.SetUserValue("column", "name")
	ctx.SetUserValue("value", "true")
	ctx.SetUserValue("ids", "1")
	res, err := fnc(ctx)
	assert.ErrorIs(t, err, apis.ErrWrongParamsList)
	assert.Equal(t, map[string]string{"column": "column 'name' isn't boolean"}, res)
}

func TestTableCard(t *testing.T) {
	users, _ := newUsers()
	fnc := TableCard(users, "id")

	t.Run("a row", func(t *testing.T) {
		users.rows = [][]any{{"Alice", true}}
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue("id", int32(3))
		res, err := fnc(ctx)
		require.NoError(t, err)
		assert.Nil(t, res)
		body := string(ctx.Response.Body())
		assert.Contains(t, body, `<a href='/table/users/form?html&id=3' target="_modal">Alice</a>`)
		assert.Contains(t, body, `<dt>Active</dt><dd>true</dd>`)
		assert.Contains(t, string(ctx.Response.Header.ContentType()), "text/html")
	})

	t.Run("no such row", func(t *testing.T) {
		users.rows = nil
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue("id", int32(9))
		_, err := fnc(ctx)
		require.NoError(t, err)
		assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	})

	t.Run("no id", func(t *testing.T) {
		_, err := fnc(&fasthttp.RequestCtx{})
		assert.ErrorIs(t, err, apis.ErrWrongParamsList)
	})
}

var _ dbEngine.Table = (*fakeTable)(nil)

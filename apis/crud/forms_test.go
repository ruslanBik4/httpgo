/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"context"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/views"
)

func TestGetNameOfTitleColumn(t *testing.T) {
	tests := []struct {
		name  string
		table dbEngine.Table
		lang  any
		want  string
	}{
		{"name", newShop().brands, nil, "name"},
		{"name in the language", newShop().brands, "ua", "name_ua"},
		{"no column of the language", newShop().brands, "de", "name"},
		{"lang of not a string", newShop().brands, 1, "name"},
		{"title", newShop().colors, "ua", "title"},
		{"name before title", newTable("t", "",
			newCol("title", "varchar"), newCol("name", "varchar")), nil, "name"},
		{"first not primary column", newTable("t", "",
			newCol("id", "int4", asPrimary), newCol("code", "varchar"), newCol("memo", "text")), nil, "code"},
		{"only primary columns", newTable("t", "", newCol("id", "int4", asPrimary)), nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetNameOfTitleColumn(tt.table, tt.lang))
		})
	}
}

func TestGetNameAccordingLang(t *testing.T) {
	brands := newShop().brands
	assert.Equal(t, "name_ua", GetNameAccordingLang(brands, "name", "ua"))
	assert.Equal(t, "name", GetNameAccordingLang(brands, "name", "en"))
	assert.Equal(t, "name", GetNameAccordingLang(brands, "name", nil))
}

func TestGetForeignName(t *testing.T) {
	s := newShop()
	ctx := &fasthttp.RequestCtx{}
	idBrands := s.goods.FindColumn("id_brands")

	assert.Equal(t, "Acme", GetForeignName(ctx, s.DB, idBrands, int32(3)))
	assert.Equal(t, "SELECT name FROM brands WHERE id=$1", s.brands.sql[0])

	assert.Nil(t, GetForeignName(ctx, s.DB, idBrands, nil), "nil value")
	assert.Nil(t, GetForeignName(ctx, s.DB, idBrands, int32(9)), "no such row")
	assert.Nil(t, GetForeignName(ctx, s.DB, s.goods.FindColumn("price"), 1.0), "not a foreign key")

	delete(s.DB.Tables, "brands")
	assert.Nil(t, GetForeignName(ctx, s.DB, idBrands, int32(3)), "no parent table")
}

func TestRoutesFromDB(t *testing.T) {
	s := newShop()
	all := []DbRouteType{DbRouteType_Insert, DbRouteType_Update, DbRouteType_Select}
	ctx := context.WithValue(context.Background(), apis.Database, s.DB)

	t.Run("no DB in the context", func(t *testing.T) {
		assert.Nil(t, RoutesFromDB(context.Background(), all, "goods"))
	})

	t.Run("routes of the listed tables only", func(t *testing.T) {
		routes := RoutesFromDB(ctx, all, "goods", "unknown")
		assert.Equal(t, []string{
			"/api/v1/table/goods/form",
			"/api/v1/table/goods/get",
			"/api/v1/table/goods/put",
			"/api/v1/table/goods/update",
		}, slices.Sorted(maps.Keys(routes)))
	})

	t.Run("no tables - no routes", func(t *testing.T) {
		assert.Empty(t, RoutesFromDB(ctx, all))
	})

	t.Run("path version of the context", func(t *testing.T) {
		ctx := context.WithValue(ctx, PathVersion, "/api/v2")
		routes := RoutesFromDB(ctx, []DbRouteType{DbRouteType_Select}, "brands")
		assert.Equal(t, []string{"/api/v2/table/brands/get"}, slices.Sorted(maps.Keys(routes)))
	})
}

func TestCreateRoutesForTable(t *testing.T) {
	s := newShop()
	defParams := []apis.InParam{ParamsLang}
	routes := createRoutesForTable(s.DB, "goods", "/table/", s.goods,
		[]DbRouteType{DbRouteType_Insert, DbRouteType_Update, DbRouteType_Select}, defParams)
	names := func(params []apis.InParam) []string {
		res := make([]string, len(params))
		for i, p := range params {
			res[i] = p.Name
		}
		return res
	}

	put := routes["/table/goods/put"]
	require.NotNil(t, put)
	assert.Equal(t, apis.POST, put.Method)
	assert.True(t, put.NeedAuth)
	assert.True(t, put.Multipart)
	assert.NotContains(t, names(put.Params), "id", "no autoincrement column on insert")
	assert.Contains(t, names(put.Params), "name")

	update := routes["/table/goods/update"]
	require.NotNil(t, update)
	assert.Contains(t, names(update.Params), "id", "update finds the row by its key")

	get := routes["/table/goods/get"]
	require.NotNil(t, get)
	assert.Equal(t, apis.GET, get.Method)
	for _, p := range get.Params {
		if p.Name != ParamsLang.Name {
			assert.False(t, p.Req, "%s: every column is optional on select", p.Name)
		}
	}

	form := routes["/table/goods/form"]
	require.NotNil(t, form)
	assert.Contains(t, names(form.Params), "id")

	t.Run("no form for select only", func(t *testing.T) {
		routes := createRoutesForTable(s.DB, "goods", "/table/", s.goods, []DbRouteType{DbRouteType_Select}, defParams)
		assert.NotContains(t, routes, "/table/goods/form")
		assert.Len(t, routes, 1)
	})
}

func TestTableForm(t *testing.T) {
	newForm := func(s *shop) apis.ApiRouteHandler {
		return TableForm(s.DB, "/table/", s.goods, []string{"id"})
	}

	t.Run("new record", func(t *testing.T) {
		s := newShop()
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue("html", true)
		ctx.SetUserValue(ParamsLang.Name, "ua")
		res, err := newForm(s)(ctx)
		require.NoError(t, err)
		assert.Nil(t, res)

		body := string(ctx.Response.Body())
		assert.Contains(t, body, "/table/goods/put")
		assert.Contains(t, body, "Insert")
		assert.Contains(t, body, `name="name"`)
		assert.Contains(t, body, `name="lang"`, "the language goes with the form")
		assert.NotContains(t, body, `name="id"`, "no autoincrement column")
	})

	t.Run("edit record", func(t *testing.T) {
		s := newShop()
		s.goods.rows = [][]any{{"Hammer", 12.5}}
		s.goods.rowCols = []dbEngine.Column{s.goods.FindColumn("name"), s.goods.FindColumn("price")}
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue("html", true)
		ctx.SetUserValue("id", int32(1))
		_, err := newForm(s)(ctx)
		require.NoError(t, err)

		body := string(ctx.Response.Body())
		assert.Contains(t, body, "/table/goods/update")
		assert.Contains(t, body, "Update")
		assert.Contains(t, body, "Hammer")
		require.NotEmpty(t, s.goods.sql)
		assert.Contains(t, s.goods.sql[0], "WHERE id=$1")
	})

	t.Run("id from the last segment of the path", func(t *testing.T) {
		s := newShop()
		s.goods.rows = [][]any{{"Hammer"}}
		s.goods.rowCols = []dbEngine.Column{s.goods.FindColumn("name")}
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue(views.ServerName, "test") // the server sets it for the JSON headers
		ctx.SetUserValue(apis.ChildRoutePath, "1")
		_, err := newForm(s)(ctx)
		require.NoError(t, err)
		assert.Contains(t, string(ctx.Response.Header.ContentType()), "json", "JSON without the html param")
		assert.Contains(t, string(ctx.Response.Body()), "/table/goods/update")
	})

	t.Run("wrong id in the path", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue(apis.ChildRoutePath, "x")
		res, err := newForm(newShop())(ctx)
		assert.ErrorIs(t, err, apis.ErrWrongParamsList)
		assert.Equal(t, apis.ChildRoutePath, res)
	})

	t.Run("no such record", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue("id", int32(9))
		res, err := newForm(newShop())(ctx)
		require.NoError(t, err)
		assert.Nil(t, res)
		assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
	})
}

func TestConverters(t *testing.T) {
	assert.Equal(t, "2026-11-14", TimeToString(time.Date(2026, 11, 14, 23, 59, 0, 0, time.UTC)))
	assert.Equal(t, "/api/v1/blob/goods?id=3&name=picture", BlobToURL("goods", "picture", 3))
	assert.Equal(t, []int32{1, 0}, Int4ArrToStrings([]pgtype.Int4{{Int32: 1, Valid: true}, {}}))
	assert.Equal(t, []int64{2, 0}, Int8ArrToStrings([]pgtype.Int8{{Int64: 2, Valid: true}, {}}))
	assert.Equal(t, []string{"a", ""}, TextArrayToStrings([]pgtype.Text{{String: "a", Valid: true}, {}}))
	assert.Empty(t, TextArrayToStrings(nil))
}

func TestUnknownArrayToStrings(t *testing.T) {
	assert.Equal(t, []string{"1", "2"}, UnknownArrayToStrings([]int{1, 2}))
	assert.Equal(t, []string{`"a"`, "true"}, UnknownArrayToStrings([]any{"a", true}), "strings as JSON")
	assert.Equal(t, UnknownArrayToStrings([]any{int64(1)}), UnknownArrayToStrings([]int64{1}))
	assert.Empty(t, UnknownArrayToStrings([]string(nil)))
}

/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"bytes"
	"go/types"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/views/templates/forms"
)

// shop is a small schema: goods refer to brands (by a foreign key),
// to photos (by a foreign key) and to colors (only by the name id_colors).
type shop struct {
	brands, photos, colors, goods *fakeTable
	DB                            *dbEngine.DB
}

func newShop() *shop {
	s := &shop{}
	s.brands = newTable("brands", "Brands",
		newCol("id", "int4", asPrimary),
		newCol("name", "varchar"),
		newCol("name_ua", "varchar"),
	)
	s.brands.byArg["3"] = "Acme"
	s.brands.byArg["[3 4]"] = []string{"Acme", "Bolt"}
	s.brands.rows = [][]any{{int32(3), "Acme"}, {int32(4), "Bolt"}}

	s.photos = newTable("photos", "Photos",
		newCol("id", "int4", asPrimary),
		newCol("blob", "bytea", withBasic(types.UnsafePointer)),
	)
	s.photos.byArg["5"] = "5"

	s.colors = newTable("colors", "Colors",
		newCol("id", "int4", asPrimary),
		newCol("title", "varchar"),
	)
	s.colors.byArg["1"] = "red"

	s.goods = newTable("goods", "Goods",
		newCol("id", "int4", asPrimary, asAutoInc),
		newCol("name", "varchar", asRequired, withComment("Name of goods")),
		newCol("price", "numeric", withBasic(types.Float64)),
		newCol("id_brands", "int4", withBasic(types.Int32), refersTo(s.brands, "id")),
		newCol("id_photos", "int4", withBasic(types.Int32), refersTo(s.photos, "id")),
		newCol("id_colors", "int4", withBasic(types.Int32)),
		newCol("tags", "_int4", withBasic(types.Int32), refersTo(s.brands, "id")),
		newCol("updateat", "timestamptz"),
		newCol("picture", "bytea", withBasic(types.UnsafePointer)),
		newCol("image", "varchar", withComment(`Image {"view_url": "/goods/{id}/image/%s"}`)),
	)
	s.DB = newDB(s.brands, s.photos, s.colors, s.goods)

	return s
}

// decor is the decor of column name of goods
func (s *shop) decor(name string) *forms.ColumnDecor {
	return forms.NewColumnDecor(s.goods.FindColumn(name), nil)
}

func TestToTableView(t *testing.T) {
	s := newShop()
	replacer := strings.NewReplacer("{id}", "7")
	tests := []struct {
		name string
		col  string
		val  any
		want string
	}{
		{"nil", "name", nil, ""},
		{"empty bytea", "picture", []byte{}, ""},
		{"view_url of the comment", "image", "a.png", "/goods/7/image/a.png"},
		{"bytea links the blob", "picture", []byte{1, 2}, "/blob?col=picture"},
		{"bytea of not bytes doesn't panic", "picture", "x", "/blob?col=picture"},
		{"plain text", "name", "Hammer", "Hammer"},
		{"number", "price", 12.5, "12.5"},
		{"foreign key", "id_brands", int32(3),
			`<a href='/table/brands/form?html&id=3' target="_modal">Acme</a>`},
		{"unknown foreign value", "id_brands", int32(9),
			`<a href='/table/brands/form?html&id=9' target="_modal">9</a>`},
		{"foreign by the name id_<table>", "id_colors", int32(1),
			`<a href='/table/colors/form?html&id=1' target="_modal">red</a>`},
		{"foreign array", "tags", []int32{3, 4},
			`<a href='/table/brands/form?html&id=[3 4]' target="_modal">[Acme Bolt]</a>`},
		{"photos as img", "id_photos", int32(5),
			`<img src='/table/photos/blob?id=5' id='imgblob5' style='max-width:100%'/>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			assert.Equal(t, tt.want, ToTableView(ctx, s.DB, "/blob?col=%s", s.decor(tt.col), tt.val, replacer))
		})
	}

}

func TestForeignValue(t *testing.T) {
	t.Run("not a foreign column", func(t *testing.T) {
		s := newShop()
		res, table, isArray := ForeignValue(&fasthttp.RequestCtx{}, s.DB, s.decor("price"), 1.0)
		assert.Nil(t, res)
		assert.Nil(t, table)
		assert.False(t, isArray)
	})

	t.Run("name id_<table> without such table", func(t *testing.T) {
		s := newShop()
		delete(s.DB.Tables, "colors")
		res, table, _ := ForeignValue(&fasthttp.RequestCtx{}, s.DB, s.decor("id_colors"), int32(1))
		assert.Nil(t, res)
		assert.Nil(t, table)
	})

	t.Run("foreign key: value and options", func(t *testing.T) {
		s := newShop()
		col := s.decor("id_brands")
		res, table, _ := ForeignValue(&fasthttp.RequestCtx{}, s.DB, col, int32(3))
		assert.Equal(t, "Acme", res)
		assert.Equal(t, "brands", table.Name())
		assert.Equal(t, map[string]forms.SelectOption{
			"set null": {},
			"Acme":     {Value: "3"},
			"Bolt":     {Value: "4"},
		}, col.SelectOptions, "a not required column may be set to null")
		assert.Equal(t, []string{
			"SELECT name FROM brands WHERE id=$1",
			"SELECT id,name FROM brands order by name",
		}, s.brands.sql)
	})

	t.Run("nil value: only the options", func(t *testing.T) {
		s := newShop()
		col := s.decor("id_brands")
		res, table, _ := ForeignValue(&fasthttp.RequestCtx{}, s.DB, col, nil)
		assert.Nil(t, res)
		assert.Equal(t, "brands", table.Name())
		assert.NotContains(t, col.SelectOptions, "set null")
		assert.Len(t, col.SelectOptions, 2)
		assert.Len(t, s.brands.sql, 1, "no search of the value")
	})

	t.Run("title in the language of the request", func(t *testing.T) {
		s := newShop()
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue(ParamsLang.Name, "ua")
		_, _, _ = ForeignValue(ctx, s.DB, s.decor("id_brands"), int32(3))
		assert.Equal(t, "SELECT name_ua FROM brands WHERE id=$1", s.brands.sql[0])
	})

	t.Run("array column aggregates the titles", func(t *testing.T) {
		s := newShop()
		res, _, _ := ForeignValue(&fasthttp.RequestCtx{}, s.DB, s.decor("tags"), []int32{3, 4})
		assert.Equal(t, "[Acme Bolt]", res)
		assert.Equal(t, "SELECT array_agg(name) FROM brands WHERE id=ANY($1)", s.brands.sql[0])
	})
}

func TestIsByteaCol_IsArrayCol(t *testing.T) {
	tests := []struct {
		typ            string
		bytea, isArray bool
	}{
		{"bytea", true, false},
		{"_bytea", true, true},
		{"int4", false, false},
		{"_int4", false, true},
		{"varchar", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			col := newCol("c", tt.typ)
			assert.Equal(t, tt.bytea, IsByteaCol(col))
			assert.Equal(t, tt.isArray, IsArrayCol(col))
		})
	}
}

func TestGetNameOfTimestampColumn(t *testing.T) {
	assert.Equal(t, "updateat", GetNameOfTimestampColumn(newShop().goods))
	assert.Equal(t, "createat", GetNameOfTimestampColumn(newTable("t", "",
		newCol("updateat", "text"), // not a timestamp
		newCol("createat", "timestamp"),
	)))
	assert.Equal(t, "", GetNameOfTimestampColumn(newShop().brands))
}

func TestToColumnDecors(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	brackets := Translator(func(_ *fasthttp.RequestCtx, s string) string { return "[" + s + "]" })
	s := newShop()

	t.Run("label of the comment translated", func(t *testing.T) {
		d := ToColumnDecors(ctx, s.goods.FindColumn("name"), nil, brackets)
		assert.Equal(t, "[Name of goods]", d.Label)
	})
	t.Run("label of the name without a comment", func(t *testing.T) {
		d := ToColumnDecors(ctx, s.goods.FindColumn("price"), nil, brackets)
		assert.Equal(t, "[price]", d.Label)
	})
	t.Run("nil translator keeps the label", func(t *testing.T) {
		d := ToColumnDecors(ctx, s.goods.FindColumn("name"), nil, nil)
		assert.Equal(t, "Name of goods", d.Label)
	})
	t.Run("foreign array is a slice", func(t *testing.T) {
		assert.True(t, ToColumnDecors(ctx, s.goods.FindColumn("tags"), nil, nil).IsSlice)
		assert.False(t, ToColumnDecors(ctx, s.goods.FindColumn("id_brands"), nil, nil).IsSlice)
	})
	t.Run("enum gives options", func(t *testing.T) {
		col := newCol("status", "order_status", withEnum("new", "paid"))
		assert.Equal(t, map[string]forms.SelectOption{
			"new":  {Value: "new"},
			"paid": {Value: "paid"},
		}, ToColumnDecors(ctx, col, nil, nil).SelectOptions)
	})
}

func TestCreateColumnDecors(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	columns := newShop().brands.Columns()

	t.Run("allocates for nil", func(t *testing.T) {
		got := CreateColumnDecors(ctx, columns, nil, nil, nil)
		require.Len(t, got, len(columns))
		for i, d := range got {
			assert.Equal(t, columns[i].Name(), d.Name())
		}
	})
	t.Run("fills a given slice, e.g. after the first '#' column", func(t *testing.T) {
		all := make([]*forms.ColumnDecor, len(columns)+1)
		CreateColumnDecors(ctx, columns, nil, all[1:], nil)
		assert.Nil(t, all[0])
		assert.Equal(t, "name_ua", all[len(columns)].Name())
	})
	t.Run("an empty slice with capacity is allocated, not indexed", func(t *testing.T) {
		got := CreateColumnDecors(ctx, columns, nil, make([]*forms.ColumnDecor, 0, 8), nil)
		assert.Len(t, got, len(columns))
	})
}

func TestToStandardColumnValueType(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "<nil>"},
		{"string", "abc", "abc"},
		{"bool", true, "true"},
		{"float", 1.5, "1.5"},
		{"bytes", []byte("raw"), "raw"},
		{"int32", int32(7), "7"},
		{"int64s", []int64{1, 2}, "[1 2]"},
		{"strings", []string{"a", "b"}, "[a b]"},
		{"map", map[string]string{"k": "v"}, "map[k:v]"},
		{"any array", []any{int64(1), "two"}, `[1 "two"]`},
		{"null interval", pgtype.Interval{}, ""},
		{"valuer", pgtype.Text{String: "txt", Valid: true}, "txt"},
		{"null valuer", pgtype.Int4{}, "<nil>"},
		{"range", pgtype.Range[pgtype.Int4]{
			Lower: pgtype.Int4{Int32: 1, Valid: true}, Upper: pgtype.Int4{Int32: 5, Valid: true},
			LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true,
		}, "[ 1, 5 )"},
		{"open range", pgtype.Range[pgtype.Int4]{
			Lower: pgtype.Int4{Int32: 1, Valid: true}, Upper: pgtype.Int4{Int32: 5, Valid: true},
			LowerType: pgtype.Exclusive, UpperType: pgtype.Inclusive, Valid: true,
		}, "( 1, 5 ]"},
		{"unknown type", struct{ A int }{1}, "{1}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToStandardColumnValueType(tt.in))
		})
	}
}

func TestPrepareSQLArgs(t *testing.T) {
	s := newShop()
	ctx := &fasthttp.RequestCtx{}
	ctx.SetUserValue("name", "Hammer")
	ctx.SetUserValue("price", 12.5)
	ctx.SetUserValue("color", "red") // no such column
	buf := &bytes.Buffer{}
	bad := map[string]string{}

	args, names := PrepareSQLArgs(ctx, s.goods, buf, bad, "name", "price", "id_brands", "color")
	assert.Equal(t, []any{"Hammer", 12.5}, args, "no param id_brands, no column color")
	assert.Equal(t, []string{"name", "price"}, names)
	assert.Equal(t, " Hammer 12.5", buf.String())
	assert.Contains(t, bad, "color")
}

func TestPhotosURL(t *testing.T) {
	assert.Equal(t, "/table/photos?id=3", PhotosURL(3))
	assert.Equal(t, "/table/goods?id=3", PhotosURL(3, "goods"))
	assert.Equal(t, "/table/goods/blob?id=3", PhotosURL(3, "goods", "blob"))
	assert.Equal(t, `<img src='/table/photos/blob?id=3' id='imgblob3' style='max-width:100%'/>`, renderImg(3, "photos", "blob"))
	assert.Equal(t, `<img src='/table/photos?id=3' id='imgblob3' style='max-width:100%'/>`, renderImg(3))
}

func TestTranslator(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	var none Translator
	assert.Equal(t, "Name", none.of(ctx, "Name"), "nil keeps the text")
	upper := Translator(func(_ *fasthttp.RequestCtx, s string) string { return "[" + s + "]" })
	assert.Equal(t, "[Name]", upper.of(ctx, "Name"))
}

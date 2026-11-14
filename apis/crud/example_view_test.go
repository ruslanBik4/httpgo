/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud_test

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/apis/crud"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
)

// brands is a table with a title column in two languages.
var brands = dbEngine.NewTableString("brands", "Brands", []dbEngine.Column{
	dbEngine.NewStringColumn("code", "", true),
	dbEngine.NewStringColumn("name", "", true),
	dbEngine.NewStringColumn("name_ua", "", false),
}, nil, nil)

func ExampleToStandardColumnValueType() {
	fmt.Println(crud.ToStandardColumnValueType([]byte("raw text")))
	fmt.Println(crud.ToStandardColumnValueType([]any{int64(1), "two"}))
	fmt.Println(crud.ToStandardColumnValueType(pgtype.Text{String: "pgx value", Valid: true}))
	fmt.Println(crud.ToStandardColumnValueType(pgtype.Range[pgtype.Int4]{
		Lower:     pgtype.Int4{Int32: 1, Valid: true},
		Upper:     pgtype.Int4{Int32: 10, Valid: true},
		LowerType: pgtype.Inclusive,
		UpperType: pgtype.Exclusive,
		Valid:     true,
	}))
	// Output:
	// raw text
	// [1 "two"]
	// pgx value
	// [ 1, 10 )
}

// A column comment may carry "view_url": the table view shows the value
// through it, the params of the request replaced first.
func ExampleToTableView_viewURL() {
	col := forms.NewColumnDecor(
		dbEngine.NewStringColumn("image", `Image {"view_url": "<img src='/goods/{id}/%s'>"}`, false), nil)
	params := strings.NewReplacer("{id}", "42")

	fmt.Println(crud.ToTableView(&fasthttp.RequestCtx{}, nil, "", col, "front.png", params))
	fmt.Printf("%q\n", crud.ToTableView(&fasthttp.RequestCtx{}, nil, "", col, nil, params))
	// Output:
	// <img src='/goods/42/front.png'>
	// ""
}

func ExamplePhotosURL() {
	fmt.Println(crud.PhotosURL(7))
	fmt.Println(crud.PhotosURL(7, "goods"))
	fmt.Println(crud.PhotosURL(7, "goods", "picture"))
	// Output:
	// /table/photos?id=7
	// /table/goods?id=7
	// /table/goods/picture?id=7
}

// The Translate of a project translates the labels of the inputs; a nil
// Translator keeps them.
func ExampleTranslator() {
	var tr crud.Translator = func(ctx *fasthttp.RequestCtx, s string) string {
		if lang, _ := ctx.UserValue(crud.ParamsLang.Name).(string); lang == "ua" && s == "Name" {
			return "Назва"
		}
		return s
	}
	ctx := &fasthttp.RequestCtx{}
	ctx.SetUserValue(crud.ParamsLang.Name, "ua")
	fmt.Println(tr(ctx, "Name"))
	// Output:
	// Назва
}

func ExampleGetNameOfTitleColumn() {
	fmt.Println(crud.GetNameOfTitleColumn(brands, nil))
	fmt.Println(crud.GetNameOfTitleColumn(brands, "ua"))
	fmt.Println(crud.GetNameOfTitleColumn(brands, "de")) // no name_de
	// Output:
	// name
	// name_ua
	// name
}

// PrepareSQLArgs takes the params of the request that are present.
func ExamplePrepareSQLArgs() {
	ctx := &fasthttp.RequestCtx{}
	ctx.SetUserValue("name", "Acme")
	buf := &bytes.Buffer{}
	badParams := map[string]string{}

	args, columns := crud.PrepareSQLArgs(ctx, brands, buf, badParams, "code", "name")
	fmt.Println(args, columns, len(badParams))
	// Output:
	// [Acme] [name] 0
}

func ExampleBlobToURL() {
	fmt.Println(crud.BlobToURL("goods", "picture", 3))
	// Output:
	// /api/v1/blob/goods?id=3&name=picture
}

func ExampleTimeToString() {
	fmt.Println(crud.TimeToString(time.Date(2026, 11, 14, 10, 30, 0, 0, time.UTC)))
	// Output:
	// 2026-11-14
}

func ExampleUnknownArrayToStrings() {
	fmt.Printf("%q\n", crud.UnknownArrayToStrings([]any{1, "two", true}))
	// Output:
	// ["1" "\"two\"" "true"]
}

/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

// Converting DB values for the REST API views (browse tables, forms):
// formerly api/internal.go generated into every gozaurus project, now
// compiled once. Everything works on *dbEngine.DB - the generated db.Database
// of a project embeds it (pass DB.DB).

import (
	sql_types "database/sql/driver"
	"fmt"
	"go/types"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/dbEngine/dbEngine/psql"
	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/logs"

	"github.com/ruslanBik4/httpgo/views/templates/forms"
)

// Translator translates a text of the UI into the language of the request
// (the Translate of a project); nil keeps the texts as they are.
type Translator func(ctx *fasthttp.RequestCtx, name string) string

func (tr Translator) of(ctx *fasthttp.RequestCtx, name string) string {
	if tr == nil {
		return name
	}
	return tr(ctx, name)
}

// ToTableView renders the value val of column col for a table view: by the
// "view_url" of the column comment, as a link to a blob (url), as the title of
// the foreign row it points to, or as plain text.
func ToTableView(ctx *fasthttp.RequestCtx, DB *dbEngine.DB, url string, col *forms.ColumnDecor, val any, paramReplacer *strings.Replacer) string {
	if b, ok := val.([]byte); val == nil || (ok && len(b) == 0 && col.BasicType() == types.UnsafePointer) {
		return ""
	}

	if viewURL := col.ExtProperties.GetStringBytes("view_url"); len(viewURL) > 0 {
		return fmt.Sprintf(paramReplacer.Replace(gotools.BytesToString(viewURL)), ToStandardColumnValueType(val))
	}

	if IsByteaCol(col) {
		return fmt.Sprintf(url, col.Name())
	}

	if v, foreign, isArray := ForeignValue(ctx, DB, col, val); v != nil {
		switch foreignName := foreign.Name(); {
		case foreignName == "photos":
			return renderImg(val.(int32), foreignName, "blob")
		case isArray:
			return fmt.Sprintf(`<a href='/table/%s/form?html&id=[%s]' target="_modal">%s</a>`,
				foreignName, ToStandardColumnValueType(val), v)
		default:
			return fmt.Sprintf(`<a href='/table/%s/form?html&id=%d' target="_modal">%v</a>`,
				foreignName, val, v)
		}
	}

	return ToStandardColumnValueType(val)
}

// IsByteaCol reports a bytea column.
func IsByteaCol(col dbEngine.Column) bool {
	return strings.HasSuffix(col.Type(), "bytea")
}

// IsArrayCol reports an array column.
func IsArrayCol(col dbEngine.Column) bool {
	return strings.HasPrefix(col.Type(), "_")
}

// ForeignValue returns the value 'res' of the foreign table the column 'col'
// refers to for 'val' (when present): by the foreign key, otherwise by a name
// like id_<table> or name_<table>.
// WARNING: col MAY BE CHANGED - its SelectOptions get the key values of the table.
// res: foreign value, nil when val is nil; table: the foreign table;
// isArray: the foreign value is an array.
func ForeignValue(ctx *fasthttp.RequestCtx, DB *dbEngine.DB, col *forms.ColumnDecor,
	val any) (res any, table dbEngine.Table, isArray bool) {

	var (
		frgPrimary string
		ok         bool
	)
	if frCol := col.Foreign(); frCol != nil {
		table = frCol.ForeignCol.Table()
		frgPrimary = frCol.Column
	} else {
		if parts := strings.Split(col.Name(), "_"); len(parts) > 1 && (parts[0] == "id" || slices.Contains(namesTitleCols, parts[0])) {
			frgPrimary = parts[0]
			table, ok = DB.Tables[parts[1]]
		}
		if !ok {
			return
		}
	}

	name := GetNameOfTitleColumn(table, ctx.UserValue(ParamsLang.Name))
	selCol := name
	if IsArrayCol(col) {
		selCol = "array_agg(" + name + ")"
	}

	// search foreign value
	if val != nil {
		switch err := table.SelectOneAndScan(ctx,
			&res,
			dbEngine.ColumnsForSelect(selCol),
			dbEngine.WhereForSelect(frgPrimary),
			dbEngine.ArgsForSelect(val),
		); err {
		case nil:
			res = ToStandardColumnValueType(res)

		case pgx.ErrNoRows:
			logs.StatusLog("unknown value %+v for column %s", val, col.Name())
			res = ToStandardColumnValueType(val)
			return

		default:
			logs.ErrorLog(err, "%s=%+v", name, val)
			return
		}
	}

	col.SelectOptions = make(map[string]forms.SelectOption)
	if val != nil && !col.Required() {
		col.SelectOptions["set null"] = forms.SelectOption{}
	}

	err := table.SelectAndRunEach(ctx,
		func(values []any, columns []dbEngine.Column) error {
			col.SelectOptions[fmt.Sprintf("%v", values[1])] = forms.SelectOption{Value: fmt.Sprintf("%v", values[0])}
			return nil
		},
		dbEngine.ColumnsForSelect(frgPrimary, name),
		dbEngine.OrderBy(name),
	)
	if err != nil {
		logs.ErrorLog(err, "")
	}

	return
}

// names of the columns that may present a row
var namesTitleCols = []string{
	"name",
	"title",
	"label",
	"fullname",
	"marka",
	"brend",
	"desc",
	"description",
}

// names of the columns that may report the last time a row was changed
var namesTimeStampCols = []string{
	"updateat",
	"updateAt",
	"modifyat",
	"modifyAt",
	"modify",
	"createat",
	"createAt",
	"create",
	"startedAt",
	"startedat",
}

// GetNameOfTimestampColumn returns the column that reports the last time a row was changed.
func GetNameOfTimestampColumn(table dbEngine.Table) string {
	for _, name := range namesTimeStampCols {
		col := table.FindColumn(name)
		if col != nil && (col.Type() == "timestamp" || col.Type() == "timestamptz") {
			return name
		}
	}

	return ""
}

func renderImg(id int32, s ...string) string {
	name := "blob"
	if len(s) > 1 {
		name = s[1]
	}
	return fmt.Sprintf(`<img src='%s' id='img%s%v' style='max-width:100%%'/>`,
		PhotosURL(id, s...), name, id)
}

// PhotosURL is the URL of picture id: of photos, of table s[0], of column s[1] of table s[0].
func PhotosURL(id int32, s ...string) string {
	const pathPhoto = "/table/"

	switch len(s) {
	case 0:
		return fmt.Sprintf(pathPhoto+"photos?id=%d", id)
	case 1:
		return fmt.Sprintf(pathPhoto+"%s?id=%d", s[0], id)
	default:
		return fmt.Sprintf(pathPhoto+"%s/%s?id=%d", s[0], s[1], id)
	}
}

// CreateColumnDecors fills colDecors (allocated when shorter than columns) with
// the decors of columns, their texts translated by tr.
func CreateColumnDecors(ctx *fasthttp.RequestCtx, columns []dbEngine.Column, patternList dbEngine.Table,
	colDecors []*forms.ColumnDecor, tr Translator) []*forms.ColumnDecor {
	if len(colDecors) < len(columns) {
		colDecors = make([]*forms.ColumnDecor, len(columns))
	}
	for i, col := range columns {
		colDecors[i] = ToColumnDecors(ctx, col, patternList, tr)
	}

	return colDecors
}

// ToColumnDecors is the decor of the input of col, its label and placeholder translated by tr.
func ToColumnDecors(ctx *fasthttp.RequestCtx, col dbEngine.Column, patternList dbEngine.Table, tr Translator) *forms.ColumnDecor {
	colDecors := forms.NewColumnDecor(col, patternList)
	if colDecors.PlaceHolder > "" {
		colDecors.PlaceHolder = tr.of(ctx, colDecors.PlaceHolder)
	}

	if colDecors.Label > "" {
		colDecors.Label = tr.of(ctx, colDecors.Label)
	} else {
		colDecors.Label = tr.of(ctx, col.Name())
	}

	definedType := col.UserDefinedType()
	switch {
	case col.Foreign() != nil || strings.HasPrefix(col.Type(), "_"):
		colDecors.IsSlice = IsArrayCol(col)

	case definedType != nil:
		if len(definedType.Enumerates) > 0 {
			colDecors.SelectOptions = maps.Collect(func(yield func(string, forms.SelectOption) bool) {
				for _, name := range definedType.Enumerates {
					if !yield(name, forms.SelectOption{Value: name}) {
						return
					}
				}
			})
		}
	}

	return colDecors
}

// ToStandardColumnValueType renders a value read from the DB as text.
func ToStandardColumnValueType(value any) string {
	switch v := value.(type) {
	case nil, string, bool, float32, float64, int16, int32, int64, []int16, []int32, []int64,
		[]string, map[string]string, map[string]any, psql.Numeric:
		return fmt.Sprintf("%v", value)

	case []byte:
		return gotools.BytesToString(v)

	case []any:
		return fmt.Sprintf("%v", UnknownArrayToStrings(v))

	case pgtype.Interval:
		// v.Value() returns (nil, nil) for a NULL/invalid interval
		val, err := v.Value()
		if err != nil {
			return err.Error()
		}
		s, _ := val.(string)

		return s

	// pgx v5 Codec types (composite/range values) implement driver.Valuer;
	// an array of them ([]*T) falls through to default: fmt renders the elements
	case pgtype.RangeValuer:
		return rangeToString(v)

	case sql_types.Valuer:
		v1, err := v.Value()
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("%v", v1)

	default:
		logs.DebugLog("undecoded type:%T, use standart output", value)
		return fmt.Sprintf("%v", value)
	}
}

func rangeToString(src pgtype.RangeValuer) string {
	lower, upper := src.Bounds()
	l, u := src.BoundTypes()
	r := '['
	if l != pgtype.Inclusive {
		r = '('
	}
	d := ']'
	if u != pgtype.Inclusive {
		d = ')'
	}

	return fmt.Sprintf("%c %s, %s %c", r, ToStandardColumnValueType(lower), ToStandardColumnValueType(upper), d)
}

// PrepareSQLArgs collects the values of the request params named columns
// (when present) and their column names for an SQL statement on table.
// A param AddColumnAndValue rejects (unknown column, bad file) is left out:
// the reason is in badParams.
func PrepareSQLArgs(ctx *fasthttp.RequestCtx, table dbEngine.Table, buf io.Writer, badParams map[string]string,
	columns ...string) ([]any, []string) {
	args := make([]any, 0, len(columns))
	colNames := make([]string, 0, len(columns))
	for _, name := range columns {
		if v := ctx.UserValue(name); v != nil {
			n, a := AddColumnAndValue(name, table, v, buf, badParams)
			if n == "" {
				continue
			}
			args = append(args, a)
			colNames = append(colNames, n)
		}
	}
	return args, colNames
}

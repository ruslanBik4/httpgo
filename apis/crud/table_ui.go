/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

// What the table view needs besides its rows (see views/js/src/table-row.js):
//
//   - switches of boolean columns: a switch saves one cell through the
//     "update" route, the filter card changes many rows at once - the visible
//     ones ("set_bool") or the whole table ("set_bool_all", admin only);
//   - the short card of the parent row a foreign key points to ("card").
//
// This file holds the parts that only build SQL and HTML; the handlers that
// run them are in table_ui_routes.go.

import (
	"context"
	"fmt"
	"go/types"
	"html"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/ruslanBik4/dbEngine/dbEngine"

	"github.com/ruslanBik4/httpgo/views/templates/forms"
)

const (
	// maxBulkIDs limits the rows one "set_bool" request may change
	maxBulkIDs = 5000
	// maxCardFields is the count of the fields of a card, the title included
	maxCardFields = 8
	// maxCardValueLen is the length of a value of a card, in characters
	maxCardValueLen = 120
)

// adminOnlyProperty is the flag of a column comment, {"admin_only": true}:
// the column is shown only to administrators in the card of a row.
const adminOnlyProperty = "admin_only"

// HasBoolColumns reports whether the table has a boolean column.
func HasBoolColumns(table dbEngine.Table) bool {
	for _, col := range table.Columns() {
		if col.BasicType() == types.Bool {
			return true
		}
	}

	return false
}

// SinglePrimary returns the name of the only primary column of the table
// (a column named "id" counts when no one is marked), "" when there is none or
// there are several of them.
func SinglePrimary(table dbEngine.Table) string {
	name := ""
	for _, col := range table.Columns() {
		if !col.Primary() {
			continue
		}
		if name != "" {
			return ""
		}
		name = col.Name()
	}
	if name == "" && table.FindColumn("id") != nil {
		return "id"
	}

	return name
}

// BuildSetBoolSQL builds the query that sets the boolean column of the table.
//
// value is "true", "false" or "toggle" (a NULL becomes true). ids is the list
// of the values of the primary column pk, separated by commas: the rows to
// change. all=true changes every row, ids is not used then, and "toggle" is
// not allowed (nothing tells what the user saw).
// The query returns one number: the count of the changed rows.
func BuildSetBoolSQL(table dbEngine.Table, pk, column, value, ids string, all bool) (string, []any, error) {
	col := table.FindColumn(column)
	switch {
	case col == nil:
		return "", nil, fmt.Errorf("unknown column '%s'", column)
	case col.BasicType() != types.Bool:
		return "", nil, fmt.Errorf("column '%s' isn't boolean", column)
	case table.FindColumn(pk) == nil:
		return "", nil, fmt.Errorf("unknown primary column '%s'", pk)
	}

	quoted := pgx.Identifier{column}.Sanitize()
	var (
		set  string
		args []any
	)
	switch value {
	case "true", "false":
		set, args = quoted+" = $1", []any{value == "true"}
	case "toggle":
		if all {
			return "", nil, fmt.Errorf("'toggle' changes only chosen rows")
		}
		set = quoted + " = NOT COALESCE(" + quoted + ", false)"
	default:
		return "", nil, fmt.Errorf("value must be 'true' or 'false'")
	}

	query := "UPDATE " + pgx.Identifier(strings.Split(table.Name(), ".")).Sanitize() + " SET " + set
	if !all {
		list, err := splitIDs(ids)
		if err != nil {
			return "", nil, err
		}
		args = append(args, list)
		query += fmt.Sprintf(" WHERE %s::text = ANY($%d)", pgx.Identifier{pk}.Sanitize(), len(args))
	}

	return "WITH changed AS (" + query + " RETURNING 1) SELECT count(*) FROM changed", args, nil
}

// splitIDs splits the list of ids separated by commas; the same id once.
func splitIDs(ids string) ([]string, error) {
	list := make([]string, 0, strings.Count(ids, ",")+1)
	seen := make(map[string]struct{})
	for _, id := range strings.Split(ids, ",") {
		if id = strings.TrimSpace(id); id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		list = append(list, id)
	}

	switch {
	case len(list) == 0:
		return nil, fmt.Errorf("no rows chosen")
	case len(list) > maxBulkIDs:
		return nil, fmt.Errorf("too many rows: %d, the limit is %d", len(list), maxBulkIDs)
	}

	return list, nil
}

// CardField is a line of the card of a row.
type CardField struct {
	Label, Value string
}

// CardColumns picks the columns of the short card of a row of the table: the
// title column first, then the others without the primary key, the foreign
// keys (they are ids), the files and the columns for administrators only
// ({"admin_only": true} in the comment) when the user isn't one.
func CardColumns(table dbEngine.Table, title string, isAdmin bool) []*forms.ColumnDecor {
	res := make([]*forms.ColumnDecor, 0, maxCardFields)
	add := func(col dbEngine.Column) {
		decor := forms.NewColumnDecor(col, nil)
		// the label is the comment up to its JSON part: "Note {...}" -> "Note"
		if decor.Label = strings.TrimSpace(decor.Label); decor.Label == "" {
			decor.Label = col.Name()
		}
		res = append(res, decor)
	}

	if col := table.FindColumn(title); col != nil {
		add(col)
	}

	for _, col := range table.Columns() {
		if len(res) >= maxCardFields {
			break
		}

		switch {
		case col.Name() == title,
			col.Primary(), col.Name() == "id", col.AutoIncrement(),
			col.Foreign() != nil,
			col.BasicType() == types.UnsafePointer, IsByteaCol(col),
			strings.HasPrefix(col.Type(), "json"):
			continue
		}

		decor := forms.NewColumnDecor(col, nil)
		if decor.IsHidden || (!isAdmin && decor.ExtProperties.GetBool(adminOnlyProperty)) {
			continue
		}
		add(col)
	}

	return res
}

// ReadCard reads the card of the row of the table whose primary column pk is id:
// the title (the value of the title column) and the other fields. found is
// false when there is no such row.
func ReadCard(ctx context.Context, table dbEngine.Table, pk string, id any, lang any, isAdmin bool) (
	title string, fields []CardField, found bool, err error) {

	titleCol := GetNameOfTitleColumn(table, lang)
	columns := CardColumns(table, titleCol, isAdmin)
	names := make([]string, len(columns))
	for i, col := range columns {
		names[i] = col.Name()
	}
	if len(names) == 0 {
		return
	}

	err = table.SelectAndRunEach(ctx,
		func(values []any, _ []dbEngine.Column) error {
			found = true
			for i, col := range columns {
				if i >= len(values) {
					break
				}
				value := cardValue(values[i])
				if col.Name() == titleCol {
					title = value
					continue
				}
				if value != "" {
					fields = append(fields, CardField{Label: col.Label, Value: value})
				}
			}

			return nil
		},
		dbEngine.ColumnsForSelect(names...),
		dbEngine.WhereForSelect(pk),
		dbEngine.ArgsForSelect(id),
	)

	return
}

// cardValue is the text of a value for the card, shortened.
func cardValue(value any) string {
	if value == nil {
		return ""
	}

	s := strings.Join(strings.Fields(ToStandardColumnValueType(value)), " ")
	if utf8.RuneCountInString(s) > maxCardValueLen {
		s = string([]rune(s)[:maxCardValueLen]) + "…"
	}

	return s
}

// RenderCard is the HTML of the card: the title as a link that opens the form
// of the row, then the fields. Texts are escaped.
func RenderCard(tableName string, id any, title string, fields []CardField) string {
	var b strings.Builder
	b.WriteString(`<dl class="tbl-card">`)
	if title == "" {
		title = "…"
	}
	fmt.Fprintf(&b, `<dt class="tbl-card__title"></dt><dd class="tbl-card__title"><a href='/table/%s/form?html&id=%s' target="_modal">%s</a></dd>`,
		url.PathEscape(tableName), url.QueryEscape(fmt.Sprint(id)), html.EscapeString(title))
	for _, f := range fields {
		fmt.Fprintf(&b, `<dt>%s</dt><dd>%s</dd>`, html.EscapeString(f.Label), html.EscapeString(f.Value))
	}
	b.WriteString(`</dl>`)

	return b.String()
}

// ForeignCell is the HTML of a cell with a foreign key: its text is the name
// of the parent row. The script of the table (table-row.js) opens the list of
// the parents on a click, to choose another one, and the card of the parent -
// from data-card + the id - on hovering.
func ForeignCell(parent dbEngine.Table, column string, id any, name any) string {
	pk := SinglePrimary(parent)
	if pk == "" {
		pk = "id"
	}
	parentName := parent.Name()

	return fmt.Sprintf(`<span class="tbl-fk" tabindex="0" role="button" data-col="%s" data-fk-table="%s" data-fk-id="%v" data-card="/table/%s/card?%s=">%v</span>`,
		html.EscapeString(column), html.EscapeString(parentName), html.EscapeString(fmt.Sprint(id)),
		url.PathEscape(parentName), url.QueryEscape(pk), html.EscapeString(fmt.Sprint(name)))
}

// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbEngine

import (
	"fmt"
	"strings"
)

type SQLBuilder struct {
	Args          []interface{}
	columns       []string
	filter        []string
	posFilter     int
	Table         Table
	SelectColumns []Column
}

func (b SQLBuilder) InsertSql() (string, error) {
	if len(b.columns) != len(b.Args) {
		return "", NewErrWrongArgsLen(b.Table.Name(), b.columns, b.Args)
	}

	return "INSERT INTO " + b.Table.Name() + "(" + b.Select() + ") VALUES (" + b.values() + ")", nil
}

func (b SQLBuilder) UpdateSql() (string, error) {
	if len(b.columns)+len(b.filter) != len(b.Args) {
		return "", NewErrWrongArgsLen(b.Table.Name(), b.columns, b.Args)
	}

	return "UPDATE " + b.Table.Name() + b.Set() + b.Where(), nil
}

func (b SQLBuilder) SelectSql() (string, error) {
	if len(b.filter) != len(b.Args) {
		return "", NewErrWrongArgsLen(b.Table.Name(), b.filter, b.Args)
	}

	return "SELECT " + b.Select() + " FROM " + b.Table.Name() + b.Where(), nil
}

func (b *SQLBuilder) Select() string {
	if len(b.columns) == 0 {
		return "*"
	}

	return strings.Join(b.columns, ",")
}

func (b *SQLBuilder) Set() string {
	s, comma := " SET ", ""
	if len(b.columns) == 0 {
		for _, col := range b.Table.Columns() {
			b.posFilter++
			s += fmt.Sprintf(comma+" %s=$%d", col.Name(), b.posFilter)
			comma = ","
		}

		return s
	}

	for _, name := range b.columns {
		b.posFilter++
		s += fmt.Sprintf(comma+" %s=$%d", name, b.posFilter)
		comma = ","
	}

	return s
}

func (b *SQLBuilder) Where() string {

	where, comma := "", ""
	for _, name := range b.filter {
		b.posFilter++

		switch pre := name[0]; pre {
		case '>', '<', '$', '~', '^':

			name = name[1:]
			switch pre {
			case '$':
				where += fmt.Sprintf(comma+" %s ~ '.*' + $%d + '$' ", name, b.posFilter)
			case '^':
				where += fmt.Sprintf(comma+" %s ~ '^.*' + $%d + '.*' ", name, b.posFilter)
			default:
				where += fmt.Sprintf(comma+" %s %s $%d", name, pre, b.posFilter)
			}
		default:
			where += fmt.Sprintf(comma+" %s=$%d", name, b.posFilter)
		}
		comma = " AND "
	}

	if where > "" {
		return " WHERE " + where
	}

	return ""
}

func (b *SQLBuilder) values() string {
	s, comma := "", ""
	for _ = range b.Args {
		b.posFilter++
		s += fmt.Sprintf("%s$%d", comma, b.posFilter)
		comma = ","
	}

	return s
}

type BuildSqlOptions func(b *SQLBuilder) error

func ColumnsForSelect(columns ...string) BuildSqlOptions {
	return func(b *SQLBuilder) error {

		if b.Table != nil {
			b.SelectColumns = make([]Column, len(columns))
			for i, name := range columns {
				if col := b.Table.FindColumn(name); col == nil {
					return NewErrNotFoundColumn(b.Table.Name(), name)
				} else {
					b.SelectColumns[i] = col
				}
			}
		}

		b.columns = columns

		return nil
	}
}

func WhereForSelect(columns ...string) BuildSqlOptions {
	return func(b *SQLBuilder) error {

		b.filter = make([]string, len(columns))
		if b.Table != nil {
			for _, name := range columns {

				switch pre := name[0]; pre {
				case '>', '<', '$', '~', '^':
					name = name[1:]
				}

				if b.Table.FindColumn(name) == nil {
					return NewErrNotFoundColumn(b.Table.Name(), name)
				}

			}
		}

		b.filter = columns

		return nil
	}
}

func ArgsForSelect(args ...interface{}) BuildSqlOptions {
	return func(b *SQLBuilder) error {

		b.Args = args

		return nil
	}
}

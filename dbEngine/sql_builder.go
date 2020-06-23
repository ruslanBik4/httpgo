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
	Filter        []string
	Table         Table
	SelectColumns []Column
}

func (b SQLBuilder) InsertSql() string {
	return "INSERT INTO " + b.Table.Name() + "(" + b.Select() + ") VALUES " + b.values()
}

func (b SQLBuilder) SelectSql() string {
	return "SELECT " + b.Select() + " FROM " + b.Table.Name() + b.Where()
}

func (b SQLBuilder) Select() string {
	if len(b.columns) == 0 {
		return "*"
	}

	return strings.Join(b.columns, ",")
}

func (b SQLBuilder) values() string {
	s, comma := "(", ""
	for i := range b.Args {
		s += fmt.Sprintf("%s$%d", comma, i)
		comma = ","
	}

	return s
}

func (b SQLBuilder) Where() string {
	if len(b.Filter) == 0 {
		return ""
	}

	return " WHERE " + strings.Join(b.Filter, " AND ")
}

type BuildSqlOptions func(b SQLBuilder) error

func ColumnsForSelect(columns ...string) BuildSqlOptions {
	return func(b SQLBuilder) error {

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
	return func(b SQLBuilder) error {

		b.Filter = make([]string, len(columns))
		if b.Table != nil {
			for i, name := range columns {
				switch pre := name[0]; pre {
				case '>', '<', '$', '~', '^':

					name = name[1:]
					if b.Table.FindColumn(name) == nil {
						return NewErrNotFoundColumn(b.Table.Name(), name)
					}

					switch pre {
					case '$':
						b.Filter[i] = fmt.Sprintf(" %s ~ '.*' + $%d + '$' ", name, i)
					case '^':
						b.Filter[i] = fmt.Sprintf(" %s ~ '^.*' + $%d + '.*' ", name, i)
					default:
						b.Filter[i] = fmt.Sprintf(" %s %s $%d", name, pre, i)
					}
				default:

					if b.Table.FindColumn(name) == nil {
						return NewErrNotFoundColumn(b.Table.Name(), name)
					}

					b.Filter[i] = fmt.Sprintf(" %s=$%d", name, i)

				}
			}
		}

		return nil
	}
}

func ArgsForSelect(args ...interface{}) BuildSqlOptions {
	return func(b SQLBuilder) error {

		if len(b.Filter) != len(args) {
			return NewErrWrongArgsLen(b.Table.Name(), b.Filter, args)
		}

		b.Args = args

		return nil
	}
}

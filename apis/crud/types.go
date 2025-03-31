/*
 * Copyright (c) 2022-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgtype"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/logs"
)

type NumRangeMarshal struct {
	*pgtype.Numrange
}

func (n *NumRangeMarshal) GetValue() any {
	return n.Numrange
}

func (n *NumRangeMarshal) NewValue() any {
	return &NumRangeMarshal{&pgtype.Numrange{}}
}

type DateRangeMarshal struct {
	*pgtype.Daterange
}

func (d *DateRangeMarshal) GetValue() any {
	return d.Daterange
}

func (d *DateRangeMarshal) NewValue() any {
	return &DateRangeMarshal{&pgtype.Daterange{}}
}

func (d *DateRangeMarshal) Value() (driver.Value, error) {
	return d.Daterange, nil
}

// Format implement Formatter interface
func (d *DateRangeMarshal) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		_, err := fmt.Fprintf(s, "%T", d)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 'g':
		_, err := fmt.Fprintf(s, "&%T{}", *d)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 's':
		_, err := fmt.Fprintf(s, "%v %v %v %v", d.LowerType, d.Lower, d.UpperType, d.UpperType)
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

func (d *DateRangeMarshal) Get() any {
	return d.GetValue()
}

func (d *DateRangeMarshal) Set(src any) error {
	if d.Daterange == nil {
		d.Daterange = &pgtype.Daterange{Status: pgtype.Null}
	}
	// untyped nil and typed nil interfaces are different
	if src == nil {
		d.Status = pgtype.Null
		return nil
	}

	switch src := src.(type) {
	case string:
		if src == "" {
			d.Status = pgtype.Undefined
			return nil
		}

		parts := strings.Split(src, ",")

		lower := strings.TrimSpace(parts[0])
		d.LowerType, lower = lowerBoundType(lower)

		err := d.Lower.Scan(lower)
		if err == nil {
			// if get one value set range to one date
			if len(parts) == 1 {
				d.Upper = d.Lower
				d.UpperType = pgtype.Inclusive
			} else {
				upper := strings.TrimSpace(parts[1])
				d.UpperType, upper = upperBoundType(upper)
				err = d.Upper.Scan(upper)
			}
		}
		if err != nil {
			logs.ErrorLog(err)
			return d.Daterange.Set(src)
		}

		d.Status = pgtype.Present

	case *pgtype.Daterange:
		d.Lower = src.Lower
		d.Upper = src.Upper
		d.LowerType = src.LowerType
		d.UpperType = src.UpperType
		d.Status = src.Status
		d.LowerType = pgtype.Inclusive
		d.UpperType = pgtype.Inclusive

	case pgtype.Daterange:
		return d.Set(&src)

	default:
		return d.Daterange.Set(src)
	}

	return nil
}

func lowerBoundType(lower string) (pgtype.BoundType, string) {
	if a, ok := strings.CutPrefix(lower, "["); ok {
		return pgtype.Inclusive, a
	} else if a, ok := strings.CutPrefix(lower, "("); ok {
		return pgtype.Exclusive, a
	}

	// inclusive border as default
	return pgtype.Inclusive, lower
}

func upperBoundType(upper string) (pgtype.BoundType, string) {
	if b, ok := strings.CutSuffix(upper, "]"); ok {
		return pgtype.Inclusive, b
	} else if b, ok := strings.CutSuffix(upper, ")"); ok {
		return pgtype.Exclusive, b
	}

	// inclusive border as default
	return pgtype.Inclusive, upper
}

func (d *DateRangeMarshal) GetPgxType() *pgtype.Daterange {
	return d.Daterange
}

type IntervalMarshal struct {
	*pgtype.Interval
}

func (i *IntervalMarshal) GetValue() any {
	return i
}

func (i *IntervalMarshal) NewValue() any {
	return &IntervalMarshal{&pgtype.Interval{}}
}

func (i *IntervalMarshal) Set(src any) error {
	switch src := src.(type) {
	case string:
		return i.Interval.DecodeText(nil, gotools.StringToBytes(src))
	default:
		return i.Interval.Scan(src)
	}
}

// Format implement Formatter interface
func (d *IntervalMarshal) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		_, err := fmt.Fprintf(s, "%T", d)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 'g':
		_, err := fmt.Fprintf(s, "&%T{}", *d)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 's':
		_, err := fmt.Fprintf(s, "%d month %d day %d", d.Months, d.Days, d.Microseconds)
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

type InetMarshal struct {
	*pgtype.Inet
}

func (i *InetMarshal) GetValue() any {
	return i
}

func (i *InetMarshal) NewValue() any {
	return &InetMarshal{&pgtype.Inet{}}
}

func (i *InetMarshal) Set(src any) error {
	switch src := src.(type) {
	case string:
		return i.DecodeText(nil, gotools.StringToBytes(src))
	default:
		return i.Scan(src)
	}
}

type NumrangeMarshal pgtype.Numrange

func (n *NumrangeMarshal) Expect() string {
	return "float"
}

func (n *NumrangeMarshal) FormatDoc() string {
	return "num_range"
}

func (n *NumrangeMarshal) RequestType() string {
	return "float"
}

func (n *NumrangeMarshal) GetValue() any {
	return pgtype.Numrange(*n)
}

func (n *NumrangeMarshal) NewValue() any {
	v := NumrangeMarshal(pgtype.Numrange{})
	return &v
}

func (n *NumrangeMarshal) Set(src any) error {
	v := pgtype.Numrange(*n)
	switch src := src.(type) {
	case string:
		return v.DecodeText(nil, gotools.StringToBytes(src))
	default:
		return v.Scan(src)
	}
}

// Format implement Formatter interface
func (n *NumrangeMarshal) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		_, err := fmt.Fprintf(s, "%T", n)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 'g':
		_, err := fmt.Fprintf(s, "&%T{}", *n)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 's':
		_, err := fmt.Fprintf(s, "%v %v %v %v", n.LowerType, n.Lower, n.UpperType, n.UpperType)
		if err != nil {
			logs.ErrorLog(err)
		}
	}
}

func (d *NumrangeMarshal) GetPgxType() pgtype.Numrange {
	return pgtype.Numrange(*d)
}

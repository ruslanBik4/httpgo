/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"database/sql/driver"
	"strings"

	"github.com/jackc/pgtype"

	"github.com/ruslanBik4/gotools"
)

type DateRangeMarshal struct {
	*pgtype.Daterange
}

func (d *DateRangeMarshal) GetValue() any {
	return d
}

func (d *DateRangeMarshal) NewValue() any {
	return &DateRangeMarshal{&pgtype.Daterange{}}
}

func (d *DateRangeMarshal) Value() (driver.Value, error) {
	return d.Daterange, nil
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

		err := d.Lower.Scan(strings.TrimSpace(parts[0]))
		if err == nil {
			err = d.Upper.Scan(strings.TrimSpace(parts[1]))
		}
		if err != nil {
			return d.Daterange.Set(src)
		}

		d.Status = pgtype.Present
		d.LowerType = pgtype.Inclusive
		d.UpperType = pgtype.Inclusive

	case pgtype.Daterange:
		d.Lower = src.Lower
		d.Upper = src.Upper
		// 	LowerType: src.LowerType,
		// 	UpperType: src.UpperType,
		// 	Status:    src.Status,
		// }
		d.Status = pgtype.Present
		d.LowerType = pgtype.Inclusive
		d.UpperType = pgtype.Inclusive

	case *pgtype.Daterange:
		return d.Set(*src)

	default:
		return d.Daterange.Set(src)
	}

	return nil
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

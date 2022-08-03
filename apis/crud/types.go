/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package crud

import (
	"database/sql/driver"
	"strings"

	"github.com/jackc/pgtype"
)

type DateMarshal struct {
	*pgtype.Daterange
}

func (d *DateMarshal) GetValue() interface{} {
	return d
}

func (d *DateMarshal) NewValue() interface{} {
	return &DateMarshal{&pgtype.Daterange{}}
}

func (d *DateMarshal) Value() (driver.Value, error) {
	return d.Daterange, nil
}

func (d *DateMarshal) Set(src interface{}) error {
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

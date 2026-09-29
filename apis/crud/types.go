/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"database/sql/driver"
	"fmt"
	"net/netip"
	"strings"
	"unsafe"

	"github.com/jackc/pgx/v5/pgtype"
	jsoniter "github.com/json-iterator/go"

	"github.com/ruslanBik4/logs"
)

type NumRangeMarshal struct {
	*pgtype.Range[pgtype.Numeric]
}

func (n *NumRangeMarshal) GetValue() any {
	return n.Range
}

func (n *NumRangeMarshal) NewValue() any {
	return &NumRangeMarshal{&pgtype.Range[pgtype.Numeric]{}}
}

type DateRangeMarshal struct {
	*pgtype.Range[pgtype.Date]
}

func NewDateRangeMarshal() *DateRangeMarshal {
	return &DateRangeMarshal{&pgtype.Range[pgtype.Date]{}}
}

func (d *DateRangeMarshal) Expect() string {
	return "string"
}

func (d *DateRangeMarshal) FormatDoc() string {
	return "date-range"
}

func (d *DateRangeMarshal) RequestType() string {
	return "string"
}

func (d *DateRangeMarshal) GetValue() any {
	return d.Range
}

func (d *DateRangeMarshal) NewValue() any {
	return NewDateRangeMarshal()
}

func (d *DateRangeMarshal) Value() (driver.Value, error) {
	return d.Range, nil
}

// Format implement Formatter interface
func (d *DateRangeMarshal) Format(s fmt.State, verb rune) {
	var err error
	switch verb {
	case 't':
		_, err = fmt.Fprintf(s, "%T", d)
	case 'g':
		_, err = fmt.Fprintf(s, "&%T{}", *d)
	case 's':
		_, err = fmt.Fprintf(s, "%v %v %v %v", d.LowerType, d.Lower, d.UpperType, d.UpperType)
	default:
		_, err = fmt.Fprintf(s, "%v %v %v %v", d.LowerType, d.Lower, d.UpperType, d.UpperType)

	}
	if err != nil {
		logs.ErrorLog(err)
	}
}

func (d *DateRangeMarshal) Get() any {
	return d.GetValue()
}
func (d *DateRangeMarshal) UnmarshalJSON(src []byte) error {
	sc := new(pgtype.RangeCodec{ElementType: &pgtype.Type{
		Codec: pgtype.DateCodec{},
		Name:  "date",
		OID:   pgtype.DateOID,
	}}).PlanScan(pgtype.NewMap(), pgtype.DateOID, pgtype.TextFormatCode, d)
	logs.StatusLog(sc, d)
	err := sc.Scan(src, d)
	logs.StatusLog("d = %s '%s'", d, src, err)
	return err
}
func DecodeDateRangeMarshal(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	val := (*DateRangeMarshal)(ptr)
	switch t := iter.WhatIsNext(); t {
	case jsoniter.ArrayValue:
		v := val.Lower
		iter.Read()
		iter.ReadArrayCB(func(iter *jsoniter.Iterator) bool {
			err := v.Scan(iter.ReadObject())
			if err != nil {
				logs.ErrorLog(err)
				return false
			}
			v = val.Upper
			return true
		})
		val.LowerType = pgtype.Inclusive
		val.UpperType = val.LowerType
		val.Valid = true

	case jsoniter.StringValue:
		src := iter.ReadString()

		if src == "" {
			val.Valid = false
			return
		}

		parts := strings.Split(src, ",")

		lower := strings.TrimSpace(parts[0])
		val.LowerType, lower = lowerBoundType(lower)

		err := val.Lower.Scan(lower)
		if err == nil {
			// if get one value set range to one date
			if len(parts) == 1 {
				val.Upper = val.Lower
				val.UpperType = pgtype.Inclusive
			} else {
				upper := strings.TrimSpace(parts[1])
				val.UpperType, upper = upperBoundType(upper)
				err = val.Upper.Scan(upper)
			}
		}
		if err != nil {
			logs.ErrorLog(err)
			return
		}

		val.Valid = true

	default:
		logs.ErrorLog(fmt.Errorf("unknown type"), t)
	}
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

func (d *DateRangeMarshal) GetPgxType() pgtype.Range[pgtype.Date] {
	return *d.Range
}

type IntervalMarshal struct {
	*pgtype.Interval
}

func NewIntervalMarshal() *IntervalMarshal {
	return &IntervalMarshal{&pgtype.Interval{}}
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
		return i.Interval.Scan(src)
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
	*netip.Addr
}

func NewInetMarshal() *InetMarshal {
	return &InetMarshal{&netip.Addr{}}
}

func (i *InetMarshal) GetValue() any {
	return i
}

func (i *InetMarshal) NewValue() any {
	return &InetMarshal{&netip.Addr{}}
}

//func (i *InetMarshal) Set(src any) error {
//	switch src := src.(type) {
//	case string:
//		return i.DecodeText(nil, gotools.StringToBytes(src))
//	default:
//		return i.Scan(src)
//	}
//}

type NumrangeMarshal pgtype.Range[pgtype.Numeric]

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
	return pgtype.Range[pgtype.Numeric](*n)
}

func (n *NumrangeMarshal) NewValue() any {
	return new(NumrangeMarshal(pgtype.Range[pgtype.Numeric]{}))
}

//func (n *NumrangeMarshal) Set(src any) error {
//	v := pgtype.Range[pgtype.Numeric](*n)
//	switch src := src.(type) {
//	case string:
//		return v.DecodeText(nil, gotools.StringToBytes(src))
//	default:
//		return v.Scan(src)
//	}
//}

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

func (d *NumrangeMarshal) GetPgxType() pgtype.Range[pgtype.Numeric] {
	return pgtype.Range[pgtype.Numeric](*d)
}

func init() {
	jsoniter.RegisterTypeDecoderFunc("crud.DateRangeMarshal", DecodeDateRangeMarshal)
	//jsoniter.RegisterTypeEncoderFunc("crud.DateRangeMarshal", EncodeDateString, IsEmptyDateString)
}

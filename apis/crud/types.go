/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/netip"
	"strings"
	"unsafe"

	"github.com/jackc/pgx/v5/pgtype"
	jsoniter "github.com/json-iterator/go"

	"github.com/ruslanBik4/gotools"
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
		// NewDateRangeMarshal(), not "&crud.DateRangeMarshal{}" - the zero-value
		// literal leaves the embedded *pgtype.Range nil, which panics the moment
		// UnmarshalJSON (or GetPgxType) touches it - same as TzString.Format.
		_, err = fmt.Fprint(s, "crud.NewDateRangeMarshal()")
	case 's':
		_, err = fmt.Fprintf(s, "%s %v %v %s", d.LowerType, d.Lower, d.Upper, d.UpperType)
	default:
		_, err = fmt.Fprintf(s, "%s %v %v %s", d.LowerType, d.Lower, d.Upper, d.UpperType)

	}
	if err != nil {
		logs.ErrorLog(err)
	}
}

func (d *DateRangeMarshal) Get() any {
	return d.GetValue()
}

// UnmarshalJSON scans a range literal in Postgres's own text syntax
// ("[2026-01-01,2026-02-01)", "empty") with pgx's RangeCodec; when that fails it
// tries the value as a single date, in any layout DateString.UnmarshalJSON
// accepts ("2026-01-01", "01.02.2026", Unix time, ...), and makes it the
// one-day range "[2026-01-01,2026-01-01]" - the same as DecodeDateRangeMarshal.
func (d *DateRangeMarshal) UnmarshalJSON(src []byte) error {
	// "&crud.DateRangeMarshal{}" would leave the embedded *pgtype.Range nil -
	// see Format's 'g' case; guarded here too for a value built any other way.
	if d.Range == nil {
		d.Range = &pgtype.Range[pgtype.Date]{}
	}

	src = bytes.Trim(src, `" `)
	sc := new(pgtype.RangeCodec{ElementType: &pgtype.Type{
		Codec: pgtype.DateCodec{},
		Name:  "date",
		OID:   pgtype.DateOID,
	}}).PlanScan(pgtype.NewMap(), pgtype.DateOID, pgtype.TextFormatCode, d)
	err := sc.Scan(src, d)
	if err == nil {
		return nil
	}

	t, dateErr := parseDate(gotools.BytesToString(src))
	if dateErr != nil {
		return fmt.Errorf("not a range (%w) nor a date (%w)", err, dateErr)
	}

	date := pgtype.Date{Time: t, Valid: true}
	*d.Range = pgtype.Range[pgtype.Date]{
		Lower:     date,
		Upper:     date,
		LowerType: pgtype.Inclusive,
		UpperType: pgtype.Inclusive,
		Valid:     true,
	}
	return nil
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

func (i *IntervalMarshal) GetPgxType() pgtype.Interval {
	return *i.Interval
}

func (i *IntervalMarshal) Set(src any) error {
	switch src := src.(type) {
	case string:
		return i.Interval.Scan(src)
	default:
		return i.Interval.Scan(src)
	}
}

// UnmarshalJSON: pgtype.Interval has no UnmarshalJSON of its own - unwrap the
// JSON string and hand it to Set, which already knows how to Scan a
// Postgres interval literal into the embedded value.
func (i *IntervalMarshal) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	return i.Set(s)
}

// Format implement Formatter interface. 'g' uses crud.NewIntervalMarshal(),
// not the zero-value literal "&crud.IntervalMarshal{}" - the latter leaves
// the embedded *pgtype.Interval nil, which panics the moment UnmarshalJSON
// (Set -> i.Interval.Scan) dereferences it - same reasoning as
// TzString/TimestampString/InetMarshal's own Format doc comments.
func (d *IntervalMarshal) Format(s fmt.State, verb rune) {
	var err error
	switch verb {
	case 't':
		_, err = fmt.Fprintf(s, "%T", d)
	case 'g':
		_, err = fmt.Fprint(s, "crud.NewIntervalMarshal()")
	case 's':
		_, err = fmt.Fprintf(s, "%d month %d day %d", d.Months, d.Days, d.Microseconds)
	}
	if err != nil {
		logs.ErrorLog(err)
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

func (i *InetMarshal) GetPgxType() netip.Addr {
	return *i.Addr
}

// UnmarshalJSON: netip.Addr has no UnmarshalJSON of its own (only
// UnmarshalText), so unwrap the JSON string first and parse it directly -
// covers both "inet" and "cidr" (inet/cidr share InetMarshal, see setType).
func (i *InetMarshal) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	addr, err := netip.ParseAddr(s)
	if err != nil {
		return err
	}

	*i.Addr = addr
	return nil
}

// Format implement Formatter interface. 'g' uses crud.NewInetMarshal(), not
// the zero-value literal "&crud.InetMarshal{}" - the latter leaves the
// embedded *netip.Addr nil, which panics the moment UnmarshalJSON above (or
// anything else) dereferences i.Addr - same reasoning as TzString/
// TimestampString's own Format doc comments.
func (i *InetMarshal) Format(s fmt.State, verb rune) {
	var err error
	switch verb {
	case 't':
		_, err = fmt.Fprintf(s, "%T", i)
	case 'g':
		_, err = fmt.Fprint(s, "crud.NewInetMarshal()")
	case 's':
		if i.Addr != nil && i.Addr.IsValid() {
			_, err = fmt.Fprint(s, i.Addr.String())
		}
	}
	if err != nil {
		logs.ErrorLog(err)
	}
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

// UnmarshalJSON reuses parseRangeLiteral (field_dto.go) - the same helper
// Int4RangeMarshal/Int8RangeMarshal/TsRangeMarshal/TsTzRangeMarshal already
// use - with a bound parser that hands the bound's text to pgtype.Numeric's
// own Scan, which already parses arbitrary-precision decimal text without
// going through float64 (see NumericString's doc comment for why that
// matters).
func (n *NumrangeMarshal) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	rng, err := parseRangeLiteral(s, func(b string) (pgtype.Numeric, error) {
		var num pgtype.Numeric
		err := num.Scan(b)
		return num, err
	})
	if err != nil {
		return err
	}

	*n = NumrangeMarshal(rng)
	return nil
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
	var err error
	switch verb {
	case 't':
		_, err = fmt.Fprintf(s, "%T", n)
	case 'g':
		_, err = fmt.Fprintf(s, "&%T{}", *n)
	case 's':
		_, err = fmt.Fprintf(s, "%v %v %v %v", n.LowerType, n.Lower, n.UpperType, n.UpperType)
	}
	if err != nil {
		logs.ErrorLog(err)
	}
}

func (d *NumrangeMarshal) GetPgxType() pgtype.Range[pgtype.Numeric] {
	return pgtype.Range[pgtype.Numeric](*d)
}

func init() {
	jsoniter.RegisterTypeDecoderFunc("crud.DateRangeMarshal", DecodeDateRangeMarshal)
	//jsoniter.RegisterTypeEncoderFunc("crud.DateRangeMarshal", EncodeDateString, IsEmptyDateString)
}

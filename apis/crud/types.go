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
	"errors"
	"fmt"
	"math"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/jackc/pgx/v5/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fastjson"

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
	default: // 's', 'v'
		// a zero-value DateRangeMarshal has a nil embedded *pgtype.Range
		if d.Range != nil {
			_, err = fmt.Fprintf(s, "%s %v %v %s", d.LowerType, d.Lower, d.Upper, d.UpperType)
		}
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
	// the codec has no state of its own and the type map comes from a pool -
	// a new codec and pgtype.NewMap() for every value were ten allocations
	m := dateRangeMaps.Get().(*pgtype.Map)
	err := dateRangeCodec.PlanScan(m, pgtype.DaterangeOID, pgtype.TextFormatCode, d).Scan(src, d)
	dateRangeMaps.Put(m)
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

var (
	dateRangeCodec = &pgtype.RangeCodec{ElementType: &pgtype.Type{
		Codec: pgtype.DateCodec{},
		Name:  "date",
		OID:   pgtype.DateOID,
	}}
	// a *pgtype.Map is not documented as safe for concurrent use, so each
	// UnmarshalJSON borrows its own
	dateRangeMaps = sync.Pool{New: func() any { return pgtype.NewMap() }}
)

// DecodeDateRangeMarshal is the jsoniter decoder of DateRangeMarshal (see
// init): jsoniter calls it instead of UnmarshalJSON. It reads
//   - a string: "2026-01-01,2026-02-01", "[2026-01-01,2026-02-01)", or one
//     date (a one-day range);
//   - an array of one or two dates: ["2026-01-01","2026-02-01"], both
//     bounds inclusive;
//   - null.
func DecodeDateRangeMarshal(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	val := (*DateRangeMarshal)(ptr)
	// jsoniter hands over a zero-value DateRangeMarshal for a struct field:
	// its embedded *pgtype.Range is nil
	if val.Range == nil {
		val.Range = &pgtype.Range[pgtype.Date]{}
	}

	switch t := iter.WhatIsNext(); t {
	case jsoniter.NilValue:
		iter.ReadNil()
		*val.Range = pgtype.Range[pgtype.Date]{}

	case jsoniter.ArrayValue:
		var bounds [2]pgtype.Date
		n := 0
		iter.ReadArrayCB(func(iter *jsoniter.Iterator) bool {
			s := iter.ReadString()
			if iter.Error != nil {
				return false
			}
			if n == len(bounds) {
				iter.ReportError("DecodeDateRangeMarshal", "a date range has at most two bounds")
				return false
			}
			if err := bounds[n].Scan(strings.TrimSpace(s)); err != nil {
				iter.ReportError("DecodeDateRangeMarshal", err.Error())
				return false
			}
			n++
			return true
		})
		if iter.Error != nil || n == 0 {
			val.Valid = false
			return
		}
		if n == 1 {
			bounds[1] = bounds[0]
		}
		*val.Range = pgtype.Range[pgtype.Date]{
			Lower:     bounds[0],
			Upper:     bounds[1],
			LowerType: pgtype.Inclusive,
			UpperType: pgtype.Inclusive,
			Valid:     true,
		}

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
		iter.Skip() // leave the iterator after the value
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
	if d == nil || d.Range == nil {
		return pgtype.Range[pgtype.Date]{} // NULL
	}
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
	return NewIntervalMarshal()
}

// GetPgxType is NULL for a zero-value IntervalMarshal ("&crud.IntervalMarshal{}",
// a struct field nothing was read into): its embedded *pgtype.Interval is nil.
func (i *IntervalMarshal) GetPgxType() pgtype.Interval {
	if i == nil || i.Interval == nil {
		return pgtype.Interval{}
	}
	return *i.Interval
}

// IntervalValue implements pgtype.IntervalValuer, Value implements
// driver.Valuer - what pgx calls when an IntervalMarshal is a query argument.
// Both exist on the embedded *pgtype.Interval already, but promoted they
// panic on the nil pointer of a zero value.
func (i IntervalMarshal) IntervalValue() (pgtype.Interval, error) {
	return i.GetPgxType(), nil
}

func (i IntervalMarshal) Value() (driver.Value, error) {
	return i.GetPgxType().Value()
}

// ScanInterval implements pgtype.IntervalScanner (pgx reading a column into
// an IntervalMarshal) - for the same reason: the promoted one writes through
// the nil pointer.
func (i *IntervalMarshal) ScanInterval(v pgtype.Interval) error {
	i.set(v)
	return nil
}

// Set is Scan under its old name.
func (i *IntervalMarshal) Set(src any) error {
	return i.Scan(src)
}

// Scan implements database/sql.Scanner. It reads the text
// pgtype.Interval.Scan reads,
//
//	[N year[s]] [N mon[s]] [N day[s]] [[-]H:M:S[.ffffff]]
//
// ("1 day 02:03:04", "-01:30:00", "2 years 3 mons") and three things that one
// does not: a clock without seconds ("01:30" - <input type="time"> drops zero
// seconds), spaces around the value, and "" as NULL. A []byte is read too -
// database/sql hands text over that way.
//
// pgtype.Interval.Scan copies the string to a []byte, that back to a string,
// then splits it into a slice per level; this one parses the text where it
// lies and allocates nothing for a value that parses (an error does allocate
// its message). The destination is left as it was when src does not parse.
func (i *IntervalMarshal) Scan(src any) error {
	switch s := src.(type) {
	case nil:
		i.set(pgtype.Interval{})
		return nil
	case string:
		return i.scanText(s)
	case []byte:
		return i.scanText(gotools.BytesToString(s)) // not kept after the call
	default:
		// reflect.TypeOf, not the %T verb: handing src itself to fmt makes it
		// escape, and then every caller pays an allocation to box its argument
		return fmt.Errorf("cannot scan %s into an interval", reflect.TypeOf(src))
	}
}

func (i *IntervalMarshal) scanText(s string) error {
	iv, err := parseInterval(s)
	if err != nil {
		return err
	}

	i.set(iv)
	return nil
}

// set stores iv; only a zero-value IntervalMarshal (nil embedded pointer)
// costs an allocation here.
func (i *IntervalMarshal) set(iv pgtype.Interval) {
	if i.Interval == nil {
		i.Interval = new(pgtype.Interval)
	}
	*i.Interval = iv
}

const (
	usPerSecond = 1_000_000
	usPerMinute = 60 * usPerSecond
	usPerHour   = 60 * usPerMinute
)

var errBadInterval = errors.New("bad interval format")

// parseInterval is the parser behind Scan. "" (or only spaces) is NULL.
func parseInterval(src string) (pgtype.Interval, error) {
	s := strings.TrimSpace(src)
	if s == "" {
		return pgtype.Interval{}, nil
	}

	var months, days, micro int64
	for s != "" {
		tok, rest := cutToken(s)
		if rest == "" {
			// the last word without a unit after it is the clock
			var ok bool
			if micro, ok = parseClock(tok); !ok {
				return pgtype.Interval{}, fmt.Errorf("%w: %q", errBadInterval, src)
			}
			break
		}

		n, ok := parseInt(tok)
		// the limit also keeps n*12 and the sums below inside an int64
		if !ok || n < math.MinInt32 || n > math.MaxInt32 {
			return pgtype.Interval{}, fmt.Errorf("%w: %q", errBadInterval, src)
		}

		var unit string
		unit, s = cutToken(rest)
		switch unit {
		case "year", "years":
			months += n * 12
		case "mon", "mons":
			months += n
		case "day", "days":
			days = n
		default:
			return pgtype.Interval{}, fmt.Errorf("%w: unit %q in %q", errBadInterval, unit, src)
		}
	}

	if months < math.MinInt32 || months > math.MaxInt32 {
		return pgtype.Interval{}, fmt.Errorf("%w: %q is out of range", errBadInterval, src)
	}

	return pgtype.Interval{Months: int32(months), Days: int32(days), Microseconds: micro, Valid: true}, nil
}

// cutToken returns the first word of s and what follows the spaces after it.
func cutToken(s string) (tok, rest string) {
	n := strings.IndexByte(s, ' ')
	if n < 0 {
		return s, ""
	}

	return s[:n], strings.TrimLeft(s[n:], " ")
}

// parseClock reads "[+-]H:M[:S[.ffffff]]" as microseconds. Like pgx it does
// not hold minutes and seconds under 60: "00:90:00" is an hour and a half.
// Digits after the sixth of the fraction are dropped - an interval has no
// place for them.
func parseClock(s string) (int64, bool) {
	negative := false
	if s != "" && (s[0] == '-' || s[0] == '+') {
		negative = s[0] == '-'
		s = s[1:]
	}

	h, rest, ok := strings.Cut(s, ":")
	if !ok {
		return 0, false
	}
	m, sec, withSeconds := strings.Cut(rest, ":")

	// hours go as far as the microseconds of a pgtype.Interval do
	hours, ok := parseClockField(h, math.MaxInt64/usPerHour)
	if !ok {
		return 0, false
	}
	minutes, ok := parseClockField(m, math.MaxInt32)
	if !ok {
		return 0, false
	}

	var seconds, fraction int64
	if withSeconds {
		whole, frac, withFraction := strings.Cut(sec, ".")
		if seconds, ok = parseClockField(whole, math.MaxInt32); !ok {
			return 0, false
		}

		if withFraction {
			if frac == "" {
				return 0, false
			}
			for n := 0; n < len(frac); n++ {
				if frac[n] < '0' || frac[n] > '9' {
					return 0, false
				}
				if n < 6 {
					fraction = fraction*10 + int64(frac[n]-'0')
				}
			}
			for n := len(frac); n < 6; n++ {
				fraction *= 10
			}
		}
	}

	micro := hours*usPerHour + minutes*usPerMinute + seconds*usPerSecond + fraction
	if micro < 0 { // past the last hour an int64 holds
		return 0, false
	}
	if negative {
		micro = -micro
	}

	return micro, true
}

// parseClockField reads the digits of an hour, minute or second, up to limit.
func parseClockField(s string, limit int64) (int64, bool) {
	if s == "" {
		return 0, false
	}

	var n int64
	for k := 0; k < len(s); k++ {
		c := s[k]
		if c < '0' || c > '9' {
			return 0, false
		}
		if n = n*10 + int64(c-'0'); n > limit {
			return 0, false
		}
	}

	return n, true
}

// parseInt reads "[+-]digits". strconv.ParseInt would do, but it allocates
// its error, and a word that is not a number is an ordinary case here.
func parseInt(s string) (int64, bool) {
	negative := false
	if s != "" && (s[0] == '-' || s[0] == '+') {
		negative = s[0] == '-'
		s = s[1:]
	}
	if s == "" {
		return 0, false
	}

	var n int64
	for k := 0; k < len(s); k++ {
		c := s[k]
		if c < '0' || c > '9' {
			return 0, false
		}
		d := int64(c - '0')
		if n > (math.MaxInt64-d)/10 {
			return 0, false
		}
		n = n*10 + d
	}
	if negative {
		n = -n
	}

	return n, true
}

// appendIntervalText writes iv the way Postgres (and pgx's text format) does:
// "1 mon 2 day 03:04:05.5" - what parseInterval reads back.
func appendIntervalText(buf []byte, iv pgtype.Interval) []byte {
	if iv.Months != 0 {
		buf = strconv.AppendInt(buf, int64(iv.Months), 10)
		buf = append(buf, " mon "...)
	}
	if iv.Days != 0 {
		buf = strconv.AppendInt(buf, int64(iv.Days), 10)
		buf = append(buf, " day "...)
	}

	micro := uint64(iv.Microseconds)
	if iv.Microseconds < 0 {
		buf = append(buf, '-')
		micro = -micro // also right for math.MinInt64
	}

	buf = appendPadded(buf, micro/usPerHour, 2)
	buf = append(buf, ':')
	buf = appendPadded(buf, micro%usPerHour/usPerMinute, 2)
	buf = append(buf, ':')
	buf = appendPadded(buf, micro%usPerMinute/usPerSecond, 2)

	if fraction := micro % usPerSecond; fraction != 0 {
		buf = append(buf, '.')
		buf = appendPadded(buf, fraction, 6)
	}

	return buf
}

// appendPadded writes n with zeros in front up to width digits.
func appendPadded(buf []byte, n uint64, width int) []byte {
	for limit := uint64(10); width > 1; width, limit = width-1, limit*10 {
		if n < limit {
			buf = append(buf, '0')
		}
	}

	return strconv.AppendUint(buf, n, 10)
}

// intervalParsers: UnmarshalJSON reads the object form with a pooled parser -
// fastjson.ParseBytes builds a new one (and its buffers) for every call.
var intervalParsers fastjson.ParserPool

// UnmarshalJSON accepts every shape an interval arrives in:
//   - the bare text of a form/query parameter, as TypeInParam.ReadValue
//     passes it: 01:30:00, 1 day 02:03:04, 01:30   -> parseInterval
//   - a JSON string with the same text: "01:30:00"  -> parseInterval
//   - an object {"Months":..,"Days":..,"Microseconds":..,"Valid":..}
//     (what pgtype.Interval itself marshals to)      -> read as the struct
//   - null, "" and a blank string                    -> NULL
//
// An empty parameter and a bare 0 / 00 are the zero interval, not NULL.
// A string and the object form allocate nothing.
func (i *IntervalMarshal) UnmarshalJSON(src []byte) error {
	src = bytes.TrimSpace(src)
	if len(src) > 0 && src[0] == '{' {
		return i.unmarshalObject(src)
	}

	switch string(src) { // compiled without a conversion
	case "", "0", "00":
		i.set(pgtype.Interval{Valid: true})
		return nil
	}

	s, null := paramText(src)
	if null {
		i.set(pgtype.Interval{})
		return nil
	}

	return i.scanText(s)
}

func (i *IntervalMarshal) unmarshalObject(src []byte) error {
	p := intervalParsers.Get()
	defer intervalParsers.Put(p)

	v, err := p.ParseBytes(src)
	if err != nil {
		return err
	}

	months, err := intervalField(v, "Months", "months", math.MinInt32, math.MaxInt32)
	if err != nil {
		return err
	}
	days, err := intervalField(v, "Days", "days", math.MinInt32, math.MaxInt32)
	if err != nil {
		return err
	}
	micro, err := intervalField(v, "Microseconds", "microseconds", math.MinInt64, math.MaxInt64)
	if err != nil {
		return err
	}

	valid := true // a missing "Valid" means a real value
	if f := jsonField(v, "Valid", "valid"); f != nil {
		if valid, err = f.Bool(); err != nil {
			return fmt.Errorf("interval field Valid: %w", err)
		}
	}

	i.set(pgtype.Interval{Months: int32(months), Days: int32(days), Microseconds: micro, Valid: valid})
	return nil
}

// jsonField is the field of v under name or its lower-case spelling (Go
// marshals a struct with capitalised keys); nil when it is absent or null.
func jsonField(v *fastjson.Value, name, lower string) *fastjson.Value {
	f := v.Get(name)
	if f == nil {
		f = v.Get(lower)
	}
	if f == nil || f.Type() == fastjson.TypeNull {
		return nil
	}

	return f
}

func intervalField(v *fastjson.Value, name, lower string, minVal, maxVal int64) (int64, error) {
	f := jsonField(v, name, lower)
	if f == nil {
		return 0, nil
	}

	n, err := f.Int64()
	if err != nil {
		return 0, fmt.Errorf("interval field %s: %w", name, err)
	}
	if n < minVal || n > maxVal {
		return 0, fmt.Errorf("interval field %s: %d is out of range", name, n)
	}

	return n, nil
}

// MarshalJSON writes the interval as the string UnmarshalJSON reads back:
// "00:00:00", "1 day 02:03:04"; NULL is null.
func (i IntervalMarshal) MarshalJSON() ([]byte, error) {
	if i.Interval == nil || !i.Valid {
		return []byte("null"), nil
	}

	buf := make([]byte, 0, 40)
	buf = append(buf, '"')
	buf = appendIntervalText(buf, *i.Interval)

	return append(buf, '"'), nil
}

// Format implement Formatter interface. 'g' uses crud.NewIntervalMarshal(),
// not the zero-value literal "&crud.IntervalMarshal{}" - the latter leaves
// the embedded *pgtype.Interval nil - same reasoning as
// TzString/TimestampString/InetMarshal's own Format doc comments.
func (i *IntervalMarshal) Format(s fmt.State, verb rune) {
	var err error
	switch verb {
	case 't':
		_, err = fmt.Fprintf(s, "%T", i)
	case 'g':
		_, err = fmt.Fprint(s, "crud.NewIntervalMarshal()")
	default: // 's', 'v'
		// nothing for a zero value: its embedded pointer is nil
		if i.Interval != nil {
			// Microseconds, not nanoseconds as a bare time.Duration would count them
			_, err = fmt.Fprintf(s, "%d month %d day %s", i.Months, i.Days,
				time.Duration(i.Microseconds)*time.Microsecond)
		}
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
	if i == nil || i.Addr == nil {
		return netip.Addr{}
	}
	return *i.Addr
}

// UnmarshalJSON reads an address either as a JSON string ("192.168.0.1", from
// a JSON body) or as the bare text of a form/query parameter (192.168.0.1,
// what TypeInParam.ReadValue passes) - see paramText. netip.Addr has no
// UnmarshalJSON of its own, only UnmarshalText.
//
// Only a host address fits: InetMarshal keeps a netip.Addr, so a network
// ("10.0.0.0/8", any cidr value and an inet with a mask) is rejected - there
// is nowhere to keep the prefix length.
func (i *InetMarshal) UnmarshalJSON(src []byte) error {
	s, _ := paramText(src)

	addr, err := netip.ParseAddr(s)
	if err != nil {
		return err
	}

	// "&crud.InetMarshal{}" leaves the embedded *netip.Addr nil
	if i.Addr == nil {
		i.Addr = new(netip.Addr)
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
	rng, err := unmarshalRange(src, func(b string) (pgtype.Numeric, error) {
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
	default: // 's', 'v': the range literal, "[1.5,3.5)"
		_, err = fmt.Fprint(s, formatRangeLiteral(convertRange(pgtype.Range[pgtype.Numeric](*n), numericText)))
	}
	if err != nil {
		logs.ErrorLog(err)
	}
}

func (d *NumrangeMarshal) GetPgxType() pgtype.Range[pgtype.Numeric] {
	return pgtype.Range[pgtype.Numeric](*d)
}

// numericText is a pgtype.Numeric as Postgres writes it: "1.5", "NaN".
func numericText(n pgtype.Numeric) string {
	v, err := n.Value()
	if err != nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func init() {
	jsoniter.RegisterTypeDecoderFunc("crud.DateRangeMarshal", DecodeDateRangeMarshal)
	//jsoniter.RegisterTypeEncoderFunc("crud.DateRangeMarshal", EncodeDateString, IsEmptyDateString)
}

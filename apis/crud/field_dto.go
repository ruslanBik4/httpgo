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
	"go/types"
	"io"
	"mime/multipart"
	"net"
	"slices"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/jackc/pgx/v5/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/logs"
)

type DTO[T any] struct {
	val T
}

func (d *DTO[T]) String() string {
	return fmt.Sprintf("&crud.DTO[%T]{}", d.val)
}

func (d *DTO[T]) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		_, err := fmt.Fprintf(s, "*crud.DTO[%T]", d.val)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 'g':
		_, err := fmt.Fprintf(s, "&crud.DTO[%T]{}", d.val)
		if err != nil {
			logs.ErrorLog(err)
		}

	case 's':
		_, err := fmt.Fprintf(s, "&crud.DTO[%T]", d.val)
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

func NewDTO[T any](val T) *DTO[T] {
	return &DTO[T]{val: val}
}

func (d *DTO[T]) GetPgxType() T {
	return d.val
}
func (d *DTO[T]) GetValue() any {
	return d.val
}

func (d *DTO[T]) NewValue() any {
	var a T
	return a
}
func (d *DTO[T]) Expect() string {
	return fmt.Sprintf("%T", d.val)
}
func (d *DTO[T]) FormatDoc() string {
	return fmt.Sprintf("%T", d.val)
}
func (d *DTO[T]) RequestType() string {
	return fmt.Sprintf("%T", d.val)
}

type DTOtype struct {
	Val string
}

func (D *DTOtype) GetValue() any {
	return D.Val
}

func (D *DTOtype) NewValue() any {
	return &DTOtype{Val: D.Val}
}
func (d *DTOtype) Format(s fmt.State, verb rune) {
	var err error
	switch verb {
	case 't', 'P':
		_, err = fmt.Fprintf(s, "%s", d.Val)
	//	for initial value
	case 'g':
		if strings.HasPrefix(d.Val, "*crud.DTO[") {
			_, err = fmt.Fprintf(s, "%s{}", strings.Replace(d.Val, "*", "&", 1))
		} else {
			_, err = fmt.Fprintf(s, "crud.NewDTO(%s{})", d.Val)
		}

	case 's':
		_, err = fmt.Fprintf(s, "%s{}", d.Val)
	default:
	}
	if err != nil {
		logs.ErrorLog(err)
	}
}

type DateTimeString time.Time

func (d *DateTimeString) Expect() string {
	return "date as string format #" + time.RFC3339
}

// Format implement Formatter interface
func (d *DateTimeString) Format(s fmt.State, verb rune) {
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
		_, err := fmt.Fprint(s, (time.Time)(*d).String())
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

func (d *DateTimeString) FormatDoc() string {
	return "date-time"
}

func (d *DateTimeString) RequestType() string {
	return "string"
}

func (d *DateTimeString) GetValue() any {
	return d
}

func (d *DateTimeString) NewValue() any {
	return &DateTimeString{}
}

func (d *DateTimeString) UnmarshalJSON(src []byte) error {
	t, err := bytesToTime(
		src,
		time.RFC3339,
		time.RFC3339Nano,
		time.DateTime,
		time.RFC1123,
		time.RFC1123Z,
		time.Stamp,
	)

	if err != nil {
		return err
	}

	*d = (DateTimeString)(t)
	return nil
}

func (d *DateTimeString) MarshalJSON() ([]byte, error) {
	return gotools.StringToBytes((*time.Time)(d).Format(time.RFC3339)), nil
}

func (d *DateTimeString) Scan(src any) error {
	switch s := src.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			logs.ErrorLog(err, "Parse(time.RFC3339")
			return nil
		}

		*d = (DateTimeString)(t)

		return nil
	case json.Number:
		t, err := time.Parse(time.DateOnly, (string)(s))
		if err != nil {
			return errors.Wrap(err, "Parse(time.DateOnly")
		}

		*d = (DateTimeString)(t)

		return nil
	default:
		return errors.Errorf("unknown type %T %[1]v", src)
	}
}

func (d *DateTimeString) CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool {
	return true
}

func (d *DateTimeString) GetPgxType() pgtype.Time {
	return pgtype.Time{
		Microseconds: (time.Time)(*d).UnixMicro(),
		Valid:        true,
	}
}

type TzString struct {
	*DateTimeString
}

func NewTzString() *TzString {
	return &TzString{&DateTimeString{}}
}

func (d *TzString) GetPgxType() pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  (time.Time)(*d.DateTimeString),
		Valid: true,
	}
}

// GetValue/NewValue are overridden (rather than left to promotion from the
// embedded *DateTimeString) because the promoted DateTimeString.NewValue would
// hand back a *DateTimeString, not a *TzString - the wrong concrete type once
// something downstream needs TzString's own GetPgxType (pgtype.Timestamptz,
// not DateTimeString's pgtype.Time).
func (d *TzString) GetValue() any { return d }
func (d *TzString) NewValue() any { return NewTzString() }

// Format implement Formatter interface
func (d *TzString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		_, err := fmt.Fprintf(s, "%T", d)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 'g':
		// NewTzString(), not "&crud.TzString{}" - the zero-value literal leaves the
		// embedded *DateTimeString nil, which panics the moment anything (UnmarshalJSON,
		// GetPgxType) touches it.
		_, err := fmt.Fprint(s, "crud.NewTzString()")
		if err != nil {
			logs.ErrorLog(err)
		}
	case 's':
		_, err := fmt.Fprint(s, (time.Time)(*d.DateTimeString).String())
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

type TimestampString struct {
	*DateTimeString
}

func NewTimestampString() *TimestampString {
	return &TimestampString{&DateTimeString{}}
}

func (d *TimestampString) GetPgxType() pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  (time.Time)(*d.DateTimeString),
		Valid: true,
	}
}

// GetValue/NewValue overridden for the same reason as TzString above - the
// promoted DateTimeString.NewValue would hand back a *DateTimeString, not a
// *TimestampString.
func (d *TimestampString) GetValue() any { return d }
func (d *TimestampString) NewValue() any { return NewTimestampString() }

// Format implement Formatter interface
func (d *TimestampString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		_, err := fmt.Fprintf(s, "%T", d)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 'g':
		// NewTimestampString(), not "&crud.TimestampString{}" - see TzString.Format above.
		_, err := fmt.Fprint(s, "crud.NewTimestampString()")
		if err != nil {
			logs.ErrorLog(err)
		}
	case 's':
		_, err := fmt.Fprint(s, (time.Time)(*d.DateTimeString).String())
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

type PgxDateString DateString

func (d *PgxDateString) NewValue() any {
	return &PgxDateString{}
}

func (d *PgxDateString) GetValue() any {
	return pgtype.Date{
		Time:  (time.Time)(*d),
		Valid: true,
	}
}

func (d *PgxDateString) PgxTypeString() string {
	return "pgtype.Date"
}

type DateString time.Time

func (d *DateString) Expect() string {
	return "date as string format #" + time.DateOnly
}

// Format implement Formatter interface
func (d *DateString) Format(s fmt.State, verb rune) {
	var err error
	switch verb {
	case 't':
		_, err = fmt.Fprintf(s, "%T", d)
	case 'g':
		_, err = fmt.Fprintf(s, "&%T{}", *d)
	case 's':
		_, err = fmt.Fprint(s, (time.Time)(*d).String())
	default:
	}
	if err != nil {
		logs.ErrorLog(err)
	}
}

func (d *DateString) FormatDoc() string {
	return "date"
}

func (d *DateString) RequestType() string {
	return "string"
}

func (d *DateString) GetValue() any {
	return d
}

func (d *DateString) NewValue() any {
	return &DateString{}
}

func (d *DateString) GetPgxType() pgtype.Date {
	return pgtype.Date{
		Time:  (time.Time)(*d),
		Valid: true,
	}
}

type PGXType[T any] interface {
	GetPgxType() T
}

// ToPgxSlice converts a slice of per-element wrapper values - DateString,
// TimestampString, PointString, DateRangeMarshal, ... anything already providing
// GetPgxType() E the way the single-value case does - into the []E slice pgx v5
// itself expects for an array (or an array-of-range) column.
//
// It exists because a slice has no GetPgxType() method of its own: for a scalar
// param, FuncAPI (endpointTpl.qtpl) can call apis.GetValue[*crud.DateString](ctx,
// param) and chain ".GetPgxType()" straight onto the result, but for an array
// param apis.GetValue[[]*crud.DateString](ctx, param) returns a slice - request
// parsing already produced one wrapper per array element - and that slice must be
// converted element-by-element instead. FuncParam calls this generic function
// (crud.ToPgxSlice) rather than emitting a per-type loop for every "standard"
// column type that has an array variant.
func ToPgxSlice[T PGXType[E], E any](vals []T) []E {
	out := make([]E, len(vals))
	for i, v := range vals {
		out[i] = v.GetPgxType()
	}

	return out
}

func ConvertUnixTime(t *time.Time, src string) error {

	fTime, err := strconv.ParseInt(src, 10, 64)
	if err != nil {
		return err
	}

	*t = time.Unix(fTime, 0)
	return nil
}

func (d *DateString) UnmarshalJSON(src []byte) error {
	t, err := bytesToTime(
		src,
		time.DateOnly,
		"2006-02-01",
		"01-02-2006",
		"02-01-2006",
	)
	if err != nil {
		return err
	}

	*d = (DateString)(t)
	return nil
}

func bytesToTime(src []byte, formats ...string) (t time.Time, err error) {
	str := gotools.BytesToString(src)
	if i := slices.IndexFunc(formats, func(f string) bool {
		t, err = time.Parse(f, str)
		return err == nil
	}); i < 0 {
		err = ConvertUnixTime(&t, str)
	}

	if err != nil {
		return t, fmt.Errorf("'%s' must be one of formats: %s or UnixTime", str, formats)
	}

	return
}

func (d *DateString) MarshalJSON() ([]byte, error) {
	return gotools.StringToBytes((*time.Time)(d).Format(time.DateOnly)), nil
}

type DtoFileField []*multipart.FileHeader

func (d *DtoFileField) GetPgxType() [][]byte {
	return slices.Collect(func(yield func([]byte) bool) {
		for _, header := range *d {
			f, err := header.Open()
			if err != nil {
				logs.ErrorLog(err, header)
				continue
			}
			all, err := io.ReadAll(f)
			_ = f.Close()
			if err != nil {
				logs.ErrorLog(err, header)
				continue
			}
			if !yield(all) {
				return
			}
		}

	})
}
func (d *DtoFileField) GetValue() any {
	return d.GetPgxType()
}

func (d *DtoFileField) NewValue() any {
	return new(DtoFileField)
}

func (d *DtoFileField) Expect() string {
	return "multipart file"
}

func (d *DtoFileField) FormatDoc() string {
	return "file"
}

func (d *DtoFileField) RequestType() string {
	return "file"
}

// Format implement Formatter interface
func (d *DtoFileField) Format(s fmt.State, verb rune) {
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
		_, err := fmt.Fprint(s, "[]*multipart.FileHeader")
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

// NewFileParam create new InParam for handling
func NewFileParam(name, desc string) apis.InParam {
	return apis.InParam{
		Name: name,
		Desc: desc,
		Type: apis.NewTypeInParam(types.UnsafePointer),
		//Type: apis.NewStructInParam(&DtoFileField{}),
	}
}

// CheckParams implement CheckDTO interface, put each params into user value on context
func (d *DtoFileField) CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool {
	for i, header := range *d {
		f, err := header.Open()
		if err != nil {
			logs.DebugLog(err, header)
			badParams[header.Filename] = errors.Wrapf(err, "%d. open file", i).Error()
		}
		_ = f.Close()
	}

	return len(badParams) == 0
}

type DtoField map[string]any

func (d *DtoField) Expect() string {
	return "JSON object {'key':'value'...}"
}

func (d *DtoField) FormatDoc() string {
	return "json"
}

func (d *DtoField) RequestType() string {
	return "string"
}

// Format implement Formatter interface
func (d *DtoField) Format(s fmt.State, verb rune) {
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
		_, err := fmt.Fprint(s, "map[string]any")
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

// CheckParams implement CheckDTO interface, put each params into user value on context
func (d *DtoField) CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool {
	for key, val := range *d {
		if strings.HasSuffix(key, "[]") {
			// key = strings.TrimSuffix(key, "[]")
			switch v := val.(type) {
			case []string:
				val = v
			case string:
				val = []string{v}
			case []any:
				s := make([]string, len(v))
				for i, str := range v {
					s[i] = fmt.Sprintf("%v", str)
				}
				val = s
			}
		}
		ctx.SetUserValue(key, val)
	}

	return true
}

func (d *DtoField) GetValue() any {
	return d
}

func (d *DtoField) NewValue() any {
	n := new(DtoField)
	return n
}

type FormActions struct {
	Typ string `json:"type"`
	Url string `json:"url"`
}

func DecodeDatetimeString(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	val := (*DateTimeString)(ptr)

	switch t := iter.WhatIsNext(); t {
	case jsoniter.StringValue:
		src := iter.ReadString()
		err := val.Scan(src)
		if err != nil {
			logs.ErrorLog(err, val, src)
		}
	case jsoniter.NumberValue:
		buf := bytes.NewBufferString("")
		for iter.Error != io.EOF {
			// iter.Error = nil
			i := iter.ReadAny()
			if iter.WhatIsNext() == jsoniter.InvalidValue {
				logs.ErrorLog(iter.Error)
				iter.Skip()
				break
			}
			_, _ = buf.WriteString((string)(i.ToString()))
		}
		src := buf.String()
		err := val.Scan(src)
		if err != nil {
			logs.ErrorLog(err, val, src)
		}
		iter.Error = nil
	case jsoniter.ObjectValue:
		err := ((*time.Time)(val)).UnmarshalText(gotools.StringToBytes(iter.ReadObject()))
		if err != nil {
			logs.ErrorLog(err, val)
		}
	default:
		logs.ErrorLog(errors.New("unknown type"), t)
		err := val.Scan(iter.Read())
		if err != nil {
			logs.ErrorLog(err, val, t)
		}
	}
}

func EncodeDateString(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	d := (*DateString)(ptr)
	stream.WriteString((*time.Time)(d).Format(time.DateOnly))
}
func IsEmptyDateString(ptr unsafe.Pointer) bool {
	d := (*DateString)(ptr)
	return (*time.Time)(d).IsZero()
}

// ===================== Geometric Types =====================

type PointString struct {
	pgtype.Point
}

func (p *PointString) GetPgxType() pgtype.Point {
	return p.Point
}

func (p *PointString) GetValue() any       { return &p.Point }
func (p *PointString) NewValue() any       { return &PointString{} }
func (p *PointString) Expect() string      { return "point" }
func (p *PointString) FormatDoc() string   { return "point" }
func (p *PointString) RequestType() string { return "string" }

func (p *PointString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", p)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *p)
	case 's':
		if p.Valid {
			fmt.Fprintf(s, "(%v,%v)", p.P.X, p.P.Y)
		}
	}
}

// ===================== Other Geometric Types =====================

// scanTextJSON is shared by every geometric wrapper below: it unmarshals a JSON
// string and hands it straight to the embedded pgtype value's own Scan, which
// already knows how to parse Postgres's native text syntax for that type (e.g.
// Box's "(x1,y1),(x2,y2)") and sets Valid on success - so none of these wrappers
// need a bespoke parser of their own. dst is satisfied automatically by every
// *LineString/*LsegString/*BoxString/*PathString/*PolygonString/*CircleString
// below via method promotion from their embedded pgtype value.
func scanTextJSON(dst interface{ Scan(any) error }, src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	return dst.Scan(s)
}

type LineString struct{ pgtype.Line }
type LsegString struct{ pgtype.Lseg }
type BoxString struct{ pgtype.Box }
type PathString struct{ pgtype.Path }
type PolygonString struct{ pgtype.Polygon }
type CircleString struct{ pgtype.Circle }

func (l *LineString) GetPgxType() pgtype.Line       { return l.Line }
func (l *LsegString) GetPgxType() pgtype.Lseg       { return l.Lseg }
func (l *BoxString) GetPgxType() pgtype.Box         { return l.Box }
func (l *PathString) GetPgxType() pgtype.Path       { return l.Path }
func (l *PolygonString) GetPgxType() pgtype.Polygon { return l.Polygon }
func (l *CircleString) GetPgxType() pgtype.Circle   { return l.Circle }

func (l *LineString) GetValue() any    { return &l.Line }
func (l *LsegString) GetValue() any    { return &l.Lseg }
func (l *BoxString) GetValue() any     { return &l.Box }
func (l *PathString) GetValue() any    { return &l.Path }
func (l *PolygonString) GetValue() any { return &l.Polygon }
func (l *CircleString) GetValue() any  { return &l.Circle }

func (l *LineString) NewValue() any    { return &LineString{} }
func (l *LsegString) NewValue() any    { return &LsegString{} }
func (l *BoxString) NewValue() any     { return &BoxString{} }
func (l *PathString) NewValue() any    { return &PathString{} }
func (l *PolygonString) NewValue() any { return &PolygonString{} }
func (l *CircleString) NewValue() any  { return &CircleString{} }

func (l *LineString) Expect() string    { return `line, e.g. "{A,B,C}"` }
func (l *LsegString) Expect() string    { return `line segment, e.g. "[(x1,y1),(x2,y2)]"` }
func (l *BoxString) Expect() string     { return `box, e.g. "(x1,y1),(x2,y2)"` }
func (l *PathString) Expect() string    { return `path, e.g. "[(x1,y1),(x2,y2),...]" or "((x1,y1),...)"` }
func (l *PolygonString) Expect() string { return `polygon, e.g. "((x1,y1),(x2,y2),...)"` }
func (l *CircleString) Expect() string  { return `circle, e.g. "<(x,y),r>"` }

func (l *LineString) FormatDoc() string    { return "line" }
func (l *LsegString) FormatDoc() string    { return "lseg" }
func (l *BoxString) FormatDoc() string     { return "box" }
func (l *PathString) FormatDoc() string    { return "path" }
func (l *PolygonString) FormatDoc() string { return "polygon" }
func (l *CircleString) FormatDoc() string  { return "circle" }

func (l *LineString) RequestType() string    { return "string" }
func (l *LsegString) RequestType() string    { return "string" }
func (l *BoxString) RequestType() string     { return "string" }
func (l *PathString) RequestType() string    { return "string" }
func (l *PolygonString) RequestType() string { return "string" }
func (l *CircleString) RequestType() string  { return "string" }

// UnmarshalJSON: each wrapper's embedded pgtype value already has a Scan(any)
// error accepting Postgres's own text format (promoted onto the pointer receiver
// below), so scanTextJSON just needs a JSON string unwrapped first.
func (l *LineString) UnmarshalJSON(src []byte) error    { return scanTextJSON(l, src) }
func (l *LsegString) UnmarshalJSON(src []byte) error    { return scanTextJSON(l, src) }
func (l *BoxString) UnmarshalJSON(src []byte) error     { return scanTextJSON(l, src) }
func (l *PathString) UnmarshalJSON(src []byte) error    { return scanTextJSON(l, src) }
func (l *PolygonString) UnmarshalJSON(src []byte) error { return scanTextJSON(l, src) }
func (l *CircleString) UnmarshalJSON(src []byte) error  { return scanTextJSON(l, src) }

// Format ('s' case): each embedded pgtype value already implements
// database/sql/driver.Valuer (promoted the same way Scan is above), rendering
// Postgres's own canonical text syntax for that type - so no per-type formatting
// code is needed here either.
func formatValuerString(v interface{ Value() (driver.Value, error) }, s fmt.State) {
	val, err := v.Value()
	if err != nil {
		logs.ErrorLog(err)
		return
	}
	if val != nil {
		fmt.Fprintf(s, "%v", val)
	}
}

func (l *LineString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", l)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *l)
	case 's':
		formatValuerString(l, s)
	}
}

func (l *LsegString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", l)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *l)
	case 's':
		formatValuerString(l, s)
	}
}

func (l *BoxString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", l)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *l)
	case 's':
		formatValuerString(l, s)
	}
}

func (l *PathString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", l)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *l)
	case 's':
		formatValuerString(l, s)
	}
}

func (l *PolygonString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", l)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *l)
	case 's':
		formatValuerString(l, s)
	}
}

func (l *CircleString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", l)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *l)
	case 's':
		formatValuerString(l, s)
	}
}

// ===================== Full-Text Search Types =====================

// TSVectorString and TSQueryString both embed pgtype.Text, which already
// implements UnmarshalJSON/MarshalJSON/Scan/Value on its own (pgx v5 represents
// both tsvector and tsquery as plain text client-side - Postgres does the actual
// full-text parsing server-side), so both get those four for free via promotion
// and only need the same companion methods every other wrapper here provides.
type TSVectorString struct{ pgtype.Text }
type TSQueryString struct{ pgtype.Text }

func (t *TSVectorString) GetPgxType() pgtype.Text { return t.Text }
func (t *TSQueryString) GetPgxType() pgtype.Text  { return t.Text }

func (t *TSVectorString) GetValue() any { return &t.Text }
func (t *TSQueryString) GetValue() any  { return &t.Text }

func (t *TSVectorString) NewValue() any { return &TSVectorString{} }
func (t *TSQueryString) NewValue() any  { return &TSQueryString{} }

func (t *TSVectorString) Expect() string { return "tsvector" }
func (t *TSQueryString) Expect() string  { return "tsquery" }

func (t *TSVectorString) FormatDoc() string { return "tsvector" }
func (t *TSQueryString) FormatDoc() string  { return "tsquery" }

func (t *TSVectorString) RequestType() string { return "string" }
func (t *TSQueryString) RequestType() string  { return "string" }

func (t *TSVectorString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", t)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *t)
	case 's':
		formatValuerString(t, s)
	}
}

func (t *TSQueryString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", t)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *t)
	case 's':
		formatValuerString(t, s)
	}
}

// ===================== Arbitrary-precision Numeric =====================

// NumericString wraps pgtype.Numeric so a "numeric"/"decimal" column parses
// straight into Postgres's arbitrary-precision representation. pgtype.Numeric
// already implements UnmarshalJSON/Scan/Value itself - its UnmarshalJSON parses
// the JSON number's own text directly (scanPlanTextAnyToNumericScanner), never
// going through float64 - so this only adds the same companion methods every
// other wrapper here needs. Without this, a "numeric"/"decimal" column falls
// through setType's default branch to col.BasicType(), which maps to float64 and
// silently rounds anything beyond float64's ~15-17 significant digits.
type NumericString struct{ pgtype.Numeric }

func (n *NumericString) GetPgxType() pgtype.Numeric { return n.Numeric }
func (n *NumericString) GetValue() any              { return &n.Numeric }
func (n *NumericString) NewValue() any              { return &NumericString{} }
func (n *NumericString) Expect() string             { return "numeric" }
func (n *NumericString) FormatDoc() string          { return "numeric" }
func (n *NumericString) RequestType() string        { return "number" }

func (n *NumericString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", n)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *n)
	case 's':
		formatValuerString(n, s)
	}
}

// ===================== Hstore =====================

// HstoreMarshal wraps pgtype.Hstore (map[string]*string) so an "hstore" column
// gets a real pgx conversion via GetPgxType instead of the previous reuse of
// DtoField (map[string]any), which has no GetPgxType at all and can't represent
// hstore's per-key NULLs (a nil *string) the way map[string]*string does. No
// custom UnmarshalJSON parser is needed beyond pointing json.Unmarshal at the
// embedded map directly - Hstore's underlying map kind isn't flattened by
// encoding/json the way an embedded struct would be, so it has to be targeted
// explicitly, but decoding into map[string]*string (including a JSON null
// becoming a nil entry) is otherwise the default behavior.
type HstoreMarshal struct{ pgtype.Hstore }

func NewHstoreMarshal() *HstoreMarshal { return &HstoreMarshal{Hstore: pgtype.Hstore{}} }

func (h *HstoreMarshal) GetPgxType() pgtype.Hstore { return h.Hstore }
func (h *HstoreMarshal) GetValue() any             { return &h.Hstore }
func (h *HstoreMarshal) NewValue() any             { return NewHstoreMarshal() }
func (h *HstoreMarshal) Expect() string            { return `hstore, e.g. {"key":"value"}` }
func (h *HstoreMarshal) FormatDoc() string         { return "hstore" }
func (h *HstoreMarshal) RequestType() string       { return "json" }

func (h *HstoreMarshal) UnmarshalJSON(src []byte) error {
	return json.Unmarshal(src, &h.Hstore)
}

func (h *HstoreMarshal) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", h)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *h)
	case 's':
		formatValuerString(h, s)
	}
}

// ===================== Range types (int4range, int8range, tsrange, tstzrange) =====================
//
// daterange and numrange already have their own hand-written wrapper further up
// this file (DateRangeMarshal/NumrangeMarshal). Postgres's other three built-in
// range types - int4range, int8range, tsrange, tstzrange - had no case in
// setType (params.go) at all and fell through to the generic composite/DTO path
// instead of getting clean parsing. pgtype.Range[T] itself has no Scan/Value/
// UnmarshalJSON (it can't - encoding a range generically requires knowing how to
// encode T, which needs a type map, not a plain method), so each concrete range
// type below still needs a parser; parseRangeLiteral is shared across all four
// since none of their bound texts can contain a literal ',' or bracket, so one
// generic split-on-first-comma parser is safe for all of them.

// parseRangeLiteral parses Postgres's range literal syntax - "[lower,upper)",
// with '[' / '(' for an inclusive/exclusive lower bound, ']' / ')' for the upper
// bound, an empty bound meaning unbounded, and the literal string "empty" for an
// empty range - using parse to convert each non-empty bound's text into E.
func parseRangeLiteral[E any](src string, parse func(string) (E, error)) (pgtype.Range[E], error) {
	var r pgtype.Range[E]

	if src == "empty" {
		r.LowerType, r.UpperType = pgtype.Empty, pgtype.Empty
		r.Valid = true
		return r, nil
	}

	if len(src) < 3 {
		return r, fmt.Errorf("invalid range literal %q", src)
	}

	switch src[0] {
	case '[':
		r.LowerType = pgtype.Inclusive
	case '(':
		r.LowerType = pgtype.Exclusive
	default:
		return r, fmt.Errorf("invalid range literal %q: expected '[' or '('", src)
	}

	switch src[len(src)-1] {
	case ']':
		r.UpperType = pgtype.Inclusive
	case ')':
		r.UpperType = pgtype.Exclusive
	default:
		return r, fmt.Errorf("invalid range literal %q: expected ']' or ')'", src)
	}

	lowerSrc, upperSrc, found := strings.Cut(src[1:len(src)-1], ",")
	if !found {
		return r, fmt.Errorf("invalid range literal %q: missing ','", src)
	}

	if lowerSrc == "" {
		r.LowerType = pgtype.Unbounded
	} else {
		v, err := parse(lowerSrc)
		if err != nil {
			return r, fmt.Errorf("invalid lower bound %q: %w", lowerSrc, err)
		}
		r.Lower = v
	}

	if upperSrc == "" {
		r.UpperType = pgtype.Unbounded
	} else {
		v, err := parse(upperSrc)
		if err != nil {
			return r, fmt.Errorf("invalid upper bound %q: %w", upperSrc, err)
		}
		r.Upper = v
	}

	r.Valid = true
	return r, nil
}

// formatRangeLiteral renders a pgtype.Range[E] back into the same syntax
// parseRangeLiteral accepts, for doc/example display (Format's 's' case) - the
// query-arg path doesn't use this, it goes through GetPgxType + pgx's own
// RangeCodec instead.
func formatRangeLiteral[E any](r pgtype.Range[E]) string {
	if !r.Valid {
		return ""
	}
	if r.LowerType == pgtype.Empty {
		return "empty"
	}

	var b strings.Builder
	if r.LowerType == pgtype.Inclusive {
		b.WriteByte('[')
	} else {
		b.WriteByte('(')
	}
	if r.LowerType != pgtype.Unbounded {
		fmt.Fprintf(&b, "%v", r.Lower)
	}
	b.WriteByte(',')
	if r.UpperType != pgtype.Unbounded {
		fmt.Fprintf(&b, "%v", r.Upper)
	}
	if r.UpperType == pgtype.Inclusive {
		b.WriteByte(']')
	} else {
		b.WriteByte(')')
	}
	return b.String()
}

type Int4RangeMarshal struct{ pgtype.Range[int32] }
type Int8RangeMarshal struct{ pgtype.Range[int64] }
type TsRangeMarshal struct{ pgtype.Range[time.Time] }
type TsTzRangeMarshal struct{ pgtype.Range[time.Time] }

func NewInt4RangeMarshal() *Int4RangeMarshal { return &Int4RangeMarshal{} }
func NewInt8RangeMarshal() *Int8RangeMarshal { return &Int8RangeMarshal{} }
func NewTsRangeMarshal() *TsRangeMarshal     { return &TsRangeMarshal{} }
func NewTsTzRangeMarshal() *TsTzRangeMarshal { return &TsTzRangeMarshal{} }

func (r *Int4RangeMarshal) GetPgxType() pgtype.Range[int32]     { return r.Range }
func (r *Int8RangeMarshal) GetPgxType() pgtype.Range[int64]     { return r.Range }
func (r *TsRangeMarshal) GetPgxType() pgtype.Range[time.Time]   { return r.Range }
func (r *TsTzRangeMarshal) GetPgxType() pgtype.Range[time.Time] { return r.Range }

func (r *Int4RangeMarshal) GetValue() any { return &r.Range }
func (r *Int8RangeMarshal) GetValue() any { return &r.Range }
func (r *TsRangeMarshal) GetValue() any   { return &r.Range }
func (r *TsTzRangeMarshal) GetValue() any { return &r.Range }

func (r *Int4RangeMarshal) NewValue() any { return NewInt4RangeMarshal() }
func (r *Int8RangeMarshal) NewValue() any { return NewInt8RangeMarshal() }
func (r *TsRangeMarshal) NewValue() any   { return NewTsRangeMarshal() }
func (r *TsTzRangeMarshal) NewValue() any { return NewTsTzRangeMarshal() }

func (r *Int4RangeMarshal) Expect() string { return `int4range, e.g. "[1,10)"` }
func (r *Int8RangeMarshal) Expect() string { return `int8range, e.g. "[1,10)"` }
func (r *TsRangeMarshal) Expect() string {
	return `tsrange, e.g. "[2024-01-01T00:00:00Z,2024-02-01T00:00:00Z)"`
}
func (r *TsTzRangeMarshal) Expect() string {
	return `tstzrange, e.g. "[2024-01-01T00:00:00Z,2024-02-01T00:00:00Z)"`
}

func (r *Int4RangeMarshal) FormatDoc() string { return "range" }
func (r *Int8RangeMarshal) FormatDoc() string { return "range" }
func (r *TsRangeMarshal) FormatDoc() string   { return "range" }
func (r *TsTzRangeMarshal) FormatDoc() string { return "range" }

func (r *Int4RangeMarshal) RequestType() string { return "string" }
func (r *Int8RangeMarshal) RequestType() string { return "string" }
func (r *TsRangeMarshal) RequestType() string   { return "string" }
func (r *TsTzRangeMarshal) RequestType() string { return "string" }

func (r *Int4RangeMarshal) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	rng, err := parseRangeLiteral(s, func(b string) (int32, error) {
		v, err := strconv.ParseInt(b, 10, 32)
		return int32(v), err
	})
	if err != nil {
		return err
	}

	r.Range = rng
	return nil
}

func (r *Int8RangeMarshal) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	rng, err := parseRangeLiteral(s, func(b string) (int64, error) {
		return strconv.ParseInt(b, 10, 64)
	})
	if err != nil {
		return err
	}

	r.Range = rng
	return nil
}

// tsBound parses one range bound's text using the same format list DateTimeString
// already accepts, so "2024-01-01T00:00:00Z", "2024-01-01 00:00:00" etc. all work.
func tsBound(b string) (time.Time, error) {
	return bytesToTime(gotools.StringToBytes(b),
		time.RFC3339,
		time.RFC3339Nano,
		time.DateTime,
	)
}

func (r *TsRangeMarshal) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	rng, err := parseRangeLiteral(s, tsBound)
	if err != nil {
		return err
	}

	r.Range = rng
	return nil
}

func (r *TsTzRangeMarshal) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	rng, err := parseRangeLiteral(s, tsBound)
	if err != nil {
		return err
	}

	r.Range = rng
	return nil
}

func (r *Int4RangeMarshal) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", r)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *r)
	case 's':
		fmt.Fprint(s, formatRangeLiteral(r.Range))
	}
}

func (r *Int8RangeMarshal) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", r)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *r)
	case 's':
		fmt.Fprint(s, formatRangeLiteral(r.Range))
	}
}

func (r *TsRangeMarshal) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", r)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *r)
	case 's':
		fmt.Fprint(s, formatRangeLiteral(r.Range))
	}
}

func (r *TsTzRangeMarshal) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", r)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *r)
	case 's':
		fmt.Fprint(s, formatRangeLiteral(r.Range))
	}
}

// ===================== UUID =====================

// UUIDString wraps pgtype.UUID, which already implements UnmarshalJSON/Scan/
// Value/String itself - previously "uuid" went through plain apis.NewTypeInParam
// (types.String) with no format validation at all, so a malformed UUID only
// surfaced as an opaque Postgres error. Wrapping it here rejects a bad UUID with
// a clear client-side error before it ever reaches the DB.
type UUIDString struct{ pgtype.UUID }

func (u *UUIDString) GetPgxType() pgtype.UUID { return u.UUID }
func (u *UUIDString) GetValue() any           { return &u.UUID }
func (u *UUIDString) NewValue() any           { return &UUIDString{} }
func (u *UUIDString) Expect() string          { return "uuid" }
func (u *UUIDString) FormatDoc() string       { return "uuid" }
func (u *UUIDString) RequestType() string     { return "string" }

func (u *UUIDString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", u)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *u)
	case 's':
		if u.Valid {
			fmt.Fprint(s, u.UUID.String())
		}
	}
}

// ===================== MAC address =====================

// MacaddrString covers both "macaddr" and "macaddr8" - pgx v5 has no dedicated
// pgtype.Macaddr; it scans/encodes both straight into net.HardwareAddr, whose
// length (6 vs 8 bytes) determines which Postgres type it matches. Like uuid
// above, this replaces the previous plain-string handling with an early,
// clear validation error instead of a round trip to Postgres.
type MacaddrString struct{ net.HardwareAddr }

func (m *MacaddrString) GetPgxType() net.HardwareAddr { return m.HardwareAddr }
func (m *MacaddrString) GetValue() any                { return &m.HardwareAddr }
func (m *MacaddrString) NewValue() any                { return &MacaddrString{} }
func (m *MacaddrString) Expect() string               { return "MAC address, e.g. 01:23:45:67:89:ab" }
func (m *MacaddrString) FormatDoc() string            { return "macaddr" }
func (m *MacaddrString) RequestType() string          { return "string" }

func (m *MacaddrString) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	addr, err := net.ParseMAC(s)
	if err != nil {
		return err
	}

	m.HardwareAddr = addr
	return nil
}

func (m *MacaddrString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", m)
	case 'g':
		fmt.Fprintf(s, "&%T{}", *m)
	case 's':
		if m.HardwareAddr != nil {
			fmt.Fprint(s, m.HardwareAddr.String())
		}
	}
}

// ===================== Enum =====================

// EnumString validates a user-defined enum column's incoming value against the
// type's own labels client-side, instead of only finding out it's invalid after
// a round trip to Postgres. ConvertDbType (params.go) constructs one with
// NewEnumString(udt.Enumerates) whenever the column's UserDefinedType has
// Enumerates set.
type EnumString struct {
	Value   string
	Allowed []string
}

func NewEnumString(allowed []string) *EnumString {
	return &EnumString{Allowed: allowed}
}

func (e *EnumString) GetPgxType() string  { return e.Value }
func (e *EnumString) GetValue() any       { return e.Value }
func (e *EnumString) NewValue() any       { return NewEnumString(e.Allowed) }
func (e *EnumString) Expect() string      { return "one of: " + strings.Join(e.Allowed, ", ") }
func (e *EnumString) FormatDoc() string   { return "enum" }
func (e *EnumString) RequestType() string { return "string" }

func (e *EnumString) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return err
	}

	e.Value = s
	return nil
}

// CheckParams implement CheckDTO interface - rejects a label Postgres doesn't
// know about with a clear message instead of letting the query fail server-side.
func (e *EnumString) CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool {
	if e.Value == "" || slices.Contains(e.Allowed, e.Value) {
		return true
	}

	badParams["value"] = fmt.Sprintf("%q must be one of: %s", e.Value, strings.Join(e.Allowed, ", "))
	return false
}

func (e *EnumString) Format(s fmt.State, verb rune) {
	switch verb {
	case 't':
		fmt.Fprintf(s, "%T", e)
	case 'g':
		fmt.Fprintf(s, "crud.NewEnumString(%#v)", e.Allowed)
	case 's':
		fmt.Fprint(s, e.Value)
	}
}

func init() {
	jsoniter.RegisterTypeDecoderFunc("crud.DateTimeString", DecodeDatetimeString)
	jsoniter.RegisterTypeEncoderFunc("crud.DateString", EncodeDateString, IsEmptyDateString)
}

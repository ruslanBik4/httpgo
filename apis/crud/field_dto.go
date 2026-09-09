/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
	"io"
	"mime/multipart"
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

// Format implement Formatter interface
func (d *TzString) Format(s fmt.State, verb rune) {
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

// Format implement Formatter interface
func (d *TimestampString) Format(s fmt.State, verb rune) {
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

type LineString struct{ pgtype.Line }
type LsegString struct{ pgtype.Lseg }

func (l *LsegString) GetValue() any {
	//TODO implement me
	panic("implement me")
}

func (l *LsegString) NewValue() any {
	//TODO implement me
	panic("implement me")
}

type BoxString struct{ pgtype.Box }

func (l *BoxString) GetValue() any {
	//TODO implement me
	panic("implement me")
}

func (l *BoxString) NewValue() any {
	//TODO implement me
	panic("implement me")
}

type PathString struct{ pgtype.Path }

func (l *PathString) GetValue() any {
	//TODO implement me
	panic("implement me")
}

func (l *PathString) NewValue() any {
	//TODO implement me
	panic("implement me")
}

type PolygonString struct{ pgtype.Polygon }

func (l *PolygonString) GetValue() any {
	//TODO implement me
	panic("implement me")
}

func (l *PolygonString) NewValue() any {
	//TODO implement me
	panic("implement me")
}

type CircleString struct{ pgtype.Circle }

func (l *CircleString) GetValue() any {
	//TODO implement me
	panic("implement me")
}

func (l *CircleString) NewValue() any {
	//TODO implement me
	panic("implement me")
}

func (l *LineString) GetPgxType() pgtype.Line       { return l.Line }
func (l *LsegString) GetPgxType() pgtype.Lseg       { return l.Lseg }
func (l *BoxString) GetPgxType() pgtype.Box         { return l.Box }
func (l *PathString) GetPgxType() pgtype.Path       { return l.Path }
func (l *PolygonString) GetPgxType() pgtype.Polygon { return l.Polygon }
func (l *CircleString) GetPgxType() pgtype.Circle   { return l.Circle }

// Common methods (can be simplified with embedding if desired)
func (l *LineString) GetValue() any       { return &l.Line }
func (l *LineString) NewValue() any       { return &LineString{} }
func (l *LineString) Expect() string      { return "line" }
func (l *LineString) FormatDoc() string   { return "line" }
func (l *LineString) RequestType() string { return "string" }

// (Repeat similar methods for LsegString, BoxString, PathString, PolygonString, CircleString)

// ===================== Full-Text Search Types =====================

type TSVectorString struct{ pgtype.Text }
type TSQueryString struct{ pgtype.Text }

func (t *TSQueryString) GetValue() any {
	//TODO implement me
	panic("implement me")
}

func (t *TSQueryString) NewValue() any {
	//TODO implement me
	panic("implement me")
}

func (t *TSVectorString) GetPgxType() pgtype.Text { return t.Text }
func (t *TSQueryString) GetPgxType() pgtype.Text  { return t.Text }

func (t *TSVectorString) GetValue() any       { return &t.Text }
func (t *TSVectorString) NewValue() any       { return &TSVectorString{} }
func (t *TSVectorString) Expect() string      { return "tsvector" }
func (t *TSVectorString) FormatDoc() string   { return "tsvector" }
func (t *TSVectorString) RequestType() string { return "string" }

func init() {
	jsoniter.RegisterTypeDecoderFunc("crud.DateTimeString", DecodeDatetimeString)
	jsoniter.RegisterTypeEncoderFunc("crud.DateString", EncodeDateString, IsEmptyDateString)
}

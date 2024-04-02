/*
 * Copyright (c) 2022-2024. Author: Ruslan Bikchentaev. All rights reserved.
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
	"strings"
	"time"
	"unsafe"

	"github.com/jackc/pgtype"
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
	case 'g', 's':
		_, err := fmt.Fprintf(s, "&crud.DTO[%T]{}", d.val)
		if err != nil {
			logs.ErrorLog(err)
		}

	}
}

func NewDTO[T any](val T) *DTO[T] {
	return &DTO[T]{val: val}
}

func (d *DTO[T]) GetValue() any {
	return d.val
}

func (d *DTO[T]) NewValue() any {
	var a T
	return a
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
	switch verb {
	case 't':
		_, err := fmt.Fprintf(s, "%s", d.Val)
		if err != nil {
			logs.ErrorLog(err)
		}
	case 'g':
		_, err := fmt.Fprintf(s, "crud.NewDTO(%s{})", d.Val)
		if err != nil {
			logs.ErrorLog(err)
		}

	case 's':
		_, err := fmt.Fprintf(s, "%s{}", d.Val)
		if err != nil {
			logs.ErrorLog(err)
		}

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

func (d *DateTimeString) UnmarshalJSON(src []byte) (err error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.DateTime,
		time.RFC1123,
		time.RFC1123Z,
		time.Stamp,
	}
	toString := gotools.BytesToString(src)
	var t time.Time
	for _, f := range formats {
		t, err = time.Parse(f, toString)
		if err == nil {
			*d = (DateTimeString)(t)
			return nil
		}
		err = errors.Wrap(err, "Parse(time)")
	}

	return
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

func (d *DateTimeString) GetPgxType() pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:   (time.Time)(*d),
		Status: pgtype.Present,
	}
}

type DateString time.Time

func (d *DateString) Expect() string {
	return "date as string format #" + time.DateOnly
}

// Format implement Formatter interface
func (d *DateString) Format(s fmt.State, verb rune) {
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
		Time:   (time.Time)(*d),
		Status: pgtype.Present,
	}
}

func (d *DateString) UnmarshalJSON(src []byte) (err error) {
	formats := []string{
		time.DateOnly,
		"2006-02-01",
		"01-02-2006",
		"02-01-2006",
	}
	toString := gotools.BytesToString(src)
	var t time.Time
	s := ""
	for _, f := range formats {
		t, err = time.Parse(f, toString)
		if err == nil {
			*d = (DateString)(t)
			return nil
		}
		s += err.Error()
	}

	return errors.Wrap(err, s)
}

func (d *DateString) MarshalJSON() ([]byte, error) {
	return gotools.StringToBytes((*time.Time)(d).Format(time.DateOnly)), nil
}

type DtoFileField []*multipart.FileHeader

func (d *DtoFileField) GetValue() any {
	return d
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
func init() {
	jsoniter.RegisterTypeDecoderFunc("crud.DateTimeString", DecodeDatetimeString)
	jsoniter.RegisterTypeEncoderFunc("crud.DateString", EncodeDateString, IsEmptyDateString)
}

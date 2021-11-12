// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"
)

type DateTimeString time.Time

func (d *DateTimeString) UnmarshalJSON(src []byte) error {
	t, err := time.Parse(time.RFC3339, string(src))
	if err != nil {
		return errors.Wrap(err, "Parse(time.RFC3339")
	}

	*d = (DateTimeString)(t)

	return nil
}

func (d *DateTimeString) MarshalJSON() ([]byte, error) {
	return []byte((*time.Time)(d).Format(time.RFC3339)), nil
}

func (d *DateTimeString) Scan(src interface{}) error {
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
		t, err := time.Parse("2006-01-02", (string)(s))
		if err != nil {
			return errors.Wrap(err, "Parse(time.RFC3339")
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

type dtoField map[string]interface{}

func (d *dtoField) CheckParams(ctx *fasthttp.RequestCtx, badParams map[string]string) bool {
	for key, val := range *d {
		if strings.HasSuffix(key, "[]") {
			// key = strings.TrimSuffix(key, "[]")
			switch v := val.(type) {
			case string:
				val = []string{v}
			case []interface{}:
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

func (d dtoField) GetValue() interface{} {
	return d
}

func (d dtoField) NewValue() interface{} {
	d = make(map[string]interface{}, 0)
	// todo check later
	return &d
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
		err := ((*time.Time)(val)).UnmarshalText([]byte(iter.ReadObject()))
		if err != nil {
			logs.ErrorLog(err, val)
		}
		// val.Valid = iter.ReadMapCB(func(iterator *jsoniter.Iterator, key string) bool {
		// 	switch strings.ToLower(key) {
		// 	case "string":
		// 		val.String = iter.ReadString()
		// 		return true
		// 	case "valid":
		// 		val.Valid = iter.ReadBool()
		// 		return val.Valid
		// 	default:
		// 		logs.ErrorLog(errors.New("unknown key of NUllString"), key)
		// 		return false
		// 	}
		// })
	default:
		logs.ErrorLog(errors.New("unknown type"), t)
		err := val.Scan(iter.Read())
		if err != nil {
			logs.ErrorLog(err, val, t)
		}
	}
}

func init() {
	jsoniter.RegisterTypeDecoderFunc("crud.DateTimeString", DecodeDatetimeString)
	jsoniter.RegisterTypeDecoderFunc("*crud.DateTimeString", DecodeDatetimeString)
}

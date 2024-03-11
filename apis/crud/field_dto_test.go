/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"bytes"
	"fmt"
	"testing"
	"time"
	"unsafe"

	"github.com/jackc/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/views/templates/json"
	"github.com/ruslanBik4/logs"
)

func TestDecodeDatetimeString(t *testing.T) {
	tests := []struct {
		name string
		want error
	}{
		// TODO: Add test cases.
		{
			"2021-01-31T22:00:00.000Z",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := DateTimeString{}
			assert.NotNil(t, jsoniter.UnmarshalFromString(tt.name, &v))
		})
	}
}

func TestDateString_Expect(t *testing.T) {
	tests := []struct {
		name string
		d    DateString
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.Expect(), "Expect()")
		})
	}
}

func TestDateString_Format(t *testing.T) {
	tests := []struct {
		name string
		d    DateString
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.Format(), "Format()")
		})
	}
}

func TestDateString_GetValue(t *testing.T) {
	tests := []struct {
		name string
		d    DateString
		want any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.GetValue(), "GetValue()")
		})
	}
}

func TestDateString_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		d       DateString
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.MarshalJSON()
			if !tt.wantErr(t, err, fmt.Sprintf("MarshalJSON()")) {
				return
			}
			assert.Equalf(t, tt.want, got, "MarshalJSON()")
		})
	}
}

func TestDateString_NewValue(t *testing.T) {
	tests := []struct {
		name string
		d    DateString
		want any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.NewValue(), "NewValue()")
		})
	}
}

func TestDateString_RequestType(t *testing.T) {
	tests := []struct {
		name string
		d    DateString
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.RequestType(), "RequestType()")
		})
	}
}

func TestDateString_UnmarshalJSON(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		d       DateString
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.d.UnmarshalJSON(tt.args.src), fmt.Sprintf("UnmarshalJSON(%v)", tt.args.src))
		})
	}
}

func TestDateTimeString_CheckParams(t *testing.T) {
	type args struct {
		ctx       *fasthttp.RequestCtx
		badParams map[string]string
	}
	tests := []struct {
		name string
		d    DateTimeString
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.CheckParams(tt.args.ctx, tt.args.badParams), "CheckParams(%v, %v)", tt.args.ctx, tt.args.badParams)
		})
	}
}

func TestDateTimeString_Expect(t *testing.T) {
	tests := []struct {
		name string
		d    DateTimeString
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.Expect(), "Expect()")
		})
	}
}

func TestDateTimeString_Format(t *testing.T) {
	tests := []struct {
		name string
		d    DateTimeString
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.Format(), "Format()")
		})
	}
}

func TestDateTimeString_GetValue(t *testing.T) {
	tests := []struct {
		name string
		d    DateTimeString
		want any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.GetValue(), "GetValue()")
		})
	}
}

func TestDateTimeString_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		d       DateTimeString
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.MarshalJSON()
			if !tt.wantErr(t, err, fmt.Sprintf("MarshalJSON()")) {
				return
			}
			assert.Equalf(t, tt.want, got, "MarshalJSON()")
		})
	}
}

func TestDateTimeString_NewValue(t *testing.T) {
	tests := []struct {
		name string
		d    DateTimeString
		want any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.NewValue(), "NewValue()")
		})
	}
}

func TestDateTimeString_RequestType(t *testing.T) {
	tests := []struct {
		name string
		d    DateTimeString
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.RequestType(), "RequestType()")
		})
	}
}

func TestDateTimeString_Scan(t *testing.T) {
	type args struct {
		src any
	}
	tests := []struct {
		name    string
		d       DateTimeString
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.d.Scan(tt.args.src), fmt.Sprintf("Scan(%v)", tt.args.src))
		})
	}
}

func TestDateTimeString_UnmarshalJSON(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		d       DateTimeString
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.d.UnmarshalJSON(tt.args.src), fmt.Sprintf("UnmarshalJSON(%v)", tt.args.src))
		})
	}
}

func TestDecodeDatetimeString1(t *testing.T) {
	type args struct {
		ptr  unsafe.Pointer
		iter *jsoniter.Iterator
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DecodeDatetimeString(tt.args.ptr, tt.args.iter)
		})
	}
}

func TestDtoField_CheckParams(t *testing.T) {
	type args struct {
		ctx       *fasthttp.RequestCtx
		badParams map[string]string
	}
	tests := []struct {
		name string
		d    DtoField
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.CheckParams(tt.args.ctx, tt.args.badParams), "CheckParams(%v, %v)", tt.args.ctx, tt.args.badParams)
		})
	}
}

func TestDtoField_Expect(t *testing.T) {
	tests := []struct {
		name string
		d    DtoField
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.Expect(), "Expect()")
		})
	}
}

func TestDtoField_Format(t *testing.T) {
	tests := []struct {
		name string
		d    DtoField
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.Format(), "Format()")
		})
	}
}

func TestDtoField_GetValue(t *testing.T) {
	tests := []struct {
		name string
		d    DtoField
		want any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.GetValue(), "GetValue()")
		})
	}
}

func TestDtoField_NewValue(t *testing.T) {
	tests := []struct {
		name string
		d    DtoField
		want any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.NewValue(), "NewValue()")
		})
	}
}

func TestDtoField_RequestType(t *testing.T) {
	tests := []struct {
		name string
		d    DtoField
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.RequestType(), "RequestType()")
		})
	}
}

func TestDtoFileField_CheckParams(t *testing.T) {
	type args struct {
		ctx       *fasthttp.RequestCtx
		badParams map[string]string
	}
	tests := []struct {
		name string
		d    DtoFileField
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.CheckParams(tt.args.ctx, tt.args.badParams), "CheckParams(%v, %v)", tt.args.ctx, tt.args.badParams)
		})
	}
}

func TestDtoFileField_Expect(t *testing.T) {
	tests := []struct {
		name string
		d    DtoFileField
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.Expect(), "Expect()")
		})
	}
}

func TestDtoFileField_Format(t *testing.T) {
	tests := []struct {
		name string
		d    DtoFileField
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.Format(), "Format()")
		})
	}
}

func TestDtoFileField_GetValue(t *testing.T) {
	tests := []struct {
		name string
		d    DtoFileField
		want interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.GetValue(), "GetValue()")
		})
	}
}

func TestDtoFileField_NewValue(t *testing.T) {
	tests := []struct {
		name string
		d    DtoFileField
		want interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.NewValue(), "NewValue()")
		})
	}
}

func TestDtoFileField_RequestType(t *testing.T) {
	tests := []struct {
		name string
		d    DtoFileField
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.RequestType(), "RequestType()")
		})
	}
}

func TestEncodeDateString(t *testing.T) {
	tests := []struct {
		name string
		args DateString
		want string
	}{
		// TODO: Add test cases.
		{
			"simple",
			DateString(time.Now()),
			`"` + time.Now().Format(time.DateOnly) + `"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			stream := jsoniter.NewStream(jsoniter.ConfigDefault, buf, 0)
			EncodeDateString(unsafe.Pointer(&tt.args), stream)
			err := stream.Flush()
			if err != nil {
				logs.ErrorLog(err, "during flash")
			}
			assert.Equal(t, tt.want, buf.String())
			buf.Reset()
			json.WriteElement(buf, tt.args)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestIsEmptyDateString(t *testing.T) {
	type args struct {
		ptr unsafe.Pointer
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsEmptyDateString(tt.args.ptr), "IsEmptyDateString(%v)", tt.args.ptr)
		})
	}
}

func TestDTO_NewValue(t *testing.T) {
	type fields struct {
		any pgtype.Point
	}
	test1 := pgtype.Point{
		P:      pgtype.Vec2{1, 2},
		Status: pgtype.Present,
	}
	test2 := pgtype.Point{
		P:      pgtype.Vec2{2, 2},
		Status: pgtype.Present,
	}
	tests := []struct {
		name   string
		fields fields
		want   any
	}{
		{
			name:   "point",
			fields: fields{test1},
			want:   test1,
		},
		{
			name:   "point",
			fields: fields{test2},
			want:   test2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDTO[pgtype.Point](tt.fields.any)

			// chg first value
			tt.fields.any.P.X = 0
			value := d.NewValue()
			assert.Equalf(t, tt.want, value, "NewValue() first value")
			t.Log(value, tt.want, tt.fields.any)

			// cng yearly value
			value = tt.fields.any
			value1 := d.NewValue()
			assert.Equalf(t, tt.want, value1, "NewValue() cng value")
			t.Log(value, value1, tt.want, tt.fields.any)
		})
	}
}

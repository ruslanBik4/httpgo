/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

// A parameter value reaches UnmarshalJSON in two spellings: as a JSON string
// from a JSON body, and as bare text from a form or query parameter
// (TypeInParam.ReadValue passes the parameter's bytes as they are). Every
// wrapper must read both; null is NULL where the type has one.

func TestParamText(t *testing.T) {
	tests := []struct {
		in       string
		want     string
		wantNull bool
	}{
		{`"abc"`, "abc", false},
		{`abc`, "abc", false},
		{`  "abc"  `, "abc", false},
		{`  abc  `, "abc", false},
		{`""`, "", false},
		{``, "", false},
		{`null`, "", true},
		{` null `, "", true},
		{`"null"`, "null", false},
		{`"a\"bA\n"`, "a\"bA\n", false},
		{`"`, `"`, false},
		{`5`, "5", false},
		{`[1,10)`, "[1,10)", false},
		// bare text that only looks like a JSON string
		{`"new york" & "times"`, `"new york" & "times"`, false},
		{`"bad \escape"`, `"bad \escape"`, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, null := paramText([]byte(tt.in))
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantNull, null)
		})
	}

	t.Run("no allocation without escapes", func(t *testing.T) {
		src := []byte(`"01:23:45:67:89:ab"`)
		assert.Zero(t, testing.AllocsPerRun(100, func() { _, _ = paramText(src) }))
	})
}

type unmarshaler interface {
	UnmarshalJSON([]byte) error
}

func TestWrappers_UnmarshalJSON_bothSpellings(t *testing.T) {
	enum := func() unmarshaler { return NewEnumString([]string{"new", "active"}) }

	tests := []struct {
		name  string
		make  func() unmarshaler
		text  string
		check func(t *testing.T, v any)
	}{
		{"InetMarshal", func() unmarshaler { return NewInetMarshal() }, "192.168.0.1", func(t *testing.T, v any) {
			assert.Equal(t, "192.168.0.1", v.(*InetMarshal).Addr.String())
		}},
		{"InetMarshal, zero value", func() unmarshaler { return &InetMarshal{} }, "::1", func(t *testing.T, v any) {
			assert.Equal(t, "::1", v.(*InetMarshal).GetPgxType().String())
		}},
		{"NumrangeMarshal", func() unmarshaler { return new(NumrangeMarshal) }, "[1.5,3.5)", func(t *testing.T, v any) {
			assert.Equal(t, "[1.5,3.5)", fmt.Sprintf("%s", v))
		}},
		{"Int4RangeMarshal", func() unmarshaler { return NewInt4RangeMarshal() }, "[1, 10)", func(t *testing.T, v any) {
			r := v.(*Int4RangeMarshal)
			assert.True(t, r.Valid)
			assert.Equal(t, int32(1), r.Lower)
			assert.Equal(t, int32(10), r.Upper)
		}},
		{"Int8RangeMarshal", func() unmarshaler { return NewInt8RangeMarshal() }, "(,9000000000]", func(t *testing.T, v any) {
			r := v.(*Int8RangeMarshal)
			assert.Equal(t, pgtype.Unbounded, r.LowerType)
			assert.Equal(t, int64(9_000_000_000), r.Upper)
		}},
		{"TsRangeMarshal", func() unmarshaler { return NewTsRangeMarshal() }, "[2024-01-01T00:00:00Z,2024-02-01 00:00:00)", func(t *testing.T, v any) {
			r := v.(*TsRangeMarshal)
			assert.True(t, r.Valid)
			assert.Equal(t, time.February, r.Upper.Month())
		}},
		{"TsTzRangeMarshal", func() unmarshaler { return NewTsTzRangeMarshal() }, "empty", func(t *testing.T, v any) {
			assert.Equal(t, pgtype.Empty, v.(*TsTzRangeMarshal).LowerType)
		}},
		{"PointString", func() unmarshaler { return &PointString{} }, "(1.5,2)", func(t *testing.T, v any) {
			assert.Equal(t, pgtype.Point{P: pgtype.Vec2{X: 1.5, Y: 2}, Valid: true}, v.(*PointString).Point)
		}},
		{"BoxString", func() unmarshaler { return &BoxString{} }, "(3,4),(1,2)", func(t *testing.T, v any) {
			assert.True(t, v.(*BoxString).Valid)
		}},
		{"CircleString", func() unmarshaler { return &CircleString{} }, "<(1,2),3>", func(t *testing.T, v any) {
			assert.Equal(t, 3.0, v.(*CircleString).R)
		}},
		{"TSQueryString", func() unmarshaler { return &TSQueryString{} }, "fat & rat", func(t *testing.T, v any) {
			assert.Equal(t, pgtype.Text{String: "fat & rat", Valid: true}, v.(*TSQueryString).Text)
		}},
		{"TSVectorString", func() unmarshaler { return &TSVectorString{} }, "a:1 b:2", func(t *testing.T, v any) {
			assert.Equal(t, "a:1 b:2", v.(*TSVectorString).String)
		}},
		{"NumericString", func() unmarshaler { return &NumericString{} }, "0.1234567890123456789", func(t *testing.T, v any) {
			assert.Equal(t, "0.1234567890123456789", fmt.Sprintf("%s", v))
		}},
		{"NumericString, NaN", func() unmarshaler { return &NumericString{} }, "NaN", func(t *testing.T, v any) {
			assert.True(t, v.(*NumericString).NaN)
		}},
		{"UUIDString", func() unmarshaler { return &UUIDString{} }, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", func(t *testing.T, v any) {
			assert.Equal(t, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", fmt.Sprintf("%s", v))
		}},
		{"MacaddrString", func() unmarshaler { return &MacaddrString{} }, "01:23:45:67:89:ab", func(t *testing.T, v any) {
			assert.Equal(t, "01:23:45:67:89:ab", v.(*MacaddrString).String())
		}},
		{"EnumString", enum, "active", func(t *testing.T, v any) {
			assert.Equal(t, "active", v.(*EnumString).Value)
		}},
		{"DateString", func() unmarshaler { return &DateString{} }, "2026-01-02", func(t *testing.T, v any) {
			assert.Equal(t, "2026-01-02", time.Time(*v.(*DateString)).Format(time.DateOnly))
		}},
		{"TzString, zero value", func() unmarshaler { return &TzString{} }, "2026-01-02T03:04:05+02:00", func(t *testing.T, v any) {
			assert.Equal(t, 2026, v.(*TzString).GetPgxType().Time.Year())
		}},
		{"TimestampString, zero value", func() unmarshaler { return &TimestampString{} }, "2026-01-02 03:04:05", func(t *testing.T, v any) {
			assert.Equal(t, 3, v.(*TimestampString).GetPgxType().Time.Hour())
		}},
		{"DateRangeMarshal", func() unmarshaler { return NewDateRangeMarshal() }, "[2026-01-01,2026-02-01)", func(t *testing.T, v any) {
			assert.True(t, v.(*DateRangeMarshal).Valid)
		}},
		{"HstoreMarshal is a JSON object in both", func() unmarshaler { return NewHstoreMarshal() }, "", nil},
	}
	for _, tt := range tests {
		if tt.check == nil {
			continue
		}
		t.Run(tt.name+"/form value", func(t *testing.T) {
			v := tt.make()
			require.NoError(t, v.UnmarshalJSON([]byte(tt.text)))
			tt.check(t, v)
		})
		t.Run(tt.name+"/JSON string", func(t *testing.T) {
			quoted, err := json.Marshal(tt.text)
			require.NoError(t, err)

			v := tt.make()
			require.NoError(t, v.UnmarshalJSON(quoted))
			tt.check(t, v)
		})
	}
}

// The text given to UnmarshalJSON belongs to the request: a wrapper that
// keeps a string must keep its own copy.
func TestWrappers_UnmarshalJSON_copyTheText(t *testing.T) {
	t.Run("TSQueryString", func(t *testing.T) {
		src := []byte(`"fat & rat"`)
		var q TSQueryString
		require.NoError(t, q.UnmarshalJSON(src))
		copy(src, `"XXXXXXXXX"`)
		assert.Equal(t, "fat & rat", q.String)
	})
	t.Run("EnumString", func(t *testing.T) {
		src := []byte(`active`)
		e := NewEnumString([]string{"new", "active"})
		require.NoError(t, e.UnmarshalJSON(src))
		copy(src, `XXXXXX`)
		assert.Equal(t, "active", e.Value)
	})
}

func TestWrappers_UnmarshalJSON_null(t *testing.T) {
	t.Run("NumericString", func(t *testing.T) {
		n := &NumericString{pgtype.Numeric{Valid: true}}
		require.NoError(t, n.UnmarshalJSON([]byte(`null`)))
		assert.False(t, n.Valid)
	})
	t.Run("UUIDString", func(t *testing.T) {
		u := &UUIDString{pgtype.UUID{Valid: true}}
		require.NoError(t, u.UnmarshalJSON([]byte(`null`)))
		assert.False(t, u.Valid)
	})
	t.Run("TSVectorString", func(t *testing.T) {
		v := &TSVectorString{pgtype.Text{String: "x", Valid: true}}
		require.NoError(t, v.UnmarshalJSON([]byte(`null`)))
		assert.Equal(t, pgtype.Text{}, v.Text)
	})
	t.Run("PointString", func(t *testing.T) {
		p := &PointString{pgtype.Point{Valid: true}}
		require.NoError(t, p.UnmarshalJSON([]byte(`null`)))
		assert.False(t, p.Valid)
	})
	t.Run("Int4RangeMarshal", func(t *testing.T) {
		r := NewInt4RangeMarshal()
		require.NoError(t, r.UnmarshalJSON([]byte(`"[1,2)"`)))
		require.NoError(t, r.UnmarshalJSON([]byte(`null`)))
		assert.False(t, r.Valid)
	})
	t.Run("MacaddrString", func(t *testing.T) {
		m := &MacaddrString{}
		require.NoError(t, m.UnmarshalJSON([]byte(`null`)))
		assert.Nil(t, m.HardwareAddr)
	})
}

func TestWrappers_UnmarshalJSON_errors(t *testing.T) {
	tests := []struct {
		name string
		v    unmarshaler
		in   string
	}{
		{"UUIDString, short", &UUIDString{}, `"a0eebc99"`},
		{"UUIDString, empty", &UUIDString{}, ``},
		{"NumericString, word", &NumericString{}, `"abc"`},
		{"MacaddrString", &MacaddrString{}, `zz:zz`},
		{"PointString", &PointString{}, `"1;2"`},
		{"Int4RangeMarshal, bound too large", NewInt4RangeMarshal(), `[1,9000000000)`},
		{"Int8RangeMarshal, no brackets", NewInt8RangeMarshal(), `"1,2"`},
		{"EnumString, unknown label", NewEnumString([]string{"new"}), `"old"`},
		{"EnumString, empty", NewEnumString([]string{"new", ""}), `""`},
		{"EnumString, null", NewEnumString([]string{"new"}), `null`},
		{"TzString, clock without a date", NewTzString(), `"15:04"`},
		{"TimestampString, clock without a date", NewTimestampString(), `15:04:05`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.v.UnmarshalJSON([]byte(tt.in)))
		})
	}
}

// A struct field of a wrapper type is a zero value when a JSON decoder
// reaches it: the pointer some wrappers embed is nil then.
func TestWrappers_zeroValueStructFields(t *testing.T) {
	type dto struct {
		Tz       TzString
		Ts       *TimestampString
		Inet     InetMarshal
		Interval IntervalMarshal
		Dates    DateRangeMarshal
	}
	in := []byte(`{"Tz":"2026-01-02T03:04:05Z","Ts":"2026-01-02 03:04:05","Inet":"10.0.0.1",` +
		`"Interval":"01:30","Dates":"[2026-01-01,2026-02-01)"}`)

	decoders := map[string]func([]byte, any) error{
		"encoding/json": json.Unmarshal,
		"jsoniter":      jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
	}
	for name, unmarshal := range decoders {
		t.Run(name, func(t *testing.T) {
			var d dto
			require.NotPanics(t, func() { require.NoError(t, unmarshal(in, &d)) })

			assert.Equal(t, 2026, d.Tz.GetPgxType().Time.Year())
			require.NotNil(t, d.Ts)
			assert.Equal(t, 5, d.Ts.GetPgxType().Time.Second())
			assert.Equal(t, "10.0.0.1", d.Inet.GetPgxType().String())
			assert.Equal(t, int64(90*usPerMinute), d.Interval.GetPgxType().Microseconds)
			assert.True(t, d.Dates.GetPgxType().Valid)
		})
	}

	t.Run("nothing was read", func(t *testing.T) {
		var d dto
		require.NotPanics(t, func() {
			assert.False(t, d.Tz.GetPgxType().Valid)
			assert.False(t, (&TimestampString{}).GetPgxType().Valid)
			assert.False(t, d.Inet.GetPgxType().IsValid())
			assert.False(t, d.Interval.GetPgxType().Valid)
			assert.False(t, d.Dates.GetPgxType().Valid)
			_ = fmt.Sprintf("%s %s %s %s %s", &d.Tz, &TimestampString{}, &d.Inet, &d.Interval, &d.Dates)
		})
	})
}

// "time" columns use DateTimeString.
func TestDateTimeString_timeOfDay(t *testing.T) {
	tests := []struct {
		in   string
		want int64 // microseconds since midnight
	}{
		{`"15:04:05"`, 15*usPerHour + 4*usPerMinute + 5*usPerSecond},
		{`15:04`, 15*usPerHour + 4*usPerMinute},
		{`"00:00"`, 0},
		{`"15:04:05.25"`, 15*usPerHour + 4*usPerMinute + 5*usPerSecond + 250_000},
		{`"2026-01-02T15:04:05Z"`, 15*usPerHour + 4*usPerMinute + 5*usPerSecond},
		{`"2026-01-02 23:59:59"`, 24*usPerHour - usPerSecond},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			var d DateTimeString
			require.NoError(t, d.UnmarshalJSON([]byte(tt.in)))
			assert.Equal(t, pgtype.Time{Microseconds: tt.want, Valid: true}, d.GetPgxType())
		})
	}
}

func TestDateStrings_MarshalJSON(t *testing.T) {
	at := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	dt, d := DateTimeString(at), DateString(at)

	got, err := dt.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `"2026-01-02T15:04:05Z"`, string(got))

	got, err = d.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `"2026-01-02"`, string(got))

	t.Run("encoding/json takes it and reads it back", func(t *testing.T) {
		type dto struct {
			At  *DateTimeString
			Day *DateString
		}
		data, err := json.Marshal(dto{&dt, &d})
		require.NoError(t, err)
		assert.Equal(t, `{"At":"2026-01-02T15:04:05Z","Day":"2026-01-02"}`, string(data))

		var back dto
		require.NoError(t, json.Unmarshal(data, &back))
		assert.True(t, at.Equal(time.Time(*back.At)))
		assert.Equal(t, "2026-01-02", time.Time(*back.Day).Format(time.DateOnly))
	})
}

func TestDtoField_Each(t *testing.T) {
	const body = `{"s":"text","esc":"a\"b","n":42,"big":9007199254740993,"neg":-7,"f":1.5,"e":1e3,` +
		`"yes":true,"no":false,"nil":null,"arr":[1,"two",[3]],"obj":{"k":"v","n":{"deep":1}}}`

	want := DtoField{
		"s": "text", "esc": `a"b`, "n": int64(42), "big": int64(9007199254740993), "neg": int64(-7),
		"f": 1.5, "e": 1000.0, "yes": true, "no": false, "nil": nil,
		"arr": []any{int64(1), "two", []any{int64(3)}},
		"obj": map[string]any{"k": "v", "n": map[string]any{"deep": int64(1)}},
	}

	fill := func(t *testing.T, d *DtoField) {
		t.Helper()
		v, err := fastjson.Parse(body)
		require.NoError(t, err)
		v.GetObject().Visit(d.Each)
	}

	t.Run("NewValue", func(t *testing.T) {
		d := (&DtoField{}).NewValue().(*DtoField)
		fill(t, d)
		assert.Equal(t, want, *d)
	})
	t.Run("nil map", func(t *testing.T) {
		d := new(DtoField)
		require.NotPanics(t, func() { fill(t, d) })
		assert.Equal(t, want, *d)
	})
	t.Run("goes to a json column as it came", func(t *testing.T) {
		d := new(DtoField)
		fill(t, d)

		data, err := json.Marshal(d)
		require.NoError(t, err)
		assert.JSONEq(t, `{"s":"text","esc":"a\"b","n":42,"big":9007199254740993,"neg":-7,"f":1.5,"e":1000,`+
			`"yes":true,"no":false,"nil":null,"arr":[1,"two",[3]],"obj":{"k":"v","n":{"deep":1}}}`, string(data))
	})
}

func TestDtoFileField_CheckParams_brokenHeader(t *testing.T) {
	// a FileHeader without content cannot be opened
	d := DtoFileField{{Filename: "a.txt"}, {Filename: "b.txt"}}
	bad := map[string]string{}

	var ok bool
	require.NotPanics(t, func() { ok = d.CheckParams(nil, bad) })
	if !ok {
		assert.Contains(t, bad, "a.txt")
		assert.Contains(t, bad, "b.txt")
	}
	_ = multipart.FileHeader{}
}

func TestParseRangeLiteral_spaces(t *testing.T) {
	r, err := parseRangeLiteral(" [ 1 , 10 ) ", func(b string) (int, error) {
		var n int
		_, err := fmt.Sscanf(b, "%d", &n)
		return n, err
	})
	require.NoError(t, err)
	assert.Equal(t, pgtype.Range[int]{Lower: 1, Upper: 10, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}, r)
}

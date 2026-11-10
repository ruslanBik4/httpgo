/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for types.go. IntervalMarshal gets the most attention: the cases come
// from real failures ("00:00:00" / "00:00" from <input type="time">, the
// {"Microseconds":..,"Days":..,"Months":..,"Valid":..} object that
// pgtype.Interval marshals to, null and "").
//
// These tests call parseDate and parseRangeLiteral (field_dto.go). They
// assume only what the doc comments in types.go promise: ISO dates
// ("2026-01-01") and Postgres range literals ("[1.5,3.5)", "(1,2]").

// ---------------------------------------------------------------------------
// IntervalMarshal
// ---------------------------------------------------------------------------

func TestIntervalMarshal_Set(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    pgtype.Interval
		wantErr bool
	}{
		{"zero clock", "00:00:00", pgtype.Interval{Valid: true}, false},
		{"zero clock, one digit hour", "0:00:00", pgtype.Interval{Valid: true}, false},
		{"hours and minutes", "01:30:00", pgtype.Interval{Microseconds: 5_400_000_000, Valid: true}, false},
		{"negative", "-01:30:00", pgtype.Interval{Microseconds: -5_400_000_000, Valid: true}, false},
		{"fraction of second", "00:00:01.5", pgtype.Interval{Microseconds: 1_500_000, Valid: true}, false},
		{"day and clock", "1 day 02:03:04", pgtype.Interval{Days: 1, Microseconds: 7_384_000_000, Valid: true}, false},
		{"days only", "3 days", pgtype.Interval{Days: 3, Valid: true}, false},
		{"month only", "1 mon", pgtype.Interval{Months: 1, Valid: true}, false},
		{"years and mons are folded into months", "2 years 3 mons 4 days 05:06:07",
			pgtype.Interval{Months: 27, Days: 4, Microseconds: 18_367_000_000, Valid: true}, false},
		{"nil is NULL", nil, pgtype.Interval{}, false},

		// pgx's text parser wants exactly H:M:S - Set adds the missing ":00"
		// (what <input type="time"> sends) and takes "" as NULL
		{"short clock gets :00", "00:00", pgtype.Interval{Valid: true}, false},
		{"short clock hours and minutes", "1:30", pgtype.Interval{Microseconds: 5_400_000_000, Valid: true}, false},
		{"empty string is NULL", "", pgtype.Interval{}, false},
		{"blank string is NULL", "  ", pgtype.Interval{}, false},
		{"garbage", "abc", pgtype.Interval{}, true},
		{"ISO 8601 is not supported", "PT0S", pgtype.Interval{}, true},
		{"int is not supported", 5, pgtype.Interval{}, true},
		{"[]byte is not supported", []byte("00:00:00"), pgtype.Interval{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewIntervalMarshal()
			err := i.Set(tt.src)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, *i.Interval)
		})
	}
}

func TestIntervalMarshal_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    pgtype.Interval
		wantErr bool
	}{
		// --- JSON string -> pgtype.Interval.Scan
		{"zero clock", `"00:00:00"`, pgtype.Interval{Valid: true}, false},
		{"zero clock with spaces around", `"  00:00:00 "`, pgtype.Interval{Valid: true}, false},
		{"hours:minutes:seconds", `"01:30:00"`, pgtype.Interval{Microseconds: 5_400_000_000, Valid: true}, false},
		{"negative", `"-01:30:00"`, pgtype.Interval{Microseconds: -5_400_000_000, Valid: true}, false},
		{"day and clock", `"1 day 00:00:00"`, pgtype.Interval{Days: 1, Valid: true}, false},
		{"days word only", `"2 days"`, pgtype.Interval{Days: 2, Valid: true}, false},

		// --- short clock: <input type="time"> drops zero seconds
		{"short zero clock", `"00:00"`, pgtype.Interval{Valid: true}, false},
		{"short clock", `"1:30"`, pgtype.Interval{Microseconds: 5_400_000_000, Valid: true}, false},
		{"short negative clock", `"-01:30"`, pgtype.Interval{Microseconds: -5_400_000_000, Valid: true}, false},

		// --- NULL
		{"null", `null`, pgtype.Interval{}, false},
		{"empty string", `""`, pgtype.Interval{}, false},
		{"blank string", `"   "`, pgtype.Interval{}, false},

		// --- object: what json.Marshal(pgtype.Interval) produces
		{"object, zero", `{"Microseconds":0,"Days":0,"Months":0,"Valid":true}`, pgtype.Interval{Valid: true}, false},
		{"object, all fields", `{"Microseconds":5400000000,"Days":2,"Months":1,"Valid":true}`,
			pgtype.Interval{Microseconds: 5_400_000_000, Days: 2, Months: 1, Valid: true}, false},
		{"object, negative", `{"Microseconds":-1000000,"Days":-1,"Months":-2,"Valid":true}`,
			pgtype.Interval{Microseconds: -1_000_000, Days: -1, Months: -2, Valid: true}, false},
		{"object, Valid=false stays false", `{"Microseconds":0,"Days":0,"Months":0,"Valid":false}`, pgtype.Interval{}, false},
		{"object, lower case keys", `{"microseconds":7,"days":8,"months":9,"valid":true}`,
			pgtype.Interval{Microseconds: 7, Days: 8, Months: 9, Valid: true}, false},
		{"object, missing Valid means a real value", `{"Days":3}`, pgtype.Interval{Days: 3, Valid: true}, false},
		{"object, null fields", `{"Days":null,"Valid":null}`, pgtype.Interval{Valid: true}, false},
		{"empty object", `{}`, pgtype.Interval{Valid: true}, false},

		// --- errors
		{"garbage string", `"abc"`, pgtype.Interval{}, true},
		{"ISO 8601 string", `"PT0S"`, pgtype.Interval{}, true},
		{"number", `5`, pgtype.Interval{}, true},
		{"bool", `true`, pgtype.Interval{}, true},
		{"array", `[1,2]`, pgtype.Interval{}, true},
		{"object, field is not a number", `{"Days":"x"}`, pgtype.Interval{}, true},
		{"object, Valid is not a bool", `{"Valid":"yes"}`, pgtype.Interval{}, true},
		{"broken json", `{"Days":`, pgtype.Interval{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// called directly, so even invalid JSON reaches UnmarshalJSON
			i := NewIntervalMarshal()
			err := i.UnmarshalJSON([]byte(tt.in))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, *i.Interval)
		})
	}
}

// NULL must clear a value that was set before, not leave it behind.
func TestIntervalMarshal_UnmarshalJSON_overwritesOldValue(t *testing.T) {
	i := NewIntervalMarshal()
	require.NoError(t, i.UnmarshalJSON([]byte(`"1 day 02:03:04"`)))
	require.True(t, i.Valid)

	require.NoError(t, i.UnmarshalJSON([]byte(`null`)))
	assert.Equal(t, pgtype.Interval{}, *i.Interval)

	require.NoError(t, i.UnmarshalJSON([]byte(`"05:00:00"`)))
	require.NoError(t, i.UnmarshalJSON([]byte(`{"Days":1,"Valid":true}`)))
	assert.Equal(t, pgtype.Interval{Days: 1, Valid: true}, *i.Interval)
}

// A zero-value IntervalMarshal has a nil embedded *pgtype.Interval (that is
// what encoding/json creates for a struct field): UnmarshalJSON must allocate
// it instead of panicking.
func TestIntervalMarshal_UnmarshalJSON_zeroValueField(t *testing.T) {
	var dst struct {
		Plain  IntervalMarshal
		Ptr    *IntervalMarshal
		Object IntervalMarshal
		Null   IntervalMarshal
	}
	in := `{"Plain":"00:00:00","Ptr":"1 day","Object":{"Days":2,"Valid":true},"Null":null}`

	require.NotPanics(t, func() { require.NoError(t, json.Unmarshal([]byte(in), &dst)) })

	require.NotNil(t, dst.Plain.Interval)
	assert.Equal(t, pgtype.Interval{Valid: true}, *dst.Plain.Interval)
	require.NotNil(t, dst.Ptr)
	assert.Equal(t, pgtype.Interval{Days: 1, Valid: true}, *dst.Ptr.Interval)
	assert.Equal(t, pgtype.Interval{Days: 2, Valid: true}, *dst.Object.Interval)
	// encoding/json does call UnmarshalJSON(null) for a non-pointer field
	require.NotNil(t, dst.Null.Interval)
	assert.Equal(t, pgtype.Interval{}, *dst.Null.Interval)
}

// The regression that started all this: the value goes out as the struct
// (IntervalMarshal has no MarshalJSON) and comes back in the next request.
func TestIntervalMarshal_JSONRoundTrip(t *testing.T) {
	for _, iv := range []pgtype.Interval{
		{Valid: true},
		{Microseconds: 5_400_000_000, Valid: true},
		{Microseconds: -1, Valid: true},
		{Days: 1, Microseconds: 7_384_000_000, Valid: true},
		{Months: 27, Days: 4, Microseconds: 18_367_500_000, Valid: true},
		{},
	} {
		t.Run(fmt.Sprintf("%+v", iv), func(t *testing.T) {
			src := &IntervalMarshal{&iv}

			data, err := json.Marshal(src)
			require.NoError(t, err)

			back := NewIntervalMarshal()
			require.NoError(t, json.Unmarshal(data, back), string(data))
			assert.Equal(t, iv, *back.Interval, string(data))
		})
	}
}

// Everything a client may send for "no time", read back through a form.
func TestIntervalMarshal_zeroSpellings(t *testing.T) {
	for _, in := range []string{
		`"00:00:00"`, `"0:00:00"`, `"00:00"`, `"0:0"`,
		`{"Microseconds":0,"Days":0,"Months":0,"Valid":true}`,
	} {
		t.Run(in, func(t *testing.T) {
			i := NewIntervalMarshal()
			require.NoError(t, json.Unmarshal([]byte(in), i))
			assert.True(t, i.Valid, "a zero interval is a value, not NULL")
			assert.Zero(t, i.Months)
			assert.Zero(t, i.Days)
			assert.Zero(t, i.Microseconds)
		})
	}
}

func TestIntervalMarshal_accessors(t *testing.T) {
	i := NewIntervalMarshal()
	require.NotNil(t, i.Interval)
	require.NoError(t, i.Set("1 day 02:03:04"))

	t.Run("GetValue is the wrapper itself", func(t *testing.T) {
		assert.Same(t, i, i.GetValue())
	})
	t.Run("NewValue is a fresh, ready to use value", func(t *testing.T) {
		v, ok := i.NewValue().(*IntervalMarshal)
		require.True(t, ok)
		require.NotNil(t, v.Interval)
		assert.NotSame(t, i, v)
		assert.NotSame(t, i.Interval, v.Interval)
		assert.False(t, v.Valid)
		require.NoError(t, v.Set("00:00:00")) // must not panic
	})
	t.Run("GetPgxType is a copy", func(t *testing.T) {
		got := i.GetPgxType()
		assert.Equal(t, *i.Interval, got)
		got.Days = 99
		assert.Equal(t, int32(1), i.Days)
	})
}

// IntervalMarshal embeds *pgtype.Interval, so database/sql's driver.Valuer
// is promoted: this is the value pgx/sql gets as a query argument.
func TestIntervalMarshal_Value(t *testing.T) {
	tests := []struct {
		src, want string
	}{
		{"00:00:00", "00:00:00"},
		{"01:30:00", "01:30:00"},
		{"-01:30:00", "-01:30:00"},
		{"1 day 02:03:04", "1 day 02:03:04"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			i := NewIntervalMarshal()
			require.NoError(t, i.Set(tt.src))
			got, err := i.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("NULL", func(t *testing.T) {
		got, err := NewIntervalMarshal().Value()
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestIntervalMarshal_Format(t *testing.T) {
	i := &IntervalMarshal{&pgtype.Interval{Months: 1, Days: 2, Microseconds: 3, Valid: true}}

	assert.Equal(t, "*crud.IntervalMarshal", fmt.Sprintf("%t", i))
	assert.Equal(t, "crud.NewIntervalMarshal()", fmt.Sprintf("%g", i))
	// microseconds are 0 here: see the known issue about the unit below
	i.Microseconds = 0
	assert.Equal(t, "1 month 2 day 0s", fmt.Sprintf("%s", i))
}

// ---------------------------------------------------------------------------
// InetMarshal
// ---------------------------------------------------------------------------

func TestInetMarshal_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"ipv4", `"192.168.0.1"`, "192.168.0.1", false},
		{"ipv4 zero", `"0.0.0.0"`, "0.0.0.0", false},
		{"ipv6", `"::1"`, "::1", false},
		{"ipv6 full", `"2001:db8::8a2e:370:7334"`, "2001:db8::8a2e:370:7334", false},

		{"empty string", `""`, "", true},
		{"not an address", `"not-an-ip"`, "", true},
		{"octet out of range", `"256.1.1.1"`, "", true},
		{"number", `5`, "", true},
		{"null for a string is accepted by encoding/json, ParseAddr rejects \"\"", `null`, "", true},
		// InetMarshal keeps only a netip.Addr, so a network (cidr) has no place to go
		{"cidr literal is not supported yet", `"10.0.0.0/8"`, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewInetMarshal()
			err := i.UnmarshalJSON([]byte(tt.in))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, i.Addr.String())
		})
	}
}

func TestInetMarshal_accessors(t *testing.T) {
	i := NewInetMarshal()
	require.NotNil(t, i.Addr)
	require.NoError(t, i.UnmarshalJSON([]byte(`"10.1.2.3"`)))

	assert.Same(t, i, i.GetValue())

	v, ok := i.NewValue().(*InetMarshal)
	require.True(t, ok)
	require.NotNil(t, v.Addr)
	assert.NotSame(t, i, v)
	assert.False(t, v.Addr.IsValid())

	got := i.GetPgxType()
	assert.Equal(t, netip.MustParseAddr("10.1.2.3"), got)
}

func TestInetMarshal_Format(t *testing.T) {
	i := NewInetMarshal()
	assert.Equal(t, "*crud.InetMarshal", fmt.Sprintf("%t", i))
	assert.Equal(t, "crud.NewInetMarshal()", fmt.Sprintf("%g", i))

	assert.Equal(t, "", fmt.Sprintf("%s", i), "an unset address prints nothing")
	assert.Equal(t, "", fmt.Sprintf("%s", &InetMarshal{}), "a nil embedded *Addr must not panic")

	require.NoError(t, i.UnmarshalJSON([]byte(`"::1"`)))
	assert.Equal(t, "::1", fmt.Sprintf("%s", i))
}

// ---------------------------------------------------------------------------
// DateRangeMarshal
// ---------------------------------------------------------------------------

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestDateRangeMarshal_docs(t *testing.T) {
	d := NewDateRangeMarshal()
	assert.Equal(t, "string", d.Expect())
	assert.Equal(t, "string", d.RequestType())
	assert.Equal(t, "date-range", d.FormatDoc())
}

func TestDateRangeMarshal_UnmarshalJSON(t *testing.T) {
	jan1, feb1 := date(2026, 1, 1), date(2026, 2, 1)

	tests := []struct {
		name         string
		in           string
		wantValid    bool
		lower, upper time.Time
		lt, ut       pgtype.BoundType
		wantErr      bool
	}{
		{"half open", `"[2026-01-01,2026-02-01)"`, true, jan1, feb1, pgtype.Inclusive, pgtype.Exclusive, false},
		{"other way round", `"(2026-01-01,2026-02-01]"`, true, jan1, feb1, pgtype.Exclusive, pgtype.Inclusive, false},
		{"closed", `"[2026-01-01,2026-02-01]"`, true, jan1, feb1, pgtype.Inclusive, pgtype.Inclusive, false},
		{"single date becomes one-day range", `"2026-01-01"`, true, jan1, jan1, pgtype.Inclusive, pgtype.Inclusive, false},
		{"single date with spaces", `" 2026-01-01 "`, true, jan1, jan1, pgtype.Inclusive, pgtype.Inclusive, false},
		{"empty range", `"empty"`, true, time.Time{}, time.Time{}, pgtype.Empty, pgtype.Empty, false},
		{"empty string is not a value", `""`, false, time.Time{}, time.Time{}, 0, 0, false},

		{"junk", `"junk"`, false, time.Time{}, time.Time{}, 0, 0, true},
		{"not a date", `"2026-13-45"`, false, time.Time{}, time.Time{}, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDateRangeMarshal()
			err := d.UnmarshalJSON([]byte(tt.in))
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not a range")
				assert.False(t, d.Valid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, d.Valid)
			if !tt.wantValid {
				return
			}
			assert.Equal(t, tt.lt, d.LowerType)
			assert.Equal(t, tt.ut, d.UpperType)
			if tt.lt != pgtype.Empty {
				assert.True(t, tt.lower.Equal(d.Lower.Time), "lower %v != %v", tt.lower, d.Lower.Time)
				assert.True(t, tt.upper.Equal(d.Upper.Time), "upper %v != %v", tt.upper, d.Upper.Time)
			}
		})
	}
}

// A zero-value DateRangeMarshal has a nil embedded *pgtype.Range.
func TestDateRangeMarshal_UnmarshalJSON_zeroValue(t *testing.T) {
	var dst struct{ R DateRangeMarshal }

	require.NotPanics(t, func() {
		require.NoError(t, json.Unmarshal([]byte(`{"R":"[2026-01-01,2026-02-01)"}`), &dst))
	})
	require.NotNil(t, dst.R.Range)
	assert.True(t, dst.R.Valid)
	assert.True(t, date(2026, 2, 1).Equal(dst.R.Upper.Time))
}

func TestDateRangeMarshal_accessors(t *testing.T) {
	d := NewDateRangeMarshal()
	require.NoError(t, d.UnmarshalJSON([]byte(`"[2026-01-01,2026-02-01)"`)))

	assert.Same(t, d.Range, d.GetValue())
	assert.Same(t, d.Range, d.Get())

	v, err := d.Value()
	require.NoError(t, err)
	assert.Same(t, d.Range, v)

	got := d.GetPgxType()
	assert.Equal(t, *d.Range, got)

	nv, ok := d.NewValue().(*DateRangeMarshal)
	require.True(t, ok)
	require.NotNil(t, nv.Range)
	assert.NotSame(t, d, nv)
	assert.False(t, nv.Valid)
}

func TestDateRangeMarshal_Format(t *testing.T) {
	d := NewDateRangeMarshal()
	require.NoError(t, d.UnmarshalJSON([]byte(`"[2026-01-01,2026-02-01)"`)))

	assert.Equal(t, "*crud.DateRangeMarshal", fmt.Sprintf("%t", d))
	assert.Equal(t, "crud.NewDateRangeMarshal()", fmt.Sprintf("%g", d))
	assert.NotPanics(t, func() {
		_ = fmt.Sprintf("%s", d)
		_ = fmt.Sprintf("%v", d)
	})
}

func TestLowerUpperBoundType(t *testing.T) {
	lowers := []struct {
		in   string
		bt   pgtype.BoundType
		rest string
	}{
		{"[2026-01-01", pgtype.Inclusive, "2026-01-01"},
		{"(2026-01-01", pgtype.Exclusive, "2026-01-01"},
		{"2026-01-01", pgtype.Inclusive, "2026-01-01"}, // default
	}
	for _, tt := range lowers {
		bt, rest := lowerBoundType(tt.in)
		assert.Equal(t, tt.bt, bt, tt.in)
		assert.Equal(t, tt.rest, rest, tt.in)
	}

	uppers := []struct {
		in   string
		bt   pgtype.BoundType
		rest string
	}{
		{"2026-02-01]", pgtype.Inclusive, "2026-02-01"},
		{"2026-02-01)", pgtype.Exclusive, "2026-02-01"},
		{"2026-02-01", pgtype.Inclusive, "2026-02-01"}, // default
	}
	for _, tt := range uppers {
		bt, rest := upperBoundType(tt.in)
		assert.Equal(t, tt.bt, bt, tt.in)
		assert.Equal(t, tt.rest, rest, tt.in)
	}
}

// DecodeDateRangeMarshal is registered for jsoniter in init().
func TestDecodeDateRangeMarshal_jsoniter(t *testing.T) {
	jan1, feb1 := date(2026, 1, 1), date(2026, 2, 1)

	tests := []struct {
		name         string
		in           string
		wantValid    bool
		lower, upper time.Time
		lt, ut       pgtype.BoundType
	}{
		{"two dates", `"2026-01-01,2026-02-01"`, true, jan1, feb1, pgtype.Inclusive, pgtype.Inclusive},
		{"brackets", `"[2026-01-01,2026-02-01)"`, true, jan1, feb1, pgtype.Inclusive, pgtype.Exclusive},
		{"parens", `"(2026-01-01,2026-02-01]"`, true, jan1, feb1, pgtype.Exclusive, pgtype.Inclusive},
		{"spaces after comma", `"2026-01-01, 2026-02-01"`, true, jan1, feb1, pgtype.Inclusive, pgtype.Inclusive},
		{"one date is a one-day range", `"2026-01-01"`, true, jan1, jan1, pgtype.Inclusive, pgtype.Inclusive},
		{"empty string", `""`, false, time.Time{}, time.Time{}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDateRangeMarshal()
			require.NoError(t, jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(tt.in), d))

			assert.Equal(t, tt.wantValid, d.Valid)
			if !tt.wantValid {
				return
			}
			assert.Equal(t, tt.lt, d.LowerType)
			assert.Equal(t, tt.ut, d.UpperType)
			assert.True(t, tt.lower.Equal(d.Lower.Time), "lower %v != %v", tt.lower, d.Lower.Time)
			assert.True(t, tt.upper.Equal(d.Upper.Time), "upper %v != %v", tt.upper, d.Upper.Time)
		})
	}

	t.Run("garbage leaves the range invalid", func(t *testing.T) {
		d := NewDateRangeMarshal()
		_ = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(`"not-a-date"`), d)
		assert.False(t, d.Valid)
	})
}

// ---------------------------------------------------------------------------
// NumRangeMarshal (pointer wrapper) and NumrangeMarshal (defined type)
// ---------------------------------------------------------------------------

func TestNumRangeMarshal_accessors(t *testing.T) {
	n := &NumRangeMarshal{&pgtype.Range[pgtype.Numeric]{Valid: true}}

	assert.Same(t, n.Range, n.GetValue())

	nv, ok := n.NewValue().(*NumRangeMarshal)
	require.True(t, ok)
	require.NotNil(t, nv.Range)
	assert.NotSame(t, n, nv)
	assert.NotSame(t, n.Range, nv.Range)
	assert.False(t, nv.Valid)
}

func numF(t *testing.T, n pgtype.Numeric) float64 {
	t.Helper()
	f, err := n.Float64Value()
	require.NoError(t, err)
	require.True(t, f.Valid)
	return f.Float64
}

func TestNumrangeMarshal_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		lower, upper float64
		lt, ut       pgtype.BoundType
		wantErr      bool
	}{
		{"half open", `"[1.5,3.5)"`, 1.5, 3.5, pgtype.Inclusive, pgtype.Exclusive, false},
		{"other way round", `"(1,2]"`, 1, 2, pgtype.Exclusive, pgtype.Inclusive, false},
		{"negative", `"[-10.25,0]"`, -10.25, 0, pgtype.Inclusive, pgtype.Inclusive, false},

		{"not a range", `"abc"`, 0, 0, 0, 0, true},
		{"number instead of string", `5`, 0, 0, 0, 0, true},
		{"bound is not a number", `"[a,b)"`, 0, 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := new(NumrangeMarshal)
			err := json.Unmarshal([]byte(tt.in), n)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.True(t, n.Valid)
			assert.Equal(t, tt.lt, n.LowerType)
			assert.Equal(t, tt.ut, n.UpperType)
			assert.InDelta(t, tt.lower, numF(t, n.Lower), 1e-9)
			assert.InDelta(t, tt.upper, numF(t, n.Upper), 1e-9)
		})
	}
}

// Numeric bounds are parsed from text, not through float64.
func TestNumrangeMarshal_UnmarshalJSON_precision(t *testing.T) {
	n := new(NumrangeMarshal)
	require.NoError(t, json.Unmarshal([]byte(`"[0.1234567890123456789,2)"`), n))

	assert.Equal(t, "0.1234567890123456789", fmt.Sprint(mustNumericString(t, n.Lower)))
}

func mustNumericString(t *testing.T, n pgtype.Numeric) string {
	t.Helper()
	v, err := n.Value()
	require.NoError(t, err)
	s, ok := v.(string)
	require.True(t, ok, "%T", v)
	return s
}

func TestNumrangeMarshal_accessors(t *testing.T) {
	n := new(NumrangeMarshal)
	require.NoError(t, json.Unmarshal([]byte(`"[1,2)"`), n))

	assert.Equal(t, "float", n.Expect())
	assert.Equal(t, "float", n.RequestType())
	assert.Equal(t, "num_range", n.FormatDoc())

	val, ok := n.GetValue().(pgtype.Range[pgtype.Numeric])
	require.True(t, ok, "GetValue returns the pgtype.Range by value")
	assert.Equal(t, pgtype.Range[pgtype.Numeric](*n), val)
	assert.Equal(t, val, n.GetPgxType())

	nv, ok := n.NewValue().(*NumrangeMarshal)
	require.True(t, ok)
	assert.NotSame(t, n, nv)
	assert.False(t, nv.Valid)
}

func TestNumrangeMarshal_Format(t *testing.T) {
	n := new(NumrangeMarshal)
	assert.Equal(t, "*crud.NumrangeMarshal", fmt.Sprintf("%t", n))
	assert.Equal(t, "&crud.NumrangeMarshal{}", fmt.Sprintf("%g", n))
	assert.NotPanics(t, func() { _ = fmt.Sprintf("%s", n) })
}

// ---------------------------------------------------------------------------
// Known problems found while writing these tests. Each subtest describes the
// behaviour it should have and is skipped until the code is fixed - remove the
// t.Skip line after fixing and the test becomes a regression test.
// ---------------------------------------------------------------------------

func TestKnownIssues(t *testing.T) {
	t.Run("DecodeDateRangeMarshal panics for a nil *DateRangeMarshal struct field", func(t *testing.T) {
		t.Skip("KNOWN BUG: jsoniter allocates a zero DateRangeMarshal whose embedded *pgtype.Range is nil; " +
			"DecodeDateRangeMarshal dereferences it. UnmarshalJSON guards this (d.Range == nil), the jsoniter decoder does not.")

		var dst struct{ R *DateRangeMarshal }
		require.NotPanics(t, func() {
			require.NoError(t, jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(`{"R":"2026-01-01"}`), &dst))
		})
		require.NotNil(t, dst.R)
		assert.True(t, dst.R.Valid)
	})

	t.Run("DecodeDateRangeMarshal array form [lower, upper]", func(t *testing.T) {
		t.Skip("KNOWN BUG: the ArrayValue branch calls iter.Read() before ReadArrayCB (consumes the whole array), " +
			"and v := val.Lower is a copy, so Lower/Upper are never written back.")

		d := NewDateRangeMarshal()
		require.NoError(t, jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(`["2026-01-01","2026-02-01"]`), d))
		assert.True(t, date(2026, 1, 1).Equal(d.Lower.Time))
		assert.True(t, date(2026, 2, 1).Equal(d.Upper.Time))
	})

	t.Run("InetMarshal accepts a cidr network", func(t *testing.T) {
		t.Skip("KNOWN GAP: the doc comment says it covers inet and cidr, but netip.ParseAddr rejects \"10.0.0.0/8\". " +
			"Either parse with netip.ParsePrefix (and keep the prefix) or drop cidr from the comment.")

		i := NewInetMarshal()
		require.NoError(t, i.UnmarshalJSON([]byte(`"10.0.0.0/8"`)))
	})

	t.Run("NumrangeMarshal Format verb s prints the upper bound", func(t *testing.T) {
		t.Skip("KNOWN BUG: Format's 's' case prints n.UpperType twice and never n.Upper.")

		n := new(NumrangeMarshal)
		require.NoError(t, json.Unmarshal([]byte(`"[1,99)"`), n))
		assert.Contains(t, fmt.Sprintf("%s", n), "99")
	})

	t.Run("IntervalMarshal Format verb s uses the right unit", func(t *testing.T) {
		t.Skip("KNOWN BUG: Format's 's' case does time.Duration(i.Microseconds).String() - a Duration counts " +
			"nanoseconds, so 1.5 s (1_500_000 us) prints as 1.5ms. Use time.Duration(i.Microseconds) * time.Microsecond.")

		i := &IntervalMarshal{&pgtype.Interval{Microseconds: 1_500_000, Valid: true}}
		assert.Equal(t, "0 month 0 day 1.5s", fmt.Sprintf("%s", i))
	})

	t.Run("IntervalMarshal Format verb v prints something", func(t *testing.T) {
		t.Skip("KNOWN GAP: Format has no default case, so the v verb, fmt.Sprint and log lines print an empty string; " +
			"'s' prints raw microseconds ('0 month 1 day 7384000000').")

		i := NewIntervalMarshal()
		require.NoError(t, i.Set("1 day 02:03:04"))
		assert.NotEmpty(t, fmt.Sprintf("%v", i))
	})
}

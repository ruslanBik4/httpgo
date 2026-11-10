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
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of IntervalMarshal.Scan and the parser behind it: it must read what
// pgtype.Interval.Scan reads, with the same result, and allocate nothing.

func TestIntervalMarshal_Scan(t *testing.T) {
	const hour = int64(usPerHour)

	tests := []struct {
		name    string
		src     any
		want    pgtype.Interval
		wantErr bool
	}{
		{"clock", "02:03:04", pgtype.Interval{Microseconds: 7_384_000_000, Valid: true}, false},
		{"clock with plus", "+02:03:04", pgtype.Interval{Microseconds: 7_384_000_000, Valid: true}, false},
		{"hours are not held under 24", "100:00:00", pgtype.Interval{Microseconds: 100 * hour, Valid: true}, false},
		{"minutes are not held under 60, as in pgx", "00:90:00", pgtype.Interval{Microseconds: 90 * usPerMinute, Valid: true}, false},
		{"one digit of fraction", "00:00:00.5", pgtype.Interval{Microseconds: 500_000, Valid: true}, false},
		{"six digits of fraction", "00:00:00.000001", pgtype.Interval{Microseconds: 1, Valid: true}, false},
		{"digits after the sixth are dropped", "00:00:00.1234567", pgtype.Interval{Microseconds: 123_456, Valid: true}, false},
		{"negative clock with fraction", "-00:00:01.25", pgtype.Interval{Microseconds: -1_250_000, Valid: true}, false},
		{"short clock", "02:03", pgtype.Interval{Microseconds: 2*hour + 3*usPerMinute, Valid: true}, false},

		{"singular units", "1 year 1 mon 1 day", pgtype.Interval{Months: 13, Days: 1, Valid: true}, false},
		{"plural units", "2 years 2 mons 2 days", pgtype.Interval{Months: 26, Days: 2, Valid: true}, false},
		{"signs, as Postgres writes a mixed interval", "-1 years -2 mons +3 days -04:05:06",
			pgtype.Interval{Months: -14, Days: 3, Microseconds: -(4*hour + 5*usPerMinute + 6*usPerSecond), Valid: true}, false},
		{"more than one space between the words", "1 day   02:03:04", pgtype.Interval{Days: 1, Microseconds: 7_384_000_000, Valid: true}, false},
		{"spaces around", "  3 days  ", pgtype.Interval{Days: 3, Valid: true}, false},
		{"largest months", "2147483647 mons", pgtype.Interval{Months: math.MaxInt32, Valid: true}, false},

		{"nil", nil, pgtype.Interval{}, false},
		{"empty", "", pgtype.Interval{}, false},
		{"[]byte", []byte("1 mon"), pgtype.Interval{Months: 1, Valid: true}, false},

		{"number without a unit", "5", pgtype.Interval{}, true},
		{"unit without a number", "day", pgtype.Interval{}, true},
		{"unknown unit", "5 weeks", pgtype.Interval{}, true},
		{"word instead of a number", "one day", pgtype.Interval{}, true},
		{"unit after the clock", "02:03:04 day", pgtype.Interval{}, true},
		{"four clock fields", "01:02:03:04", pgtype.Interval{}, true},
		{"empty hour", ":02:03", pgtype.Interval{}, true},
		{"empty seconds", "01:02:", pgtype.Interval{}, true},
		{"empty fraction", "01:02:03.", pgtype.Interval{}, true},
		{"letters in the fraction", "01:02:03.5x", pgtype.Interval{}, true},
		{"sign inside the clock", "01:-02:03", pgtype.Interval{}, true},
		{"sign only", "-", pgtype.Interval{}, true},
		{"months do not fit an int32", "2147483648 mons", pgtype.Interval{}, true},
		{"years do not fit an int32 of months", "178956971 years", pgtype.Interval{}, true},
		{"days do not fit an int32", "-2147483649 days", pgtype.Interval{}, true},
		{"number does not fit an int64", "99999999999999999999 days", pgtype.Interval{}, true},
		{"hours do not fit", "2562047789:00:00", pgtype.Interval{}, true},
		{"the clock does not fit", "2562047788:59:00", pgtype.Interval{}, true},
		{"minutes do not fit", "00:2147483648:00", pgtype.Interval{}, true},
		{"int", 5, pgtype.Interval{}, true},
		{"pgtype.Interval is not a database/sql value", pgtype.Interval{Valid: true}, pgtype.Interval{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := pgtype.Interval{Months: 7, Days: 7, Microseconds: 7, Valid: true}
			i := &IntervalMarshal{new(old)}

			err := i.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, old, *i.Interval, "the value is kept when src does not parse")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, *i.Interval)
		})
	}
}

// A zero-value IntervalMarshal has a nil embedded *pgtype.Interval: every
// method must take it.
func TestIntervalMarshal_zeroValue(t *testing.T) {
	t.Run("Scan", func(t *testing.T) {
		var i IntervalMarshal
		require.NoError(t, i.Scan("1 day"))
		assert.Equal(t, pgtype.Interval{Days: 1, Valid: true}, *i.Interval)
	})
	t.Run("Scan nil", func(t *testing.T) {
		var i IntervalMarshal
		require.NoError(t, i.Scan(nil))
		require.NotNil(t, i.Interval)
		assert.False(t, i.Valid)
	})
	t.Run("ScanInterval", func(t *testing.T) {
		var i IntervalMarshal
		var _ pgtype.IntervalScanner = &i
		require.NoError(t, i.ScanInterval(pgtype.Interval{Days: 2, Valid: true}))
		assert.Equal(t, int32(2), i.Days)
	})
	t.Run("query argument", func(t *testing.T) {
		var i IntervalMarshal
		var _ pgtype.IntervalValuer = i
		var _ pgtype.IntervalValuer = &i

		assert.Equal(t, pgtype.Interval{}, i.GetPgxType())
		iv, err := i.IntervalValue()
		require.NoError(t, err)
		assert.False(t, iv.Valid)

		v, err := i.Value()
		require.NoError(t, err)
		assert.Nil(t, v)

		assert.Equal(t, pgtype.Interval{}, (*IntervalMarshal)(nil).GetPgxType())
	})
	t.Run("MarshalJSON and Format", func(t *testing.T) {
		var i IntervalMarshal
		data, err := json.Marshal(i)
		require.NoError(t, err)
		assert.Equal(t, "null", string(data))
		assert.Equal(t, "", fmt.Sprintf("%s", &i))
	})
}

// randomIntervalText builds a text pgtype.Interval.Scan accepts.
func randomIntervalText(r *rand.Rand) string {
	var parts []string
	num := func(limit int) int { return r.Intn(2*limit+1) - limit }
	unit := func(n int, one, many string) string {
		if r.Intn(2) == 0 {
			return fmt.Sprintf("%d %s", n, one)
		}
		return fmt.Sprintf("%+d %s", n, many)
	}

	if r.Intn(2) == 0 {
		parts = append(parts, unit(num(1000), "year", "years"))
	}
	if r.Intn(2) == 0 {
		parts = append(parts, unit(num(100_000), "mon", "mons"))
	}
	if r.Intn(2) == 0 {
		parts = append(parts, unit(num(1_000_000), "day", "days"))
	}
	if r.Intn(4) != 0 || len(parts) == 0 {
		clock := fmt.Sprintf("%02d:%02d:%02d", r.Intn(100_000), r.Intn(60), r.Intn(60))
		if r.Intn(2) == 0 {
			clock = "-" + clock
		}
		if r.Intn(2) == 0 {
			clock += "." + fmt.Sprintf("%06d", r.Intn(1_000_000))[:1+r.Intn(6)]
		}
		parts = append(parts, clock)
	}

	return strings.Join(parts, " ")
}

// Whatever pgx reads, Scan reads the same.
func TestIntervalMarshal_Scan_sameAsPgx(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	for n := 0; n < 20_000; n++ {
		src := randomIntervalText(r)

		var want pgtype.Interval
		require.NoError(t, want.Scan(src), src)

		got := NewIntervalMarshal()
		require.NoError(t, got.Scan(src), src)
		require.Equal(t, want, *got.Interval, src)
	}
}

// MarshalJSON writes what pgx writes as the interval's text, and Scan reads
// its own output back.
func TestIntervalMarshal_MarshalJSON_sameAsPgx(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	values := []pgtype.Interval{
		{Valid: true},
		{Microseconds: 1, Valid: true},
		{Microseconds: -1, Valid: true},
		{Microseconds: 9 * usPerHour, Valid: true},
		{Microseconds: math.MaxInt64, Valid: true},
		{Microseconds: math.MinInt64 + 1, Valid: true},
		{Months: math.MinInt32, Days: math.MaxInt32, Valid: true},
	}
	for n := 0; n < 5_000; n++ {
		values = append(values, pgtype.Interval{
			Months:       int32(r.Intn(2001) - 1000),
			Days:         int32(r.Intn(2001) - 1000),
			Microseconds: r.Int63n(2_000_000*usPerHour) - 1_000_000*usPerHour,
			Valid:        true,
		})
	}

	for _, iv := range values {
		text, err := iv.Value()
		require.NoError(t, err)
		want, err := json.Marshal(text)
		require.NoError(t, err)

		got, err := IntervalMarshal{&iv}.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, string(want), string(got), "%+v", iv)

		back := NewIntervalMarshal()
		require.NoError(t, back.UnmarshalJSON(got), string(got))
		require.Equal(t, iv, *back.Interval, string(got))
	}
}

// The bare text of a form/query parameter - TypeInParam.ReadValue passes it
// to UnmarshalJSON without quotes.
func TestIntervalMarshal_UnmarshalJSON_formValue(t *testing.T) {
	tests := []struct {
		in      string
		want    pgtype.Interval
		wantErr bool
	}{
		{`00:00:00`, pgtype.Interval{Valid: true}, false},
		{`00:00`, pgtype.Interval{Valid: true}, false},
		{`01:30`, pgtype.Interval{Microseconds: 5_400_000_000, Valid: true}, false},
		{`01:30:00`, pgtype.Interval{Microseconds: 5_400_000_000, Valid: true}, false},
		{`-01:30:00`, pgtype.Interval{Microseconds: -5_400_000_000, Valid: true}, false},
		{`1 day 02:03:04`, pgtype.Interval{Days: 1, Microseconds: 7_384_000_000, Valid: true}, false},
		{` 2 years 3 mons `, pgtype.Interval{Months: 27, Valid: true}, false},
		// an empty parameter and a bare zero are the zero interval, not NULL
		{``, pgtype.Interval{Valid: true}, false},
		{`0`, pgtype.Interval{Valid: true}, false},
		{`00`, pgtype.Interval{Valid: true}, false},

		{`abc`, pgtype.Interval{}, true},
		{`1 day 02:03:04"`, pgtype.Interval{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
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

func TestIntervalMarshal_UnmarshalJSON_objectRange(t *testing.T) {
	for _, in := range []string{
		`{"Months":2147483648}`,
		`{"Days":-2147483649}`,
		`{"Microseconds":1.5}`,
	} {
		t.Run(in, func(t *testing.T) {
			assert.Error(t, NewIntervalMarshal().UnmarshalJSON([]byte(in)))
		})
	}
}

// The reason Scan exists: a value that parses costs no allocation.
func TestIntervalMarshal_noAllocations(t *testing.T) {
	i := NewIntervalMarshal()
	text := "2 years 3 mons 4 days -05:06:07.5"
	raw := []byte(text)
	quoted := []byte(`"` + text + `"`)
	short, null := []byte(`"01:30"`), []byte(`null`)
	var sink error

	tests := []struct {
		name string
		fn   func()
	}{
		{"Scan(string)", func() { sink = i.Scan(text) }},
		{"Scan([]byte)", func() { sink = i.Scan(raw) }},
		{"Set(string)", func() { sink = i.Set(text) }},
		{"UnmarshalJSON, form value", func() { sink = i.UnmarshalJSON(raw) }},
		{"UnmarshalJSON, JSON string", func() { sink = i.UnmarshalJSON(quoted) }},
		{"UnmarshalJSON, short clock", func() { sink = i.UnmarshalJSON(short) }},
		{"UnmarshalJSON, null", func() { sink = i.UnmarshalJSON(null) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Zero(t, testing.AllocsPerRun(200, tt.fn))
			require.NoError(t, sink)
		})
	}

	require.NoError(t, i.Scan(text))
	assert.Equal(t, pgtype.Interval{Months: 27, Days: 4, Microseconds: -18_367_500_000, Valid: true}, *i.Interval)
}

var benchIntervalTexts = []string{"01:30:00", "1 day 02:03:04", "2 years 3 mons 4 days -05:06:07.5"}

func BenchmarkIntervalScan(b *testing.B) {
	for _, text := range benchIntervalTexts {
		b.Run("pgtype.Interval/"+text, func(b *testing.B) {
			var iv pgtype.Interval
			b.ReportAllocs()
			for b.Loop() {
				if err := iv.Scan(text); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("IntervalMarshal/"+text, func(b *testing.B) {
			i := NewIntervalMarshal()
			b.ReportAllocs()
			for b.Loop() {
				if err := i.Scan(text); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkIntervalMarshal_JSON(b *testing.B) {
	inputs := map[string][]byte{
		"string": []byte(`"1 day 02:03:04"`),
		"form":   []byte(`1 day 02:03:04`),
		"object": []byte(`{"Microseconds":7384000000,"Days":1,"Months":0,"Valid":true}`),
	}
	for name, src := range inputs {
		b.Run("Unmarshal/"+name, func(b *testing.B) {
			i := NewIntervalMarshal()
			b.ReportAllocs()
			for b.Loop() {
				if err := i.UnmarshalJSON(src); err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	b.Run("Marshal", func(b *testing.B) {
		i := IntervalMarshal{&pgtype.Interval{Days: 1, Microseconds: 7_384_000_000, Valid: true}}
		b.ReportAllocs()
		for b.Loop() {
			if _, err := i.MarshalJSON(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

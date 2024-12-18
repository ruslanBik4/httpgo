/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Package json формирование JSON из разного вида данных и выдача текста в поток
package json

import (
	"database/sql"
	"math"
	"math/big"
	"sort"
	"time"
	"unsafe"

	"github.com/jackc/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/quicktemplate"

	"github.com/ruslanBik4/logs"
)

var Json = jsoniter.ConfigFastest

type Number interface {
	int | int64 | int32 | float32 | float64
}

func StreamWrap(w *quicktemplate.Writer, value any) {

	stream := jsoniter.NewStream(Json, w.W(), int(unsafe.Sizeof(value)))
	stream.WriteVal(value)

	if err := stream.Flush(); err != nil {
		logs.ErrorLog(err, "during stream %v", value)
	}
}

func StreamSlice[T any](w *quicktemplate.Writer, value []T) {
	w.N().S(`[`)
	for key, v := range value {
		if key > 0 {
			w.N().S(`,`)
		}
		StreamElement(w, v)
	}
	w.N().S(`]`)
}

func StreamMap[E comparable, T any](w *quicktemplate.Writer, value map[E]T) {
	if value == nil {
		w.N().S("nil")
		return
	}
	sortList := make([]E, 0, len(value))
	for name := range value {
		sortList = append(sortList, name)
	}

	sort.Slice(sortList, func(i, j int) bool {
		return i < j
	})

	w.N().S(`{`)

	for key, name := range sortList {
		if key > 0 {
			w.N().S(`,`)
		}
		w.N().S(`"`)
		switch name := any(name).(type) {
		case string:
			w.N().J(name)
		case int:
			w.N().D(name)
		default:
			w.N().V(name)
		}
		w.N().S(`":`)
		StreamElement(w, value[name])
	}
	w.N().S(`}`)
}

func init() {
	jsoniter.RegisterTypeEncoderFunc("map[string]string",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			m := *(*map[string]string)(ptr)
			WriteStringJSON(stream, m)
		}, func(pointer unsafe.Pointer) bool {
			return false
		})
	jsoniter.RegisterTypeEncoderFunc("map[string]interface{}",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			m := *(*map[string]interface{})(ptr)
			WriteAnyJSON(stream, m)
		}, func(pointer unsafe.Pointer) bool {
			return false
		})
	jsoniter.RegisterTypeEncoderFunc("interface{}",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			m := *(*interface{})(ptr)
			WriteElement(stream, m)
		}, func(pointer unsafe.Pointer) bool {
			return false
		})
	jsoniter.RegisterTypeEncoderFunc("sql.NullString",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			val := (*sql.NullString)(ptr)
			if val.Valid {
				WriteString(stream, val.String)
			} else {
				stream.WriteNil()
			}
		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("sql.NullInt32",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			val := (*sql.NullInt32)(ptr)
			if val.Valid {
				stream.WriteInt32(val.Int32)
			} else {
				stream.WriteNil()
			}
		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("sql.NullInt64",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			val := (*sql.NullInt64)(ptr)
			if val.Valid {
				stream.WriteInt64(val.Int64)
			} else {
				stream.WriteNil()
			}
		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("sql.NullFloat64",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			val := (*sql.NullFloat64)(ptr)
			if val.Valid {
				stream.WriteFloat64Lossy(val.Float64)
			} else {
				stream.WriteNil()
			}
		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.Int4Array",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			accArray := (*pgtype.Int4Array)(ptr)
			stream.WriteArrayStart()

			for i, val := range accArray.Elements {
				if i > 0 {
					stream.WriteMore()
				}
				stream.WriteInt32(val.Int)
			}

			stream.WriteArrayEnd()

		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.Int8Array",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			accArray := (*pgtype.Int8Array)(ptr)
			stream.WriteArrayStart()

			for i, val := range accArray.Elements {
				if i > 0 {
					stream.WriteMore()
				}
				stream.WriteInt64(val.Int)
			}

			stream.WriteArrayEnd()

		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.Float4Array",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			accArray := (*pgtype.Float4Array)(ptr)
			stream.WriteArrayStart()

			for i, val := range accArray.Elements {
				if i > 0 {
					stream.WriteMore()
				}
				stream.WriteFloat32Lossy(val.Float)
			}

			stream.WriteArrayEnd()

		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.Float8Array",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			accArray := (*pgtype.Float8Array)(ptr)
			stream.WriteArrayStart()

			for i, val := range accArray.Elements {
				if i > 0 {
					stream.WriteMore()
				}
				stream.WriteFloat64Lossy(val.Float)
			}

			stream.WriteArrayEnd()

		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.Numeric",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			val := (*pgtype.Numeric)(ptr)
			divider := math.Pow10(int(val.Exp))
			stream.WriteFloat64Lossy(float64(val.Int.Int64()) * divider)
		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeDecoderFunc("pgtype.Numeric",
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			val := iter.ReadFloat64()

			f := big.NewFloat(val)
			err := (*(*pgtype.Numeric)(ptr)).Scan(f.String())
			if err != nil {
				logs.ErrorLog(err, val)
			}
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.NumericArray",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			accArray := (*pgtype.NumericArray)(ptr)
			stream.WriteArrayStart()

			for i, val := range accArray.Elements {
				if i > 0 {
					stream.WriteMore()
				}
				divider := math.Pow10(int(val.Exp))
				stream.WriteFloat64Lossy(float64(val.Int.Int64()) * divider)
			}

			stream.WriteArrayEnd()

		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("*pgtype.Date",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			val := (*pgtype.Date)(ptr)
			if val.Status == pgtype.Present {
				stream.WriteString(val.Time.Format(time.DateOnly))
			} else {
				stream.WriteNil()
			}
		},
		func(ptr unsafe.Pointer) bool {
			val := (*pgtype.Date)(ptr)
			return val.Status != pgtype.Present || val.Time.IsZero()
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.Daterange",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			val := (*pgtype.Daterange)(ptr)
			stream.WriteArrayStart()

			stream.WriteInt64(val.Lower.Time.Unix())
			stream.WriteMore()
			stream.WriteInt64(val.Upper.Time.Unix())

			stream.WriteArrayEnd()
		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.VarcharArray",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			accArray := (*pgtype.VarcharArray)(ptr)
			stream.WriteArrayStart()

			for i, val := range accArray.Elements {
				if i > 0 {
					stream.WriteMore()
				}
				stream.WriteString(val.String)
			}

			stream.WriteArrayEnd()

		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.BPCharArray",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			accArray := (*pgtype.BPCharArray)(ptr)
			stream.WriteArrayStart()

			for i, val := range accArray.Elements {
				if i > 0 {
					stream.WriteMore()
				}
				stream.WriteString(val.String)
			}

			stream.WriteArrayEnd()

		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeEncoderFunc("pgtype.TextArray",
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			accArray := (*pgtype.TextArray)(ptr)
			stream.WriteArrayStart()

			for i, val := range accArray.Elements {
				if i > 0 {
					stream.WriteMore()
				}
				stream.WriteString(val.String)
			}

			stream.WriteArrayEnd()

		},
		func(pointer unsafe.Pointer) bool {
			return false
		})

}

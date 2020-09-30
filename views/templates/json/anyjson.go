// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// формирование JSON из разного вида данных и выдача текста в поток
package json

import (
	"database/sql"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/quicktemplate"
)

func StreamWrap(w *quicktemplate.Writer, value interface{}) {
	enc := jsoniter.NewEncoder(w.W())
	_ = enc.Encode(value)
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
				WriteElement(stream, val.String)
			} else {
				stream.WriteString("nil")
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
				stream.WriteString("nil")
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
				stream.WriteString("nil")
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
				stream.WriteString("nil")
			}
		},
		func(pointer unsafe.Pointer) bool {
			return false
		})
}

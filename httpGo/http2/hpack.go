/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"bytes"

	"golang.org/x/net/http2/hpack"
)

// HPACKDecoder decodes request header blocks. It is owned by the read loop.
//
// The HPACK state machine (static/dynamic tables, Huffman) is delegated to
// golang.org/x/net/http2/hpack — it's the reference implementation used by
// net/http, is fuzzed, and has no shared global state, so it's race-free when
// each connection owns its own instance.
type HPACKDecoder struct {
	dec  *hpack.Decoder
	emit func(name, value string)
}

// NewHPACKDecoder creates a decoder. maxStringLen bounds a single header field.
func NewHPACKDecoder(maxStringLen int) *HPACKDecoder {
	d := &HPACKDecoder{}
	d.dec = hpack.NewDecoder(defaultHPACKTblLen, func(f hpack.HeaderField) {
		d.emit(f.Name, f.Value)
	})
	d.dec.SetMaxStringLength(maxStringLen)
	return d
}

// Decode decodes a complete header block and calls emit for each field.
// Any error is a connection-level COMPRESSION_ERROR (HPACK state is lost).
func (d *HPACKDecoder) Decode(block []byte, emit func(name, value string)) error {
	d.emit = emit
	defer func() { d.emit = nil }()
	if _, err := d.dec.Write(block); err != nil {
		return connErr(ErrCodeCompression, "hpack: %v", err)
	}
	if err := d.dec.Close(); err != nil {
		return connErr(ErrCodeCompression, "hpack: %v", err)
	}
	return nil
}

// HPACKEncoder encodes response header blocks.
//
// Encoding mutates the dynamic table, so the encoded block MUST be written to
// the wire in the same order it was encoded: Conn encodes and writes under writeMu.
type HPACKEncoder struct {
	buf bytes.Buffer
	enc *hpack.Encoder
	// lower caches lower-cased header names ("Content-Type" -> "content-type").
	// Lookups with m[string(b)] don't allocate; only misses do.
	lower map[string]string
}

// NewHPACKEncoder creates an encoder.
func NewHPACKEncoder() *HPACKEncoder {
	e := &HPACKEncoder{lower: make(map[string]string, 32)}
	e.enc = hpack.NewEncoder(&e.buf)
	return e
}

// Reset starts a new header block.
func (e *HPACKEncoder) Reset() { e.buf.Reset() }

// Bytes returns the encoded block (valid until the next Reset).
func (e *HPACKEncoder) Bytes() []byte { return e.buf.Bytes() }

// SetMaxDynamicTableSizeLimit applies the peer's SETTINGS_HEADER_TABLE_SIZE.
func (e *HPACKEncoder) SetMaxDynamicTableSizeLimit(v uint32) { e.enc.SetMaxDynamicTableSizeLimit(v) }

// WriteField encodes one field. name must already be lower-case.
func (e *HPACKEncoder) WriteField(name, value string) {
	// hpack.Encoder keeps strings in its dynamic table, so they must be real
	// (immutable) strings, never unsafe views over reusable []byte.
	_ = e.enc.WriteField(hpack.HeaderField{Name: name, Value: value})
}

// WriteFieldBytes lower-cases name (cached) and encodes the field.
func (e *HPACKEncoder) WriteFieldBytes(name, value []byte) {
	n, ok := e.lower[string(name)]
	if !ok {
		n = string(bytes.ToLower(name))
		if len(e.lower) < 256 {
			e.lower[string(name)] = n
		}
	}
	e.WriteField(n, string(value))
}

/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

// FrameType is an HTTP/2 frame type (RFC 9113 §6).
type FrameType uint8

const (
	FrameData         FrameType = 0x0
	FrameHeaders      FrameType = 0x1
	FramePriority     FrameType = 0x2
	FrameRSTStream    FrameType = 0x3
	FrameSettings     FrameType = 0x4
	FramePushPromise  FrameType = 0x5
	FramePing         FrameType = 0x6
	FrameGoAway       FrameType = 0x7
	FrameWindowUpdate FrameType = 0x8
	FrameContinuation FrameType = 0x9
)

// Flags are frame flags; their meaning depends on the frame type.
type Flags uint8

const (
	FlagEndStream  Flags = 0x1 // DATA, HEADERS
	FlagAck        Flags = 0x1 // SETTINGS, PING
	FlagEndHeaders Flags = 0x4 // HEADERS, CONTINUATION
	FlagPadded     Flags = 0x8 // DATA, HEADERS
	FlagPriority   Flags = 0x20
)

func (f Flags) Has(v Flags) bool { return f&v == v }

// SettingID is an HTTP/2 SETTINGS parameter identifier (RFC 9113 §6.5.2).
type SettingID uint16

const (
	SettingHeaderTableSize      SettingID = 0x1
	SettingEnablePush           SettingID = 0x2
	SettingMaxConcurrentStreams SettingID = 0x3
	SettingInitialWindowSize    SettingID = 0x4
	SettingMaxFrameSize         SettingID = 0x5
	SettingMaxHeaderListSize    SettingID = 0x6
)

// Setting is a single SETTINGS parameter.
type Setting struct {
	ID  SettingID
	Val uint32
}

const (
	frameHeaderLen     = 9
	minMaxFrameSize    = 1 << 14   // 16 KiB, protocol default and minimum
	maxMaxFrameSize    = 1<<24 - 1 // protocol maximum
	defaultWindowSize  = 65535
	maxWindowSize      = 1<<31 - 1
	clientPreface      = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
	defaultHPACKTblLen = 4096
)

// errIdle means the read deadline expired between frames (no bytes lost).
var errIdle = errors.New("http2: idle timeout")

// Frame is a raw HTTP/2 frame. Payload aliases the Framer read buffer and is
// valid only until the next ReadFrame call.
type Frame struct {
	Length   uint32
	Type     FrameType
	Flags    Flags
	StreamID uint32
	Payload  []byte
}

// Framer reads and writes HTTP/2 frames.
//
// Reading is used only by the connection read loop; writing must be
// serialized by the caller (Conn holds writeMu). The Framer itself does no locking.
type Framer struct {
	r       *bufio.Reader
	w       *bufio.Writer
	hdr     [frameHeaderLen]byte // read side only
	whdr    [frameHeaderLen]byte // write side only (never share with read side: different goroutines)
	readBuf []byte
	// MaxReadFrameSize is the SETTINGS_MAX_FRAME_SIZE we advertised.
	MaxReadFrameSize uint32
	frame            Frame
}

// NewFramer creates a Framer over r and w.
func NewFramer(r io.Reader, w io.Writer) *Framer {
	return &Framer{
		r:                bufio.NewReaderSize(r, 32<<10),
		w:                bufio.NewWriterSize(w, 32<<10),
		MaxReadFrameSize: minMaxFrameSize,
	}
}

// ReadFrame reads the next frame. The returned frame is reused by subsequent calls.
func (fr *Framer) ReadFrame() (*Frame, error) {
	if n, err := io.ReadFull(fr.r, fr.hdr[:]); err != nil {
		if n == 0 {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				return nil, errIdle // nothing consumed: safe to retry after resetting the deadline
			}
		}
		return nil, err
	}
	f := &fr.frame
	f.Length = uint32(fr.hdr[0])<<16 | uint32(fr.hdr[1])<<8 | uint32(fr.hdr[2])
	f.Type = FrameType(fr.hdr[3])
	f.Flags = Flags(fr.hdr[4])
	f.StreamID = binary.BigEndian.Uint32(fr.hdr[5:]) & (1<<31 - 1)
	if f.Length > fr.MaxReadFrameSize {
		return nil, connErr(ErrCodeFrameSize, "frame length %d exceeds max %d", f.Length, fr.MaxReadFrameSize)
	}
	if cap(fr.readBuf) < int(f.Length) {
		fr.readBuf = make([]byte, f.Length, max(int(f.Length), minMaxFrameSize))
	}
	f.Payload = fr.readBuf[:f.Length]
	if _, err := io.ReadFull(fr.r, f.Payload); err != nil {
		return nil, err
	}
	return f, nil
}

func (fr *Framer) writeHeader(length int, t FrameType, flags Flags, streamID uint32) {
	h := fr.whdr[:0:frameHeaderLen]
	h = append(h, byte(length>>16), byte(length>>8), byte(length), byte(t), byte(flags))
	h = binary.BigEndian.AppendUint32(h, streamID&(1<<31-1))
	_, _ = fr.w.Write(h) // bufio errors are sticky and surface on Flush
}

// WriteSettings writes a SETTINGS frame.
func (fr *Framer) WriteSettings(settings ...Setting) error {
	fr.writeHeader(6*len(settings), FrameSettings, 0, 0)
	var b [6]byte
	for _, s := range settings {
		binary.BigEndian.PutUint16(b[:2], uint16(s.ID))
		binary.BigEndian.PutUint32(b[2:], s.Val)
		_, _ = fr.w.Write(b[:])
	}
	return nil
}

// WriteSettingsAck writes an empty SETTINGS frame with the ACK flag.
func (fr *Framer) WriteSettingsAck() error {
	fr.writeHeader(0, FrameSettings, FlagAck, 0)
	return nil
}

// WritePing writes a PING frame.
func (fr *Framer) WritePing(ack bool, data [8]byte) error {
	var flags Flags
	if ack {
		flags = FlagAck
	}
	fr.writeHeader(8, FramePing, flags, 0)
	_, err := fr.w.Write(data[:])
	return err
}

// WriteWindowUpdate writes a WINDOW_UPDATE frame.
func (fr *Framer) WriteWindowUpdate(streamID, incr uint32) error {
	fr.writeHeader(4, FrameWindowUpdate, 0, streamID)
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], incr&(1<<31-1))
	_, err := fr.w.Write(b[:])
	return err
}

// WriteRSTStream writes a RST_STREAM frame.
func (fr *Framer) WriteRSTStream(streamID uint32, code ErrCode) error {
	fr.writeHeader(4, FrameRSTStream, 0, streamID)
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(code))
	_, err := fr.w.Write(b[:])
	return err
}

// WriteGoAway writes a GOAWAY frame.
func (fr *Framer) WriteGoAway(lastStreamID uint32, code ErrCode, debug []byte) error {
	fr.writeHeader(8+len(debug), FrameGoAway, 0, 0)
	var b [8]byte
	binary.BigEndian.PutUint32(b[:4], lastStreamID&(1<<31-1))
	binary.BigEndian.PutUint32(b[4:], uint32(code))
	_, _ = fr.w.Write(b[:])
	_, err := fr.w.Write(debug)
	return err
}

// WriteHeaders writes a header block as HEADERS followed by as many
// CONTINUATION frames as needed to respect maxFrameSize.
func (fr *Framer) WriteHeaders(streamID uint32, endStream bool, block []byte, maxFrameSize uint32) error {
	t := FrameHeaders
	var flags Flags
	if endStream {
		flags = FlagEndStream
	}
	for first := true; first || len(block) > 0; first = false {
		chunk := block
		if uint32(len(chunk)) > maxFrameSize {
			chunk = chunk[:maxFrameSize]
		}
		block = block[len(chunk):]
		f := flags
		if len(block) == 0 {
			f |= FlagEndHeaders
		}
		fr.writeHeader(len(chunk), t, f, streamID)
		if _, err := fr.w.Write(chunk); err != nil {
			return err
		}
		t, flags = FrameContinuation, 0
	}
	return nil
}

// WritePushPromise writes PUSH_PROMISE (+ CONTINUATION frames as needed) on
// the associated client stream, promising promisedID.
func (fr *Framer) WritePushPromise(streamID, promisedID uint32, block []byte, maxFrameSize uint32) error {
	first := block
	if uint32(len(first)) > maxFrameSize-4 {
		first = first[:maxFrameSize-4]
	}
	block = block[len(first):]
	var flags Flags
	if len(block) == 0 {
		flags = FlagEndHeaders
	}
	fr.writeHeader(4+len(first), FramePushPromise, flags, streamID)
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], promisedID&(1<<31-1))
	_, _ = fr.w.Write(b[:])
	if _, err := fr.w.Write(first); err != nil {
		return err
	}
	for len(block) > 0 {
		chunk := block
		if uint32(len(chunk)) > maxFrameSize {
			chunk = chunk[:maxFrameSize]
		}
		block = block[len(chunk):]
		flags = 0
		if len(block) == 0 {
			flags = FlagEndHeaders
		}
		fr.writeHeader(len(chunk), FrameContinuation, flags, streamID)
		if _, err := fr.w.Write(chunk); err != nil {
			return err
		}
	}
	return nil
}

// WriteData writes a single DATA frame. The caller is responsible for
// respecting flow control and the peer's max frame size.
func (fr *Framer) WriteData(streamID uint32, endStream bool, data []byte) error {
	var flags Flags
	if endStream {
		flags = FlagEndStream
	}
	fr.writeHeader(len(data), FrameData, flags, streamID)
	_, err := fr.w.Write(data)
	return err
}

// Flush flushes buffered frames to the connection.
func (fr *Framer) Flush() error { return fr.w.Flush() }

// Buffered reports whether there is unflushed output.
func (fr *Framer) Buffered() int { return fr.w.Buffered() }

// stripPadding removes the Pad Length byte and trailing padding.
func stripPadding(f *Frame) ([]byte, error) {
	p := f.Payload
	if !f.Flags.Has(FlagPadded) {
		return p, nil
	}
	if len(p) < 1 {
		return nil, connErr(ErrCodeFrameSize, "padded frame too short")
	}
	padLen := int(p[0])
	p = p[1:]
	if padLen > len(p) {
		return nil, connErr(ErrCodeProtocol, "padding exceeds payload")
	}
	return p[:len(p)-padLen], nil
}

// headerBlockFragment extracts the header block fragment from a HEADERS frame,
// dropping padding and the (deprecated) priority fields.
func headerBlockFragment(f *Frame) (frag []byte, dependsOn uint32, err error) {
	p, err := stripPadding(f)
	if err != nil {
		return nil, 0, err
	}
	if f.Flags.Has(FlagPriority) {
		if len(p) < 5 {
			return nil, 0, connErr(ErrCodeFrameSize, "HEADERS priority too short")
		}
		dependsOn = binary.BigEndian.Uint32(p) & (1<<31 - 1)
		p = p[5:]
	}
	return p, dependsOn, nil
}

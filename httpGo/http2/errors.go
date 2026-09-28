/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"errors"
	"fmt"
)

// ErrCode is an HTTP/2 error code (RFC 9113 §7).
type ErrCode uint32

const (
	ErrCodeNo                 ErrCode = 0x0
	ErrCodeProtocol           ErrCode = 0x1
	ErrCodeInternal           ErrCode = 0x2
	ErrCodeFlowControl        ErrCode = 0x3
	ErrCodeSettingsTimeout    ErrCode = 0x4
	ErrCodeStreamClosed       ErrCode = 0x5
	ErrCodeFrameSize          ErrCode = 0x6
	ErrCodeRefusedStream      ErrCode = 0x7
	ErrCodeCancel             ErrCode = 0x8
	ErrCodeCompression        ErrCode = 0x9
	ErrCodeConnect            ErrCode = 0xa
	ErrCodeEnhanceYourCalm    ErrCode = 0xb
	ErrCodeInadequateSecurity ErrCode = 0xc
	ErrCodeHTTP11Required     ErrCode = 0xd
)

var errCodeName = [...]string{
	"NO_ERROR", "PROTOCOL_ERROR", "INTERNAL_ERROR", "FLOW_CONTROL_ERROR",
	"SETTINGS_TIMEOUT", "STREAM_CLOSED", "FRAME_SIZE_ERROR", "REFUSED_STREAM",
	"CANCEL", "COMPRESSION_ERROR", "CONNECT_ERROR", "ENHANCE_YOUR_CALM",
	"INADEQUATE_SECURITY", "HTTP_1_1_REQUIRED",
}

func (e ErrCode) String() string {
	if int(e) < len(errCodeName) {
		return errCodeName[e]
	}
	return fmt.Sprintf("UNKNOWN_ERROR_0x%x", uint32(e))
}

// ConnError is a connection-level error: we send GOAWAY and close the connection.
type ConnError struct {
	Code   ErrCode
	Reason string
}

func (e ConnError) Error() string {
	return "http2: connection error " + e.Code.String() + ": " + e.Reason
}

// StreamError is a stream-level error: we send RST_STREAM and keep the connection alive.
type StreamError struct {
	StreamID uint32
	Code     ErrCode
	Reason   string
}

func (e StreamError) Error() string {
	return fmt.Sprintf("http2: stream %d error %s: %s", e.StreamID, e.Code, e.Reason)
}

var (
	// ErrProtocol is returned when the client preface is invalid.
	ErrProtocol = ConnError{ErrCodeProtocol, "invalid client preface"}
	// ErrClientGoAway is returned by Serve when the client sent GOAWAY and all streams are done.
	ErrClientGoAway = errors.New("http2: client sent GOAWAY")
	// errStreamClosed is used internally when a stream was reset while the handler was writing.
	errStreamClosed = errors.New("http2: stream closed")
	errConnClosed   = errors.New("http2: connection closed")
)

func connErr(code ErrCode, format string, args ...any) error {
	return ConnError{Code: code, Reason: fmt.Sprintf(format, args...)}
}

func streamErr(id uint32, code ErrCode, reason string) error {
	return StreamError{StreamID: id, Code: code, Reason: reason}
}

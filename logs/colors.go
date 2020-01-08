// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logs

import (
	"fmt"
)

const (
	LogPutColor = "\033["
	LogEndColor = "\033[0m"
)

type Level int

const (
	CRITICAL Level = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

type color int

const (
	colorBlack = (iota + 30)
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite
)

var (
	colors = []string{
		CRITICAL: colorSeq(colorMagenta),
		ERROR:    colorSeq(colorRed),
		WARNING:  colorSeq(colorYellow),
		NOTICE:   colorSeq(colorGreen),
		DEBUG:    colorSeq(colorCyan),
	}
	boldcolors = []string{
		CRITICAL: colorSeqBold(colorMagenta),
		ERROR:    colorSeqBold(colorRed),
		WARNING:  colorSeqBold(colorYellow),
		NOTICE:   colorSeqBold(colorGreen),
		DEBUG:    colorSeqBold(colorCyan),
	}
)

func colorSeq(color color) string {
	return fmt.Sprintf(LogPutColor+"%dm", int(color))
}

func colorSeqBold(color color) string {
	return fmt.Sprintf(LogPutColor+"%d;1m", int(color))
}

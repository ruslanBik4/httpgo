// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package logs output logs and advanced debug information
package logs

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

var (
	fDebug   = flag.Bool("debug", false, "debug mode")
	fStatus  = flag.Bool("status", true, "status mode")
	logErr   = NewWrapKitLogger(colors[ERROR]+"ERROR"+"\033[0m", 1)
	logStat  = NewWrapKitLogger("INFO", 3)
	logDebug = NewWrapKitLogger(colors[DEBUG]+"DEBUG"+"\033[0m", 3)
)

// LogsType - interface for print logs record
type LogsType interface {
	PrintToLogs() string
}

type wrapKitLogger struct {
	*log.Logger
	calldepth int
	funcName  string
	typeLog   string
	toSentry  bool
	toOther   io.Writer
}

const logFlags = log.Lshortfile | log.Ltime

var stackBeginWith = 1

func NewWrapKitLogger(pref string, depth int) *wrapKitLogger {
	return &wrapKitLogger{
		Logger:    log.New(os.Stdout, "[["+pref+"]]", logFlags),
		typeLog:   pref,
		calldepth: depth,
	}
}

// SetDebug set debug level for log, return old value
func SetDebug(d bool) bool {
	old := *fDebug
	*fDebug = d

	return old
}

// SetStatus set status level for log, return old value
func SetStatus(s bool) bool {
	old := *fStatus
	*fStatus = s

	return old
}

type fgLogWriter int8

const (
	fgAll fgLogWriter = iota
	fgErr
	fgInfo
	fgDebug
)

func (logger *wrapKitLogger) addWriters(newWriter io.Writer) {
	if logger.toOther == nil {
		logger.toOther = newWriter
	} else {
		logger.toOther = io.MultiWriter(logger.toOther, newWriter)
	}
}

// SetWriters for logs
func SetWriters(newWriter io.Writer, logFlag fgLogWriter) {
	
	switch logFlag {
	case fgAll:
		logErr.addWriters(newWriter)
		logStat.addWriters(newWriter)
		logDebug.addWriters(newWriter)
	case fgErr:
		logErr.addWriters(newWriter)
	case fgInfo:
		logStat.addWriters(newWriter)
	case fgDebug:
		logDebug.addWriters(newWriter)
	}
}

// SetSentry set SetSentry output for error
func SetSentry(dns string) error {
	err := sentry.Init(sentry.ClientOptions{Dsn: dns})
	if err != nil {
		return errors.Wrap(err, "sentry.Init")
	}

	logErr.toSentry = true

	return nil
}

// SetLogFlags set logger flags & return old flags
func SetLogFlags(f int) int {

	flags := logErr.Flags()

	logErr.SetFlags(f)
	logDebug.SetFlags(f)
	logStat.SetFlags(f)

	return flags
}

// SetStackBeginWith set stackBeginWith level for log, return old value
func SetStackBeginWith(s int) int {
	old := stackBeginWith
	stackBeginWith = s

	return old
}

type logMess struct {
	Message string    `json:"message"`
	Now     time.Time `json:"@timestamp"`
	Level   string    `json:"level"`
	//vars    [] interface{} `json:"vars"`
}

func NewlogMess(mess string, logger *wrapKitLogger) *logMess {
	return &logMess{mess, time.Now(), logger.typeLog}
}

func (logger *wrapKitLogger) Printf(vars ...interface{}) {

	checkPrint, checkType := vars[0].(errLogPrint)

	if checkType == true {
		vars = vars[1:]
	}

	mess := getArgsString(vars...)
	if checkType && (checkPrint == true) {
		fmt.Printf(mess)
	} else {
		if logger.funcName > "" {
			mess = logger.funcName + "();" + mess
		}

		logger.Output(logger.calldepth, mess)
	}

	if logger.toOther != nil {
		go logger.toOther.Write([]byte(mess))
	}

}

func getArgsString(args ...interface{}) (message string) {

	if len(args) < 1 {
		return ""
	}
	// first param may by format string
	if format, ok := args[0].(string); ok && (strings.Index(format, "%") > -1) {
		return fmt.Sprintf(format, args[1:]...)
	}

	comma := ""
	for _, arg := range args {

		switch val := arg.(type) {
		case nil:
			message += comma + " is nil"
		case string:
			message += comma + val
		case []string:
			for _, value := range val {
				message += value + "\n"
			}
		case LogsType:
			message += comma + val.PrintToLogs()
		case time.Time:
			message += comma + val.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
		case []interface{}:
			message += getArgsString(val...)
		case error:
			message += comma + fmt.Sprintf("%v", arg)
		default:

			message += comma + fmt.Sprintf("%#v", arg)
		}

		comma = ", "
	}

	return message
}

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
	return fmt.Sprintf("\033[%dm", int(color))
}

func colorSeqBold(color color) string {
	return fmt.Sprintf("\033[%d;1m", int(color))
}

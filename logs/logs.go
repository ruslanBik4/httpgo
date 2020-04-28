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
	logErr   = NewWrapKitLogger(colors[ERROR]+"ERROR"+LogEndColor, 1)
	logStat  = NewWrapKitLogger("INFO", 3)
	logDebug = NewWrapKitLogger(colors[DEBUG]+"DEBUG"+LogEndColor, 3)
)

// LogsType - interface for print logs record
type LogsType interface {
	PrintToLogs() string
}

type wrapKitLogger struct {
	*log.Logger
	calldepth int
	line      int
	fileName  string
	funcName  string
	typeLog   string
	toSentry  bool
	sentryOrg string
	toOther   io.Writer
}

const logFlags = log.Lshortfile | log.Ltime

var stackBeginWith = 1

func NewWrapKitLogger(pref string, depth int) *wrapKitLogger {
	return &wrapKitLogger{
		Logger:    log.New(os.Stdout, "[["+pref+"]]", logFlags),
		typeLog:   pref,
		calldepth: depth,
		toOther:   &multiWriter{},
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

type FgLogWriter int8

const (
	FgAll FgLogWriter = iota
	FgErr
	FgInfo
	FgDebug
)

func (logger *wrapKitLogger) addWriter(newWriters ...io.Writer) {
	loggermultiwriter := logger.toOther.(*multiWriter)
	for _, newWriter := range newWriters {
		if newWriter != nil {
			loggermultiwriter.Append(newWriter)
		}
	}
}

func (logger *wrapKitLogger) deleteWriter(writersToDelete ...io.Writer) {
	loggermultiwriter := logger.toOther.(*multiWriter)
	for _, newWriter := range writersToDelete {
		loggermultiwriter.Remove(newWriter)
	}
}

// SetWriters for logs
func SetWriters(newWriter io.Writer, logFlags ...FgLogWriter) {
	// todo: можно поменять местами аргументы и дать возможность добавлять неограниченное количество врайтеров

	for _, logFlag := range logFlags {
		switch logFlag {
		case FgAll:
			logErr.addWriter(newWriter)
			logStat.addWriter(newWriter)
			logDebug.addWriter(newWriter)
		case FgErr:
			logErr.addWriter(newWriter)
		case FgInfo:
			logStat.addWriter(newWriter)
		case FgDebug:
			logDebug.addWriter(newWriter)
		}
	}

}

// DeleteWriters deletes mentioned writer from writers for mentioned logFlag
func DeleteWriters(writerToDelete io.Writer, logFlags ...FgLogWriter) {

	for _, logFlag := range logFlags {
		switch logFlag {
		case FgAll:
			logErr.deleteWriter(writerToDelete)
			logStat.deleteWriter(writerToDelete)
			logDebug.deleteWriter(writerToDelete)
		case FgErr:
			logErr.deleteWriter(writerToDelete)
		case FgInfo:
			logStat.deleteWriter(writerToDelete)
		case FgDebug:
			logDebug.deleteWriter(writerToDelete)
		}
	}
}

// SetSentry set SetSentry output for error
func SetSentry(dns string, org string) error {
	err := sentry.Init(sentry.ClientOptions{Dsn: dns})
	if err != nil {
		return errors.Wrap(err, "sentry.Init")
	}

	logErr.toSentry = true
	logErr.sentryOrg = org

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
	if checkType && bool(checkPrint) {
		fmt.Printf(mess + "\n")
	} else {
		_ = logger.Output(logger.calldepth, mess)
		mess = fmt.Sprintf("%s%s:%d %s",
			timeLogFormat(),
			logger.fileName,
			logger.line,
			mess)

	}

	if logger.toOther != nil {
		go func() {
			_, err := logger.toOther.Write([]byte(mess))
			_ = logger.Output(logger.calldepth, getArgsString("Write toOther: %v,", err))
		}()
	}
}

func getArgsString(args ...interface{}) string {

	// if first param is formatting string
	if format, c := getFormatString(args); c > 0 {
		return fmt.Sprintf(format, args[1:]...)
	}

	return argsToString(args...)
}

func argsToString(args ...interface{}) string {
	comma, message := "", ""
	for _, arg := range args {
		message += comma + argToString(arg)
		comma = ", "
	}

	return message
}

func argToString(arg interface{}) string {
	switch val := arg.(type) {
	case nil:
		return " is nil"
	case string:
		return strings.TrimPrefix(val, "ERROR:")
	case []string:
		return strings.Join(val, "\n")
	case LogsType:
		return val.PrintToLogs()
	case time.Time:
		return val.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
	case []interface{}:
		if len(val) > 1 {
			return getArgsString(val...)
		} else if len(val) > 0 {
			return argToString(val[0])
		}

		return ""
	case error:
		return strings.TrimPrefix(val.Error(), "ERROR:")
	default:
		return fmt.Sprintf("%#v", arg)
	}
}

func getFormatString(args []interface{}) (string, int) {
	if len(args) < 2 {
		return "", 0
	}

	if format, ok := args[0].(string); ok {
		c := strings.Count(format, "%")
		if c < len(args)-1 {
			format += strings.Repeat(", %v", len(args)-c-1)
		}

		return format, c
	}

	return "", 0
}

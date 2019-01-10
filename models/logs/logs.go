// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package logs output logs and advanced debug information
package logs

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	fDebug                    = flag.Bool("debug", false, "debug mode")
	fStatus                   = flag.Bool("status", true, "status mode")
	logErr, logStat, logDebug *wrapKitLogger
	//	todo: add datetime mark as option
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
}

const logFlags = log.Lshortfile | log.Ltime

var stackBeginWith = 0

func NewWrapKitLogger(pref string, depth int) *wrapKitLogger {
	return &wrapKitLogger{
		Logger:    log.New(os.Stdout, "[["+pref+"]]", logFlags),
		typeLog:   pref,
		calldepth: depth,
	}
}

func SetDebug(d bool) {
	*fDebug = d
}

func SetStatus(s bool) {
	*fStatus = s
}

func SetStackBeginWith(s int) {
	stackBeginWith = s
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

	mess := NewlogMess(getArgsString(vars...), logger)
	if logger.funcName > "" {
		mess.Message = logger.funcName + "();" + mess.Message
	}

	logger.Output(logger.calldepth, mess.Message)

	//if indService == nil {
	//
	//	return
	//}
	//put, err := client.Index().Index("reports").Type("test").BodyJson(mess).Do(context.Background())
	//if err != nil {
	//	logger.Output(2, fmt.Sprintf("%s. %#v", err.Error(), put))
	//}
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
		default:

			message += comma + fmt.Sprintf("%#v", arg)
		}

		comma = ", "
	}

	return message
}

func init() {
	logErr = NewWrapKitLogger("ERROR", 1)
	logStat = NewWrapKitLogger("INFO", 3)
	logDebug = NewWrapKitLogger("DEBUG", 3)

}

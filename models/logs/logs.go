// Copyright 2017 Author: Sergey Litvinov. All rights reserved.

// Package logs output logs and advanced debug information
package logs

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// LogsType - interface for print logs record
type LogsType interface {
	PrintToLogs() string
}

var fDebug = flag.Bool("debug", false, "debug mode")
var fStatus = flag.Bool("status", true, "status mode")

// DebugLog output formated(function and line calls) debug information
func DebugLog(args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(1)
	if *fDebug {

		log.Output(2, fmt.Sprintf("[DEBUG];%s;%v", changeShortName(runtime.FuncForPC(pc).Name()),
			getArgsString(args)))
	}
}

// StatusLog output formated information for status
func StatusLog(args ...interface{}) {
	//_, fn, line, _ := runtime.Caller(1)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if *fStatus {
		log.Output(2, fmt.Sprintf("[STATUS];;;;%s", getArgsString(args...)))
	}
}

// ErrorLog - output formated(function and line calls) error information
func ErrorLog(err error, args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(1)
	calldepth := 2

	funcName := changeShortName(runtime.FuncForPC(pc).Name())
	if funcName == "views.RenderInternalError" {
		pc, _, _, _ = runtime.Caller(2)
		funcName = changeShortName(runtime.FuncForPC(pc).Name())
		calldepth++
	}

	log.Output(calldepth, fmt.Sprintf("[ERROR];%s;%s;%s", funcName,
		err.Error(), getArgsString(args...)))

}

// ErrorLogHandler - output formated(function and line calls) error information
func ErrorLogHandler(err error, args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(6)

	log.Output(6, fmt.Sprintf("[ERROR];%s;%s;%s", changeShortName(runtime.FuncForPC(pc).Name()),
		err, getArgsString(args...)))

}

// ErrorStack - output formated(function and line calls) error runtime stack information
func ErrorStack() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	i := 1

	for pc, _, _, ok := runtime.Caller(i); ok; pc, _, _, ok = runtime.Caller(i) {
		i++
		funcName := changeShortName(runtime.FuncForPC(pc).Name())
		// пропускаем рендер ошибок
		if funcName == "views.RenderInternalError" {
			continue
		}
		log.Output(i, fmt.Sprintf("[ERROR_STACK];%s;;;;", funcName))
	}
}

// Fatal - output formated (function and line calls) fatal information
func Fatal(err error, args ...interface{}) {
	pc, _, _, _ := runtime.Caller(2)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Output(2, fmt.Sprintf("[FATAL];%v;%v;%v", changeShortName(runtime.FuncForPC(pc).Name()),
		err, getArgsString(args...)))
	os.Exit(1)

}
func changeShortName(file string) (short string) {
	short = file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	return short
}
func getArgsString(args ...interface{}) (message string) {

	comma := ""
	for _, arg := range args {

		switch val := arg.(type) {
		case nil:
			message += comma + " is nil"
		case LogsType:
			message += comma + val.PrintToLogs()
		case time.Time:
			message += comma + val.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
		default:

			message += comma + fmt.Sprintf("%#v", arg)
		}

		comma = ", "
	}

	return message
}

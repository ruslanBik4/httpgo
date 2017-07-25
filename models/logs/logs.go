// Copyright 2017 Author: Sergey Litvinov. All rights reserved.
//@package logs Output logs and advanced debug information
//ErrorLog(err error, args ...interface{}) - output formated(function and line calls) error information
//ErrorStack() - output formated(function and line calls) error runtime stack information
//FatalLog(err error, args ...interface{}) - output formated (function and line calls) fatal information
//StatusLog(err error, args ...interface{}) - output formated information for status
package logs

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

type LogsType interface {
	PrintToLogs() string
}

var F_debug = flag.Bool("debug", false, "debug mode")
var F_status = flag.Bool("status", true, "status mode")

//var l =
//DebugLog( args ...interface{}) - output formated(function and line calls) debug information
//@version 1.1 2017-05-31 Sergey Litvinov - Remote requred args
func DebugLog(args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(1)
	if *F_debug {

		log.Output(2, fmt.Sprintf("[DEBUG];%s;%v", changeShortName(runtime.FuncForPC(pc).Name()),
			getArgsString(args)))
	}
}

//StatusLog(err error, args ...interface{}) - output formated information for status
//@version 1.0 2017-05-31 Sergey Litvinov - Create
func StatusLog(args ...interface{}) {
	//_, fn, line, _ := runtime.Caller(1)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if *F_status {
		log.Output(2, fmt.Sprintf("[STATUS];;;;%s", getArgsString(args...)))
	}
}

//ErrorLog(err error, args ...interface{}) - output formated(function and line calls) error information
//@version 1.1 2017-05-31 Sergey Litvinov - Remote requred advanced arg
func ErrorLog(err error, args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(1)
	calldepth   := 2

	funcName := changeShortName(runtime.FuncForPC(pc).Name())
	if funcName == "views.RenderInternalError" {
		pc, _, _, _ = runtime.Caller(2)
		funcName = changeShortName(runtime.FuncForPC(pc).Name())
		calldepth++
	}

	log.Output(calldepth, fmt.Sprintf("[ERROR];%s;%s;%s", funcName,
		err.Error(), getArgsString(args...)))

}

//ErrorLog(err error, args ...interface{}) - output formated(function and line calls) error information
//@version 1.1 2017-05-31 Sergey Litvinov - Remote requred advanced arg
func ErrorLogHandler(err error, args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(5)

	log.Output(5, fmt.Sprintf("[ERROR];%s;%s;%s", changeShortName(runtime.FuncForPC(pc).Name()),
		err, getArgsString(args...)))

}

//ErrorStack() - output formated(function and line calls) error runtime stack information
//@version 1.00 2017-06-02 Sergey Litvinov - Create
func ErrorStack() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	i := 0
	for {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		funcName := changeShortName(runtime.FuncForPC(pc).Name())
		// пропускаем рендер ошибок
		if funcName == "views.RenderInternalError" {
			continue
		}
		log.Output(i, fmt.Sprintf("[ERROR_STACK];%s;;;;", funcName) )
		i++
	}
}

//FatalLog(err error, args ...interface{}) - output formated (function and line calls) fatal information
//@version 1.0 2017-05-31 Sergey Litvinov - Create
func Fatal(err error, args ...interface{}) {
	pc, _, _, _ := runtime.Caller(2)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Output(2, fmt.Sprintf("[FATAL];%v;%v;%v", changeShortName(runtime.FuncForPC(pc).Name()),
		err, getArgsString(args...)))
	os.Exit(1)

}

//changeShortName(file string) (short string) - return Short Name
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

//changeShortName(file string) (short string) - Convert args to string
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

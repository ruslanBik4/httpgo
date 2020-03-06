// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logs

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

var (
	ignoreFunc = []string{
		"views.RenderHandlerError",
		"views.RenderInternalError",
		"RenderHandlerError",
		"RenderInternalError",
		"ErrorStack",
		"ErrorLogHandler",
		"v1.Catch",
		"runtime.gopanic",
		"runtime.panicindex",
		"runtime.call32",
		"runtime.panicdottypeE",
		"v1.WrapAPIHandler.func1",
		"fasthttp.(*workerPool).workerFunc",
		"apis.(*Apis).Handler",
		"apis.(*Apis).Handler-fm",
		"apis.(*Apis).renderError",
	}
	ignoreFiles = []string{
		"asm_amd64",
		"asm_amd64.s",
		"iface.go",
		"map_fast32.go",
		"panic.go",
		"server.go",
		"signal_unix.go",
		"testing.go",
		"workerpool.go",
	}
)

type errLogPrint bool

// Fatal - output formated (function and line calls) fatal information
func Fatal(err error, args ...interface{}) {
	pc, _, _, _ := runtime.Caller(2)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logErr.Output(2, fmt.Sprintf("[FATAL];%v;%v;%v", changeShortName(runtime.FuncForPC(pc).Name()),
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

// DebugLog output formated(function and line calls) debug information
func DebugLog(args ...interface{}) {
	if *fDebug {
		pc, _, _, _ := runtime.Caller(logDebug.calldepth - 2)
		logDebug.funcName = changeShortName(runtime.FuncForPC(pc).Name())

		logDebug.Printf(args...)
	}
}

// TraceLog output formatted(function and line calls) debug information
func TraceLog(args ...interface{}) {
	//var message string
	if *fDebug {
		logDebug.Printf(append([]interface{}{"%+v"}, args...)...)
	}
	logDebug.Printf("%v %s: %s:", time.Now(), args[0], args[1])
}

// StatusLog output formatted information for status
func StatusLog(args ...interface{}) {
	if *fStatus {
		logStat.Printf(args...)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func timeFormating() string {
	hh, mm, ss := time.Now().Clock()
	return fmt.Sprintf("%.2d:%.2d:%.2d", hh, mm, ss)
}

// ErrorLog - output formatted (function and line calls) error information
func ErrorLog(err error, args ...interface{}) {

	if err == nil {
		return
	}

	var (
		message         = ""
		timeFrameString string
		errorPrint      errLogPrint
	)

	errorPrint = true

	if logErr.toSentry {
		defer sentry.Flush(2 * time.Second)
		message = fmt.Sprintf("https://sentry.io/organizations/%s/?query=%s",
			logErr.sentryOrg, string(*(sentry.CaptureException(err))))
	}

	isIgnore := true

	ErrFmt, ok := err.(stackTracer)
	if ok {
		frames := ErrFmt.StackTrace()
		for _, frame := range frames {
			file := fmt.Sprintf("%s", frame)
			fncName := fmt.Sprintf("%n", frame)
			if !isIgnoreFile(file) && !isIgnoreFunc(fncName) {
				if logErr.Flags()&log.Ltime != 0 {
					timeFrameString = fmt.Sprintf("%s %s:%d: %s()", timeFormating(), file, frame, fncName)
				} else {
					timeFrameString = fmt.Sprintf("%s:%d: %s()", file, frame, fncName)
				}
				logErr.Printf(errorPrint, logErr.Prefix()+timeFrameString, err, args)
				return
			}
		}
	}

	calldepth := 1
	for pc, _, _, ok := runtime.Caller(calldepth); ok && isIgnore; pc, _, _, ok = runtime.Caller(calldepth) {
		logErr.funcName = changeShortName(runtime.FuncForPC(pc).Name())
		// пропускаем рендер ошибок
		isIgnore = isIgnoreFunc(logErr.funcName)
		calldepth++
	}

	logErr.calldepth = calldepth + 1

	logErr.Printf(message+" "+logErr.funcName+"()", err, args)
}

const prefErrStack = "[[ERR_STACK]]"

// ErrorStack - output formatted(function and line calls) error runtime stack information
func ErrorStack(err error, args ...interface{}) {

	i := stackBeginWith
	err = logErr.Output(i+1, fmt.Sprintf("%s; %s", err, getArgsString(args...)))
	if err != nil {
		fmt.Printf("%s during log printing", err)
		return
	}

	ErrFmt, ok := err.(stackTracer)
	if ok {
		frames := ErrFmt.StackTrace()
		for _, frame := range frames[:len(frames)-2] {
			printStackLine(fmt.Sprintf("%s", frame), fmt.Sprintf("%d", frame), fmt.Sprintf("%n", frame))
		}

		return
	}

	for pc, file, line, ok := runtime.Caller(i); ok; pc, file, line, ok = runtime.Caller(i) {
		i++
		// пропускаем рендер ошибок
		printStackLine(changeShortName(file), strconv.Itoa(line), changeShortName(runtime.FuncForPC(pc).Name()))
	}
}

func printStackLine(file string, line string, fncName string) {
	if !isIgnoreFile(file) && !isIgnoreFunc(fncName) {
		if *fDebug {
			hh, mm, ss := time.Now().Clock()
			fmt.Printf("%s %d:%d:%d %s:%s: %s()", prefErrStack, hh, mm, ss, file, line, fncName)
			fmt.Println()
		} else {
			fmt.Printf("%s:%s: %s()", file, line, fncName)
		}
	}
}

func isIgnoreFile(runFile string) bool {
	for _, name := range ignoreFiles {
		if (runFile == name) || (strings.HasPrefix(runFile, name)) {
			return true
		}
	}
	return false
}

func isIgnoreFunc(funcName string) bool {
	for _, name := range ignoreFunc {
		if (funcName == name) || (strings.HasSuffix(funcName, "."+name)) {
			return true
		}
	}
	return false
}

// ErrorLogHandler - output formated(function and line calls) error information
func ErrorLogHandler(err error, args ...interface{}) {
	ErrorStack(err, args...)
}

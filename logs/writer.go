// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logs

import (
	"fmt"
	"log"
	"os"
	"runtime"
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

func timeLogFormat() string {
	if logErr.Flags()&log.Ltime != 0 {
		hh, mm, ss := time.Now().Clock()
		return fmt.Sprintf("%.2d:%.2d:%.2d ", hh, mm, ss)
	}

	return ""
}

// ErrorLog - output formatted (function and line calls) error information
func ErrorLog(err error, args ...interface{}) {

	if err == nil {
		return
	}

	format, c := getFormatString(args)
	if c > 0 {
		// add format for error
		if c < len(args) {
			format = argToString(err) + "," + format
			args = args[1:]
		} else {
			args[0] = err
		}
	} else {
		format = argsToString(err, args)
		args = args[:0]
	}

	if logErr.toSentry {
		defer sentry.Flush(2 * time.Second)
		args = append(args, logErr.sentryOrg, string(*(sentry.CaptureException(err))))
		format += " https://sentry.io/organizations/%s/?query=%s"
	}

	ErrFmt, ok := err.(stackTracer)
	if ok {
		errorPrint := errLogPrint(true)
		frames := ErrFmt.StackTrace()
		for _, frame := range frames {
			file := fmt.Sprintf("%s", frame)
			fncName := fmt.Sprintf("%n", frame)
			if !isIgnoreFile(file) && !isIgnoreFunc(fncName) {

				args = append([]interface{}{
					errorPrint,
					logErr.Prefix() + "%s%s:%d: %s() " + format,
					timeLogFormat(),
					file,
					frame,
					fncName,
				},
					args...)

				break
			}
		}
	} else {

		calldepth := 1
		isIgnore := true

		for pc, _, _, ok := runtime.Caller(calldepth); ok && isIgnore; pc, _, _, ok = runtime.Caller(calldepth) {
			logErr.funcName = changeShortName(runtime.FuncForPC(pc).Name())
			// пропускаем рендер ошибок
			isIgnore = isIgnoreFunc(logErr.funcName)
			calldepth++
		}

		logErr.calldepth = calldepth + 1

		args = append([]interface{}{
			logErr.funcName + "() " + format,
		},
			args...)
	}

	logErr.Printf(args...)
}

const prefErrStack = "[[ERR_STACK]]"

// ErrorStack - output formatted(function and line calls) error runtime stack information
func ErrorStack(err error, args ...interface{}) {

	i := stackBeginWith
	logErr.calldepth = i + 1
	logErr.Printf(err, args)

	stackLines := []string{}
	stackline := ""
	sep := ""
	ErrFmt, ok := err.(stackTracer)
	if ok {
		frames := ErrFmt.StackTrace()
		for _, frame := range frames[:len(frames)-2] {
			file := fmt.Sprintf("%s", frame)
			fncName := fmt.Sprintf("%n", frame)
			if !isIgnoreFile(file) && !isIgnoreFunc(fncName) {
				newLine := fmt.Sprintf("%s%s:%d %s()", sep, file, frame, fncName)
				stackLines = append(stackLines, newLine)
				stackline += newLine
				sep = " <- "
			}
		}
		printStackLine(stackLines, stackline)
		return
	}

	for pc, file, line, ok := runtime.Caller(i); ok; pc, file, line, ok = runtime.Caller(i) {
		i++
		fileName := changeShortName(file)
		fncName := changeShortName(runtime.FuncForPC(pc).Name())
		// пропускаем рендер ошибок
		if !isIgnoreFile(fileName) && !isIgnoreFunc(fncName) {
			newLine := fmt.Sprintf("%s%s:%d %s(), ", sep, fileName, line, fncName)
			stackLines = append(stackLines, newLine)
			stackline += newLine
			sep = " <- "
		}
		printStackLine(stackLines, stackline)
	}
}

func printStackLine(stackLines []string, line string) {
	logErr.Printf("%s %s", prefErrStack, line)
	if *fDebug {
		for i, fncLine := range stackLines {
			logDebug.Printf("%s%s %s [%d stackLevel]", prefErrStack, timeLogFormat(), fncLine, i)
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

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
		"apis.(*Apis).Handler.func1()",
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
		pc, _, _, _ := runtime.Caller(logDebug.callDepth - 2)
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
	// todo get flags of current log !
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

		callDepth := 1
		isIgnore := true

		for pc, file, line, ok := runtime.Caller(callDepth); ok && isIgnore; pc, file, line, ok = runtime.Caller(callDepth) {
			logErr.fileName = changeShortName(file)
			logErr.funcName = changeShortName(runtime.FuncForPC(pc).Name())
			logErr.line = line
			// пропускаем рендер ошибок
			isIgnore = isIgnoreFile(logErr.fileName) || isIgnoreFunc(logErr.funcName)
			callDepth++
		}

		logErr.callDepth = callDepth + 1

		args = append([]interface{}{
			logErr.funcName + "() " + format,
		},
			args...)
	}

	logErr.Printf(args...)
}

const prefErrStack = "[[ERR_STACK]]"

// ErrorStack - output formatted (function and line calls) error runtime stack information
func ErrorStack(err error, args ...interface{}) {

	stackLine, c := getFormatString(args)
	if c > 0 {
		// add format for error
		if c < len(args) {
			stackLine = prefErrStack + argToString(err) + "," + stackLine
			args = args[1:]
		} else {
			args[0] = err
		}
	} else {
		stackLine = argsToString(err, args)
		args = args[:0]
	}

	stackLine += "\n"

	ErrFmt, ok := err.(stackTracer)
	if ok {
		frames := ErrFmt.StackTrace()
		for _, frame := range frames[:len(frames)-2] {
			fileName := fmt.Sprintf("%s", frame)
			fncName := fmt.Sprintf("%n", frame)
			if !isIgnoreFile(fileName) && !isIgnoreFunc(fncName) {
				stackLine += fmt.Sprintf("%s:%d %s %s()\n", fileName, frame, prefErrStack, fncName)
			}
		}
	} else {
		stackLine = GetStack(stackBeginWith, stackLine)
	}

	logErr.Printf(errLogPrint(true), stackLine)
}

func GetStack(i int, stackLine string) string {
	for pc, file, line, ok := runtime.Caller(i); ok; pc, file, line, ok = runtime.Caller(i) {
		i++
		fileName := changeShortName(file)
		fncName := changeShortName(runtime.FuncForPC(pc).Name())
		// пропускаем рендер ошибок
		if !isIgnoreFile(fileName) && !isIgnoreFunc(fncName) {
			stackLine += fmt.Sprintf("%s:%d %s %s()\n", fileName, line, prefErrStack, fncName)
		}
	}

	return stackLine
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

// ErrorLogHandler - output formatted(function and line calls) error information
func ErrorLogHandler(err error, args ...interface{}) {
	ErrorStack(err, args...)
}

func CustomLog(level Level, prefix, fileName string, line int, msg string, logFlags ...FgLogWriter) {
	args := []interface{}{
		errLogPrint(true),
		"[[%s%d%s%s]]%s %s:%d: %s",
		LogPutColor,
		boldcolors[level],
		prefix,
		LogEndColor,
		timeLogFormat(),
		fileName,
		line,
		msg,
	}

	for _, logFlag := range logFlags {
		switch logFlag {
		case FgAll:
			logErr.Printf(args...)
			logStat.Printf(args...)
			logDebug.Printf(args...)
		case FgErr:
			logErr.Printf(args...)
		case FgInfo:
			logStat.Printf(args...)
		case FgDebug:
			logDebug.Printf(args...)
		}

	}
}

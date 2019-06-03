// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logs

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
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

// DebugLog output formated(function and line calls) debug information
func TraceLog(args ...interface{}) {
	if *fDebug {
		logDebug.Printf(append([]interface{}{"%+v"}, args...)...)
	}
	logDebug.Printf("%v %s: %s:", time.Now(), args[0], args[1])
}

// StatusLog output formated information for status
func StatusLog(args ...interface{}) {
	if *fStatus {
		logStat.Printf(args...)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ErrorLog - output formatted (function and line calls) error information
func ErrorLog(err error, args ...interface{}) {
	calldepth := logErr.calldepth

	isIgnore := true

	ErrFmt, ok := err.(stackTracer)
	if ok {
		frames := ErrFmt.StackTrace()
		for _, frame := range frames {
			if !isIgnoreFunc(fmt.Sprintf("%s", frame)) {
				fmt.Printf("%s %v:%[1]n: %v", logErr.Prefix(), frame, err)
				return
			}
		}
	}

	for pc, _, _, ok := runtime.Caller(calldepth); ok && isIgnore; pc, _, _, ok = runtime.Caller(calldepth) {
		calldepth++
		logErr.funcName = changeShortName(runtime.FuncForPC(pc).Name())
		// пропускаем рендер ошибок
		isIgnore = isIgnoreFunc(logErr.funcName)
	}
	//todo: print standart
	var message string
	if err != nil {
		message = fmt.Sprintf("%s;%s;%s", logErr.funcName,
			err.Error(), getArgsString(args...))
	} else {
		message = fmt.Sprintf("%s;%s", logErr.funcName,
			getArgsString(args...))

	}

	logErr.Output(calldepth, message)

}

const prefErrStack = "[ERROR_STACK]"

// ErrorStack - output formatted(function and line calls) error runtime stack information
func ErrorStack(err error, args ...interface{}) {

	i := stackBeginWith
	logErr.Output(i+1, fmt.Sprintf("%s;%s;%s ", prefErrStack, err, getArgsString(args...)))

	ErrFmt, ok := err.(stackTracer)
	if ok {
		frames := ErrFmt.StackTrace()
		for _, frame := range frames[:len(frames)-2] {
			file := fmt.Sprintf("%s", frame)
			fncName := fmt.Sprintf("%n", frame)
			if !isIgnoreFile(file) && !isIgnoreFunc(fncName) {

				hh, mm, ss := time.Now().Clock()
				fmt.Printf("%s %d:%d:%d %s:%d: %s()", prefErrStack, hh, mm, ss, file, frame, fncName)
				if *fDebug {
					fmt.Println()
				}
			}
		}

		return
	}

	for pc, file, line, ok := runtime.Caller(i); ok; pc, file, line, ok = runtime.Caller(i) {
		i++
		// пропускаем рендер ошибок
		printStack(file, pc, i, line)
	}
}

func printStack(file string, pc uintptr, i int, line int) {
	runFile := changeShortName(file)
	isIgnore := isIgnoreFile(runFile)
	if !isIgnore {
		funcName := changeShortName(runtime.FuncForPC(pc).Name())
		if !isIgnoreFunc(funcName) {
			logErr.Output(i, fmt.Sprintf("%s:%d: %s; ", runFile, line, funcName))
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

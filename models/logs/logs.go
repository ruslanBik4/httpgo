// Copyright 2017 Author: Sergey Litvinov. All rights reserved.
// Выдача логов с дополнительной информацией

//ErrorLog - output formatted(function and line calls) error information
//ErrorLog(err error, args ...interface{}) - output formatted(function and line calls) error information
package logs

import (
    "runtime"
    "log"
    "flag"
)

var F_debug    = flag.String("debug","","debug mode")

//DebugLog( args ...interface{}) - output formatted(function and line calls) debug information
//@version 1.1 2017-05-31 Sergey Litvinov - Remote requred args
func DebugLog( args ...interface{}) {
    _, fn, line, _ := runtime.Caller(1)
    if *F_debug > "" {
        log.Printf("[DEBUG], %s, in line, %d, %v", fn, line, args)
    }
}

//ErrorLog(err error, args ...interface{}) - output formatted(function and line calls) error information
//@version 1.1 2017-05-31 Sergey Litvinov - Remote requred advanced arg
func ErrorLog(err error, args ...interface{}) {
    _, fn, line, _ := runtime.Caller(1)
    log.Printf("[ERROR], %s, in line, %d, %v, %v", fn, line, err, args)
}

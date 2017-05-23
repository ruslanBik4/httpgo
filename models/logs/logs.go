package logs

import (
    "runtime"
    "log"
)

func DebugLog(title, data interface{}) {
    _, fn, line, _ := runtime.Caller(1)
    log.Printf("[DEBUG], %s, in line %d.  %s= `%v`", fn, line, title, data)
}

func ErrorLog(err error, data interface{}) {
    _, fn, line, _ := runtime.Caller(1)
    log.Printf("[ERROR] %s, in line %d. Error -  %v. Data= %v", fn, line, err, data)
}


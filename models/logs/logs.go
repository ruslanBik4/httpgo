// Copyright 2017 Author: Sergey Litvinov. All rights reserved.
//@package logs Output logs and advanced debug information
//ErrorLog(err error, args ...interface{}) - output formated(function and line calls) error information
//ErrorStack() - output formated(function and line calls) error runtime stack information
//FatalLog(err error, args ...interface{}) - output formated (function and line calls) fatal information
//StatusLog(err error, args ...interface{}) - output formated information for status
package logs

import (
	"runtime"
	"log"
	"flag"
	"fmt"
	"os"
)

var F_debug = flag.String("debug", "", "debug mode")
var F_status = flag.String("status", " ", "status mode")

//var l =
//DebugLog( args ...interface{}) - output formated(function and line calls) debug information
//@version 1.1 2017-05-31 Sergey Litvinov - Remote requred args
func DebugLog(args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(1)
	i:=0
	if *F_debug > "" {
		for _, arg := range args {
			log.Output(2, fmt.Sprintf("[DEBUG];%s;%d;%v", changeShortName(runtime.FuncForPC(pc).Name()),i, arg))
			i++
		}
	}
}

//StatusLog(err error, args ...interface{}) - output formated information for status
//@version 1.0 2017-05-31 Sergey Litvinov - Create
func StatusLog(args ...interface{}) {
	//_, fn, line, _ := runtime.Caller(1)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if *F_status > "" {
		log.Output(2, fmt.Sprintf("[STATUS];;;;%v", args))
	}
}

//ErrorLog(err error, args ...interface{}) - output formated(function and line calls) error information
//@version 1.1 2017-05-31 Sergey Litvinov - Remote requred advanced arg
func ErrorLog(err error, args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(1)


		//log.Print(errLog)
		i:=0
		for _, arg := range args {
			log.Output(2, fmt.Sprintf("[ERROR];%s;%s;%d;%v", changeShortName(runtime.FuncForPC(pc).Name()), err,i, arg))
			i++
		}


}

//ErrorLog(err error, args ...interface{}) - output formated(function and line calls) error information
//@version 1.1 2017-05-31 Sergey Litvinov - Remote requred advanced arg
func ErrorLogHandler(err error, args ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	pc, _, _, _ := runtime.Caller(5)

	//errLog := log.Output(5, fmt.Sprintf("%s;%s", changeShortName(runtime.FuncForPC(pc).Name()), err, args))
	//if errLog != nil {
		//log.Print(errLog)
	i:=0
		for _, arg := range args {
			log.Output(5, fmt.Sprintf("[ERROR];%s;%s;%d;%v", changeShortName(runtime.FuncForPC(pc).Name()), err,i, arg))
			i++
		}
	//}

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
		log.Output(i+1, fmt.Sprintf("[ERROR_STACK];%s;;;;", changeShortName(runtime.FuncForPC(pc).Name())))
		i++
	}
}

//FatalLog(err error, args ...interface{}) - output formated (function and line calls) fatal information
//@version 1.0 2017-05-31 Sergey Litvinov - Create
func Fatal(err error, args ...interface{}) {
	//pc, _, _, _ := runtime.Caller(2)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Output(2, fmt.Sprintf("[FATAL]%v;%v;%v", err, args))

	os.Exit(1)
	//log.Fatalf("[FATAL];%v;%v;%v", changeShortName(runtime.FuncForPC(pc).Name()), err, args)
}

func changeShortName(file string) (short string) {
	short = file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	//file1 = short
	return short
}

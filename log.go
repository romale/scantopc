// log
package main

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"time"
)

var (
	TRACE   *log.Logger
	INFO    *log.Logger
	WARNING *log.Logger
	ERROR   *log.Logger

	traceGoRoutines bool = false
)

func logInit(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	flagFile := 0
	if paramModeTrace {
		flagFile = log.Lshortfile
		//go TraceGoRoutine()
	}
	TRACE = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|flagFile)

	INFO = log.New(infoHandle,
		"",
		log.Ldate|log.Ltime|flagFile)

	WARNING = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|flagFile)

	ERROR = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|flagFile)
}

func Trace(f string, v ...interface{}) string {
	if traceGoRoutines {
		TRACE.Println("GO Routines", runtime.NumGoroutine())
	}
	TRACE.Output(3, fmt.Sprint("Entering ", f, v, "\n"))
	//TRACE.Println("Entering:", f, v)
	return f
}

func Un(f string) {
	//	TRACE.Println("Leaving:", f)
	TRACE.Output(3, fmt.Sprintf("Leaving %v\n", f))
	if traceGoRoutines {
		TRACE.Println("GO Routines", runtime.NumGoroutine())
	}
}

func TraceGoRoutine() {
	if traceGoRoutines {
		ticker := time.NewTicker(10 * time.Second)
		for _ = range ticker.C {
			TRACE.Println("GO Routines:", runtime.NumGoroutine())
		}
		ticker.Stop()
	}
}

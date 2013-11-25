// log
package main

import (
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
	TRACE.Println("GO Routines", runtime.NumGoroutine())
	TRACE.Println("Entering:", f, v)
	return f
}

func Un(f string) {
	TRACE.Println("Leaving:", f)
	TRACE.Println("GO Routines", runtime.NumGoroutine())
}

func TraceGoRoutine() {
	ticker := time.NewTicker(10 * time.Second)
	for _ = range ticker.C {
		TRACE.Println("GO Routines:", runtime.NumGoroutine())
	}
	ticker.Stop()
}

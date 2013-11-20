// log
package main

import (
	"io"
	"log"
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

	TRACE = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	INFO = log.New(infoHandle,
		"",
		log.Ldate|log.Ltime)

	WARNING = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime)

	ERROR = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime)
}

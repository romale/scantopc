// log
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	TRACE   *log.Logger
	INFO    *log.Logger
	WARNING *log.Logger
	ERROR   *log.Logger

	traceGoRoutines bool = false
	logFile         *os.File
)
var (
	pid      = os.Getpid()
	program  = filepath.Base(os.Args[0])
	host     = "unknownhost"
	userName = "unknownuser"
)

//Provide Log at the very begining of execution
func init() {
	h, err := os.Hostname()
	if err == nil {
		host = h
	}

	current, err := user.Current()
	if err == nil {
		userName = current.Username
	}

	// Sanitize userName since it may contain filepath separators on Windows.
	userName = strings.Replace(userName, `\`, "_", -1)

	logInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
}

func LogBanner() {
	INFO.Println("Starting\n\tProgram:", program, "\n\tPID:", pid, "\n\tHost:", host, "\n\tUser:", userName)
}

func InitLogFiles() {
	var err error
	logFile, err = os.Create("/var/log/scantopc.log")
	if err != nil {
		fmt.Println(err)
		logFile, err = os.Create(os.TempDir() + "/scantopc.log")
		if err != nil {
			panic(err)
		}
	}
	if !paramModeTrace {
		logInit(ioutil.Discard, logFile, logFile, logFile)
	} else {
		logInit(logFile, logFile, logFile, logFile)
		TRACE.Println("Trace enabled")
	}
	LogBanner()
}

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

	if paramModeTrace {
		INFO = log.New(infoHandle,
			"INFO: ",
			log.Ldate|log.Ltime|flagFile)
	} else {
		INFO = log.New(infoHandle,
			"",
			log.Ldate|log.Ltime|flagFile)
	}
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

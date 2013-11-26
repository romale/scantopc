// scantopc project main.go
// Implements ScanToPC for HP printers on linux

package main

import (
	"flag"
	"fmt"
	"github.com/simulot/srvloc"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const VERSION = "0.2.0 DEV"

func CheckError(context string, err error) {
	if err != nil {
		ERROR.Panicln("panic", context, "->", err)
	}
}

func httpGet(url string) (resp *http.Response, err error) {
	resp, err = http.Get(url)
	if flagTraceHTTP > 0 {
		if err != nil {
			ERROR.Panicln("http.Get(", url, ") ->", err)
			return nil, err
		} else {
			TRACE.Println("http.Get(", url, ") ->", resp.Status, err)
			if flagTraceHTTP > 1 {
				for h, v := range resp.Header {
					TRACE.Println("\t", h, " --> ", v)
				}
			}
		}
	}
	return
}

func httpPost(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	resp, err = http.Post(url, bodyType, body)
	if flagTraceHTTP > 0 {
		TRACE.Println("http.Post(", url, ",", bodyType, ") ->", resp.Status, err)
		if flagTraceHTTP > 1 {
			for h, v := range resp.Header {
				TRACE.Println("\t", h, " --> ", v)
			}
		}
	}
	return
}

func ioutilReadAll(r io.Reader) (ByteArray []byte, err error) {
	ByteArray, err = ioutil.ReadAll(r)
	if flagTraceHTTP > 1 {
		TRACE.Println("\tResponse:\n", string(ByteArray))
	}
	return
}

// Extract UUID placed at the right end of the URI
// Will be used to check wich client is concerned
func getUUIDfromURI(uri string) string {
	return uri[strings.LastIndex(uri, "/")+1:]
}

////////////////////////////////////////////////////////////////////////////////*

func hostname() string {
	s, _ := os.Hostname()
	return s
}

var (
	flagTraceHTTP     int         = 1
	filePERM          os.FileMode = 0777
	fileUserGroup     string      = ""
	paramModeTrace    bool
	paramComputerName string
	paramPrinterURL   string
	paramFolderPatern string
	paramDoubleSide   bool
)

func init() {
	flag.BoolVar(&paramModeTrace, "trace", false, "Enable traces")
	flag.StringVar(&paramComputerName, "name", hostname(), "Name of the computer visible on the printer (default: $hostname)")
	flag.StringVar(&paramPrinterURL, "printer", "", "Printer URL like http://1.2.3.4:8080, when omitted, the device is searched on the network")
	flag.StringVar(&paramFolderPatern, "destination", "", "Folder where images are strored (see help for tokens)")
	flag.StringVar(&paramFolderPatern, "d", "", "shorthand for -destination")
	flag.BoolVar(&paramDoubleSide, "D", true, "shorthand for -doubleside")
	flag.BoolVar(&paramDoubleSide, "doubleside", true, "enable double side scanning with one side scannig")
	//paramModeTrace = true

}

func usage() {
	// Fprintf allows us to print to a specifed file handle or stream
	fmt.Fprintf(os.Stderr, "\nUsage of %s:\n", os.Args[0])
	// PrintDefaults() may not be exactly what we want, but it could be
	flag.PrintDefaults()
	fmt.Println("\nExemple:")
	fmt.Println("\t", os.Args[0], "-destination ~/Documents/%Y/%Y.%m/%Y.%m.%d-%H.%M.%S")
	s, _ := ExpandString("~/Documents/%Y/%Y.%m/%Y.%m.%d-%H.%M.%S.pdf", time.Now())
	fmt.Println("\twill generate files like", s)
	TokensUsage()
	os.Exit(1)
}

func main() {
	flag.Parse()
	if !paramModeTrace {
		logInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	} else {
		logInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
		TRACE.Println("Trace enabled")
	}
	if paramComputerName == "" {
		paramComputerName, _ = os.Hostname()
	}

	if paramFolderPatern == "" {
		WARNING.Println("No destination given, assuming: -destination=./%Y%m%d-%H%M%S")
		paramFolderPatern = "./%Y%m%d-%H%M%S"
	} else {
		// Test the pattern to detect issues immediatly
		s, err := ExpandString(paramFolderPatern, time.Now())
		if err != nil {
			ERROR.Println(err)
			usage()
		}
		TRACE.Println("Save to ", s)
	}

	INFO.Println(os.Args[0], "version", VERSION, "started")
	MainLoop()
	INFO.Println(os.Args[0], "stopped")

}

////////////////////////////////////////////////////////////////////////////////
/*
  Check each XX minutes if an HP printer appears on the network.
  if the printer is detected, the print will be pulled for events.

*/
func MainLoop() {
	defer Un(Trace("MainLoop"))

	for {
		printer := paramPrinterURL
		if printer == "" {
			INFO.Println("Searching printer on the network")
			device, err := srvloc.ProbeHPPrinter()
			TRACE.Printf("%+v\n", device)
			if err == nil {
				// We have one
				printer = fmt.Sprintf("http://%s:8080", device.IPAddress)
			} else {
				ERROR.Println("Device not found", err)
			}
		}

		if printer != "" {
			INFO.Println("Connecting to", printer)
			d := NewDevice("printer", printer, DefaultDestinationSettings)
			d.DeviceEventLoop()
		}
		INFO.Println("Connection to ", printer, "lost. Sleeping...")
		printer = ""
		time.Sleep(time.Minute * 1)
	}
}

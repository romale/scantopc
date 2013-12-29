// scantopc project main.go
// Implements ScanToPC for HP printers on linux

package main

import (
	"flag"
	"fmt"
	"github.com/simulot/hpdevices"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const VERSION = "0.4.0 DEV"

func CheckError(context string, err error) {
	if err != nil {
		ERROR.Panicln("panic", context, "->", err)
	}
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
	flagTraceHTTP     int         = 0
	filePERM          os.FileMode = 0777
	fileUserGroup     string      = ""
	paramModeTrace    bool
	paramComputerName string
	paramPrinterURL   string
	paramFolderPatern string
	paramDoubleSide   bool
	paramOCR          bool
	paramPFDTool      string
)

func main() {
	banner()
	GetParameters()
	MainLoop()
	INFO.Println(os.Args[0], "stopped")

}

func init() {
	flag.BoolVar(&paramModeTrace, "trace", false, "Enable traces")
	flag.StringVar(&paramComputerName, "name", hostname(), "Name of the computer visible on the printer (default: $hostname)")
	flag.StringVar(&paramPrinterURL, "printer", "", "Printer URL like http://1.2.3.4:8080, when omitted, the device is searched on the network")
	flag.StringVar(&paramFolderPatern, "destination", "", "Folder where images are strored (see help for tokens)")
	flag.StringVar(&paramFolderPatern, "d", "", "shorthand for -destination")
	flag.BoolVar(&paramDoubleSide, "D", true, "shorthand for -doubleside")
	flag.BoolVar(&paramDoubleSide, "doubleside", true, "enable double side scanning with one side scannig")
	flag.StringVar(&paramPFDTool, "pdftool", "", "precise which tool to be used when joining pages (supported: pdftk,pdfunite)")
	flag.BoolVar(&paramOCR, "ocr", true, "enable/disable OCR functionality")
	//paramModeTrace = true

}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Println("\nExemple:")
	fmt.Println("\t", os.Args[0], "-destination ~/Documents/%Y/%Y.%m/%Y.%m.%d-%H.%M.%S")
	s, _ := ExpandString("~/Documents/%Y/%Y.%m/%Y.%m.%d-%H.%M.%S.pdf", time.Now())
	fmt.Println("\twill generate files like", s)
	TokensUsage()
	os.Exit(1)
}

func banner() {
	INFO.Println(os.Args[0], "version", VERSION, "started")
}

func GetParameters() {
	flag.Parse()
	if !paramModeTrace {
		logInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	} else {
		logInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
		TRACE.Println("Trace enabled")
	}
	hpdevices.InitLogger(TRACE, INFO, WARNING, ERROR)
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
	if CheckOCRDependencies() {
		ERROR.Println("One or many depencies are not found. Please check your setup")
		usage()
	}
}

////////////////////////////////////////////////////////////////////////////////

func MainLoop() {
	defer Un(Trace("MainLoop"))

	for {
		var (
			Scanner *hpdevices.HPDevice
			err     error
		)

		if paramPrinterURL == "" {
			TRACE.Println("Searching printer on the network")
			Scanner, err = hpdevices.LocalizeDevice()
		} else {
			TRACE.Println("Connection to the printer")
			Scanner, err = hpdevices.NewHPDevice(paramPrinterURL)
		}
		if err != nil {
			ERROR.Println(err)
		}
		time.Sleep(time.Second * 5)
		if err == nil {
			INFO.Println("Found device at", Scanner.URL)
			d := []hpdevices.DestinationSettings{
				hpdevices.DestinationSettings{
					Name:        "OCR",
					FilePattern: &paramFolderPatern,
					DoOCR:       true,
					Verso:       false,
					Resolution:  300,
					ColorSpace:  "Gray",
				},
				hpdevices.DestinationSettings{
					Name:        "OCR (Verso)",
					FilePattern: &paramFolderPatern,
					DoOCR:       true,
					Verso:       true,
					Resolution:  300,
					ColorSpace:  "Gray",
				},
			}

			_, err := hpdevices.NewScanToPC(Scanner, NewOCRBatchImageManager, paramComputerName, d)
			if err != nil {
				ERROR.Println(err)
			}
		}
	}
}

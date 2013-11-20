// scantopc project main.go
// Implements ScanToPC for HP printers on linux

package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/simulot/srvloc"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const VERSION = "0.1.1"

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

type deviceInfo struct {
	url      string     // device address:port
	ScanCaps *HPScanCap // Scanner capabilities
}

type clientInfo struct {
	name       string // Name displayed on the printer
	uri        string // Client URI
	uuid       string // Specific UUDI for the registered Scan to computer location
	registered bool   // True when client registered
	device     *deviceInfo
	scanJob    *ScanJob
}

var (
	ErrNotADevice         error = errors.New("The device is not reachable")
	ErrRegistrationFailed error = errors.New("Regisgtration has failed")
	ErrTimeOut            error = errors.New("Operation Time out")
)

const (
	DiscoveryTree                = "/DevMgmt/DiscoveryTree.xml"
	WalkupScanToCompDestinations = "/WalkupScanToComp/WalkupScanToCompDestinations"
	EventTable                   = "/EventMgmt/EventTable"
	EventTimeOut                 = "timeout=1200"
)

// Extract UUID placed at the right end of the URI
// Will be used to check wich client is concerned
func getUUIDfromURI(uri string) string {
	return uri[strings.LastIndex(uri, "/")+1:]
}

func registerScanToComp(d *deviceInfo) (client *clientInfo, err error) {
	client = new(clientInfo)
	client.device = d
	// Call discovery tree
	resp, err := httpGet(d.url + DiscoveryTree)
	defer resp.Body.Close()
	if err == nil && resp.StatusCode == 200 {
		tree := new(HPDiscoveryTree)
		ByteArray, err := ioutilReadAll(resp.Body)
		CheckError("ReadBody", err)
		err = xml.Unmarshal(ByteArray, tree)
		resp.Body.Close()
		CheckError("Unmarshal", err)
		// Get ScanToComputer destinations
		{
			resp, err = httpGet(d.url + WalkupScanToCompDestinations)
			defer resp.Body.Close()
			CheckError("", err)
			if resp.StatusCode == 200 {
				destinations := new(HPWalkupScanToCompDestinations)
				s, err := ioutil.ReadAll(resp.Body)
				CheckError("ReadBody", err)
				err = xml.Unmarshal(s, destinations)
				CheckError("Unmarshal", err)
				//}
			}
		}
		{
			Me := &HPPostDestination{
				Name:     *computerName,
				Hostname: *computerName,
				LinkType: "Network",
			}

			ByteArray, err = xml.Marshal(Me)
			CheckError("Marshal Me", err)
			r := bytes.NewBufferString(XMLHeader + string(ByteArray))

			resp, err = httpPost(d.url+WalkupScanToCompDestinations, "text/xml", r)
			defer resp.Body.Close()
			CheckError("Post WalkupScanToCompDestinations", err)
			if resp.StatusCode == 201 {
				// SuccessFull registration
				client.uri = resp.Header.Get("Location")
				client.uuid = getUUIDfromURI(client.uri)
				INFO.Println("Registration successful:", Me.Hostname, client.uuid)
				client.registered = true
			} else {
				err = ErrRegistrationFailed
				ERROR.Println("Registration failed: bad response", err, resp.Status)
				client.registered = false
				os.Exit(1)
			}
		}
	} else {
		INFO.Println("Printer not accessible: ", err)
		os.Exit(1)

	}

	return client, err
}

func pullDeviceEvents(c *clientInfo) {
	lastEtag := ""
	for c.registered {
		// Loop while the connection is established
		resp, err := http.Get(c.device.url + EventTable)
		if err != nil {
			ERROR.Println(err)
			c.registered = false
			continue
		}
		Etag := resp.Header.Get("Etag")
		if Etag != lastEtag {
			TRACE.Println("Event(", Etag, ")")
			switch {
			case err != nil:
				ERROR.Println("Error polling Event table", err)
			case resp.StatusCode == 304:
				// Not modified: nothing to do.
			default:
				// 	We have an event
				EventTable := new(HPEventTable)
				ByteArray, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				CheckError("ReadBody EventTable ", err)
				err = xml.Unmarshal(ByteArray, &EventTable)
				CheckError("Unmarshal HPEventTable", err)

				// Parse evenCompletedt list for ScanEvent
				for _, Event := range EventTable.Events {
					err = doHandleEvent(c, &Event)
					if err != nil {
						c.registered = false
					}
				}

			}
		}
		resp.Body.Close()
		time.Sleep(1000 * time.Millisecond)
		lastEtag = Etag
	}
}

var lastScanEventAgingStamp = ""

func doHandleEvent(clientInfo *clientInfo, event *HPEvent) error {
	switch event.UnqualifiedEventCategory {
	case "ScanEvent":
		uri := ""
		if event.AgingStamp != lastScanEventAgingStamp {
			for _, payload := range event.Payloads {
				if payload.ResourceType == "wus:WalkupScanToCompDestination" {
					uri = payload.ResourceURI
				}
			}
			if clientInfo.uuid == getUUIDfromURI(uri) {
				TRACE.Println("Event ScanEvent accepted")
				doGetWalkupScanToCompEvent(clientInfo)
				lastScanEventAgingStamp = event.AgingStamp
			} else {
				TRACE.Println("Event ScanEvent ignored, uri doesn't match")
			}
		} else {
			TRACE.Println("Event ScanEvent ignored, AgingStamp already processed")
		}
		return nil
	case "PoweringDownEvent":
		INFO.Println("Event", event.UnqualifiedEventCategory, "... closing connection.")
		return errors.New("PoweringDownEvent")
	default:
		TRACE.Println("Event", event.UnqualifiedEventCategory, "is ignored")
		return nil
	}
}

func doGetWalkupScanToCompEvent(clientInfo *clientInfo) {
	resp, err := httpGet(clientInfo.device.url + "/WalkupScanToComp/WalkupScanToCompEvent")
	defer resp.Body.Close()

	if err == nil && resp.StatusCode == 200 {
		event := new(HPWalkupScanToCompEvent)
		s, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		CheckError("ReadBody", err)
		err = xml.Unmarshal(s, event)
		CheckError("Unmarshal WalkupScanToCompEvent", err)
		TRACE.Println("\tWalkupScanToCompEvent", event.WalkupScanToCompEventType)
		switch event.WalkupScanToCompEventType {
		case "HostSelected":
			// That's for us...
			TRACE.Println("WalkupScanToCompEvent", event.WalkupScanToCompEventType, "accepted, empty scanjob created")
		case "ScanRequested", "ScanNewPageRequested":
			// Scan Job start

			// Call doGetWalkupScanToCompDestScanEventinations with uri given in Event
			// we have to call WalkupScanToCompDestination
			resp, err := httpGet(clientInfo.uri)
			defer resp.Body.Close()
			if err == nil && resp.StatusCode == 200 {
				s, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				destination := new(HPWalkupScanToCompDestination)
				err = xml.Unmarshal(s, destination)
				CheckError("Unmarshal HPWalkupScanToCompDestination", err)

				if destination.WalkupScanToCompSettings != nil {
					// Get Scanning source
					source, err := doGetSource(clientInfo)
					CheckError("doGetScanStatus", err)
					if clientInfo.scanJob == nil ||
						event.WalkupScanToCompEventType == "ScanRequested" ||
						(event.WalkupScanToCompEventType == "ScanNewPageRequested" && time.Since(clientInfo.scanJob.lastPageTime) > time.Duration(2*time.Minute)) {
						// Create a new scanJob
						clientInfo.scanJob = createNewScanJob(destination.WalkupScanToCompSettings.Shortcut, destination.WalkupScanToCompSettings.ScanSettings.ScanPlexMode, source)
					}
					if clientInfo.scanJob.HPScanCap == nil {
						doGetScanCaps(clientInfo)
					}
					doPostScanJob(clientInfo)
					doJobHandling(clientInfo)
				}
			} else {
				ERROR.Println("Error ", err)
			}

		case "ScanPagesComplete":
			// Job is done
			if clientInfo.scanJob != nil && clientInfo.scanJob.documentType == "SavePDF" && clientInfo.scanJob.Status != "Canceled" {
				doSaveAsPDF(clientInfo.scanJob)
			}
			clientInfo.scanJob = nil
		default:
			TRACE.Println("WalkupScanToCompEvent", event.WalkupScanToCompEventType, "is ignored")
		}
	} else {
		ERROR.Println("Error ", err)
	}
}

func doGetScanCaps(clientInfo *clientInfo) {
	resp, err := httpGet(clientInfo.device.url + "/Scan/ScanCaps.xml")
	defer resp.Body.Close()
	HPScanCap := new(HPScanCap)
	s, err := ioutilReadAll(resp.Body)
	err = xml.Unmarshal(s, HPScanCap)
	CheckError("Unmarshal HPScanCap", err)
	clientInfo.scanJob.HPScanCap = HPScanCap
}

func createNewScanJob(Shortcut, scanPlexMode, scanSource string) *ScanJob {
	scanJob := new(ScanJob)

	now := time.Now()
	scanJob.FileName, _ = ExpandString(folderPatern, now)
	scanJob.documentType = Shortcut
	scanJob.scanPlexMode = scanPlexMode
	scanJob.scanSource = scanSource

	INFO.Println("New scanjob accepted, with", scanSource, Shortcut, scanPlexMode)

	return scanJob
}

type ScanJob struct {
	scanPlexMode string // Duplex / Simplex
	documentType string // SavePDF / SaveJPEG
	scanSource   string // Platen / Adf
	FileName     string // Image file
	jobURL       string
	pages        []string
	pageCount    int
	lastPageTime time.Time
	HPScanCap    *HPScanCap
	Status       string
}

type scanningOptions struct {
	AgingStamp   string
	ScanPlexMode string
	SaveAs       string
}

func doGetSource(clientInfo *clientInfo) (source string, err error) {
	HPScanStatus, err := doGetScanStatus(clientInfo)
	if err != nil {
		return "", err
	}

	if HPScanStatus.AdfState == "Empty" {
		source = "Platen"
	} else {
		source = "Adf"
	}
	return source, nil
}

func doGetScanStatus(clientInfo *clientInfo) (*HPScanStatus, error) {
	// Get Status
	resp, err := httpGet(clientInfo.device.url + "/Scan/Status")
	defer resp.Body.Close()
	if err == nil && resp.StatusCode == 200 {
		status := new(HPScanStatus)
		s, err := ioutil.ReadAll(resp.Body)
		CheckError("ReadBody", err)
		err = xml.Unmarshal(s, status)
		CheckError("Unmarshal HPScanStatus", err)
		TRACE.Printf("Scan/Status: %+v\n", status)
		return status, nil
	} else {
		ERROR.Println("get /Scan/Status has failed", err, resp.StatusCode)
		return nil, err
	}
}

func doPostScanJob(clientInfo *clientInfo) {
	var MaxResolution int
	var ColorColorSpace string
	var HPScanSourceCap *HPScanSourceCap
	var HPResolution *HPResolution
	var ContentType string

	switch clientInfo.scanJob.documentType {
	case "SaveJPEG":
		MaxResolution = 300
		ColorColorSpace = "Color"
		ContentType = "Photo"
	case "SavePDF":
		MaxResolution = 200
		ColorColorSpace = "Gray"
		ContentType = "Document"
	default:
		log.Println("Unknown save mode", clientInfo.scanJob.documentType)
	}

	switch clientInfo.scanJob.scanSource {
	case "Adf":
		HPScanSourceCap = &clientInfo.scanJob.HPScanCap.Adf.InputSourceCaps
	case "Platen":
		HPScanSourceCap = &clientInfo.scanJob.HPScanCap.Platen.InputSourceCaps
	default:
		log.Println("Unknown ScanSource", clientInfo.scanJob.scanSource)
	}
	// Check most adapted resolution

	for _, xHPResolution := range HPScanSourceCap.SupportedResolutions {
		HPResolution = &xHPResolution
		if HPResolution.XResolution > MaxResolution {
			break
		}
	}

	ScanSetting := HPScanSettings{
		XResolution:        HPResolution.XResolution,
		YResolution:        HPResolution.YResolution,
		XStart:             0,
		YStart:             0,
		Width:              2481,
		Height:             3507,
		Format:             "Jpeg",
		CompressionQFactor: 15,
		ColorSpace:         ColorColorSpace,
		BitDepth:           8,
		InputSource:        clientInfo.scanJob.scanSource,
		GrayRendering:      "NTSC",
		Gamma:              1000,
		Brightness:         1000,
		Contrast:           1000,
		Highlite:           179,
		Shadow:             25,
		Threshold:          0,
		SharpeningLevel:    128,
		NoiseRemoval:       0,
		ContentType:        ContentType, //		ContentType:        "Photo","Document"
	}
	ByteArray, err := xml.Marshal(ScanSetting)
	CheckError("Marshal ScanSetting", err)
	r := bytes.NewBufferString(XMLHeader + string(ByteArray))
	resp, err := httpPost(clientInfo.device.url+"/Scan/Jobs", "text/xml", r)
	defer resp.Body.Close()
	if err != nil || resp.StatusCode != 201 {
		ERROR.Println("Error when posting scan job", err, resp.Status)
	} else {
		clientInfo.scanJob.jobURL = resp.Header.Get("Location")
		INFO.Println("Job created, url:", clientInfo.scanJob.jobURL)
	}
}

func doJobHandling(clientInfo *clientInfo) {
	/*
		JobState:
			Processing
				PrescanPage
					PageState : PreparingScan,
					PageState : ReadyToUpload
						GET /Scan/Jobs/1/Pages/1
						Image JPEG
		JobState
			Completed
					PageState : UploadCompleted
		Completed
	*/
	JobCompleted, JobError, JobCancled := false, false, false
	previousJobState := ""
	for !JobCompleted && !JobError && !JobCancled {
		TRACE.Println("!JobCompleted && !JobError && !JobCancled", !JobCompleted && !JobError && !JobCancled)
		resp, err := httpGet(clientInfo.scanJob.jobURL)
		defer resp.Body.Close()
		if err == nil && resp.StatusCode == 200 {
			job := new(HPJob)
			s, err := ioutil.ReadAll(resp.Body)
			CheckError("ReadBody", err)
			err = xml.Unmarshal(s, job)
			resp.Body.Close()
			CheckError("Unmarshal HPJob", err)
			clientInfo.scanJob.Status = job.JobState
			switch clientInfo.scanJob.Status {
			case "Processing":
				switch {
				case job.ScanJob.PreScanPage != nil:
					if previousJobState != "Processing" {
						TRACE.Print("JobState ", job.JobState)
						TRACE.Print("PageState ", job.ScanJob.PreScanPage.PageState, ", PageNumber ", job.ScanJob.PreScanPage.PageNumber, "\n")
						TRACE.Print("Processing")
					}
					if job.ScanJob.PreScanPage.PageState == "ReadyToUpload" {
						TRACE.Print("JobState ", job.JobState, job.ScanJob.PreScanPage.PageState)
						doUploadImage(clientInfo.device.url+job.ScanJob.PreScanPage.BinaryURL, clientInfo.scanJob, job)

					}
				case job.ScanJob.PostScanPage != nil:
					if job.ScanJob.PostScanPage.PageState == "CanceledByDevice" {
						ERROR.Println("Scan job cancled by device")
						JobCancled = true
					}
					TRACE.Print("JobState ", job.JobState)
					TRACE.Println("Batch end,", job.ScanJob.PostScanPage.PageNumber, "page(s) scanned")
				}
			case "Completed":
				TRACE.Print("JobState ", job.JobState, clientInfo.scanJob)
				INFO.Println("Job completed")
				JobCompleted = true
			case "Canceled":
				TRACE.Print("JobState ", job.JobState, clientInfo.scanJob)
				ERROR.Println("Job canceled")
				JobCancled = true
				continue
			}
			previousJobState = job.JobState

		} else {
			ERROR.Println("Error ", resp.Status, err)
		}
		time.Sleep(500 * time.Millisecond)

		HPScanStatus, err := doGetScanStatus(clientInfo)
		if err != nil && strings.Contains(HPScanStatus.ScannerState, "Error") {
			ERROR.Println("Scanner error:", HPScanStatus.ScannerState, HPScanStatus.AdfState)
			JobError = true
		}
	}
	TRACE.Println("Quitting doJobHandling")
}

func doUploadImage(uri string, scanJob *ScanJob, HPJob *HPJob) {
	imageFile := fmt.Sprintf("%s.%04d.jpg", scanJob.FileName, scanJob.pageCount)
	err := CheckFolder(imageFile)
	CheckError("Create deistination folder", err)
	scanJob.pageCount++
	out, err := os.Create(imageFile)
	CheckError("CreateFile", err)
	defer out.Close()
	resp, err := httpGet(uri)
	defer resp.Body.Close()
	if err == nil && resp.StatusCode == 200 {
		size, err := CopyAndFixJPEG(out, resp.Body, HPJob.ScanJob.PreScanPage.BufferInfo.ImageHeight)
		if err == nil {
			log.Printf("Image downloaded, size %d\n", size)
			scanJob.lastPageTime = time.Now()
			scanJob.pages = append(scanJob.pages, imageFile)
		}
	} else {
		log.Println("Get", uri, "failed!", resp.Status, err)
	}
}

////////////////////////////////////////////////////////////////////////////////*

func hostname() string {
	s, _ := os.Hostname()
	return s
}

var (
	flagTraceHTTP int         = 1
	folderPatern  string      // File name pattern
	filePERM      os.FileMode = 0777
	fileUserGroup string      = ""

	//flagTraceHTTP     = flag.Int("http-trace", 0, "trace level for http, from 0: no trace (default)) to 2:detailed trace")
	printerURL   = flag.String("printer", "", "Printer URL like http://1.2.3.4:8080, when omitted, the device is searched on the network")
	computerName = flag.String("name", hostname(), "Name of the computer visible on the printer (default: $hostname)")
	modeTrace    = flag.Bool("trace", false, "Enable traces")
)

func init() {
	flag.StringVar(&folderPatern, "destination", "", "Folder where images are strored (see help for tokens)")
	flag.StringVar(&folderPatern, "d", "", "shorthand for -destination")
	//flag.Int64Var(&optionFilePERM, "permission", 0777, "file permission, default 0777")

	flag.StringVar(&fileUserGroup, "owner", "", "userid:groupid (ex: 1000:1000)")
	//*modeTrace = true
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
	if !*modeTrace {
		logInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	} else {
		logInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	}
	TRACE.Println("Trace enabled")
	if *computerName == "" {
		*computerName, _ = os.Hostname()
	}
	CheckIDs(fileUserGroup)

	if folderPatern == "" {
		WARNING.Println("No destination given, assuming: -destination = ./%Y%m%d-%H%M%S")
		folderPatern = "./%Y%m%d-%H%M%S"
		//ERROR.Println("Parameter -destination can't be empty")
		//usage()
	} else {
		// Test the pattern to detect issues immediatly
		s, err := ExpandString(folderPatern, time.Now())
		if err != nil {
			ERROR.Println(err)
			usage()
		}
		TRACE.Println("Save to ", s)
	}

	INFO.Println(os.Args[0], "version", VERSION, "started")
	SearchPrinter()
	INFO.Println(os.Args[0], "stopped")

}

/*
  Check each XX minutes if an HP printer appears on the network.
  if the printer is detected, the print will be pulled for events.

*/
func SearchPrinter() {
	for {
		printer := printerURL
		if *printer == "" {
			INFO.Println("Searching printer on the network")
			device, err := srvloc.ProbeHPPrinter()
			TRACE.Printf("%+v\n", device)
			if err == nil {
				// We have one
				*printer = fmt.Sprintf("http://%s:8080", device.IPAddress)
			} else {
				ERROR.Println("Device not found", err)
			}
		}

		if *printer != "" {
			INFO.Println("Connecting to", *printer)
			deviceInfo := new(deviceInfo)
			deviceInfo.url = *printer
			clientInfo, err := registerScanToComp(deviceInfo)
			if err == nil {
				pullDeviceEvents(clientInfo)
			}
		}
		INFO.Println("Connection to ", *printer, "lost. Sleeping...")
		*printer = ""
		time.Sleep(time.Minute * 2)
	}
}

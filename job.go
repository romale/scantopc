// job.go
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type Job struct {
	Device       *Device      // Scanner device
	Destination  *Destination // WalkupDestination
	Source       string       // Scan source : Platten,Adf
	DocumentType string       // Document type: Document,Photo
	TempDir      string       // Temporary folder for images
	ImageList    []string     // List of scanned images
	ImageNumber  int          //
	//	HPJob              *HPJob          // Structure of HP Job
	HPScanCap      *HPScanCap      // Scanner capabilities
	HPScanSettings *HPScanSettings // Settings for the scan job
	Now            time.Time       // Job time for generating file name
	//	FilePath, FileName string          // Final path and image name(s)
}

// Creates job structure, call Scanner Cap and determine Settings for ScanJob
func NewJob(Device *Device, Destination *Destination, Source string) *Job {
	defer Un(Trace("NewJob", Source, Destination.DestinationSettings.Name))
	j := new(Job)

	j.Device = Device
	j.Destination = Destination
	j.Source = Source
	j.Now = time.Now()
	//TODO: Make this OS independant
	t, err := ioutil.TempDir("", "scantopc")
	CheckError("ioutil.TempDir", err)
	j.TempDir = t

	j.GetScanCap()
	TRACE.Printf("New Job: %+v\n", j)
	return j
}

//Get Scanner capabilties
func (j *Job) GetScanCap() {
	defer Un(Trace("Job.GetScanCap"))
	resp, err := httpGet(j.Device.URL + "/Scan/ScanCaps.xml")
	defer resp.Body.Close()
	j.HPScanCap = new(HPScanCap)
	s, err := ioutilReadAll(resp.Body)
	err = xml.Unmarshal(s, j.HPScanCap)
	CheckError("Unmarshal HPScanCap", err)
	//TRACE.Println(string(s))
}

func (j *Job) CheckScanJobSettings() {
	defer Un(Trace("Job.CheckScanJobSettings"))

	var MaxResolution int
	var HPScanSourceCap *HPScanSourceCap
	var HPResolution *HPResolution

	MaxResolution = j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].Resolution

	// Select Scanner cap for the srource
	switch j.Source {
	case "Adf":
		HPScanSourceCap = &j.HPScanCap.Adf.InputSourceCaps
	case "Platen":
		HPScanSourceCap = &j.HPScanCap.Platen.InputSourceCaps
	}

	// Choose the highest available resolution for the source and for the desired job.
	// TODO: Resolutions must be sorted
	for _, xHPResolution := range HPScanSourceCap.SupportedResolutions {
		HPResolution = &xHPResolution
		if HPResolution.XResolution > MaxResolution {
			break
		}
	}

	TRACE.Printf("Source %s, Format %s,\nSettings %+v\n", j.Source, j.DocumentType, j.Destination.DestinationSettings.SourceDocument[j.Source])

	// Create Setting structure for the ScanJob.
	j.HPScanSettings = &HPScanSettings{
		XResolution:        HPResolution.XResolution,
		YResolution:        HPResolution.YResolution,
		XStart:             0,
		YStart:             0,
		Width:              2481,
		Height:             3507,
		Format:             "Jpeg",
		CompressionQFactor: j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].Compression,
		ColorSpace:         j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].ColorSpace,
		BitDepth:           j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].BitDepth,
		InputSource:        j.Source,
		GrayRendering:      "NTSC",
		Gamma:              j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].Gamma,
		Brightness:         j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].Brightness,
		Contrast:           j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].Contrast,
		Highlite:           j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].Highlite,
		Shadow:             j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].Shadow,
		Threshold:          j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].Threshold,
		SharpeningLevel:    j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].SharpeningLevel,
		NoiseRemoval:       j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].NoiseRemoval,
		ContentType:        j.Destination.DestinationSettings.SourceDocument[j.Source][j.DocumentType].ContentType,
	}
	TRACE.Printf("%+v\n", j.HPScanSettings)
}

// Post scan job and handle it
func (j *Job) Scan(DocumentType string) {
	defer Un(Trace("Job.Scan", DocumentType))

	j.DocumentType = DocumentType[4:]
	j.CheckScanJobSettings()

	ByteArray, err := xml.MarshalIndent(j.HPScanSettings, "", "  ")
	r := bytes.NewBufferString(XMLHeader + string(ByteArray))
	resp, err := httpPost(j.Device.URL+"/Scan/Jobs", "text/xml", r)
	defer resp.Body.Close()
	if err != nil || resp.StatusCode != 201 {
		ERROR.Println("Error when posting scan job", err, resp.Status)
		return
	}
	jobURL := resp.Header.Get("Location")
	TRACE.Println("Job created, url:", jobURL)

	ticker := time.NewTicker(3 * time.Second)

	Status := "Processing"

	// Call JobUrl until JobState is Completed or Canceled
JobLoop:
	for _ = range ticker.C {
		j.Device.DocumentProcessor.SendEvent(EventProcessingJob, j)
		resp, err := httpGet(jobURL)
		defer resp.Body.Close()
		if err == nil && resp.StatusCode == 200 {
			job := new(HPJob)
			s, _ := ioutil.ReadAll(resp.Body)
			err = xml.Unmarshal(s, job)
			resp.Body.Close()
			Status = job.JobState
			TRACE.Printf("HPJob.JobState %s,%+v\n", Status, job)
			switch Status {
			case "Processing":

				// During PreScan phase, check if a page is ready to upload
				if job.ScanJob.PreScanPage != nil && job.ScanJob.PreScanPage.PageState == "ReadyToUpload" {
					TRACE.Print("JobState ", job.JobState, job.ScanJob.PreScanPage.PageState)
					j.DownloadImage(job)
				}
				if job.ScanJob.PostScanPage != nil && job.ScanJob.PostScanPage.PageState == "CanceledByDevice" {
					ERROR.Println("Scan job cancled by device")
					TRACE.Print("JobState ", job.JobState)
					TRACE.Println("Batch end,", job.ScanJob.PostScanPage.PageNumber, "page(s) scanned")
				}
			case "Canceled":
				TRACE.Print("JobState ", job.JobState)
				ERROR.Println("Job canceled")
				break JobLoop
			case "Completed":
				TRACE.Print("JobState ", job.JobState)
				INFO.Println("Job completed")
				break JobLoop
			}

		} else {
			ERROR.Println("Error ", resp.Status, err)
			Status = "Error"
		}

	}
	TRACE.Println("Quitting doJobHandling:", Status)
	return
}

// Download image form the scanner, fix JPEG structure and save it into a file
func (j *Job) DownloadImage(HPJob *HPJob) {
	defer Un(Trace("Job.DownloadImage"))

	imageFile := j.TempDir + "/" + fmt.Sprintf("page-%04d.jpg", j.ImageNumber)
	out, err := os.Create(imageFile)
	defer out.Close()
	CheckError("Create file", err)

	// Take image download link from job
	uri := j.Device.URL + HPJob.ScanJob.PreScanPage.BinaryURL
	resp, err := httpGet(uri)
	defer resp.Body.Close()
	if err == nil && resp.StatusCode == 200 {
		// Read HTML response, then fix the JPEG structure and save it into a file
		size, err := CopyAndFixJPEG(out, resp.Body, HPJob.ScanJob.PreScanPage.BufferInfo.ImageHeight)
		if err == nil {
			INFO.Printf("Image downloaded in %s,  size %d\n", imageFile, size)
			j.ImageList = append(j.ImageList, imageFile)
			j.ImageNumber++
		}
	} else {
		ERROR.Println("Get", uri, "failed!", resp.Status, err)
	}

}

func (j *Job) Path() string {
	defer Un(Trace("Job.Path"))
	p, err := ExpandString(paramFolderPatern, j.Now)
	CheckError("ExpandString", err)
	return p
}

func (j *Job) Finalize() {
	defer Un(Trace("Job.Finalize"))
	os.RemoveAll(j.TempDir)
}

func (j *Job) PageNumber() int {
	return len(j.ImageList)
}

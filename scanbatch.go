// scanbatch.go

package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"time"
)

type ScanBatch struct {
	d            *Device
	CurrentState string // HostSelected,
	t            time.Time
	Destination  *Destination // WalkupDestination
	Source       string       // Scan source : Platten,Adf
	Document     *Document
	TimeoutTimer *time.Timer
	Previous     *ScanBatch
}

func NewScanBatch(d *Device, Destination *Destination, Previous *ScanBatch) (sb *ScanBatch, err error) {
	defer Un(Trace("ScanBatch.NewScanBatch"))
	sb = new(ScanBatch)
	sb.d = d
	sb.Destination = Destination
	sb.t = time.Now()
	if Previous != nil {
		if Previous.Previous != nil {
			Previous.Previous.CleanUp()
			Previous.Previous = nil
		}
		sb.Previous = Previous
	}
	return

}

func (sb *ScanBatch) HostSelected() error {
	defer Un(Trace("ScanBatch.HostSelected"))
	var PrevDoc *Document
	if sb.Previous != nil {
		PrevDoc = sb.Previous.Document
	}
	Document, err := NewDocument(sb.Destination, PrevDoc)
	sb.Document = Document
	return err
}

func (sb *ScanBatch) ScanRequested(shortcut string) error {
	defer Un(Trace("ScanBatch.ScanRequested"))
	documenttype := shortcut[4:]
	sb.Document.SetFileType(documenttype)
	source, err := sb.GetSource()
	if err == nil {
		err = sb.Scan(source, documenttype)
	}
	return err
}

func (sb *ScanBatch) ScanNewPageRequested(shortcut string) error {
	defer Un(Trace("ScanBatch.ScanNewPageRequested"))
	documenttype := shortcut[4:]
	sb.Document.SetFileType(documenttype)
	source, err := sb.GetSource()
	if err == nil {
		err = sb.Scan(source, documenttype)
	}
	return err
}

func (sb *ScanBatch) ScanPagesComplete() error {
	defer Un(Trace("ScanBatch.ScanPagesComplete"))
	err := sb.Document.Save()
	return err
}

func (sb *ScanBatch) CleanUp() {
	defer Un(Trace("ScanBatch.CleanUp"))
	if sb.Document != nil {
		sb.Document.CleanUp()
		sb.Document = nil
	}
}

func (sb *ScanBatch) ForceClose() error {
	return DeviceError("ScanBatch", "ForceClose not implemented", nil)
}

func (sb *ScanBatch) GetStatus() (*HPScanStatus, error) {
	defer Un(Trace("ScanBatch.GetStatus"))
	resp, err := HttpGet(sb.d.URL + "/Scan/Status")
	if err != nil {
		return nil, DeviceError("ScanBatch", "GetStatus", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, DeviceError("ScanBatch", "GetStatus: Unexpected status"+resp.Status, err)
	}
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, DeviceError("ScanBatch", "GetStatus", err)
	}
	HPScanStatus := new(HPScanStatus)
	err = xml.Unmarshal(s, HPScanStatus)
	if err != nil {
		return nil, DeviceError("ScanBatch", "GetStatus", err)
	}
	return HPScanStatus, err
}

func (sb *ScanBatch) GetSource() (source string, err error) {
	defer Un(Trace("ScanBatch.GetSource"))

	HPScanStatus, err := sb.GetStatus()
	if err == nil {
		if HPScanStatus.AdfState == "Empty" {
			source = "Platen"
		} else {
			source = "Adf"
		}
		return source, err
	}
	return "", err
}

func (sb *ScanBatch) Scan(source string, documenttype string) (err error) {
	defer Un(Trace("ScanBatch.Scan", source, documenttype))

	HPScanSettings, err := sb.CheckScanJobSettings(source, documenttype)
	if err != nil {
		return err
	}

	ByteArray, err := xml.MarshalIndent(HPScanSettings, "", "  ")
	r := bytes.NewBufferString(XMLHeader + string(ByteArray))
	resp, err := HttpPost(sb.d.URL+"/Scan/Jobs", "text/xml", r)
	if err != nil {
		return DeviceError("ScanBatch", "Scan", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return DeviceError("ScanBatch", "Scan Post job unexpected status code"+resp.Status, nil)
	}

	resp.Body.Close()
	jobURL := resp.Header.Get("Location")
	TRACE.Println("Job created, url:", jobURL)

	return sb.ScanJobLoop(jobURL)

}

func (sb *ScanBatch) GetScanCap() (ScanCap *HPScanCap, err error) {
	defer Un(Trace("ScanBatch.GetScanCap"))

	resp, err := HttpGet(sb.d.URL + "/Scan/ScanCaps.xml")
	if err != nil {
		return nil, DeviceError("ScanBatch", "GetScanCap", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, DeviceError("ScanBatch", "GetScanCap Unknown status"+resp.Status, nil)
	}

	ScanCap = new(HPScanCap)
	s, err := ioutilReadAll(resp.Body)
	if err != nil {
		return nil, DeviceError("ScanBatch", "GetScanCap", err)
	}
	err = xml.Unmarshal(s, ScanCap)
	if err != nil {
		return nil, DeviceError("ScanBatch", "GetScanCap", err)
	}
	return
}

func (sb *ScanBatch) CheckScanJobSettings(source string, documenttype string) (ScanSettings *HPScanSettings, err error) {
	defer Un(Trace("ScanBatch.CheckScanJobSettings"))

	var MaxResolution int
	var HPScanSourceCap *HPScanSourceCap
	var HPResolution *HPResolution

	// Take max resolution from default settings
	MaxResolution = sb.Destination.DestinationSettings.ScanSettings.Resolution
	if MaxResolution == 0 {
		return nil, DeviceError("ScanBatch", "No scan settings for ["+source+","+documenttype+"]", nil)
	}

	HPScanCap, err := sb.GetScanCap()
	if err != nil {
		return nil, err
	}

	// Select Scanner cap for the srource
	switch source {
	case "Adf":
		HPScanSourceCap = &HPScanCap.Adf.InputSourceCaps
	case "Platen":
		HPScanSourceCap = &HPScanCap.Platen.InputSourceCaps
	}

	// Choose the highest available resolution for the source and for the desired job.
	// TODO: Resolutions must be sorted
	for _, xHPResolution := range HPScanSourceCap.SupportedResolutions {
		HPResolution = &xHPResolution
		if HPResolution.XResolution > MaxResolution {
			break
		}
	}

	//TRACE.Printf("Source %s, Format %s,\nSettings %+v\n", source, documenttype, sb.Destination.DestinationSettings.ScanSettings)

	// Create Setting structure for the ScanJob.
	ScanSettings = &HPScanSettings{
		XResolution:        HPResolution.XResolution,
		YResolution:        HPResolution.YResolution,
		XStart:             0,
		YStart:             0,
		Width:              2481,
		Height:             3507,
		Format:             "Jpeg",
		CompressionQFactor: sb.Destination.DestinationSettings.ScanSettings.Compression,
		ColorSpace:         sb.Destination.DestinationSettings.ScanSettings.ColorSpace,
		BitDepth:           sb.Destination.DestinationSettings.ScanSettings.BitDepth,
		InputSource:        source,
		GrayRendering:      "NTSC",
		Gamma:              sb.Destination.DestinationSettings.ScanSettings.Gamma,
		Brightness:         sb.Destination.DestinationSettings.ScanSettings.Brightness,
		Contrast:           sb.Destination.DestinationSettings.ScanSettings.Contrast,
		Highlite:           sb.Destination.DestinationSettings.ScanSettings.Highlite,
		Shadow:             sb.Destination.DestinationSettings.ScanSettings.Shadow,
		Threshold:          sb.Destination.DestinationSettings.ScanSettings.Threshold,
		SharpeningLevel:    sb.Destination.DestinationSettings.ScanSettings.SharpeningLevel,
		NoiseRemoval:       sb.Destination.DestinationSettings.ScanSettings.NoiseRemoval,
		ContentType:        sb.Destination.DestinationSettings.ScanSettings.ContentType,
	}
	return ScanSettings, err
}

func (sb *ScanBatch) ScanJobLoop(jobURL string) (err error) {
	defer Un(Trace("ScanBatch.ScanJobLoop"))

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for _ = range ticker.C {
		resp, err := HttpGet(jobURL)
		if err != nil {
			return DeviceError("ScanBatch", "ScanJobLoop", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return DeviceError("ScanBatch", "ScanJobLoop Unexpected status "+resp.Status, nil)
		}
		job := new(HPJob)
		s, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return DeviceError("ScanBatch", "ScanJobLoop", err)
		}

		err = xml.Unmarshal(s, job)
		if err != nil {
			return DeviceError("ScanBatch", "ScanJobLoop", err)
		}
		resp.Body.Close()
		Status := job.JobState
		TRACE.Printf("HPJob.JobState %s,%+v\n", Status, job)
		switch Status {
		case "Processing":
			// During PreScan phase, check if a page is ready to upload
			if job.ScanJob.PreScanPage != nil && job.ScanJob.PreScanPage.PageState == "ReadyToUpload" {
				sb.DownloadImage(job.ScanJob.PreScanPage.BinaryURL, job.ScanJob.PreScanPage.BufferInfo.ImageHeight)
			}
			// During PostScan phase, check if job is not canceled
			if job.ScanJob.PostScanPage != nil && job.ScanJob.PostScanPage.PageState == "CanceledByDevice" {
				return DeviceError("ScanBatch.ScanJobLoop", "CanceledByDevice", nil)
			}
		case "Canceled":
			return DeviceError("ScanBatch.ScanJobLoop", "Canceled", nil)
		case "Completed":
			return nil
		}
	}

	return err
}

func (sb *ScanBatch) DownloadImage(imageURL string, ImageHeight int) (err error) {
	defer Un(Trace("ScanBatch.DownloadImage"))

	// Take image download link from job
	uri := sb.d.URL + imageURL

	resp, err := HttpGetWithTimeout(uri, nil, 2*time.Second, 5*time.Minute)
	if err != nil {
		return DeviceError("ScanBatch.DownloadImage", "Get "+uri, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return DeviceError("ScanBatch.DownloadImage", "Unexpected status "+resp.Status, nil)
	}
	err = sb.Document.WriteImage(resp.Body, ImageHeight)
	sb.ResetTimeOut()
	return nil
}

func (sb *ScanBatch) ResetTimeOut() {
	defer Un(Trace("ScanBatch.ResetTimeOut"))
}

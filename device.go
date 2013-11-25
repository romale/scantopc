// device.go
package main

/*
Manage events from a device
*/

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"
)

type Device struct {
	Name                    string
	URL                     string
	Destinations            map[string]*Destination
	LastEtag                string
	LastScanEventAgingStamp string
	Connected               bool
	HPScanStatus            HPScanStatus
	Job                     *Job
}

type Destination struct {
	DestinationSettings *DestinationSettings
	UUID                string
	URI                 string
}

func NewDevice(name, url string, settings MapOfDestinationSettings) *Device {
	d := new(Device)
	d.Name = name
	d.URL = url
	d.Destinations = make(map[string]*Destination)
	d.Register(settings)
	return d
}

func (d *Device) Register(settings MapOfDestinationSettings) {
	defer Un(Trace("Device.Register"))

	// TODO: Is this call necessary?
	resp, err := httpGet(d.URL + "/DevMgmt/DiscoveryTree.xml")

	if err != nil || resp.StatusCode != 200 {
		d.Connected = false
		return
	}

	tree := new(HPDiscoveryTree)
	ByteArray, err := ioutilReadAll(resp.Body)
	err = xml.Unmarshal(ByteArray, tree)
	resp.Body.Close()

	// TODO: Is this call necessary?
	resp, err = httpGet(d.URL + "/WalkupScanToComp/WalkupScanToCompDestinations")
	if err != nil || resp.StatusCode != 200 {
		d.Connected = false
		return
	}

	destinations := new(HPWalkupScanToCompDestinations)
	s, err := ioutil.ReadAll(resp.Body)
	CheckError("ReadBody", err)
	err = xml.Unmarshal(s, destinations)

	for Name, DestinationSettings := range settings {
		// Create Post stucture of register a destination
		Me := &HPPostDestination{
			Name:     paramComputerName + "(" + Name + ")",
			Hostname: paramComputerName + "(" + Name + ")",
			LinkType: "Network",
		}

		ByteArray, err = xml.Marshal(Me)
		CheckError("Marshal Me", err)
		r := bytes.NewBufferString(XMLHeader + string(ByteArray))

		resp, err = httpPost(d.URL+"/WalkupScanToComp/WalkupScanToCompDestinations", "text/xml", r)
		defer resp.Body.Close()

		CheckError("Post WalkupScanToCompDestinations", err)

		if err != nil || resp.StatusCode != 201 {
			d.Connected = false
			return
		}

		// SuccessFull registration
		uri := resp.Header.Get("Location")
		uuid := getUUIDfromURI(uri)

		// Create internal Destination structure
		d.Destinations[uuid] = &Destination{
			DestinationSettings: DestinationSettings,
			UUID:                uuid,
			URI:                 uri,
		}
		INFO.Println("Registration successful:", Me.Hostname, uuid)
	}

}

func (d *Device) DeviceEventLoop() {
	defer Un(Trace("Device.DeviceEventLoop"))
	d.Connected = true
	d.LastEtag = "" // Prevent double handling of event
	ticker := time.NewTicker(1 * time.Second)

	for _ = range ticker.C {
		// Loop while the connection is established
		resp, err := http.Get(d.URL + "/EventMgmt/EventTable")
		if err != nil {
			// We have lost contact with the device
			ERROR.Println(err)
			d.Connected = false
			break
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case 304:
			//NotChanged since If-Modified-Since or If-Match.
			//TODO: manage If-Modified-Since or If-Match.
		case 200:
			// OK
			Etag := resp.Header.Get("Etag")
			if Etag != d.LastEtag {
				// Etag has changed
				d.LastEtag = Etag
				EventTable := new(HPEventTable)
				ByteArray, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				err = xml.Unmarshal(ByteArray, &EventTable)

				// Parse evenCompletedt list foHPScanStatusr ScanEvent
				for _, Event := range EventTable.Events {
					d.HandleEvent(&Event)
				}
			}

		default:
			// Log status, do nothing
		}
		resp.Body.Close() // Close response
		// Check if the device is still connected
		if !d.Connected {
			break
		}
	}
	ticker.Stop()
}

// Handle a single event
func (d *Device) HandleEvent(HPEvent *HPEvent) {
	defer Un(Trace("Device.HandleEvent"))
	switch HPEvent.UnqualifiedEventCategory {
	case "ScanEvent":
		uri := ""
		if HPEvent.AgingStamp != d.LastScanEventAgingStamp {
			d.LastScanEventAgingStamp = HPEvent.AgingStamp
			for _, payload := range HPEvent.Payloads {
				if payload.ResourceType == "wus:WalkupScanToCompDestination" {
					uri = payload.ResourceURI
				}
			}

			// Check if the ScanToComp event is for one of our destinations
			if dest, ok := d.Destinations[getUUIDfromURI(uri)]; ok {
				TRACE.Println("Event ScanEvent accepted")
				d.ScanEvent(dest, uri)
			} else {
				TRACE.Println("Event ScanEvent ignored, uri doesn't match")
			}
		} else {
			TRACE.Println("Event ScanEvent ignored, AgingStamp already processed")
		}
	case "PoweringDownEvent":
		INFO.Println("Event", HPEvent.UnqualifiedEventCategory, ", closing connection.")
		d.Connected = false
	default:
		TRACE.Println("Event", HPEvent.UnqualifiedEventCategory, "is ignored")
	}
}

func (d *Device) GetSource() (source string) {
	defer Un(Trace("Device.GetSource"))
	if d.GetStatus() {
		if d.HPScanStatus.AdfState == "Empty" {
			source = "Platen"
		} else {
			source = "Adf"
		}
		return source
	} else {
		return ""
	}
}

func (d *Device) GetStatus() bool {
	defer Un(Trace("Device.GetStatus"))
	// Get Status
	resp, err := httpGet(d.URL + "/Scan/Status")
	defer resp.Body.Close()
	if err == nil && resp.StatusCode == 200 {
		s, _ := ioutil.ReadAll(resp.Body)
		err = xml.Unmarshal(s, &d.HPScanStatus)
		return true
	} else {
		d.Connected = false
		ERROR.Println("get /Scan/Status has failed", err)
		return false
	}
}

// Handle a scan event
func (d *Device) ScanEvent(Destination *Destination, uri string) {
	defer Un(Trace("Device.ScanEvent"))
	resp, err := httpGet(d.URL + "/WalkupScanToComp/WalkupScanToCompEvent")
	defer resp.Body.Close()

	if err != nil {
		d.Connected = false
		return
	}

	if resp.StatusCode == 200 {
		event := new(HPWalkupScanToCompEvent)
		s, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		err = xml.Unmarshal(s, event)
		TRACE.Println("WalkupScanToCompEvent", event.WalkupScanToCompEventType)

		switch event.WalkupScanToCompEventType {
		case "HostSelected":
			// That's for us...
			// Nothing to do... We'R ready

		case "ScanRequested":
			// 	ScanRequested : Start Adf scanning or 1st page on Platten scanning
			dest := d.GetWalkupScanToCompDestinations(uri)
			source := d.GetSource()
			d.Job = NewJob(d, Destination, source, dest.WalkupScanToCompSettings.Shortcut)
			d.Job.Scan()

		case "ScanNewPageRequested":
			//	ScanNewPageRequested: Subsequent pages on Platten
			_ = d.GetWalkupScanToCompDestinations(uri)
			d.Job.Scan()

		case "ScanPagesComplete":
			if d.Job != nil {
				d.Job.SaveImage()
				d.Job.Finalize()
				d.Job = nil
			}
		}

	}
}

func (d *Device) GetWalkupScanToCompDestinations(uri string) *HPWalkupScanToCompDestination {
	defer Un(Trace("Device.GetWalkupScanToCompDestinations"))
	resp, err := httpGet(d.URL + uri)
	defer resp.Body.Close()
	if err == nil && resp.StatusCode == 200 {
		s, _ := ioutilReadAll(resp.Body)
		dest := new(HPWalkupScanToCompDestination)
		_ = xml.Unmarshal(s, dest)
		//TRACE.Println(string(s))
		return dest
	} else {
		return nil
	}
}

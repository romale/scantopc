// device.go
package main

/*
	Manage events from a device and detect scan event
*/

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"time"
)

type Device struct {
	Name         string
	URL          string
	Destinations map[string]*Destination
	ErrorChan    chan error
	EventChan    chan HPEventTable
	ScanBatch    *ScanBatch
	AgingStamp   string
}

type deviceError struct {
	Err       error
	Operation string
	Message   string
}

func (e deviceError) Error() string {
	if e.Err != nil {
		return e.Operation + ":" + e.Message + "," + e.Err.Error()
	} else {
		return e.Operation + ":" + e.Message + ",<nil>"
	}
}

func DeviceError(Op, Mes string, Err ...error) error {
	err := deviceError{nil, Op, Mes}
	if len(Err) > 0 {
		err.Err = Err[0]
	}
	ERROR.Println(err)
	return err
}

func (d *Device) SetError(err error) error {
	defer Un(Trace("Device.SetError"))
	d.ErrorChan <- err
	return err
}

func NewDeviceManager(name, url string) (*Device, error) {
	d := new(Device)
	d.Name = name
	d.URL = url
	d.Destinations = make(map[string]*Destination)
	d.ErrorChan = make(chan error, 10)
	d.EventChan = make(chan HPEventTable, 10)
	// Start the manager
	return d, d.Start()
}

func (d *Device) Start() (err error) {
	defer Un(Trace("Device.Start"))
	if err = d.Register(DefaultDestinationSettings); err == nil {
		go d.StartEventChannel()
		err = d.MainLoop()
	}
	return err
}

func (d *Device) Stop() {
	TRACE.Panicln("STOP!")
	//d.stop <- true
}

func (d *Device) Close() {
	//TODO: Close everything open
	d.Stop()
}

func (d *Device) Error(err error) {
	ERROR.Println(err.Error())
	d.Stop()
}

type Destination struct {
	DestinationSettings *DestinationSettings
	UUID                string
	URI                 string
}

func (d *Device) Register(settings MapOfDestinationSettings) (err error) {
	defer Un(Trace("Device.Register"))

	// TODO: Is this call necessary?
	resp, err := HttpGet(d.URL + "/DevMgmt/DiscoveryTree.xml")
	if err != nil {
		return DeviceError("Register", "DiscoveryTree", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return DeviceError("Register", "Unexpected Status "+resp.Status, err)
	}

	tree := new(HPDiscoveryTree)
	ByteArray, err := ioutilReadAll(resp.Body)
	if err = xml.Unmarshal(ByteArray, tree); err != nil {
		return err
	}
	resp.Body.Close()

	resp, err = HttpGet(d.URL + "/WalkupScanToComp/WalkupScanToCompDestinations")
	if err != nil {
		return DeviceError("Register", "WalkupScanToCompDestinations", err)
	}
	if resp.StatusCode != 200 {
		return DeviceError("Register", "Unexpected Status "+resp.Status, err)
	}

	destinations := new(HPWalkupScanToCompDestinations)
	s, err := ioutil.ReadAll(resp.Body)
	CheckError("ReadBody", err)
	if err = xml.Unmarshal(s, destinations); err != nil {
		return err
	}
	resp.Body.Close()

	for Name, DestinationSettings := range settings {
		// Create Post stucture of register a destination
		Me := &HPPostDestination{
			Name:     paramComputerName + "(" + Name + ")",
			Hostname: paramComputerName + "(" + Name + ")",
			LinkType: "Network",
		}

		if ByteArray, err = xml.Marshal(Me); err != nil {
			return err
		}

		r := bytes.NewBufferString(XMLHeader + string(ByteArray))

		if resp, err = HttpPost(d.URL+"/WalkupScanToComp/WalkupScanToCompDestinations", "text/xml", r); err != nil {
			return DeviceError("Register / Post WalkupScanToCompDestinations", "HTTP", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 201 {
			return DeviceError("Register /  Post WalkupScanToCompDestinations", "Unexpected Status "+resp.Status, err)
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
	return err
}

func (d *Device) MainLoop() error {
	defer Un(Trace("Device.MainLoop"))
	var err error

	for err == nil {
		TRACE.Println("Waiting an event")
		select {
		case err = <-d.ErrorChan:
			ERROR.Println(err)
			return err
		case event := <-d.EventChan:
			err = d.ParseEventTable(&event)
		}
	}
	return err
}

func (d *Device) StartEventChannel() {
	defer Un(Trace("Device.StartEventChannel"))

	var err error = nil
	headers := make(map[string]string, 1)
	Etag := ""

	// on call to get firts events and e-tag
	resp, err := HttpGet(d.URL + "/EventMgmt/EventTable")
	if err != nil {
		d.SetError(DeviceError("Device.StartEventChannel", "", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		d.SetError(DeviceError("Device.StartEventChannel", "Unexpected Status "+resp.Status, err))
		return
	}

	Etag = resp.Header.Get("Etag")

	EventTable := new(HPEventTable)
	ByteArray, _ := ioutil.ReadAll(resp.Body)
	err = xml.Unmarshal(ByteArray, &EventTable)
	resp.Body.Close()
	d.EventChan <- *EventTable

	for err == nil {
		headers["If-None-Match"] = Etag
		resp, err := HttpGetWithTimeout(d.URL+"/EventMgmt/EventTable?timeout=1200", &headers, 2*time.Second, 150*time.Second)
		if err != nil {
			err = DeviceError("Device.StartEventChannel.goroutine", "GET", err)
		} else {
			switch resp.StatusCode {
			case 304:
				//NotChanged since last call
			case 200:
				// OK, there is a new event
				Etag = resp.Header.Get("Etag")
				ByteArray, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					err = DeviceError("Device.StartEventChannel.goroutine", "ReadAll", err)
				} else {
					err = xml.Unmarshal(ByteArray, &EventTable)
					if err != nil {
						err = DeviceError("Device.StartEventChannel.goroutine", "Unmarshal", err)
					}
					resp.Body.Close()
					TRACE.Println("Send Event to channel...")
					d.EventChan <- *EventTable
				}
			default: // Other return codes denote a problem
				err = DeviceError("Device.StartEventChannel.goroutine", "Unexpected status"+resp.Status, err)
			}
		}
	}
	d.SetError(err)
}

func (d *Device) ParseEventTable(HPEventTable *HPEventTable) (err error) {
	defer Un(Trace("Device.ParseEventTable"))
	// Parse evenCompletet list ScanEvent
	for _, HPEvent := range HPEventTable.Events {
		switch HPEvent.UnqualifiedEventCategory {
		case "ScanEvent":
			err = d.ScanEvent(HPEvent)
		case "PoweringDownEvent":
			//TODO: Close pending ScanBatch
			d.SetError(DeviceError("ParseEventTable", "Powerdown event recieved", nil))
		default:
			// Ignore silently all other events
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) ScanEvent(HPEvent HPEvent) (err error) {
	TRACE.Println("Device.ScanEvent AgingStamp", d.AgingStamp, HPEvent.AgingStamp)
	if HPEvent.AgingStamp > d.AgingStamp {
		uri := ""
		for _, payload := range HPEvent.Payloads {
			switch payload.ResourceType {
			case "wus:WalkupScanToCompDestination":
				uri = payload.ResourceURI
			}
		}

		// Check if the WalkupScanToComp event is for one of our destinations
		if dest, ok := d.Destinations[getUUIDfromURI(uri)]; ok {
			HPWalkupScanToCompDestination, err := d.GetWalkupScanToCompDestinations(uri)
			if err == nil {
				d.AgingStamp = HPEvent.AgingStamp
				TRACE.Println("Event ScanEvent accepted")
				d.WalkupScanToCompEvent(dest, HPWalkupScanToCompDestination)
			} else {
				return err
			}
		} else {
			TRACE.Println("Device.ScanEvent uknown destination")
		}
		return err
	}
	return nil
}

func (d *Device) WalkupScanToCompEvent(Destination *Destination, HPWalkupScanToCompDestination *HPWalkupScanToCompDestination) error {
	defer Un(Trace("Device.WalkupScanToCompEvent"))
	// Handle a scan event
	resp, err := HttpGet(d.URL + "/WalkupScanToComp/WalkupScanToCompEvent")
	if err != nil {
		return DeviceError("Device.WalkupScanToCompEvent", "", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return DeviceError("Device.WalkupScanToCompEvent", "Unexpected Status "+resp.Status, err)
	}
	defer resp.Body.Close()

	event := new(HPWalkupScanToCompEvent)
	s, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err = xml.Unmarshal(s, event); err != nil {
		return err
	}

	switch event.WalkupScanToCompEventType {
	case "HostSelected": // That's for us...
		INFO.Println("Scan job started")
		if d.ScanBatch, err = NewScanBatch(d, Destination, d.ScanBatch); err != nil {
			return err
		}
		err = d.ScanBatch.HostSelected()

	case "ScanRequested": // Start Adf scanning or 1st page on Platen scanning
		if d.ScanBatch == nil {
			return DeviceError("Device.WalkupScanToCompEvent", "recieved ScanRequested, but ScanBatch is nil", nil)
		}
		err = d.ScanBatch.ScanRequested(HPWalkupScanToCompDestination.WalkupScanToCompSettings.Shortcut)

	case "ScanNewPageRequested": //Subsequent pages on Platten
		if d.ScanBatch == nil {
			return DeviceError("Device.WalkupScanToCompEvent", "recieved ScanNewPageRequested, but ScanBatch is nil", nil)
		}
		err = d.ScanBatch.ScanNewPageRequested(HPWalkupScanToCompDestination.WalkupScanToCompSettings.Shortcut)

	case "ScanPagesComplete": //End of ScanBatch
		if d.ScanBatch == nil {
			return DeviceError("Device.WalkupScanToCompEvent", "recieved ScanPagesComplete, but ScanBatch is nil", nil)
		}
		err = d.ScanBatch.ScanPagesComplete()
		INFO.Println("Scan job ended")
	default:
		err = DeviceError("ScanEvent", "Unknown event"+event.WalkupScanToCompEventType, nil)
	}
	return err

}

func (d *Device) GetWalkupScanToCompDestinations(uri string) (*HPWalkupScanToCompDestination, error) {
	defer Un(Trace("Device.GetWalkupScanToCompDestinations"))

	resp, err := HttpGet(d.URL + "/WalkupScanToComp/WalkupScanToCompDestinations")

	if err != nil {
		return nil, DeviceError("Device.WalkupScanToCompDestinations", "", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, DeviceError("Device.WalkupScanToCompDestinations", "Unexpected Status "+resp.Status, err)
	}

	s, err := ioutilReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	dest := new(HPWalkupScanToCompDestination)
	err = xml.Unmarshal(s, dest)

	// Call the given URI
	resp, err = HttpGet(d.URL + uri)

	if err != nil {
		return nil, DeviceError("Device.WalkupScanToCompDestinations", "", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, DeviceError("Device.WalkupScanToCompDestinations", "Unexpected Status "+resp.Status, err)
	}

	s, err = ioutilReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	dest = new(HPWalkupScanToCompDestination)
	err = xml.Unmarshal(s, dest)
	return dest, err
}

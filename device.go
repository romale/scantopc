// device.go
package main

/*
	Manage events from a device and detect scan event
*/

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	EventTimeout    = 120 * time.Second // Give timout for HP GetEvent
	RegisterTimeOut = 60 * time.Minute  // Give register interval
)

type Device struct {
	Name          string
	URL           string
	Destinations  map[string]*Destination
	ScanBatch     *ScanBatch
	AgingStamp    string
	events        chan HPEventTable
	errors        chan error
	registerTimer *time.Timer
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
	defer Un(Trace("Device.SetError", err))
	d.errors <- err
	return err
}

func NewDeviceManager(name, url string) (*Device, error) {
	d := new(Device)
	d.Name = name
	d.URL = url
	d.Destinations = make(map[string]*Destination)
	// Start the manager
	return d, d.MainLoop()
}

func (d *Device) Error(err error) {
	ERROR.Println(err.Error())
}

type Destination struct {
	DestinationSettings *DestinationSettings
	UUID                string
	URI                 string
}

func (d *Device) MainLoop() error {
	defer Un(Trace("Device.MainLoop"))
	var err error

	var innerloopclose chan bool
	d.events = make(chan HPEventTable)
	d.errors = make(chan error)
	err = d.Register(DefaultDestinationSettings)
	if err == nil {
		innerloopclose, err = d.StartEventChannel()
	}
	if err == nil {
		d.registerTimer = time.NewTimer(RegisterTimeOut)
		for {
			select {
			case Events := <-d.events:
				TRACE.Println("Device.MainLoop", "Got events!")
				err = d.ParseEventTable(&Events)
				if err == nil {
					d.registerTimer.Reset(RegisterTimeOut)
				}
			case <-d.registerTimer.C:
				// Timer is ser at each loop step.
				TRACE.Println("Device.MainLoop", "Got timeout!")
				innerloopclose <- true // Send a bullet to Event goroutine
				d.registerTimer.Reset(RegisterTimeOut)
				if err == nil {
					innerloopclose, err = d.StartEventChannel()
				}
			case err = <-d.errors:
				// Error!
				TRACE.Println("Device.MainLoop", "Got Error!")
				return err
			}
			if err != nil {
				TRACE.Println("Device.MainLoop", "Error !", err)
				return err
			}
		}
	}
	return err
}

func (d *Device) StartEventChannel() (chan bool, error) {
	defer Un(Trace("Device.StartEventChannel"))

	closeme := make(chan bool)

	go func() {
		defer Un(Trace("Device.StartEventChannel.StartGoRoutine"))

		var err error
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
		resp.Body.Close()
		err = xml.Unmarshal(ByteArray, &EventTable)
		if err != nil {
			d.SetError(DeviceError("Device.StartEventChannel", "Unmarshal", err))
			return
		}

		// Send firts events.
		d.events <- *EventTable
		for {
			select {
			case <-closeme:
				// Sorry, I must dye
				TRACE.Println("Device.StartEventChannel.goroutine", "closeme !")
				return
			default:
				//Dont block the loop
			}

			headers["If-None-Match"] = Etag
			resp, err := HttpGetWithTimeout(d.URL+"/EventMgmt/EventTable?timeout="+fmt.Sprintf("%d", int(EventTimeout.Seconds())*10), &headers, 2*time.Second, 2*EventTimeout)
			if err != nil {
				d.SetError(DeviceError("Device.StartEventChannel.goroutine", "GET", err))
				return
			} else {
				switch resp.StatusCode {
				case 304:
					//NotChanged since last call
					resp.Body.Close()
				case 200:
					// OK, there is a new event
					Etag = resp.Header.Get("Etag")
					TRACE.Println("Got Etag", Etag)
					ByteArray, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					if err != nil {
						d.SetError(DeviceError("Device.StartEventChannel.goroutine", "ReadAll", err))
						return
					} else {
						err = xml.Unmarshal(ByteArray, &EventTable)
						if err != nil {
							d.SetError(DeviceError("Device.StartEventChannel.goroutine", "Unmarshal", err))
						}
						resp.Body.Close()
						TRACE.Println("Send Event to channel...")
						d.events <- *EventTable
					}
				default: // Other return codes denote a problem
					d.SetError(DeviceError("Device.StartEventChannel.goroutine", "Unexpected status"+resp.Status, err))
				}

			}
		}
	}()

	return closeme, nil
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
			err = DeviceError("ParseEventTable", "Powerdown event recieved", nil)
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

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return DeviceError("Device.WalkupScanToCompEvent", "Unexpected Status "+resp.Status, err)
	}

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

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, DeviceError("Device.WalkupScanToCompDestinations", "Unexpected Status "+resp.Status, err)
	}

	s, err := ioutilReadAll(resp.Body)
	resp.Body.Close()
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

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, DeviceError("Device.WalkupScanToCompDestinations", "Unexpected Status "+resp.Status, err)
	}

	s, err = ioutilReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	dest = new(HPWalkupScanToCompDestination)
	err = xml.Unmarshal(s, dest)
	return dest, err
}

func (d *Device) Register(settings MapOfDestinationSettings) (err error) {
	defer Un(Trace("Device.Register"))

	// TODO: Is this call necessary?
	resp, err := HttpGet(d.URL + "/DevMgmt/DiscoveryTree.xml")
	if err != nil {
		return DeviceError("Register", "DiscoveryTree", err)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
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
		resp.Body.Close()
		return DeviceError("Register", "Unexpected Status "+resp.Status, err)
	}

	destinations := new(HPWalkupScanToCompDestinations)
	s, err := ioutil.ReadAll(resp.Body)
	CheckError("ReadBody", err)
	if err = xml.Unmarshal(s, destinations); err != nil {
		resp.Body.Close()
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
		resp.Body.Close()
	}
	return err
}

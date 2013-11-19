// strutures.go
package main

import (
	"encoding/xml"
	"fmt"
)

const XMLHeader = `<?xml version="1.0" encoding="utf-8"?>`

// XML structures used by the printer web interface

type HPDiscoveryTree struct {
	XMLName        xml.Name          `xml:"http://www.hp.com/schemas/imaging/con/ledm/2007/09/21 DiscoveryTree"`
	Revision       string            `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Version>Revision"`
	Date           string            `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Version>Date"`
	SupportedTrees []HPSupportedTree `xml:"http://www.hp.com/schemas/imaging/con/ledm/2007/09/21 SupportedTree"`
	SupportedIfcs  []HPSupportedIfc  `xml:"http://www.hp.com/schemas/imaging/con/ledm/2007/09/21 SupportedIfc"`
}

type HPSupportedTree struct {
	XMLName      xml.Name //`xml:"http://www.hp.com/schemas/imaging/con/ledm/2007/09/21 SupportedTree"`
	ResourceURI  string   `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ ResourceURI"`
	ResourceType string   `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ ResourceType"`
	Revision     string   `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Revision"`
}

type HPSupportedIfc struct {
	XMLName      xml.Name `xml:"http://www.hp.com/schemas/imaging/con/ledm/2007/09/21 SupportedIfc"`
	ManifestURI  string   `xml:"http://www.hp.com/schemas/imaging/con/ledm/2007/09/21 ManifestURI"`
	ResourceType string   `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ ResourceType"`
	Revision     string   `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Revision"`
}

type HPWalkupScanToCompDestinations struct {
	XMLName                      xml.Name                        `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 WalkupScanToCompDestinations"`
	WalkupScanToCompDestinations []HPWalkupScanToCompDestination `xml:"WalkupScanToCompDestination"`
}

type HPWalkupScanToCompDestination struct {
	XMLName                  xml.Name                    `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 WalkupScanToCompDestination"`
	ResourceURI              string                      `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ ResourceURI"`
	Name                     string                      `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Name"`
	Hostname                 string                      `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Hostname"`
	LinkType                 string                      `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ LinkType"`
	WalkupScanToCompSettings *HPWalkupScanToCompSettings `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 WalkupScanToCompSettings"`
}

type HPWalkupScanToCompSettings struct {
	XMLName      xml.Name   `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 WalkupScanToCompSettings"`
	ScanSettings HPScanType `xml:"http://www.hp.com/schemas/imaging/con/ledm/scantype/2008/03/17 ScanSettings"`
	Shortcut     string     `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 Shortcut"`
}

type HPPostDestination struct {
	XMLName                  xml.Name                    `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 WalkupScanToCompDestination"`
	Name                     string                      `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Name"`
	Hostname                 string                      `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/2009/04/06 Hostname"`
	LinkType                 string                      `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 LinkType"`
	WalkupScanToCompSettings *HPWalkupScanToCompSettings `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 WalkupScanToCompSettings"`
}

type HPScanType struct {
	XMLName      xml.Name `xml:"http://www.hp.com/schemas/imaging/con/ledm/scantype/2008/03/17 ScanSettings"`
	ScanPlexMode string   `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ ScanPlexMode"`
}

type HPEventTable struct {
	XMLName  xml.Name  `xml:"http://www.hp.com/schemas/imaging/con/ledm/events/2007/09/16 EventTable"`
	Revision string    `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Version>Revision"`
	Date     string    `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ Version>Date"`
	Events   []HPEvent `xml:"http://www.hp.com/schemas/imaging/con/ledm/events/2007/09/16 Event"`
}

type HPEvent struct {
	XMLName                  xml.Name    `xml:"http://www.hp.com/schemas/imaging/con/ledm/events/2007/09/16 Event"`
	UnqualifiedEventCategory string      `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ UnqualifiedEventCategory"`
	AgingStamp               string      `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ AgingStamp"`
	Payloads                 []HPPayload `xml:"http://www.hp.com/schemas/imaging/con/ledm/events/2007/09/16 Payload"`
}

type HPPayload struct {
	XMLName      xml.Name `xml:"http://www.hp.com/schemas/imaging/con/ledm/events/2007/09/16 Payload"`
	ResourceURI  string   `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ ResourceURI"`
	ResourceType string   `xml:"http://www.hp.com/schemas/imaging/con/dictionaries/1.0/ ResourceType"`
}

type HPScanSettings struct {
	XMLName            xml.Name `xml:"http://www.hp.com/schemas/imaging/con/cnx/scan/2008/08/19 ScanSettings"`
	XResolution        int
	YResolution        int
	XStart             int
	YStart             int
	Width              int
	Height             int
	Format             string
	CompressionQFactor int
	ColorSpace         string
	BitDepth           int
	InputSource        string
	GrayRendering      string
	Gamma              int `xml:"ToneMap>Gamma"`
	Brightness         int `xml:"ToneMap>Brightness"`
	Contrast           int `xml:"ToneMap>Contrast"`
	Highlite           int `xml:"ToneMap>Highlite"`
	Shadow             int `xml:"ToneMap>Shadow"`
	Threshold          int `xml:"ToneMap>Threshold"`
	SharpeningLevel    int
	NoiseRemoval       int
	ContentType        string
}

type HPScanCap struct {
	XMLName          xml.Name       `xml:"http://www.hp.com/schemas/imaging/con/cnx/scan/2008/08/19 ScanCaps"`
	ModelName        string         `xml:"DeviceCaps>ModelName"`
	DerivativeNumber int            `xml:"DeviceCaps>DerivativeNumber"`
	ColorEntries     []HPColorEntry `xml:"ColorEntries>ColorEntry"`
	Platen           HPScanSource   `xml:"Platen"`
	Adf              HPScanSource   `xml:"Adf"`
}

type HPColorEntry struct {
	XMLName         xml.Name `xml:"ColorEntry"`
	ColorType       string
	Formats         []string `xml:"Formats>Format"`
	ImageTransforms []string `xml:"ImageTransforms>ImageTransform"`
	GrayRenderings  []string `xml:"GrayRenderings>GrayRendering"`
}

type HPScanSource struct {
	XMLName         xml.Name //`xml:`
	InputSourceCaps HPScanSourceCap
	FeederCapacity  int
	AdfOptions      []string `xml:"AdfOptions>AdfOption"`
}

type HPScanSourceCap struct {
	MinWidth              int
	MinHeight             int
	MaxWidth              int
	MaxHeight             int
	RiskyLeftMargin       int
	RiskyRightMargin      int
	RiskyTopMargin        int
	RiskyBottomMargin     int
	MinResolution         int
	MaxOpticalXResolution int
	MaxOpticalYResolution int
	SupportedResolutions  []HPResolution `xml:"SupportedResolutions>Resolution"`
}

type HPResolution struct {
	XResolution int
	YResolution int
	NumCcd      int
	ColorTypes  []string `xml:"ColorTypes>ColorType"`
}

type HPWalkupScanToCompEvent struct {
	XMLName                   xml.Name `xml:"http://www.hp.com/schemas/imaging/con/ledm/walkupscan/2010/09/28 WalkupScanToCompEvent"`
	WalkupScanToCompEventType string
}

type HPScanStatus struct {
	XMLName      xml.Name `xml:"http://www.hp.com/schemas/imaging/con/cnx/scan/2008/08/19 ScanStatus"`
	ScannerState string   // AdfError,BusyWithScanJob
	AdfState     string   // Empty,Loaded,Jammed
}

type HPJob struct {
	XMLName        xml.Name `xml:"http://www.hp.com/schemas/imaging/con/ledm/jobs/2009/04/30 Job"`
	JobUrl         string
	JobCategory    string
	JobState       string //Canceled,Completed,Processing
	JobStateUpdate string
	ScanJob        HPScanJob `xml:"http://www.hp.com/schemas/imaging/con/cnx/scan/2008/08/19 ScanJob"`
}

type HPScanJob struct {
	XMLName      xml.Name `xml:"http://www.hp.com/schemas/imaging/con/cnx/scan/2008/08/19 ScanJob"`
	PreScanPage  *HPPreScanPage
	PostScanPage *HPPostScanPage
}

type HPPreScanPage struct {
	XMLName          xml.Name `xml:"PreScanPage"`
	PageNumber       int
	PageState        string //PreparingScan
	BufferInfo       HPBufferInfo
	BinaryURL        string
	ImageOrientation string // Normal
}
type HPPostScanPage struct {
	XMLName    xml.Name `xml:"PostScanPage"`
	PageNumber int
	PageState  string //UploadCompleted,CanceledByDevice
	TotalLines int
}

type HPBufferInfo struct {
	ScanSettings HPScanSettings
	ImageWidth   int
	ImageHeight  int
	BytesPerLine int
	Cooked       string //"Enabled"
}

func Structure_test() {
	const XML = ` <?xml version="1.0" encoding="UTF-8"?>
<!-- THIS DATA SUBJECT TO DISCLAIMER(S) INCLUDED WITH THE PRODUCT OF ORIGIN. -->
<ScanCaps xmlns="http://www.hp.com/schemas/imaging/con/cnx/scan/2008/08/19">
	<DeviceCaps>
		<ModelName>h711g</ModelName>
		<DerivativeNumber>5</DerivativeNumber>
	</DeviceCaps>
	<ColorEntries>
		<ColorEntry>
			<ColorType>K1</ColorType>
			<Formats>
				<Format>Raw</Format>
			</Formats>
			<ImageTransforms>
				<ImageTransform>ToneMap</ImageTransform>
				<ImageTransform>Sharpening</ImageTransform>
				<ImageTransform>NoiseRemoval</ImageTransform>
			</ImageTransforms>
			<GrayRenderings>
				<GrayRendering>GrayCcdEmulated</GrayRendering>
			</GrayRenderings>
		</ColorEntry>
		<ColorEntry>
			<ColorType>Gray8</ColorType>
			<Formats>
				<Format>Raw</Format>
				<Format>Jpeg</Format>
			</Formats>
			<ImageTransforms>
				<ImageTransform>ToneMap</ImageTransform>
				<ImageTransform>Sharpening</ImageTransform>
				<ImageTransform>NoiseRemoval</ImageTransform>
			</ImageTransforms>
			<GrayRenderings>
				<GrayRendering>NTSC</GrayRendering>
				<GrayRendering>GrayCcdEmulated</GrayRendering>
			</GrayRenderings>
		</ColorEntry>
		<ColorEntry>
			<ColorType>Color8</ColorType>
			<Formats>
				<Format>Raw</Format>
				<Format>Jpeg</Format>
			</Formats>
			<ImageTransforms>
				<ImageTransform>ToneMap</ImageTransform>
				<ImageTransform>Sharpening</ImageTransform>
				<ImageTransform>NoiseRemoval</ImageTransform>
			</ImageTransforms>
		</ColorEntry>
	</ColorEntries>
	<Platen>
		<InputSourceCaps>
			<MinWidth>8</MinWidth>
			<MinHeight>8</MinHeight>
			<MaxWidth>2550</MaxWidth>
			<MaxHeight>3508</MaxHeight>
			<RiskyLeftMargin>50</RiskyLeftMargin>
			<RiskyRightMargin>18</RiskyRightMargin>
			<RiskyTopMargin>50</RiskyTopMargin>
			<RiskyBottomMargin>24</RiskyBottomMa1896rgin>
			<MinResolution>75</MinResolution>
			<MaxOpticalXResolution>2400</MaxOpticalXResolution>
			<MaxOpticalYResolution>2400</MaxOpticalYResolution>
			<SupportedResolutions>
				<Resolution>
					<XResolution>75</XResolution>
					<YResolution>75</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>100</XResolution>
					<YResolution>100</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>200</XResolution>
					<YResolution>200</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>300</XResolution>
					<YResolution>300</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>600</XResolution>
					<YResolution>600</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>1200</XResolution>
					<YResolution>1200</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>2400</XResolution>
					<YResolution>2400</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
			</SupportedResolutions>
		</InputSourceCaps>
	</Platen>
	<Adf>
		<InputSourceCaps>
			<MinWidth>8</MinWidth>
			<MinHeight>8</MinHeight>
			<MaxWidth>2550</MaxWidth>
			<MaxHeight>4200</MaxHeight>
			<RiskyLeftMargin>16</RiskyLeftMargin>
			<RiskyRightMargin>0</RiskyRightMargin>
			<RiskyTopMargin>35</RiskyTopMargin>
			<RiskyBottomMargin>35</RiskyBottomMargin>
			<MinResolution>75</MinResolution>
			<MaxOpticalXResolution>600</MaxOpticalXResolution>
			<MaxOpticalYResolution>600</MaxOpticalYResolution>
			<SupportedResolutions>
				<Resolution>
					<XResolution>75</XResolution>
					<YResolution>75</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>100</XResolution>
					<YResolution>100</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>200</XResolution>
					<YResolution>200</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>300</XResolution>
					<YResolution>300</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
				<Resolution>
					<XResolution>600</XResolution>
					<YResolution>600</YResolution>
					<NumCcd>1</NumCcd>
					<ColorTypes>
						<ColorType>K1</ColorType>
						<ColorType>Gray8</ColorType>
						<ColorType>Color8</ColorType>
					</ColorTypes>
				</Resolution>
			</SupportedResolutions>
		</InputSourceCaps>
		<FeederCapacity>35</FeederCapacity>
		<AdfOptions>
			<AdfOption>DetectPaperLoaded</AdfOption>
		</AdfOptions>
	</Adf>
</ScanCaps>`

	v := new(HPScanCap)
	err := xml.Unmarshal([]byte(XML), v)
	CheckError("Unmarshal HPJob", err)
	fmt.Printf("%+v\n", v)
	//option.SaveAs = Settings.Shortcut
	s, _ := xml.MarshalIndent(v, "", "  ")
	fmt.Println(string(s))
}

package main

import (
//"encoding/xml"
)

type DestinationSettings struct {
	//XMLName        xml.Name `xml:DestinationSettings"`
	Name         string
	FilePattern  *string
	DoOCR        bool
	ScanSettings *ScanSettings
}

type ScanSettings struct {
	Resolution      int    // 75,100,200,300,600,1200 dpi
	ColorSpace      string // Color,Gray
	Compression     int    // 15
	BitDepth        int    // 8
	GrayRendering   string // NTSC
	Gamma           int    // 1000,
	Brightness      int    // 1000,
	Contrast        int    // 1000,
	Highlite        int    // 179,
	Shadow          int    // 25,
	Threshold       int    // 0,
	SharpeningLevel int    // 128,
	NoiseRemoval    int    // 0,
	ContentType     string // "Photo","Document"
}

var OCRScanSettings = ScanSettings{
	Resolution:      200,
	ColorSpace:      "Gray",
	Compression:     0,
	BitDepth:        8,
	GrayRendering:   "NTSC",
	Gamma:           1000,
	Brightness:      1000,
	Contrast:        1000,
	Highlite:        179,
	Shadow:          25,
	Threshold:       0,
	SharpeningLevel: 128,
	NoiseRemoval:    0,
	ContentType:     "Document",
}

var NormalScanSettings = ScanSettings{
	Resolution:      200,
	ColorSpace:      "Gray",
	Compression:     15,
	BitDepth:        8,
	GrayRendering:   "NTSC",
	Gamma:           1000,
	Brightness:      1000,
	Contrast:        1000,
	Highlite:        179,
	Shadow:          25,
	Threshold:       0,
	SharpeningLevel: 128,
	NoiseRemoval:    0,
	ContentType:     "Document",
}

var ColorScanSettings = ScanSettings{
	Resolution:      200,
	ColorSpace:      "Color",
	Compression:     15,
	BitDepth:        8,
	GrayRendering:   "NTSC",
	Gamma:           1000,
	Brightness:      1000,
	Contrast:        1000,
	Highlite:        179,
	Shadow:          25,
	Threshold:       0,
	SharpeningLevel: 128,
	NoiseRemoval:    0,
	ContentType:     "Photo",
}

type MapOfDestinationSettings map[string]*DestinationSettings

var DefaultDestination = MapOfDestinationSettings{
	"OCR": &DestinationSettings{
		Name:         "OCR",
		FilePattern:  &paramFolderPatern,
		DoOCR:        true,
		ScanSettings: &OCRScanSettings,
	},
	"Normal": &DestinationSettings{
		Name:         "Normal",
		FilePattern:  &paramFolderPatern,
		DoOCR:        true,
		ScanSettings: &NormalScanSettings,
	},
	"Color": &DestinationSettings{
		Name:         "Color",
		FilePattern:  &paramFolderPatern,
		DoOCR:        true,
		ScanSettings: &ColorScanSettings,
	},
}

type OCRSettings struct {
	UseScantailor bool
	UseTesseract  bool
	UseHocr2Pdf   bool
}

var DefaultOCRSettings = OCRSettings{
	UseScantailor: true,
	UseTesseract:  true,
	UseHocr2Pdf:   true,
}

package main

import (
//"encoding/xml"
)

type DestinationSettings struct {
	//XMLName        xml.Name `xml:DestinationSettings"`
	Name           string
	FilePattern    *string
	DoOCR          bool
	SourceDocument map[string]map[string]ScanSettings // Per Source, per Format
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

type MapOfDestinationSettings map[string]*DestinationSettings

var DefaultDestinationSettings = MapOfDestinationSettings{

	"Normal": &DestinationSettings{
		Name:        "Normal",
		FilePattern: &paramFolderPatern,
		SourceDocument: map[string]map[string]ScanSettings{
			"Platen": map[string]ScanSettings{
				"JPEG": ScanSettings{
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
				},
				"PDF": ScanSettings{
					Resolution:      300,
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
				},
			},
			"Adf": map[string]ScanSettings{
				"JPEG": ScanSettings{
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
				},
				"PDF": ScanSettings{
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
				},
			},
		},
	},
	"LowRes": &DestinationSettings{
		Name:        "LowRes",
		FilePattern: &paramFolderPatern,
		SourceDocument: map[string]map[string]ScanSettings{
			"Platen": map[string]ScanSettings{
				"JPEG": ScanSettings{
					Resolution:      75,
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
				},
				"PDF": ScanSettings{
					Resolution:      75,
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
				},
			},
			"Adf": map[string]ScanSettings{
				"JPEG": ScanSettings{
					Resolution:      75,
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
				},
				"PDF": ScanSettings{
					Resolution:      75,
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
				},
			},
		},
	},
}

// document.go

package main

import (
	"code.google.com/p/gofpdf"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type Document struct {
	t        time.Time // Used to determine document name
	FileType string    // JPEG, PDF

	TempDir   string   // Temporary folder for images
	ImageList []string // List of scanned images
	Previous  *Document
	timeout   *time.Timer
	Clean     bool
}

func NewDocument(previous *Document) (*Document, error) {
	defer Un(Trace("Document.NewDocument"))

	d := new(Document)
	t, err := ioutil.TempDir("", "scantopc")
	if err != nil {
		return nil, DeviceError("Document", "NewDocument", err)
	}

	d.TempDir = t
	d.t = time.Now()
	if previous != nil {
		if previous.Previous != nil {
			previous.Previous.CleanUp()
			previous.Previous = nil
		}
		d.Previous = previous
	}

	return d, err
}

func (d *Document) SetFileType(filetype string) error {
	defer Un(Trace("Document.SetFileType"))
	if filetype == "JPEG" || filetype == "PDF" {
		d.FileType = filetype
		return nil
	}
	return DeviceError("Document", "SetFileType : Unknown file type"+filetype, nil)
}

func (d *Document) CheckFolder(filename string) error {
	defer Un(Trace("Document.CheckFolder"))
	dir, _ := path.Split(filename)
	if dir != "" {
		err := os.MkdirAll(dir, filePERM)
		return err
	}
	return nil
}

func (d *Document) NewImageWritter() (io.WriteCloser, error) {
	defer Un(Trace("Document.NewImageWritter"))

	ImageName := d.TempDir + "/" + fmt.Sprintf("page-%04d.jpg", len(d.ImageList))
	out, err := os.Create(ImageName)
	if err != nil {
		return nil, DeviceError("Document.NewImageWritter", "os.Create "+ImageName, err)
	}
	d.ImageList = append(d.ImageList, ImageName)
	TRACE.Println("Image will be saved in ", ImageName)
	return out, err
}

func (d *Document) Save() (err error) {
	defer Un(Trace("Document.Save"))

	TRACE.Println("Document.Previous", d.Previous)

	if d.Previous != nil && d.Previous.Clean == false {
		if d.FileType == "PDF" && d.Previous.FileType == "PDF" && len(d.ImageList) == len(d.Previous.ImageList) {
			// Check if previous document has same page number... If so, let's create a double sided document in addtion to single side document...
			// The User will choose later on which one he wants to keep.
			err = d.SaveDoubleSidePDF()

			d.Previous.CleanUp()
			d.Previous = nil
			d.CleanUp()
		} else {
			d.Previous.CleanUp()
			d.Previous = nil
			err = d.SaveSingleSide()
		}
	} else {
		err = d.SaveSingleSide()
	}
	return
}

func (d *Document) SaveSingleSide() (err error) {
	defer Un(Trace("Document.SaveSingleSide"))

	switch d.FileType {
	case "PDF":
		err = d.SaveSingleSidePDF()
	case "JPEG":
		err = d.SaveJPEG()
	default:
		err = DeviceError("Document", "SaveSingleSide: Unknown file type["+d.FileType+"]", nil)
	}
	return err
}

func (d *Document) SaveDoubleSidePDF() error {
	defer Un(Trace("Document.SaveDoubleSide"))

	fileName, err := ExpandString(paramFolderPatern, d.Previous.t)
	if err != nil {
		return DeviceError("Document", "SaveDoubleSide", err)
	}
	fileName += "-doubleside.pdf"
	images := make([]string, 2*len(d.ImageList))
	for p := 0; p < len(d.ImageList); p++ {
		images[2*p] = d.Previous.ImageList[p]
		images[2*p+1] = d.ImageList[len(d.ImageList)-p-1]
	}
	err = d.SaveAsPDF(fileName, images)
	if err != nil {
		return err
	}
	return err
}

func (d *Document) SaveSingleSidePDF() error {
	defer Un(Trace("Document.SaveSingleSidePDF"))

	fileName, err := ExpandString(paramFolderPatern, d.t)
	if err != nil {
		return DeviceError("Document", "SaveSingleSidePDF", err)
	}
	fileName += ".pdf"
	err = d.SaveAsPDF(fileName, d.ImageList)
	if err != nil {
		return err
	}
	return err
}

func (d *Document) SaveAsPDF(filename string, images []string) error {
	defer Un(Trace("Document.SaveAsPDF", filename))

	err := d.CheckFolder(filename)
	if err != nil {
		return DeviceError("Document", "SaveAsPDF", err)
	}

	out, err := os.Create(filename)
	if err != nil {
		return DeviceError("Document", "SaveAsPDF", err)
	}

	defer out.Close()
	pdf := gofpdf.New("P", "mm", "A4", "")
	for _, page := range images {
		TRACE.Println("\tAdd image", page)
		pdf.AddPage()
		pdf.Image(page, 0, 0, 210, 297, false, "", 0, "")
	}

	err = pdf.OutputAndClose(out)
	if err != nil {
		return DeviceError("Document", "SaveAsPDF", err)
	}
	INFO.Println("Document saved", filename)
	return nil
}

func (d *Document) SaveJPEG() error {
	defer Un(Trace("Document.SaveJPEG"))

	fileName, err := ExpandString(paramFolderPatern, d.t)
	err = d.CheckFolder(fileName)
	if err != nil {
		return DeviceError("Document", "SaveJPEG", err)
	}

	//prepare filname pattern for JPEGs
	fileName, err = ExpandString(paramFolderPatern, d.t)
	fileName += "-%04d.jpg"

	for p, page := range d.ImageList {
		dest := fmt.Sprintf(fileName, p)
		_, err = CopyFile(page, dest)
		if err != nil {
			break
		}
	}

	if err != nil {
		return DeviceError("Document", "SaveJPEG", err)
	}

	err = d.CleanUp()
	return err
}

func (d *Document) CleanUp() error {
	defer Un(Trace("Document.CleanUp"))
	err := os.RemoveAll(d.TempDir)
	if err != nil {
		return DeviceError("Document", "CleanUp", err)
	}
	d.Clean = true
	return err
}

// Utility
func CopyFile(src, dst string) (int64, error) {
	defer Un(Trace("CopyFile", src, dst))
	sf, err := os.Open(src)
	defer sf.Close()
	if err != nil {
		CheckError(src, err)
		return 0, err
	}
	df, err := os.Create(dst)
	defer df.Close()
	if err != nil {
		CheckError(dst, err)
		return 0, err
	}
	return io.Copy(df, sf)
}

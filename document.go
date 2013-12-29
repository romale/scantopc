package main

/*
	Implemente a batch of documents

*/

import (
	"fmt"
	"github.com/simulot/hpdevices"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

type DocumentError struct {
	Operation string
	Message   string
	Err       error
}

// Implement Error
func (e DocumentError) Error() string {
	if e.Err != nil {
		return e.Operation + ": " + e.Message + ", " + e.Err.Error()
	} else {
		return e.Operation + ": " + e.Message
	}
}

func NewDocumentError(operation, message string, err ...error) error {
	e := DocumentError{operation, message, nil}
	if len(err) > 0 {
		e.Err = err[0]
	}
	fmt.Println("NewDocumentError ", e)
	return e
}

type OCRBatchImageManager struct {
	tempfolder    string
	settings      *hpdevices.DestinationSettings
	doctype       string
	format        string
	previousbatch *OCRBatchImageManager
	imagelist     []*imageJob
	imageJobChan  chan *imageJob
	filename      string // Final document name
	when          time.Time
}

func NewOCRBatchImageManager(doctype string, destination *hpdevices.DestinationSettings, format string, previousbatch hpdevices.DocumentBatchHandler) (bh hpdevices.DocumentBatchHandler, err error) {
	bm := new(OCRBatchImageManager)

	INFO.Println("New scan batch started:", destination.Name, doctype)

	bm.settings = destination
	bm.doctype = doctype
	switch doctype {
	case "Jpeg":
		bm.format = ".jpg"
	case "PDF":
		bm.format = ".pdf"
	default:
		err = NewDocumentError("NewOCRBatchImageManager", "Unknown format "+format, nil)
		return nil, err
	}

	bm.tempfolder, err = ioutil.TempDir("", "scantopc")
	if err != nil {
		return nil, NewDocumentError("NewOCRBatchImageManager", "", err)
	}

	TRACE.Println("Temp folder for this batch is", bm.tempfolder)
	bm.when = time.Now()
	if previousbatch != nil {
		if bm.settings.Verso {
			TRACE.Println("Verso batch and previous batch known")
		}
		bm.previousbatch = previousbatch.(*OCRBatchImageManager)
	}
	bm.imagelist = make([]*imageJob, 0)
	bm.imageJobChan = make(chan *imageJob)
	return hpdevices.DocumentBatchHandler(bm), nil
}

func (bm *OCRBatchImageManager) NewImageWriter() (file io.WriteCloser, err error) {
	ij, err := NewImageJob(bm.tempfolder+"/"+fmt.Sprintf("page-%04d.jpg", len(bm.imagelist)), bm.imageJobChan)
	INFO.Println("Recieving page from scanner:", ij.filename)
	bm.imagelist = append(bm.imagelist, ij)
	return ij, nil
}

func (bm *OCRBatchImageManager) CloseDocumentBatch() error {
	TRACE.Println("Last page recieved")
	// Put here code to generate final pdf
	go bm.FinalizeDocumentBatch()
	return nil
}

func (bm *OCRBatchImageManager) FinalizeDocumentBatch() {
	defer Un(Trace("OCRBatchImageManager.FinalizeDocumentBatch"))
	// This code is placed in a go routine to allow starting a new scan job while finishing OCR

	// Wait for all image treatment finished
	nbErr := 0
	for i := 0; i < len(bm.imagelist); i++ {
		TRACE.Println("OCRBatchImageManager.FinalizeDocumentBatch", "waiting image", i)
		ij := <-bm.imageJobChan
		if ij.err == nil {
			TRACE.Println("Image", ij.filename, "is ready")
		} else {
			TRACE.Println("Image", ij.filename, "has failed with error", ij.err)
			nbErr++
		}
	}
	INFO.Println("Last treatment for batch is finished")
	_ = nbErr
	// At that point, all scanned images have been processed or are errored
	if bm.previousbatch != nil {
		prevBatch := bm.previousbatch
		if bm.settings.Verso && len(prevBatch.imagelist) == len(bm.imagelist) {
			newImageList := make([]*imageJob, 2*len(bm.imagelist))
			for i := 0; i < len(bm.imagelist); i++ {
				newImageList[2*i] = prevBatch.imagelist[i]
				newImageList[2*i+1] = bm.imagelist[len(bm.imagelist)-i-1]
			}
			prevBatch.CombinePages(newImageList)
			prevBatch.CleanUp()
		} else {
			prevBatch.CleanUp()
			bm.CombinePages(bm.imagelist)
		}
	} else {
		bm.CombinePages(bm.imagelist)
	}
}

/*
	Clean all temporary files created by the process in TMP folder
*/

func (bm *OCRBatchImageManager) CleanUp() error {
	return os.RemoveAll(bm.tempfolder)
}

/*
	Clean up tempory folder and errase final file.
	This is used when the previous image batch is combined with current

*/

func (bm *OCRBatchImageManager) Erase() error {
	os.RemoveAll(bm.tempfolder)
	return os.Remove(bm.filename)
}

func (bm *OCRBatchImageManager) CombinePages(imagelist []*imageJob) {
	var err error
	bm.filename, err = ExpandString(*bm.settings.FilePattern, bm.when)
	if err != nil {
		ERROR.Print("Name pattern is incorrect. Job discarded", err)
	}
	if err == nil {
		bm.filename += bm.format
		bm.CreatePDF(imagelist)
	}
}

func (bm *OCRBatchImageManager) CreatePDF(imagelist []*imageJob) {
	switch paramPFDTool {
	case "pdfunite":
		CreatePDFUsingPDFunite(bm.filename, imagelist)
	case "pdftk":
		CreatePDFUsingPDFTK(bm.filename, imagelist)
	}
}

func CreatePDFUsingPDFTK(filename string, images []*imageJob) error {
	argList := make([]string, 0)
	for i := 0; i < len(images); i++ {
		p := images[i]
		d, f := path.Split(p.filename)
		argList = append(argList, d+"ocr-"+f+".pdf")
	}
	arglist := append(argList, "cat", "output", filename)
	cmd := exec.Command("pdftk", arglist...)
	out, err := cmd.CombinedOutput()
	fmt.Println("pdftk", filename, "processed\n", err, "\n", string(out))
	return err
}

func CreatePDFUsingPDFunite(filename string, images []*imageJob) error {
	if len(images) > 1 {
		argList := make([]string, 0)
		for i := 0; i < len(images); i++ {
			p := images[i]
			d, f := path.Split(p.filename)
			argList = append(argList, d+"ocr-"+f+".pdf")
		}
		arglist := append(argList, filename)
		cmd := exec.Command("pdfunite", arglist...)
		out, err := cmd.CombinedOutput()
		fmt.Println("pdfunite", filename, "processed\n", err, "\n", string(out))
		return err
	}
	d, f := path.Split(images[0].filename)
	_, err := CopyFile(d+"ocr-"+f+".pdf", filename)
	return err

}

/*
Check if all dependencies are met
*/

func CheckOCRDependencies() (r bool) {
	r = false
	if paramOCR {
		path, err := exec.LookPath("convert")
		TRACE.Println("convert", path, err)
		if err != nil {
			r = r || true
			ERROR.Print("convert executable not found. Please check imagemagick installation.")
		}
		path, err = exec.LookPath("tesseract")
		TRACE.Println("tesseract", path, err)
		if err != nil {
			r = r || true
			ERROR.Print("tesseract executable not found. Please check tesseract installation.")
		}
		path, err = exec.LookPath("hocr2pdf")
		TRACE.Println("hocr2pdf", path, err)
		if err != nil {
			r = r || true
			ERROR.Print("hocr2pdf executable not found. Please check installation (http://www.exactcode.com/site/open_source/exactimage/hocr2pdf/).")
		}
		path1, err1 := exec.LookPath("pdftk")
		TRACE.Println("pdftk", path1, err1)
		path2, err2 := exec.LookPath("pdfunite")
		TRACE.Println("pdfunite", path2, err1)
		if err1 != nil && err2 != nil {
			r = r || true
			ERROR.Print("Neither pdftk or pdfunit (from poppler-utils package) executable are not found. Please check installation.")
		}
		if paramPFDTool == "" && err2 == nil {
			paramPFDTool = "pdfunite"
			INFO.Println("PDF tool to be used", path2)
		}
		if paramPFDTool == "" && err1 == nil {
			paramPFDTool = "pdftk"
			INFO.Println("PDF tool to be used", path1)
		}
	}
	return r
}

// Utility
func CopyFile(src, dst string) (int64, error) {
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

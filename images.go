package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"
)

type imageJob struct {
	file     *os.File
	filename string
	err      error
	endChan  chan<- *imageJob
}

func (ij *imageJob) Write(b []byte) (int, error) {
	return ij.file.Write(b)
}

func NewImageJob(filename string, endchan chan<- *imageJob) (ij *imageJob, err error) {
	ij = new(imageJob)
	ij.file, err = os.Create(filename)
	ij.filename = filename
	ij.endChan = endchan
	if err != nil {
		return nil, NewDocumentError("NewImageJob", "", err)
	}
	TRACE.Println("Opening", filename, "for recieving image")
	return ij, nil
}

func (ij *imageJob) Close() (err error) {
	TRACE.Println("Closing", ij.filename)
	ij.err = ij.file.Close()
	if ij.err == nil {
		go ij.ImageProcessing()
	}
	return ij.err
}

func (ij *imageJob) ImageProcessing() {
	TRACE.Println("Processing", ij.filename)
	err := ij.ImproveImage()
	if err == nil {
		err = ij.OCRImage()
	}
	if err == nil {
		err = ij.CombineHOCRandPDF()
	}
	TRACE.Println("Processed", ij.filename)
	ij.endChan <- ij
	TRACE.Println("Processed event sent", ij.filename)
}

func (ij *imageJob) ImproveImage() (err error) {
	dir, file := path.Split(ij.filename)
	cmd := exec.Command("convert",
		dir+file,
		"-background", "white",
		"-fuzz", "75%",
		"-deskew", "50%",
		dir+"ocr-"+file)
	out, err := TimeOutCombinedOutput(time.Minute, cmd)
	if err != nil {
		ERROR.Println("imageJob.ImproveImage", "Command "+cmd.Path+" has failed", err)
		fmt.Println("Command output", string(out))
	}
	return err
}

func (ij *imageJob) OCRImage() (err error) {
	dir, file := path.Split(ij.filename)
	//TODO: tesseract language should be a parameter
	cmd := exec.Command("tesseract",
		dir+"ocr-"+file,
		dir+"ocr-"+file,
		"-l", "fra",
		"hocr")
	out, err := TimeOutCombinedOutput(time.Minute, cmd)
	if err != nil {
		ERROR.Println("imageJob.OCRImage", "Command "+cmd.Path+" has failed", err)
		fmt.Println("Command output", string(out))
	}
	return err
}

func (ij *imageJob) CombineHOCRandPDF() (err error) {
	dir, file := path.Split(ij.filename)
	cmd := exec.Command("hocr2pdf",
		"--input", dir+"ocr-"+file,
		"--output", dir+"ocr-"+file+".pdf")
	in, err := os.Open(dir + "ocr-" + file + ".html")
	if err != nil {
		err = NewDocumentError("imageJob.CombineHOCRandPDF", "Reading hocr file", err)
		return
	}
	defer in.Close()
	cmd.Stdin = in
	var out []byte
	if err == nil {
		out, err = TimeOutCombinedOutput(time.Minute, cmd)
	}
	if err != nil {
		ERROR.Println("imageJob.CombineHOCRandPDF", "Command "+cmd.Path+" has failed", err)
		fmt.Println("Command output", string(out))
	}
	return err
}

// Launch a command and kill it when timeout
func TimeOutCombinedOutput(to time.Duration, cmd *exec.Cmd) (out []byte, err error) {

	done := make(chan bool)
	go func() {
		defer Un(Trace("TimeOutCombinedOutput.goroutine"))
		out, err = cmd.CombinedOutput()
		done <- true
	}()

	select {
	case <-time.After(to):
		if err = cmd.Process.Kill(); err != nil {
			err = NewDocumentError("TimeOutCombinedOutput", "Kill error "+cmd.Path, err)
		} else {
			err = NewDocumentError("TimeOutCombinedOutput", "Command "+cmd.Path+" ran too long")
		}
		<-done // eat the true value generated when the excecution of goroutine as ended.
	case <-done:
	}
	close(done)
	return
}

//   Reformatted by   jeanf    samedi 4 janvier 2014, 19:08:31 (UTC+0100)

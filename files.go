// files.go
package main

import (
	"code.google.com/p/gofpdf"
	"fmt"
	"io"
	"os"
	"path"
)

func CheckFolder(filename string) error {
	dir, _ := path.Split(filename)
	if dir != "" {
		err := os.MkdirAll(dir, filePERM)
		return err
	}
	return nil
}

func SaveAsPDFSimplex(Job *Job) {
	Un(Trace("SaveAsPDFSimplex", Job))

	pdfFile := Job.Path() + ".pdf"
	CheckFolder(pdfFile)
	out, err := os.Create(pdfFile)
	CheckError("Create "+pdfFile, err)
	defer out.Close()
	TRACE.Println("SaveAsPDF", pdfFile)
	pdf := gofpdf.New("P", "mm", "A4", "")
	for _, page := range Job.ImageList {
		TRACE.Println("\tAdd image", page)
		pdf.AddPage()
		pdf.Image(page, 0, 0, 210, 297, false, "", 0, "")
	}
	pdf.OutputAndClose(out)
	INFO.Println("Document saved", pdfFile)
}

func SaveAsJPEG(Job *Job) {
	Un(Trace("SaveAsJPEG", Job))
	fileStub := Job.Path()
	CheckFolder(fileStub + ".jpg")
	for i, page := range Job.ImageList {
		name := fmt.Sprintf("%s-%04d.jpeg", fileStub, i)
		_, err := CopyFile(page, name)
		CheckError("CopyFile", err)
	}
	INFO.Println("Image(s) saved")
}

func CopyFile(src, dst string) (int64, error) {
	Un(Trace("CopyFile", src, dst))
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

// pdf
package main

import (
	"code.google.com/p/gofpdf"
	"os"
)

func doSaveAsPDF(ScanJob *ScanJob) {
	pdfFile := ScanJob.FileName + ".pdf"
	out, err := os.Create(pdfFile)
	CheckError("Create "+pdfFile, err)
	defer out.Close()
	TRACE.Println("doSaveAsPDF", pdfFile)
	pdf := gofpdf.New("P", "mm", "A4", "")
	for _, page := range ScanJob.pages {
		TRACE.Println("\tAdd image", page)
		pdf.AddPage()
		pdf.Image(page, 0, 0, 210, 297, false, "", 0, "")
	}
	pdf.OutputAndClose(out)
	for _, page := range ScanJob.pages {
		os.Remove(page)
	}
	INFO.Println("Document saved", pdfFile)
}

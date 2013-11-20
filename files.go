// files.go
package main

import (
	"code.google.com/p/gofpdf"
	"os"
	"path"
	//"strconv"
	//"strings"
)

var uID, gID int = 0, 0

func CheckFolder(filename string) error {
	dir, _ := path.Split(filename)
	if dir != "" {
		err := os.MkdirAll(dir, filePERM)
		//if fileUserGroup != "" {
		//	err = os.Chown(filename, uID, gID)
		//}
		return err
	}
	return nil
}

func CheckIDs(fileUserGroup string) {
	//if fileUserGroup != "" {
	//	s := strings.Split(fileUserGroup, ":")
	//	if len(s) == 2 {
	//		var err error
	//		if uID, err = strconv.Atoi(s[0]); err != nil {
	//			ERROR.Fatalln("UserID must be numerical")
	//		}
	//		if gID, err = strconv.Atoi(s[0]); err != nil {
	//			ERROR.Fatalln("GroupID must be numerical")
	//		}
	//	} else {
	//		ERROR.Fatalln("User ID and Group ID should be provided like following: uid:gid")
	//	}
	//} else {
	//	uID = os.Getuid()
	//	gID = os.Getgid()
	//}

}

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
	//err = os.Chown(pdfFile, uID, gID)
	INFO.Println("Document saved", pdfFile)
}

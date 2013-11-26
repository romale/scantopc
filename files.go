// files.go
package main

import (
	"code.google.com/p/gofpdf"
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

const timeout = time.Second * 30

func CheckFolder(filename string) error {
	dir, _ := path.Split(filename)
	if dir != "" {
		err := os.MkdirAll(dir, filePERM)
		return err
	}
	return nil
}

const (
	StatusIdling                = "StatusIdling"
	StatusIdlingExpectingDuplex = "StatusIdlingExpectingDuplex"
	StatusProcessingSimplex     = "StatusProcessingSimplex"
	StatusProcessingDuplex      = "StatusProcessingDuplex"
)

const (
	EventStartJob      = "EventStartJob"
	EventProcessingJob = "EventProcessingJob"
	EventEndJob        = "EventEndJob"
	EventTimeOut       = "EventTimeOut"
)

func (s *StateMachine) ResetTimer() {
	//	defer Un(Trace("StateMachine.ResetTimer", s.CurrentStatus))
	s.Timeout.Reset(timeout)
}
func (s *StateMachine) StopTimer() {
	//	defer Un(Trace("StateMachine.StopTimer", s.CurrentStatus))
	s.Timeout.Stop()
}

func (s *StateMachine) JobStart(event *Event) string {
	defer Un(Trace("StateMachine.StartJob", s.CurrentStatus, *event))
	nextStatus := s.CurrentStatus
	switch s.CurrentStatus {
	case StatusIdling:
		s.Recto = event.j
		s.ResetTimer()
		nextStatus = StatusProcessingSimplex
	case StatusIdlingExpectingDuplex, StatusProcessingSimplex:
		s.ResetTimer()
		s.Verso = event.j
		nextStatus = StatusProcessingDuplex
	case StatusProcessingDuplex:
		go SaveDuplex(s.Recto, s.Verso)
		s.ResetTimer()
		s.Verso = nil
		s.Recto = event.j
		nextStatus = StatusProcessingSimplex
	}
	return nextStatus
}

func (s *StateMachine) JobProcessing(event *Event) string {
	defer Un(Trace("StateMachine.ProcessingJob", s.CurrentStatus, *event))

	nextStatus := s.CurrentStatus
	switch s.CurrentStatus {
	case StatusIdling:
		s.StopTimer()
		nextStatus = StatusIdling
	default:
		s.ResetTimer()
		nextStatus = s.CurrentStatus
	}
	return nextStatus
}

func (s *StateMachine) JobEnd(event *Event) string {
	defer Un(Trace("StateMachine.EndJob", s.CurrentStatus, *event))

	nextStatus := s.CurrentStatus

	switch s.CurrentStatus {
	case StatusIdling:
		ERROR.Println("Should not having EndJob while StatusIdling ")
		nextStatus = StatusIdling
	case StatusProcessingSimplex:
		if s.Recto.DocumentType == "PDF" {
			// Potential duplex job
			nextStatus = StatusIdlingExpectingDuplex
			s.ResetTimer()
		} else {
			go SaveSimplex(s.Recto)
			s.Recto = nil
			s.StopTimer()
			nextStatus = StatusIdling
		}
	case StatusIdlingExpectingDuplex:
		ERROR.Println("Should not having EndJob while StatusIdlingExpectingDuplex ")
		nextStatus = StatusIdling
	case StatusProcessingDuplex:
		go SaveDuplex(s.Recto, s.Verso)
		s.Verso = nil
		s.Recto = nil
		s.StopTimer()
		nextStatus = StatusIdling
	}
	return nextStatus
}

func (s *StateMachine) JobTimeOut(event *Event) string {
	defer Un(Trace("StateMachine.EventTimeOut", s.CurrentStatus, *event))

	nextStatus := s.CurrentStatus
	switch s.CurrentStatus {
	case StatusIdling:
		s.StopTimer()
		nextStatus = s.CurrentStatus
	case StatusProcessingSimplex, StatusIdlingExpectingDuplex:
		go SaveSimplex(s.Recto)
		s.StopTimer()
		nextStatus = StatusIdling
	case StatusProcessingDuplex:
		go SaveDuplex(s.Recto, s.Verso)
		s.StopTimer()
		nextStatus = StatusIdling
	}
	return nextStatus
}

type Event struct {
	id string
	t  time.Time
	j  *Job
}

func NewEvent(id string, job *Job) *Event {
	return &Event{id, time.Now(), job}
}

type StateMachine struct {
	CurrentStatus string
	Recto, Verso  *Job
	Timeout       *time.Timer
	Channel       chan *Event
}

func NewStateMaching(name string) *StateMachine {
	s := new(StateMachine)
	s.CurrentStatus = StatusIdling
	s.Timeout = time.NewTimer(60 * time.Second)
	s.Timeout.Stop()
	s.Channel = make(chan *Event)
	go s.Loop()
	return s
}

func (s *StateMachine) SendEvent(id string, job *Job) {
	defer Un(Trace("SendEvent", id))
	s.Channel <- NewEvent(id, job)
}

// Machine state loop, unify external events with internal timeouts
func (s *StateMachine) Loop() {
	defer Un(Trace("StateMachine.StateMachine.Loop"))
	for {
		select {
		case event := <-s.Channel:
			s.ProcessEvent(event)

		case <-s.Timeout.C:
			s.ProcessEvent(NewEvent(EventTimeOut, nil))
		}
	}
}

func (s *StateMachine) ProcessEvent(event *Event) {
	defer Un(Trace("StateMachine.ProcessEvent", *event))

	switch event.id {
	case EventStartJob:
		s.CurrentStatus = s.JobStart(event)
	case EventProcessingJob:
		s.CurrentStatus = s.JobProcessing(event)
	case EventEndJob:
		s.CurrentStatus = s.JobEnd(event)
	case EventTimeOut:
		s.CurrentStatus = s.JobTimeOut(event)
	}
	TRACE.Println("Next status", s.CurrentStatus)
}

func SaveDuplex(recto, verso *Job) {
	defer Un(Trace("SaveDuplex", recto, verso))

	if recto.DocumentType == verso.DocumentType && recto.DocumentType == "PDF" {

		pdfFile := recto.Path() + ".pdf"
		CheckFolder(pdfFile)
		out, err := os.Create(pdfFile)
		CheckError("Create "+pdfFile, err)
		defer out.Close()

		TRACE.Println("SaveAsPDF", pdfFile)
		pdf := gofpdf.New("P", "mm", "A4", "")
		for p := 0; p < recto.PageNumber(); p++ {
			TRACE.Println("\tAdd image", recto.ImageList[p])
			pdf.AddPage()
			pdf.Image(recto.ImageList[p], 0, 0, 210, 297, false, "", 0, "")
			TRACE.Println("\tAdd image", verso.ImageList[verso.PageNumber()-p-1])
			pdf.AddPage()
			pdf.Image(verso.ImageList[verso.PageNumber()-p-1], 0, 0, 210, 297, false, "", 0, "")
		}
		pdf.OutputAndClose(out)
		INFO.Println("Document saveEndJobd", pdfFile)
		recto.Finalize()
		verso.Finalize()
	} else {
		SaveSimplex(recto)
		SaveSimplex(verso)
	}

}

func SaveSimplex(Job *Job) {
	defer Un(Trace("SaveSimplex", Job))
	if Job.DocumentType == "PDF" {
		SaveSimplexPDF(Job)
	} else {
		SaveSimplexJPG(Job)
	}
}

func SaveSimplexJPG(Job *Job) {
	defer Un(Trace("SaveSimplexJPG", Job))

	jpgFile := Job.Path() + ".jpg"
	CheckFolder(jpgFile)
	for p, page := range Job.ImageList {
		dest := fmt.Sprintf("%s-%04d.jpg", p)
		_, err := CopyFile(page, dest)
		TRACE.Println("Save", dest, err)
	}
	Job.Finalize()
}
func SaveSimplexPDF(Job *Job) {
	defer Un(Trace("SaveAsPDFSimplex", Job))

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
	Job.Finalize()
}

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

package main

import (
	"fmt"
)

type Runner interface {
	Run() Runner
	Error() error
}

const (
	NbWKProcess = 2
)

type ImageProcessor struct {
	jobChannel chan Runner
	jobResult  chan Runner

	submittedJobs int
	endedJobs     int
}

func NewImageProcessor() *ImageProcessor {
	defer Un(Trace("NewImageProcessor"))
	ip := new(ImageProcessor)
	ip.jobChannel = make(chan Runner, 1000)
	ip.jobResult = make(chan Runner, 1000)

	// Launch workrers has goroutines
	for i := 0; i < NbWKProcess; i++ {
		go ip.Worker(i)
	}
	TRACE.Println("NewImageProcessor:", NbWKProcess, "workers ready")
	return ip
}

func (ip *ImageProcessor) Worker(id int) {
	defer Un(Trace("ImageProcessor.Worker", id))
	for j := range ip.jobChannel {
		TRACE.Println("ImageProcessor.Worker", id, "Launch job", j)
		ip.jobResult <- j.Run()
		TRACE.Println("ImageProcessor.Worker", id, "Job done", j)
	}
	fmt.Println("Worker ", id, " is ended")
}

func (ip *ImageProcessor) PushJob(j Runner) {
	defer Un(Trace("ImageProcessor.PushJob", j))
	ip.submittedJobs++
	ip.jobChannel <- j
}

func (ip *ImageProcessor) WaitWorkersResults() {
	defer Un(Trace("ImageProcessor.CloseWorkers"))
	close(ip.jobChannel)
	for i := 0; i < ip.submittedJobs; i++ {
		j := <-ip.jobResult
		if j.Error() != nil {
			ERROR.Println("Job for", j, " has failed and returned error:", j.Error())
		} else {
			TRACE.Println("Job for", j, " has successed")
		}
	}
}

package main

import (
	"fmt"
	"testing"
	"time"
)

func Test_ScanToPC(t *testing.T) {
	go main()

	time.Sleep(15 * 60 * time.Second)
	fmt.Println("Done")
}

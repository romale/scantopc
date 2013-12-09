package main

// build ignore

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"testing"
	//"time"
	"bytes"
)

func Test_Desination(t *testing.T) {
	configfile, err := os.Create("scantopc.cfg")
	defer configfile.Close()

	if err != nil {
		t.Error("Create", err)
	}
	array, err := xml.Marshal(DefaultDestinationSettings["Normal"].SourceDocument["Platen"]["JPEG"])
	if err != nil {
		fmt.Println(err)
		t.Error("Marshal", err)
	}
	r := bytes.NewBufferString(XMLHeader + string(array))
	io.Copy(configfile, r)
	configfile.Close()
}

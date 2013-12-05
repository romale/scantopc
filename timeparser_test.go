// timeparser_test.go
package main

import (
	"fmt"
	//"testing"
	"time"
)

func ExampleParser() {
	//t := time.Now()
	t := time.Date(2009, time.March, 10, 15, 53, 10, 0, time.Local)
	s, err := ExpandString("%Y/%m/%d", t)
	fmt.Println(s)
	s, err = ExpandString("%H%M%S", t)
	fmt.Println(s)
	s, err = ExpandString("%Y/%Y.%m/%Y.%m.%d-%H.%M.%S", t)
	fmt.Println(s)
	s, err = ExpandString("%Y\\%Y.%m\\%Y.%m.%d-%H.%M.%S", t)
	fmt.Println(s)
	s, err = ExpandString("%Y/%W", t)
	fmt.Println(err)
	s, err = ExpandString("%Y/%m.%", t)
	fmt.Println(err)

	// Output:
	// 2009/03/10
	// 155310
	// 2009/2009.03/2009.03.10-15.53.10
	// 2009\2009.03\2009.03.10-15.53.10
	// Unknown token %W
	// % can't be the last layout's character

}

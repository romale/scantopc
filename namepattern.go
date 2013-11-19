// namepattern.go
package main

import (
	"errors"
	"fmt"
	"time"
)

func TokensUsage() {
	fmt.Println("\nAllowed tokens are:\n",
		` 	%Y  Year (4 digits):      2014
	%y  Year (2 digits):      14	   
	%d  Day (2 digits):       03
	%A  Weekday (long):       Monday
	%a  Weekday (short):      Mon
	%m  Month (2 digits):     February
	%I  Hour (12 hour):       05
	%H  Hour (24 hour):       17
	%M  Minute (2 digits):    54
	%S  Second (2 digits):    20
	%p  AM / PM:              PM
	`)
}

func ExpandString(layout string, t time.Time) (value string, err error) {
	value = ""
	err = nil
	for i := 0; i < len(layout); {
		c := layout[i]
		i++
		if c != '%' {
			value += string(c)
		} else {
			if i == len(layout) {
				err = errors.New("% can't be the last layout character")
				break
			}
			c = layout[i]
			i++
			switch c {
			case 'Y': // Year, full
				value += t.Format("2006")
			case 'y': // Year, 2 digits
				value += t.Format("06")
			case 'd': // Day, 2 digits
				value += t.Format("02")
			case 'A': // Day Name full
				value += t.Format("Monday")
			case 'a': // Day short
				value += t.Format("Mon")
			case 'm': // Month 2 digits
				value += t.Format("01")
			case 'b': // Month short
				value += t.Format("Jan")
			case 'B': // Month short
				value += t.Format("January")
			case 'I': // Hour 12 format
				value += t.Format("03")
			case 'H': // Hour 12 format
				value += t.Format("15")
			case 'M': // Minutes
				value += t.Format("04")
			case 'S': // Second
				value += t.Format("05")
			case 'p': // am / pm
				value += t.Format("pm")
			default:
				err = errors.New(fmt.Sprintf("Unknown token %%%c", c))
				break
			}
		}
	}
	return value, err
}

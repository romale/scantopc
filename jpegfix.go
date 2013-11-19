// jpegfix.go
package main

/* jpegfix gets jpeg stream delivered by HP scanner when using ADF
Segment DCT 0xFFC0 has a bad Lines field.

*/

import (
	//	"bufio"
	"errors"
	"io"
)

func CopyAndFixJPEG(w io.Writer, r io.Reader, ActualLineNumber int) (written int64, err error) {
	i := 0
	buf := make([]byte, 256)
	l, err := r.Read(buf)
	if l != len(buf) || (buf[0] != 0xff && buf[1] != 0xd8) {
		return 0, errors.New("Not a JPEG stream")
	}
	i = 2
	for i < len(buf) && buf[i] == 0xff && buf[i+1] != 0xc0 {
		size := int(buf[i+2])<<8 + int(buf[i+3]) + 2
		if size < 0 {
			return 0, errors.New("Frame lengh invalid")
		}
		i += size
	}
	if i >= len(buf) || (buf[i] != 0xff && buf[i+1] != 0xc0) {
		return 0, errors.New("SOF marker not found in the header")
	}

	// Fix line number
	i += 5
	if buf[i] == 0xff && buf[i+1] == 0xff {
		// Affected image
		buf[i] = byte(ActualLineNumber >> 8)
		buf[i+1] = byte(ActualLineNumber & 0x00ff)
	}
	// Write header
	l, err = w.Write(buf)
	written += int64(l)

	// stream
	if l == len(buf) && err == nil {
		buf = make([]byte, 32*1024)
		for {
			nr, er := r.Read(buf)
			if nr > 0 {
				nw, ew := w.Write(buf[0:nr])
				if nw > 0 {
					written += int64(nw)
				}
				if ew != nil {
					err = ew
					break
				}
				if nr != nw {
					err = io.ErrShortWrite
					break
				}
			}
			if er == io.EOF {
				break
			}
			if er != nil {
				err = er
				break
			}
		}
	}
	return written, err
}

// hocr2txt project main.go
package main

import (
	"code.google.com/p/go.net/html"
	"io"
	"os"
	"strings"
)

func hocr2html(infile string) (err error) {
	defer Un(Trace("hocr2html", infile))
	in, err := os.Open(infile)
	if err != nil {
		return err
	}
	out, err := os.Create(infile + ".txt")
	if err != nil {
		return err
	}
	defer out.Close()
	doc, err := html.Parse(in)
	if err != nil {
		return err
	}
	NodeWrite(doc, out)
	return
}

var depth = 0

func NodeWrite(n *html.Node, out io.Writer) {
	if n.Type == html.TextNode {
		out.Write([]byte(StripSpaces(n.Data)))
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		NodeWrite(c, out)
	}
	if n.Type == html.ElementNode {
		switch n.Data {
		case "div":
			//out.Write([]byte("\n"))
		case "p":
			//out.Write([]byte("\n"))
		}
	}

}

func StripSpaces(s string) (r string) {
	r = strings.Replace(s, "\n", " ", -1)
	r = strings.Replace(s, "—", "-", -1)
	for strings.Index(r, "  ") >= 0 {
		r = strings.Replace(r, "  ", " ", -1)
	}
	return
}

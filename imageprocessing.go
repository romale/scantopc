// imageprocessing.go

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
)

type Job struct {
	Image string
	DPI   int
	err   error
}

// Statisfies String interface
func (j *Job) String() string {
	return j.Image
}

func NewImageJob(Image string) Runner {
	j := new(Job)
	j.Image = Image
	return j
}

//Statisfies Runner interface
func (j *Job) Error() error {
	return j.err
}

func (j *Job) Run() Runner {
	defer Un(Trace("Job.Run"))
	var (
		cmd *exec.Cmd
		out []byte
	)

	dir, file := path.Split(j.Image)

	cmd = exec.Command("convert",
		//run convert $a -set filename:f "%t" -background white -fuzz 75% -deskew 50% +repage out/%[filename:f]_cropped.png;
		dir+file,
		"-background", "white",
		"-fuzz", "75%",
		"-deskew", "50%",
		dir+"ocr/"+file)
	fmt.Println(cmd.Args)
	out, j.err = cmd.CombinedOutput()
	fmt.Println("Image", j.err, j.Image, "processed\n", string(out))

	if j.err == nil {
		// run tesseract "$base.tif" "$base" -l $LANG hocr
		cmd = exec.Command("tesseract",
			dir+"ocr/"+file,
			dir+"ocr/"+file,
			"-l", "fra",
			"hocr")
		out, j.err = cmd.CombinedOutput()
		fmt.Println("Image", j.err, j.Image, "processed\n", string(out))

	}
	if j.err == nil {
		// Extract text from HOCR file
		j.err = hocr2html(dir + "ocr/" + file + ".html")
		fmt.Println(j.err)
	}
	if j.err == nil {
		// run 	hocr2pdf -i "$base.tif" -s -o "$base.pdf" < "$base.html"

		var hocr *os.File
		hocr, j.err = os.Open(dir + "ocr/" + file + ".html")
		defer hocr.Close()

		cmd = exec.Command("hocr2pdf",
			"-i", dir+"ocr/"+file,
			"-s",
			"-o", dir+"ocr/"+file+".pdf")
		cmd.Stdin = hocr // Inject hocr in subprocess stdin

		out, j.err = cmd.CombinedOutput()
		fmt.Println("Image", j.err, j.Image, "processed\n", string(out))
	}

	return Runner(j)
}

func CheckOCRDependencies() (r bool) {
	r = false
	path, err := exec.LookPath("convert")
	TRACE.Println("convert", path, err)
	if err != nil {
		r = r || true
		ERROR.Print("convert executable not found. Please check imagemagick installation.")
	}
	path, err = exec.LookPath("tesseract")
	TRACE.Println("tesseract", path, err)
	if err != nil {
		r = r || true
		ERROR.Print("tesseract executable not found. Please check tesseract installation.")
	}
	path, err = exec.LookPath("hocr2pdf")
	TRACE.Println("hocr2pdf", path, err)
	if err != nil {
		r = r || true
		ERROR.Print("hocr2pdf executable not found. Please check installation (http://www.exactcode.com/site/open_source/exactimage/hocr2pdf/).")
	}
	path, err = exec.LookPath("pdftk")
	TRACE.Println("pdftk", path, err)
	if err != nil {
		r = r || true
		ERROR.Print("pdftk executable not found. Please check pdftk installation.")
	}
	return r
}

func CreatePDFUsingPDFTK(filename string, images []string) error {
	defer Un(Trace("CreatePDFUsingPDFTK", filename, images))
	argList := make([]string, 0)
	for i := 0; i < len(images); i++ {
		p := images[i]
		d, f := path.Split(p)
		argList = append(argList, d+"ocr/"+f+".pdf")
	}
	arglist := append(argList, "cat", "output", filename)
	for i, arg := range arglist {
		TRACE.Println(i, arg)
	}
	cmd := exec.Command("pdftk", arglist...)
	out, err := cmd.CombinedOutput()
	fmt.Println("pdftk", filename, "processed\n", err, "\n", string(out))
	return err
}

func CreateTextIndex(filename string, images []string) error {
	defer Un(Trace("CreateTextIndex", filename, images))
	out, err := os.Create(filename + ".txt")
	defer out.Close()
	for i := 0; i < len(images); i++ {
		p := images[i]
		d, f := path.Split(p)
		in, err := os.Open(d + "ocr/" + f + ".html.txt")
		if err != nil {
			continue
		}
		_, err = io.Copy(out, in)
		in.Close()
	}
	fmt.Println(filename+".txt", "processed\n", err)
	return err
}

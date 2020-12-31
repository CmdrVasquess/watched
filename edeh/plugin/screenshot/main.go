package main

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/CmdrVasquess/watched/edeh/plugin"
	"github.com/CmdrVasquess/watched/examples/screenshot"
)

var (
	scrns = screenshot.Screenshot{
		EDPicDir: screenshot.DefaultPicsDir(),
	}
	fAspect string
)

func setAspect() {
	if fAspect == "" {
		scrns.Aspect = 0
		return
	}
	var err error
	scrns.Aspect, err = strconv.ParseFloat(fAspect, 64)
	if err != nil {
		sep := strings.IndexByte(fAspect, ':')
		if sep < 0 {
			log.Fatalf("Invalid aspect '%s'", fAspect)
		}
		w, err := strconv.ParseFloat(fAspect[:sep], 64)
		if err != nil {
			log.Fatalf("Invalid aspect width '%s': %s", fAspect[:sep], err)
		}
		h, err := strconv.ParseFloat(fAspect[sep+1:], 64)
		if err != nil {
			log.Fatalf("Invalid aspect height '%s': %s", fAspect[:sep], err)
		}
		scrns.Aspect = w / h
	}
}

func main() {
	flag.StringVar(&scrns.EDPicDir, "p", scrns.EDPicDir,
		"ED pictures directory")
	flag.StringVar(&scrns.OutRoot, "d", scrns.OutRoot,
		"Output directory for converted pictures")
	flag.IntVar(&scrns.JpegQuality, "q", 90,
		"Set JPEG output quality (1–100)")
	flag.StringVar(&fAspect, "a", "",
		"Set aspect ratio of output")
	flag.BoolVar(&scrns.SubstOrig, "s", false,
		"Put a converted substitue into ED ictures directory")
	flag.BoolVar(&scrns.RmOrig, "rm", false,
		"Remove original BMP after conversion")
	flag.Parse()
	setAspect()
	defer scrns.Close()
	plugin.RunRecv(&scrns, nil)
}

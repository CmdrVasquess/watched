package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/CmdrVasquess/watched/edeh/plugins"
	"github.com/CmdrVasquess/watched/examples/screenshot"
)

var (
	scrns = screenshot.Screenshot{
		EDPicDir: screenshot.DefaultPicsDir(),
		OutRoot:  ".",
		Lat:      math.NaN(), Lon: math.NaN(),
	}
	fAspect string
)

func flags() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&scrns.EDPicDir, "p", scrns.EDPicDir,
		"ED pictures directory")
	flag.StringVar(&scrns.OutRoot, "d", scrns.OutRoot,
		"Output directory for converted pictures")
	flag.BoolVar(&scrns.CmdrDir, "cmdr", scrns.CmdrDir,
		"Use commander-specific subdirectory in output directory")
	flag.IntVar(&scrns.JpegQuality, "q", 90,
		"Set JPEG output quality (1–100)")
	flag.StringVar(&fAspect, "a", "",
		"Set aspect ratio of output")
	flag.BoolVar(&scrns.SubstOrig, "s", false,
		"Put a converted substitue into ED pictures directory")
	flag.BoolVar(&scrns.RmOrig, "rm", false,
		"Remove original BMP after conversion")
	flag.BoolVar(&scrns.AddTags, "tags", false,
		"Add EXIF tags to JPEG output")
	flag.Parse()
}

func main() {
	flags()
	log.Println("start edeh plugin 'screenshot'")
	setAspect()
	defer scrns.Close()
	err := plugins.RunRecv(&scrns, nil, slog.Default())
	if err != nil {
		log.Fatal(err)
	}
}

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

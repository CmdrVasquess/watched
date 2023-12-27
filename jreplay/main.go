package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var doTee = flag.Bool("vv", false,
	"Show copied lines on stdout")
var progress = flag.Bool("v", false,
	"Show '.' per written event")
var targetDir = flag.String("j", ".",
	"Target directory to replay to")
var pause = flag.Duration("p", 0,
	"Pause between events")
var force = flag.Bool("f", false,
	"Force overwriting in target directory")
var interactive = flag.Bool("i", false,
	"Interactive, i.e. must press enter to send an event")
var tShift = flag.String("t", "",
	"Patch event times")
var verbosity = 0
var dt time.Duration

func open(filename string) (io.ReadCloser, error) {
	is, err := os.Open(filename)
	if err != nil {
		return is, err
	}
	switch {
	case strings.HasSuffix(filename, ".gz"):
		return gzip.NewReader(is)
		//	case strings.HasSuffix(filename, ".bz2"):
		//		return bzip2.NewReader(is), nil
	default:
		return os.Open(filename)
	}
}

const tsFmt = "2006-01-02T15:04:05Z"

var tsLabel = []byte(`"timestamp":"`)

func locateTs(line []byte) (start, end int) {
	start = bytes.Index(line, tsLabel)
	if start < 0 {
		return -1, -1
	}
	start += len(tsLabel)
	end = bytes.IndexByte(line[start:], '"')
	end += start
	return start, end
}

func readTs(line []byte, start, end int) time.Time {
	tss := string(line[start:end])
	res, err := time.Parse(tsFmt, tss)
	if err != nil {
		panic(err)
	}
	return res
}

func patchTs(line []byte, start int, t time.Time) {
	ts := t.Format(tsFmt)
	copy(line[start:], ts)
}

func shiftTime(line []byte) {
	if *tShift == "" {
		return
	}
	ts, te := locateTs(line)
	t := readTs(line, ts, te)
	if dt == 0 {
		var tt time.Time
		if *tShift == "now" {
			tt = time.Now()
		} else {
			var err error
			tt, err = time.Parse(tsFmt, *tShift)
			if err != nil {
				panic(err)
			}
		}
		dt = tt.Sub(t)
	}
	t = t.Add(dt)
	patchTs(line, ts, t)
}

func replay(sfNm string, tDir string) {
	tfNm := filepath.Join(tDir, filepath.Base(sfNm))
	if _, err := os.Stat(tfNm); !os.IsNotExist(err) {
		if *force {
			if err := os.Remove(tfNm); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Printf("skip %s, target exists: %s", sfNm, tfNm)
			return
		}
	}
	sf, err := open(sfNm) //os.Open(sfNm)
	if err != nil {
		log.Printf("source: %s", err)
		return
	}
	defer sf.Close()
	tf, err := os.Create(tfNm)
	if err != nil {
		log.Printf("target: %s", err)
		return
	}
	scn := bufio.NewScanner(sf)
	lnCount := 0
	for scn.Scan() {
		line := scn.Bytes()
		shiftTime(line)
		if *interactive {
			if len(line) > 78 {
				fmt.Printf("%s…\n", string(line[:78]))
			} else {
				fmt.Println(string(line))
			}
			fmt.Print("Press enter to send next event:")
			rd := bufio.NewReader(os.Stdin)
			rd.ReadLine()
		}
		tf.Write(line)
		fmt.Fprintln(tf)
		lnCount++
		if !*interactive {
			switch verbosity {
			case 1:
				fmt.Print(".")
			case 2:
				fmt.Println(string(line))
			}
			if *pause > 0 {
				time.Sleep(*pause)
			}
		}
	}
	if verbosity == 1 {
		fmt.Println()
	}
	fmt.Printf("wrote %d lines\n", lnCount)
}

func main() {
	flag.Parse()
	switch {
	case *doTee:
		verbosity = 2
	case *progress:
		verbosity = 1
	}
	*targetDir = filepath.Clean(*targetDir)
	for _, jfn := range flag.Args() {
		fmt.Printf("replay: %s → %s\n", jfn, *targetDir)
		replay(jfn, *targetDir)
		fmt.Printf("done: %s\n", jfn)
	}
}

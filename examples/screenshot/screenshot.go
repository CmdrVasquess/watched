package screenshot

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.fractalqb.de/fractalqb/eloc"
	"git.fractalqb.de/fractalqb/eloc/must"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/edj"
	"github.com/CmdrVasquess/watched/eds"
)

type Screenshot struct {
	EDPicDir    string
	OutRoot     string
	CmdrDir     bool
	Aspect      float64
	JpegQuality int
	SubstOrig   bool
	RmOrig      bool
	AddTags     bool

	fid      string
	cmdr     string
	Lat, Lon float64
	Alt      int
}

func (scrns *Screenshot) OutDir() string {
	path := scrns.EDPicDir
	if scrns.OutRoot != "" {
		path = scrns.OutRoot
	}
	if scrns.CmdrDir && scrns.cmdr != "" {
		cmdr := strings.ReplaceAll(scrns.cmdr, ".", "")
		cmdr = fielnameCleaner.Replace(cmdr)
		path = filepath.Join(path, cmdr)
	}
	return path
}

var fielnameCleaner = strings.NewReplacer(
	"/", "_",
	"\\", "_",
	":", "_",
	";", "_",
	" ", "_",
	"'", "",
	"\"", "",
	"\t", "_",
	"\r", "",
	"\n", "_",
)

func (scrns *Screenshot) OutFilePat(t time.Time, sys, body string) string {
	const sep = "."
	base := t.Format("060102150405.%d")
	if sys != "" {
		base += sep + fielnameCleaner.Replace(sys)
	}
	if body != "" && body != sys {
		if strings.HasPrefix(body, sys) {
			tmp := body[len(sys):]
			tmp = strings.TrimSpace(tmp)
			tmp = fielnameCleaner.Replace(tmp)
			base += sep + tmp
		} else {
			base += sep + fielnameCleaner.Replace(body)
		}
	}
	base += ".jpg"
	return base
}

func outFileIn(dir, pat string) string {
	i := 0
	for {
		res := filepath.Join(dir, fmt.Sprintf(pat, i))
		if _, err := os.Stat(res); os.IsNotExist(err) {
			return res
		}
		i++
	}
}

func (scrns *Screenshot) OnJournalEvent(e watched.JounalEvent) (err error) {
	evt, err := e.Event.PeekEvent()
	if err != nil {
		return err
	}
	if hdl := jehdl[evt]; hdl != nil {
		defer eloc.RecoverAs(&err)
		hdl(scrns, e.Event)
	}
	return nil
}

func (scrns *Screenshot) OnStatusEvent(e watched.StatusEvent) error {
	evt, err := e.Event.PeekEvent()
	if err != nil {
		return err
	}
	switch evt {
	case eds.StatusTag:
		return scrns.sStatus(e.Event)
	}
	return nil
}

func (scrns *Screenshot) Close() error { return nil }

var jehdl = map[string]func(*Screenshot, []byte){
	"Commander":  (*Screenshot).jeCommander,
	"LoadGame":   (*Screenshot).jeLoadGame,
	"Shutdown":   (*Screenshot).jeShutdown,
	"Screenshot": (*Screenshot).jeScreenshot,
}

func (scrns *Screenshot) jeCommander(e []byte) {
	var evt edj.Commander
	must.Do(json.Unmarshal(e, &evt))
	if evt.FID != scrns.fid {
		log.Printf("switch to commander %s", evt.Name)
	}
	scrns.fid = evt.FID
	scrns.cmdr = evt.Name
}

func (scrns *Screenshot) jeLoadGame(e []byte) {
	var evt edj.LoadGame
	must.Do(json.Unmarshal(e, &evt))
	if evt.FID != scrns.fid {
		log.Printf("switch to commander %s", evt.Commander)
	}
	scrns.fid = evt.FID
	scrns.cmdr = evt.Commander
}

func (scrns *Screenshot) jeShutdown(e []byte) {
	scrns.fid = ""
	scrns.cmdr = ""
	log.Println("switch to no commander")
}

func (scrns *Screenshot) jeScreenshot(e []byte) {
	var evt edj.Screenshot
	must.Do(json.Unmarshal(e, &evt))
	fnm := evt.FilenameToOS()
	fnm = filepath.Base(fnm)
	if ext := strings.ToLower(filepath.Ext(fnm)); ext != ".bmp" {
		panic(eloc.Errorf("illegal screenshot file extension: '%s'", ext))
	}
	input := filepath.Join(scrns.EDPicDir, fnm)
	img := must.Ret(readBMP(input))
	fnpat := scrns.OutFilePat(evt.Timestamp, evt.System, evt.Body)

	if scrns.SubstOrig {
		subst := outFileIn(filepath.Dir(input), fnpat)
		log.Printf("subst: %s", subst)
		if scrns.AddTags {
			must.Do(writeJPEGFile(subst, img, scrns.JpegQuality, &imageTags{
				cmdr:       scrns.cmdr,
				Screenshot: &evt,
				lat:        scrns.Lat,
				lon:        scrns.Lon,
				alt:        scrns.Alt,
			}))
		} else {
			must.Do(writeJPEGFile(subst, img, scrns.JpegQuality, nil))
		}
	}

	if scrns.Aspect > 0 {
		img = adjustAspect(img, scrns.Aspect)
	}

	outdir := scrns.OutDir()
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		log.Printf("MkDirAll %s", outdir)
		must.Do(os.MkdirAll(outdir, 0777))
	}
	output := outFileIn(outdir, fnpat)
	log.Printf("convert to: %s", output)
	if scrns.AddTags {
		must.Do(writeJPEGFile(output, img, scrns.JpegQuality, &imageTags{
			cmdr:       scrns.cmdr,
			Screenshot: &evt,
			lat:        scrns.Lat,
			lon:        scrns.Lon,
			alt:        scrns.Alt,
		}))
	} else {
		must.Do(writeJPEGFile(output, img, scrns.JpegQuality, nil))
	}

	if scrns.RmOrig {
		err := os.Remove(input)
		if err != nil {
			log.Println(err)
		}
	}
}

func (scrns *Screenshot) sStatus(e []byte) error {
	os.Stdout.Write(e)
	fmt.Println()
	var evt eds.Status
	if err := json.Unmarshal(e, &evt); err != nil {
		return err
	}
	if evt.AnyFlag(eds.StatusHasLatLon) {
		scrns.Lat = evt.Latitude
		scrns.Lon = evt.Longitude
		scrns.Alt = evt.Altitude
	} else {
		scrns.Lat = math.NaN()
		scrns.Lon = math.NaN()
		scrns.Alt = 0
	}
	return nil
}

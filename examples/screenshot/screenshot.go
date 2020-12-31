package screenshot

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/watched"
)

type Screenshot struct {
	EDPicDir    string
	OutRoot     string
	Aspect      float64
	JpegQuality int
	SubstOrig   bool
	RmOrig      bool
	FID         string
	Cmdr        string
}

func (scrns *Screenshot) ScrShot(name string) (path string) {
	return filepath.Join(scrns.EDPicDir, name)
}

func (scrns *Screenshot) OutDir(t time.Time, sys, body string) string {
	path := scrns.EDPicDir
	if scrns.OutRoot != "" {
		path = scrns.OutRoot
	}
	if scrns.Cmdr != "" {
		cmdr := strings.ReplaceAll(scrns.Cmdr, " ", "_")
		path = filepath.Join(path, cmdr)
	}
	return path
}

func (scrns *Screenshot) OutFilePat(t time.Time, sys, body string) string {
	base := t.Format("060102150405.%d")
	if sys != "" {
		sys = strings.ReplaceAll(sys, " ", "_")
		base += "-" + sys
	}
	if body != "" {
		body = strings.ReplaceAll(body, " ", "_")
		base += "-" + body
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

func (scrns *Screenshot) Journal(e watched.JounalEvent) (err error) {
	evt, err := e.Event.PeekEvent()
	if err != nil {
		return err
	}
	if hdl := jehdl[evt]; hdl != nil {
		bare := make(ggja.BareObj)
		if err = json.Unmarshal(e.Event, &bare); err != nil {
			return err
		}
		defer func() {
			if p := recover(); p != nil {
				switch x := p.(type) {
				case error:
					err = x
				default:
					err = fmt.Errorf("panic: %+v", p)
				}
			}
		}()
		hdl(scrns, ggja.Obj{Bare: bare})
	}
	return nil
}

func (scrns *Screenshot) Status(e watched.StatusEvent) error { return nil }

func (scrns *Screenshot) Close() error { return nil }

var jehdl = map[string]func(*Screenshot, ggja.Obj){
	"Commander":  jeCommander,
	"LoadGame":   jeLoadGame,
	"Shutdown":   jeShutdown,
	"Screenshot": jeScreenshot,
}

func jeCommander(scrns *Screenshot, e ggja.Obj) {
	scrns.FID = e.MStr("FID")
	scrns.Cmdr = e.MStr("Name")
}

func jeLoadGame(scrns *Screenshot, e ggja.Obj) {
	scrns.FID = e.MStr("FID")
	scrns.Cmdr = e.MStr("Commander")
}

func jeShutdown(scrns *Screenshot, e ggja.Obj) {
	scrns.FID = ""
	scrns.Cmdr = ""
}

func jeScreenshot(scrns *Screenshot, e ggja.Obj) {
	fnm := e.MStr("Filename")
	path := strings.Split(fnm, "\\")
	if len(path) < 1 {
		panic(fmt.Errorf("invalid screenshot filename '%s'", fnm))
	}
	input := scrns.ScrShot(path[len(path)-1])
	ts := e.MTime("timestamp")
	sys := e.Str("System", "")
	body := e.Str("Body", "")
	fpat := scrns.OutFilePat(ts, sys, body)
	img, err := imgio.Open(input)
	if err != nil {
		panic(err)
	}
	if scrns.SubstOrig {
		subst := outFileIn(filepath.Dir(input), fpat)
		log.Printf("subst: %s", subst)
		err = imgio.Save(subst, img, imgio.JPEGEncoder(scrns.JpegQuality))
		if err != nil {
			panic(err)
		}
	}
	if scrns.Aspect > 0 {
		img = adjustAspect(img, scrns.Aspect)
	}
	outdir := scrns.OutDir(ts, sys, body)
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		log.Printf("MkDirAll %s", outdir)
		os.MkdirAll(outdir, 0777)
	}
	output := outFileIn(outdir, fpat)
	log.Printf("convert: %s", output)
	err = imgio.Save(output, img, imgio.JPEGEncoder(scrns.JpegQuality))
	if err != nil {
		panic(err)
	}
	if scrns.RmOrig {
		err = os.Remove(input)
		if err != nil {
			log.Println(err)
		}
	}
}

func adjustAspect(img image.Image, outAspect float64) image.Image {
	imgAspect := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
	if math.Abs(imgAspect-outAspect) > 0.001 {
		if imgAspect > outAspect {
			outWidth := int(outAspect * float64(img.Bounds().Dy()))
			var rect image.Rectangle
			rect = cropWidth(img, outWidth)
			img = transform.Crop(img, rect)
		} else if imgAspect < outAspect {
			outHeight := int(float64(img.Bounds().Dx()) / outAspect)
			var rect image.Rectangle
			rect = cropHeight(img, outHeight)
			img = transform.Crop(img, rect)
		}
	}
	return img
}

func cropWidth(img image.Image, w int) image.Rectangle {
	dw2 := (img.Bounds().Dx() - w) / 2
	res := image.Rectangle{
		Min: image.Point{
			X: img.Bounds().Min.X + dw2,
			Y: img.Bounds().Min.Y,
		},
		Max: image.Point{
			X: img.Bounds().Max.X - dw2,
			Y: img.Bounds().Max.Y,
		},
	}
	return res
}

func cropHeight(img image.Image, h int) image.Rectangle {
	dh2 := (img.Bounds().Dy() - h) / 2
	res := image.Rectangle{
		Min: image.Point{
			X: img.Bounds().Min.X,
			Y: img.Bounds().Min.Y + dh2,
		},
		Max: image.Point{
			X: img.Bounds().Max.X,
			Y: img.Bounds().Max.Y - dh2,
		},
	}
	return res
}

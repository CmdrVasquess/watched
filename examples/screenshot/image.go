package screenshot

// For EXIF content: https://github.com/dsoprea/go-exif
//                   github.com/sfomuseum/go-exif-update
import (
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.fractalqb.de/fractalqb/eloc"
	"git.fractalqb.de/fractalqb/eloc/must"
	"github.com/CmdrVasquess/watched/edj"
	"github.com/anthonynsimon/bild/transform"
	exif "github.com/sfomuseum/go-exif-update"
	"golang.org/x/image/bmp"
)

func readBMP(file string) (image.Image, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return bmp.Decode(r)
}

func adjustAspect(img image.Image, outAspect float64) image.Image {
	imgAspect := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
	if math.Abs(imgAspect-outAspect) > 0.001 {
		if imgAspect > outAspect {
			outWidth := int(outAspect * float64(img.Bounds().Dy()))
			rect := cropWidth(img, outWidth)
			img = transform.Crop(img, rect)
		} else if imgAspect < outAspect {
			outHeight := int(float64(img.Bounds().Dx()) / outAspect)
			rect := cropHeight(img, outHeight)
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

type imageTags struct {
	cmdr string
	*edj.Screenshot
	lat, lon float64
	alt      int
}

func gameTime(t time.Time) time.Time {
	t = t.UTC()
	Y, M, D := t.Date()
	h, m, s := t.Clock()
	return time.Date(Y+1286, M, D, h, m, s, 0, time.UTC)
}

func writeJPEGFile(name string, img image.Image, q int, tags *imageTags) (err error) {
	wr, err := os.Create(name)
	if err != nil {
		return eloc.At(err)
	}
	if err := jpeg.Encode(wr, img, &jpeg.Options{Quality: q}); err != nil {
		wr.Close()
		return eloc.At(err)
	}
	if err := wr.Close(); err != nil {
		return eloc.At(err)
	}

	if tags != nil {
		defer eloc.RecoverAs(&err)
		et := make(map[string]any)
		prepare := func(tag, value string) {
			et[tag] = must.Ret(exif.PrepareTag(tag, value))
		}
		prepare("Software", "EDEH screenshot")
		prepare("DateTime", gameTime(tags.Timestamp).Format(time.RFC1123))
		if tags.cmdr != "" {
			prepare("Artist", tags.cmdr)
		}
		var desc strings.Builder
		fmt.Fprintf(&desc, "System: %s", tags.System)
		if body := tags.Body; body != "" {
			if body != tags.System && strings.HasPrefix(body, tags.System) {
				body = body[len(tags.System):]
				body = strings.TrimSpace(body)
			}
			fmt.Fprintf(&desc, "; Body: %s", body)
		}
		prepare("ImageDescription", desc.String())
		if !math.IsNaN(tags.lat) { // expect consistency, doesn't check the rest
			must.Do(exif.AppendGPSPropertiesWithLatitudeAndLongitude(et, tags.lat, tags.lon))
			// TODO how to convert GPS altitude
			// switch {
			// case tags.alt > 0:
			// 	prepare("GPSAltitude", strconv.Itoa(tags.alt))
			// 	prepare("GPSAltitude", "0")
			// case tags.alt < 0:
			// 	prepare("GPSAltitude", strconv.Itoa(-tags.alt))
			// 	prepare("GPSAltitude", "1")
			// }
		}

		r := must.Ret(os.Open(name))
		defer r.Close()
		dir, tmpat := filepath.Split(name)
		if dir == "" {
			dir = "."
		}
		if ext := filepath.Ext(tmpat); ext == "" {
			tmpat += "-"
		} else {
			tmpat += "-*" + ext
		}
		w := must.Ret(os.CreateTemp(dir, tmpat))
		defer w.Close()
		must.Do(exif.UpdateExif(r, w, et))
		must.Do(w.Close())
		must.Do(r.Close())
		return os.Rename(w.Name(), name)
	}

	return nil
}

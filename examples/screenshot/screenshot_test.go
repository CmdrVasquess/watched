package screenshot

import (
	"fmt"
	"image"
	"image/jpeg"
	"testing"
	"time"

	"github.com/CmdrVasquess/watched/edj"
)

func Example_outFileIn() {
	var scrns Screenshot
	pat := scrns.OutFilePat(time.Time{}, "SYS", "BODY")
	file := outFileIn(".", pat)
	fmt.Println(file)
	// Output:
	// 010101000000.0.SYS.BODY.jpg
}

func TestWriteJPEGFile(t *testing.T) {
	img := image.NewRGBA(image.Rectangle{
		image.Point{},
		image.Point{16, 16},
	})
	err := writeJPEGFile(t.Name()+".jpg", img, jpeg.DefaultQuality, &imageTags{
		Screenshot: &edj.Screenshot{
			Event:  edj.Event{Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.Local)},
			System: "Sol",
			Body:   "Sol Earth",
		},
		cmdr: "J. Jameson",
	})
	if err != nil {
		t.Fatal(err)
	}
}

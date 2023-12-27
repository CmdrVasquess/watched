package screenshot

import (
	"fmt"
	"image"
	"image/jpeg"
	"testing"
	"time"
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
	err := writeJPEGFile(t.Name()+".jpg", img, jpeg.DefaultQuality, &ImageTags{
		Time: time.Date(3300, 1, 1, 12, 0, 0, 0, time.UTC),
		CMDR: "J. Jameson",
	})
	if err != nil {
		t.Fatal(err)
	}
}

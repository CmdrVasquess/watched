package screenshot

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
)

func ExampleJFIFScanner() {
	img := image.NewRGBA(image.Rectangle{
		image.Point{},
		image.Point{16, 16},
	})
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		fmt.Println(err)
	}
	scn := NewJFIFScanner(&buf)
	for scn.Scan() {
		if scn.Err != nil {
			fmt.Println(scn.Err)
			break
		}
		if scn.Tag.Segment() {
			fmt.Printf("%s %d", scn.Tag, scn.Size)
		} else {
			fmt.Print(scn.Tag.String())
		}
		n, err := scn.Consume(scn.Segment())
		fmt.Printf(" consumed %d bytes\n", n)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	// Output:
	// SOI (D8) consumed 0 bytes
	// DQT (DB) 132 consumed 130 bytes
	// SOF0 (C0) 17 consumed 15 bytes
	// SOF4 (C4) 418 consumed 416 bytes
	// SOS (DA) consumed 18 bytes
}

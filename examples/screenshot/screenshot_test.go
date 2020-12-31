package screenshot

import (
	"fmt"
	"time"
)

func ExampleOutFileIn() {
	var scrns Screenshot
	pat := scrns.OutFilePat(time.Time{}, "SYS", "BODY")
	file := outFileIn(".", pat)
	fmt.Println(file)
	// Output:
	// 010101000000.0-SYS-BODY.jpg
}

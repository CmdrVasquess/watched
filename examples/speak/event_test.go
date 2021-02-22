package speak

import (
	"encoding/json"
	"fmt"
	"log"

	"git.fractalqb.de/fractalqb/ggja"
)

func must(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func ExampleEvent_Text() {
	var evt Event
	evt.Speak.Format = "From \"%s\": %s"
	evt.Speak.Args = ggja.BareArr{
		"From",
		ggja.BareArr{"Message_Localised", "Message"},
	}
	jevt := make(ggja.BareObj)
	must(json.Unmarshal([]byte(`{
		"From": "John Doe",
		"Message": "RoC Commander, o7!"
	}`), &jevt))
	txt := evt.Text(ggja.Obj{Bare: jevt})
	fmt.Println(txt)
	// Output:
	// _
}

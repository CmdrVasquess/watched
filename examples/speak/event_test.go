package speak

import (
	"encoding/json"
	"fmt"
)

func must(err error) {
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
}

func Example_defaultEvent() {
	var evt = defaultEvent{}
	evt.Speak.Template = `From "{{.From}}": {{.Message}}`
	must(evt.Speak.configure("test"))
	jevt := make(map[string]any)
	must(json.Unmarshal([]byte(`{
		"From": "John Doe",
		"Message": "RoC Commander, o7!"
	}`), &jevt))
	txt, args := evt.message(nil, jevt)
	fmt.Println(txt, args)
	// Output:
	// From "John Doe": RoC Commander, o7! []
}

package speak

import (
	"fmt"
	"log"

	"git.fractalqb.de/fractalqb/ggja"
)

type Event struct {
	If    interface{} `json:",omitempty"`
	Flags []string    `json:",omitempty"`
	Speak struct {
		Format string
		Args   ggja.BareArr
	}
}

func (evt *Event) Check(jevt ggja.Obj) bool {
	return true
}

func (evt *Event) Text(jevt ggja.Obj) string {
	var parts []interface{}
	for _, arg := range evt.Speak.Args {
		switch av := arg.(type) {
		case string:
			parts = append(parts, jevt.Str(av, ""))
		case ggja.BareArr:
			for _, path := range av {
				p, err := ggja.Get(jevt, path)
				if err == nil {
					parts = append(parts, p)
					break
				}
			}
		default:
			log.Printf("cannot resolve text argument: '%+v'", arg)
		}
	}
	if len(parts) == 0 {
		return evt.Speak.Format
	}
	return fmt.Sprintf(evt.Speak.Format, parts...)
}

package speak

import (
	"fmt"
	"log"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/mitchellh/mapstructure"
)

type evtHandler interface {
	Configure(evt any) error
	Message(stat *status, evt ggja.Obj) (txt string, flags []string)
}

var handlers = map[string]evtHandler{
	"ReceiveText":         new(receiveText),
	"FSSSignalDiscovered": new(fssSignalDiscovered),
}

type EventMsg struct {
	Format string
	Args   []any
}

func (em *EventMsg) Text(jevt ggja.Obj) string {
	var parts []interface{}
	for _, arg := range em.Args {
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
		return em.Format
	}
	return fmt.Sprintf(em.Format, parts...)
}

// DefaultEvent will be used on every configured Events element that has no
// pre-registered handler
type DefaultEvent struct {
	Flags []string `json:",omitempty"`
	Speak EventMsg
}

func (evt *DefaultEvent) Configure(cfg any) error {
	return mapstructure.Decode(cfg, evt)
}

func (evt *DefaultEvent) Message(stat *status, jevt ggja.Obj) (string, []string) {
	return evt.Speak.Text(jevt), evt.Flags
}

type receiveText struct {
	DefaultEvent
	Channels map[string]DefaultEvent
}

func (evt *receiveText) Configure(cfg any) error {
	return mapstructure.Decode(cfg, evt)
}

func (evt *receiveText) Message(stat *status, jevt ggja.Obj) (string, []string) {
	chn := jevt.Str("Channel", "")
	if chnCfg, ok := evt.Channels[chn]; ok {
		return chnCfg.Speak.Text(jevt), chnCfg.Flags
	}
	return evt.Speak.Text(jevt), evt.Flags
}

type fssSignalDiscovered struct {
	DefaultEvent
}

func (evt *fssSignalDiscovered) Message(stat *status, jevt ggja.Obj) (string, []string) {
	if stat.population > 0 {
		return "", nil
	}
	if stn := jevt.Bool("IsStation", false); stn {
		return "", nil
	}
	return evt.Speak.Text(jevt), evt.Flags
}

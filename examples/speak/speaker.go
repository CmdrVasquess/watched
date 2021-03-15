package speak

import (
	"encoding/json"
	"log"
	"os/exec"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/watched"
)

type Speaker struct {
	Exe     string
	Args    []string `json:",omitempty"`
	Verbose bool     `json:",omitempty"`
	Events  map[string]*Event
}

func (spk *Speaker) OnJournalEvent(e watched.JounalEvent) (err error) {
	event := make(ggja.BareObj)
	if err = json.Unmarshal(e.Event, &event); err != nil {
		return err
	}
	jevt := ggja.Obj{Bare: event}
	ename := jevt.MStr("event")
	evt := spk.Events[ename]
	if evt != nil && evt.Check(jevt) {
		text := evt.Text(jevt)
		args := append(evt.Flags, text)
		cmd := exec.Command(spk.Exe, args...)
		if spk.Verbose {
			log.Printf("event %s: '%s'", ename, text)
		}
		err = cmd.Run()
	}
	return err
}

func (spk *Speaker) OnStatusEvent(e watched.StatusEvent) error { return nil }

func (spk *Speaker) Close() error { return nil }

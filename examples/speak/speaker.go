package speak

import (
	"encoding/json"
	"log"
	"os/exec"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/watched"
)

type Speaker struct {
	TTSExe  string
	Args    []string `json:",omitempty"`
	Verbose bool     `json:",omitempty"`
	Events  map[string]any
	stat    status
}

type status struct {
	population int
}

func (spk *Speaker) Configure() error {
	for e, cfg := range spk.Events {
		h := handlers[e]
		if h == nil {
			dh := new(DefaultEvent)
			if err := dh.Configure(cfg); err != nil {
				log.Fatal(err)
			}
			handlers[e] = dh
			h = dh
		} else {
			if err := h.Configure(cfg); err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("cfg '%s'-handler: %+v", e, h)
	}
	return nil
}

func (spk *Speaker) OnJournalEvent(e watched.JounalEvent) (err error) {
	event := make(ggja.BareObj)
	if err = json.Unmarshal(e.Event, &event); err != nil {
		return err
	}
	jevt := ggja.Obj{Bare: event}
	ename := jevt.MStr("event")
	if ename == "FSDJump" {
		spk.stat.population = jevt.Int("Population", 0)
		return nil
	}
	eh := handlers[ename]
	if eh == nil {
		return nil
	}
	text, flags := eh.Message(&spk.stat, jevt)
	if text == "" {
		return nil
	}
	args := append(flags, text)
	cmd := exec.Command(spk.TTSExe, args...)
	if spk.Verbose {
		log.Printf("event %s: '%s' (%v)", ename, text, args)
	}
	return cmd.Run()
}

func (spk *Speaker) OnStatusEvent(e watched.StatusEvent) error { return nil }

func (spk *Speaker) Close() error { return nil }

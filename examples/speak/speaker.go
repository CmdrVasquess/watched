package speak

import (
	"encoding/json"
	"os/exec"
	"strings"

	"git.fractalqb.de/fractalqb/daq"
	"github.com/CmdrVasquess/watched"
)

type Speaker struct {
	TTSExe  string         `yaml:"TTSExe"`
	Args    []string       `json:",omitempty" yaml:"Args"`
	Verbose bool           `json:",omitempty" yaml:"Verbose"`
	Events  map[string]any `yaml:"Events"`
	stat    status
}

type status struct {
	population int
}

func (spk *Speaker) Configure() error {
	for e, cfg := range spk.Events {
		h := handlers[e]
		if h == nil {
			dh := new(defaultEvent)
			if err := dh.configure(e, cfg); err != nil {
				logFatal(err)
			}
			handlers[e] = dh
			h = dh
		} else {
			if err := h.configure(e, cfg); err != nil {
				logFatal(err)
			}
		}
	}
	return nil
}

func (spk *Speaker) OnJournalEvent(e watched.JounalEvent) error {
	event := make(map[string]any)
	if err := json.Unmarshal(e.Event, &event); err != nil {
		return err
	}
	jevt := daq.Map(event)
	ename, err := daq.AsString(jevt.Get("event"))
	if err != nil {
		return err
	}
	if ename == "FSDJump" {
		spk.stat.population = daq.Val(daq.AsInt, 0)(jevt.Get("Population"))
		return nil
	}
	eh := handlers[ename]
	if eh == nil {
		return nil
	}
	text, flags := eh.message(&spk.stat, jevt)
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	args := append(flags, text)
	cmd := exec.Command(spk.TTSExe, args...)
	if spk.Verbose {
		log.Info("`event` `text` with `flags`",
			`event`, ename,
			`text`, text,
			`flags`, flags,
		)
	}
	return cmd.Run()
}

func (spk *Speaker) OnStatusEvent(e watched.StatusEvent) error { return nil }

func (spk *Speaker) Close() error { return nil }

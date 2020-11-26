package main

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/watched"
)

type Speaker struct {
	Exe string
}

func (spk *Speaker) Journal(e watched.JounalEvent) (err error) {
	event := make(ggja.BareObj)
	if err = json.Unmarshal(e.Event, &event); err != nil {
		return err
	}
	evt := ggja.Obj{Bare: event}
	switch evt.MStr("event") {
	case "ReceiveText":
		text := fmt.Sprintf("From \"%s\": %s", evt.MStr("From"), evt.MStr("Message"))
		cmd := exec.Command(spk.Exe, text)
		err = cmd.Run()
	}
	return err
}

func (spk *Speaker) Status(e watched.StatusEvent) error { return nil }

func (spk *Speaker) Close() error { return nil }

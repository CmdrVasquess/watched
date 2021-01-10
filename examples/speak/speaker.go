package speak

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/watched"
)

type Speaker struct {
	Exe     string
	Verbose bool
}

func (spk *Speaker) OnJournalEvent(e watched.JounalEvent) (err error) {
	event := make(ggja.BareObj)
	if err = json.Unmarshal(e.Event, &event); err != nil {
		return err
	}
	evt := ggja.Obj{Bare: event}
	switch evt.MStr("event") {
	case "ReceiveText":
		from := evt.Str("From_Localised", "")
		if from == "" {
			from = evt.MStr("From")
		}
		msg := evt.Str("Message_Localised", "")
		if msg == "" {
			msg = evt.MStr("Message")
		}
		text := fmt.Sprintf("From \"%s\": %s", from, msg)
		//mchn := evt.MStr("Channel") // squadron npc local player starsystem
		if spk.Verbose {
			log.Println(text)
		}
		cmd := exec.Command(spk.Exe, text)
		err = cmd.Run()
	}
	return err
}

func (spk *Speaker) OnStatusEvent(e watched.StatusEvent) error { return nil }

func (spk *Speaker) Close() error { return nil }

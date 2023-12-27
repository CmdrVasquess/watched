package watched

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.fractalqb.de/fractalqb/qblog"
	"git.fractalqb.de/fractalqb/testerr"
)

type testEvtCollect struct {
	jes []JounalEvent
	ses []StatusEvent
}

func (ec *testEvtCollect) OnJournalEvent(e JounalEvent) error {
	ec.jes = append(ec.jes, e.Clone())
	return nil
}

func (ec *testEvtCollect) OnStatusEvent(e StatusEvent) error {
	ec.ses = append(ec.ses, e.Clone())
	return nil
}

func (ec *testEvtCollect) Clear() {
	ec.jes = nil
	ec.ses = nil
}

func (ec *testEvtCollect) Close() error {
	ec.Clear()
	return nil
}

func (ec *testEvtCollect) JEq(idx int, file string, eno int, evt string) error {
	if idx >= len(ec.jes) {
		return fmt.Errorf("no journal event with index %d", idx)
	}
	e := ec.jes[idx]
	if !(e.File == file && e.EventNo == eno && string(e.Event) == evt) {
		return fmt.Errorf("journal: '%s':%d [%s]", e.File, e.EventNo, string(e.Event))
	}
	return nil
}

func (ec *testEvtCollect) SEq(idx int, t StatusType, eno int, evt string) error {
	if idx >= len(ec.ses) {
		return fmt.Errorf("no status event with index %d", idx)
	}
	e := ec.ses[idx]
	if !(e.Type == t && string(e.Event) == evt) {
		return fmt.Errorf("status: %s [%s]", e.Type, string(e.Event))
	}
	return nil
}

func TestEvents(t *testing.T) {
	if testing.Verbose() {
		qblog.DefaultConfig.SetLevel(qblog.LevelTrace)
	}
	const pause = 500 * time.Millisecond

	dir := t.Name() + ".d"
	testerr.Do(os.RemoveAll(dir)).ShallBeNil(t)
	testerr.Do(os.Mkdir(dir, 0700)).ShallBeNil(t)

	var ec testEvtCollect
	evts := NewJournal(dir, &ec, nil)

	go func() {
		time.Sleep(pause)
		const file = "Journal.log"
		jf := testerr.Ret(os.Create(filepath.Join(dir, file))).ShallBeNil(t)
		defer jf.Close()

		t.Run("journal with newline", func(t *testing.T) {
			fmt.Fprintln(jf, "this is not an event")
			time.Sleep(pause)
			if err := ec.JEq(0, file, 1, "this is not an event"); err != nil {
				t.Errorf("journal event 1: %s", err)
			}
		})

		t.Run("journal without newline", func(t *testing.T) {
			fmt.Fprint(jf, "this is not an event")
			time.Sleep(pause)
			if err := ec.JEq(0, file, 1, "this is not an event"); err != nil {
				t.Errorf("journal event 1: %s", err)
			}
		})

		evts.Stop()
	}()

	testerr.Do(evts.Start()).ShallBeNil(t)
}

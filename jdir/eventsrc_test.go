package jdir

import (
	"testing"

	"github.com/CmdrVasquess/watched"
)

func TestIsNewJournalEvent(t *testing.T) {
	t.Run("init zero", func(t *testing.T) {
		var esrc Events
		if ok, err := esrc.checkNewJournalEvent(4711); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Error("zero EDEvents considers 1st 4711 to be old event")
		}
		if ok, err := esrc.checkNewJournalEvent(4711); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Error("zero EDEvents considers 2nd 4711 to be old event")
		}
	})
	t.Run("second before", func(t *testing.T) {
		var esrc Events
		esrc.lastSer = 4711 * watched.JESequenceMask
		for i := 0; i < watched.JESequenceMask+1; i++ {
			if ok, err := esrc.checkNewJournalEvent(4710); err != nil {
				t.Fatal(err)
			} else if !ok {
				t.Fatalf("%d repetition considered new", i+1)
			}
		}
	})
	t.Run("repeat 1st in second", func(t *testing.T) {
		var esrc Events
		esrc.lastSer, _ = new(watched.JEIDCounter).CountUnix(4711)
		if ok, err := esrc.checkNewJournalEvent(4711); err != nil {
			t.Fatal(err)
		} else if ok {
			t.Fatal("1st repetition is considered new")
		}
		if ok, err := esrc.checkNewJournalEvent(4711); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Fatal("2nd repetition is considered old")
		}
	})
}

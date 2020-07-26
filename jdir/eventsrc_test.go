package jdir

import (
	"testing"
)

func TestIsNewJournalEvent(t *testing.T) {
	t.Run("init zero", func(t *testing.T) {
		var esrc EDEvents
		if !esrc.checkNewJournalEvent(4711) {
			t.Error("zero EDEvents considers 1st 4711 to be old event")
		}
		if !esrc.checkNewJournalEvent(4711) {
			t.Error("zero EDEvents considers 2nd 4711 to be old event")
		}
	})
	t.Run("second before", func(t *testing.T) {
		var esrc EDEvents
		esrc.setLastJSerial(4711 * ljeSeqMax)
		for i := 0; i < ljeSeqMax+1; i++ {
			if esrc.checkNewJournalEvent(4710) {
				t.Fatalf("%d repetition considered new", i+1)
			}
		}
	})
	t.Run("repeat 1st in second", func(t *testing.T) {
		var esrc EDEvents
		esrc.checkNewJournalEvent(4711)
		esrc.setLastJSerial(esrc.LastJSerial())
		if esrc.checkNewJournalEvent(4711) {
			t.Fatal("1st repetition is considered new")
		}
		if !esrc.checkNewJournalEvent(4711) {
			t.Fatal("2nd repetition is considered old")
		}
	})
}

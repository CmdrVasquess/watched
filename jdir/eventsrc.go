package jdir

import (
	"io/ioutil"
	"time"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal"
)

type EDEvents struct {
	watched.EDEvents
	Stop   chan internal.StopEvent
	jdir   *JournalDir
	jq     chan watched.JounalEvent
	sq     chan watched.StatusEvent
	ljeSec int64 // last journal event UNIX seconds
	ljeSeq int   // seq withn ljeSec
	djeSeq int
}

func (ede *EDEvents) Start(withJournal string) {
	defer func() {
		close(ede.jq)
		close(ede.sq)
	}()
	ede.jdir.Watch(withJournal)
}

func (ede *EDEvents) LastJSerial() watched.JEventID {
	return ljeSeqMax*ede.ljeSec + int64(ede.ljeSeq)
}

func (ede *EDEvents) setLastJSerial(s watched.JEventID) {
	ede.ljeSec = s / ljeSeqMax
	ede.ljeSeq = int(s % ljeSeqMax)
	ede.djeSeq = -1
}

func (ede *EDEvents) onJournal(raw []byte) {
	t, err := watched.PeekTime(raw)
	if err != nil {
		log.Errore(err)
		return
	}
	if ede.checkNewJournalEvent(t.Unix()) {
		ede.jq <- watched.JounalEvent{
			Serial: ede.LastJSerial(),
			Event:  raw,
		}
	}
}

func (ede *EDEvents) onStat(event watched.StatusType, file string) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errore(err) // TODO be more descriptive
		return
	}
	ede.sq <- watched.StatusEvent{
		Type:  event,
		Event: raw,
	}
}

const ljeSeqMax = 1000

func (ede *EDEvents) checkNewJournalEvent(uxsec int64) bool {
	if uxsec < ede.ljeSec {
		return false
	} else if uxsec > ede.ljeSec {
		ede.ljeSec = uxsec
		ede.ljeSeq = 0
		ede.djeSeq = 0
		return true
	}
	ede.djeSeq++
	if ede.djeSeq <= ede.ljeSeq {
		return false
	}
	ede.ljeSeq = ede.djeSeq
	return true
}

type Options struct {
	PollWaitMin time.Duration
	PollWaitMax time.Duration
	JournalQLen int
	StatusQLen  int
	JSerial     int64
}

func NewEDEvents(dir string, opt *Options) *EDEvents {
	jdir := &JournalDir{
		Dir:  dir,
		Stop: make(chan internal.StopEvent),
	}
	res := &EDEvents{
		Stop: jdir.Stop,
		jdir: jdir,
	}
	jdir.PerJLine = res.onJournal
	jdir.OnStatChg = res.onStat
	if opt != nil {
		jdir.PollWaitMin = opt.PollWaitMin
		jdir.PollWaitMax = opt.PollWaitMax
		res.setLastJSerial(opt.JSerial)
		res.jq = make(chan watched.JounalEvent, opt.JournalQLen)
		res.sq = make(chan watched.StatusEvent, opt.StatusQLen)
	} else {
		res.jq = make(chan watched.JounalEvent)
		res.sq = make(chan watched.StatusEvent)
	}
	res.EDEvents.Journal = res.jq
	res.EDEvents.Status = res.sq
	return res
}

package jdir

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal"
)

type Events struct {
	Stop   chan internal.StopEvent
	recv   watched.EventRecv
	jdir   *JournalDir
	ljeSec int64 // last journal event UNIX seconds
	ljeSeq int   // seq withn ljeSec
	djeSeq int
}

func NewEvents(dir string, r watched.EventRecv, opt *Options) *Events {
	jdir := &JournalDir{Dir: dir}
	res := &Events{
		Stop: make(chan internal.StopEvent),
		recv: r,
		jdir: jdir,
	}
	jdir.PerJLine = res.onJournal
	jdir.OnStatChg = res.onStat
	if opt != nil {
		jdir.PollWaitMin = opt.PollWaitMin
		jdir.PollWaitMax = opt.PollWaitMax
		res.setLastJSerial(opt.JSerial)
	}
	return res
}

func (ede *Events) Start(withJournal string) {
	ede.jdir.Stop = make(chan internal.StopEvent)
	go ede.jdir.Watch(withJournal)
	<-ede.Stop
	ede.jdir.Stop <- watched.Stop
	<-ede.jdir.Stop
	ede.recv.Close()
	close(ede.Stop)
}

func (ede *Events) LastJSerial() watched.JEventID {
	return ljeSeqMax*ede.ljeSec + int64(ede.ljeSeq)
}

func (ede *Events) setLastJSerial(s watched.JEventID) {
	if s < 0 {
		ede.ljeSec = time.Now().Unix()
		ede.ljeSeq = 0
	} else {
		ede.ljeSec = s / ljeSeqMax
		ede.ljeSeq = int(s % ljeSeqMax)
	}
	ede.djeSeq = -1
}

func (ede *Events) onJournal(raw []byte) {
	t, err := watched.PeekTime(raw)
	if err != nil {
		log.Errore(err)
		return
	}
	if ede.checkNewJournalEvent(t.Unix()) {
		ede.recv.Journal(watched.JounalEvent{
			Serial: ede.LastJSerial(),
			Event:  bytes.Repeat(raw, 1),
		})
	}
}

var (
	statReplaceNl  = []byte{'\n'}
	statReplaceCr  = []byte{'\r'}
	statReplaceSpc = []byte{' '}
)

func (ede *Events) onStat(event watched.StatusType, file string) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errore(err) // TODO be more descriptive
		return
	}
	raw = bytes.ReplaceAll(raw, statReplaceNl, statReplaceSpc)
	raw = bytes.ReplaceAll(raw, statReplaceCr, statReplaceSpc)
	ede.recv.Status(watched.StatusEvent{
		Type:  event,
		Event: bytes.Repeat(raw, 1),
	})
}

const ljeSeqMax = 1000

func (ede *Events) checkNewJournalEvent(uxsec int64) bool {
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
	JSerial     int64
}

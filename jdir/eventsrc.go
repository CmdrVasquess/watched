package jdir

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal"
)

type Events struct {
	Stop     chan internal.StopEvent
	recv     watched.EventRecv
	jdir     *JournalDir
	serGen   watched.JEIDCounter
	lastSer  watched.JEventID
	serIndep []string
}

func NewEvents(dir string, r watched.EventRecv, opt *Options) *Events {
	jdir := &JournalDir{Dir: dir}
	res := &Events{
		Stop:     make(chan internal.StopEvent),
		recv:     r,
		jdir:     jdir,
		serIndep: opt.SerialIndependent,
	}
	jdir.PerJLine = res.onJournal
	jdir.OnStatChg = res.onStat
	if opt != nil {
		jdir.PollWaitMin = opt.PollWaitMin
		jdir.PollWaitMax = opt.PollWaitMax
		res.serGen.SetLast(opt.JSerial)
	}
	return res
}

func (ede *Events) Start(withJournal string) {
	ede.jdir.Stop = MakeStopChan()
	go ede.jdir.Watch(withJournal)
	<-ede.Stop
	ede.jdir.Stop <- watched.Stop
	<-ede.jdir.Stop
	ede.recv.Close()
	close(ede.Stop)
}

func (ede *Events) LastJSerial() watched.JEventID {
	return ede.serGen.Last()
}

func (ede *Events) onJournal(raw []byte) {
	t, err := watched.PeekTime(raw)
	if err != nil {
		log.Errore(err)
		return
	}
	ok, err := ede.checkNewJournalEvent(t.Unix())
	if err != nil {
		log.Warne(err)
	}
	if ok {
		ede.recv.OnJournalEvent(watched.JounalEvent{
			Serial: ede.LastJSerial(),
			Event:  bytes.Repeat(raw, 1),
		})
	} else if e, err := watched.PeekEvent(raw); err != nil {
		log.Errore(err)
	} else if ede.isSerIndep(e) {
		ede.recv.OnJournalEvent(watched.JounalEvent{
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
	raw = bytes.TrimSpace(raw)
	ede.recv.OnStatusEvent(watched.StatusEvent{
		Type:  event,
		Event: bytes.Repeat(raw, 1),
	})
}

const ljeSeqMax = 1000

func (ede *Events) checkNewJournalEvent(uxsec int64) (bool, error) {
	ser, err := ede.serGen.CountUnix(uxsec)
	if err != nil {
		return false, err
	}
	if ser <= ede.lastSer {
		return false, nil
	}
	ede.lastSer = ser
	return true, nil
}

func (ede *Events) isSerIndep(evt string) bool {
	for _, si := range ede.serIndep {
		if si == evt {
			return true
		}
	}
	return false
}

type Options struct {
	PollWaitMin       time.Duration
	PollWaitMax       time.Duration
	JSerial           int64
	SerialIndependent []string
}

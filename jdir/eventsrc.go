package jdir

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"

	"git.fractalqb.de/fractalqb/c4hgol"
	"github.com/rjeczalik/notify"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal"
)

func IsStatusFile(name string) watched.StatusType {
	return statsFiles[name]
}

type Events struct {
	recv     watched.EventRecv
	jdir     string
	serGen   watched.JEIDCounter
	lastSer  watched.JEventID
	serIndep []string
	stop     chan internal.StopEvent
	runs     int32
}

type Options struct {
	JSerial           int64
	SerialIndependent []string
}

func NewEvents(dir string, r watched.EventRecv, opt *Options) *Events {
	res := &Events{
		recv: r,
		jdir: dir,
		stop: make(chan internal.StopEvent),
	}
	if opt != nil {
		res.serGen.SetLast(opt.JSerial)
		res.serIndep = opt.SerialIndependent
	}
	return res
}

func (ede *Events) Start(withJournal string) (err error) {
	if !atomic.CompareAndSwapInt32(&ede.runs, 0, 1) {
		if atomic.LoadInt32(&ede.runs) > 0 {
			return errors.New("jdir events already running")
		}
		return errors.New("cannot restart stopped jdir events")
	}
	log.Infov("Start watching files in `dir`", ede.jdir)
	fsevents := make(chan notify.EventInfo, 1)
	if err := notify.Watch(ede.jdir, fsevents, notify.Write); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			log.Errore(err)
		}
		if fsevents != nil {
			notify.Stop(fsevents)
		}
		close(ede.stop)
		log.Infov("Stopped watching files in `dir`", ede.jdir)
	}()
	var jfpos int64
	var jfile *os.File
	if withJournal != "" {
		jpath := filepath.Join(ede.jdir, withJournal)
		stat, err := os.Stat(jpath)
		if err != nil {
			return err
		}
		jfpos = stat.Size()
		jfile, err = os.Open(jpath)
		if err != nil {
			return err
		}
	}
EVENT_LOOP:
	for {
		select {
		case <-ede.stop:
			break EVENT_LOOP
		case e := <-fsevents:
			log.Tracev("FS `event`", e)
			file := filepath.Base(e.Path())
			if IsJournalFile(file) {
				if file != withJournal {
					log.Debugv("Switch `from` to `journal`", withJournal, file)
					var err error
					if jfile != nil {
						err = jfile.Close()
						if err != nil {
							log.Errore(err)
						}
					}
					jfile, err = os.Open(e.Path())
					if err != nil {
						log.Errore(err)
						withJournal = ""
						continue
					}
					withJournal = file
					jfpos = 0
				}
				stat, err := jfile.Stat()
				if err != nil {
					log.Errore(err)
					continue
				}
				if stat.Size() > jfpos {
					jfile.Seek(jfpos, io.SeekStart)
					lrd := io.LimitReader(jfile, stat.Size()-jfpos)
					scn := bufio.NewScanner(lrd)
					for scn.Scan() {
						data := scn.Bytes()
						data = bytes.TrimSpace(data)
						if len(data) > 0 {
							if log.WouldLog(c4hgol.Trace) {
								log.Tracef("journal data [%s]", string(data))
							}
							ede.onJournal(data)
						}
					}
					jfpos = stat.Size()
				}
			} else if sft := IsStatusFile(file); sft > 0 {
				ede.onStatus(sft, e.Path())
			} else {
				log.Tracev("Ignore FS event `on`", file)
			}
		}
	}
	notify.Stop(fsevents)
	fsevents = nil
	return nil
}

func (ede *Events) Stop() {
	if !atomic.CompareAndSwapInt32(&ede.runs, 1, -1) {
		return
	}
	ede.stop <- internal.StopEvent{}
	<-ede.stop
}

func (ede *Events) LastJSerial() watched.JEventID {
	return ede.serGen.Last()
}

func (ede *Events) onJournal(raw []byte) int {
	t, err := watched.PeekTime(raw)
	if err != nil {
		log.Errore(err)
		return len(raw)
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
	return len(raw)
}

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

var (
	statReplaceNl  = []byte("\n")
	statReplaceCr  = []byte("\r")
	statReplaceSpc = []byte(" ")
)

func (ede *Events) onStatus(t watched.StatusType, file string) {
	raw, err := os.ReadFile(file)
	if err != nil {
		log.Errore(fmt.Errorf("%w reading %s", err, file))
		return
	}
	raw = bytes.ReplaceAll(raw, statReplaceNl, statReplaceSpc)
	raw = bytes.ReplaceAll(raw, statReplaceCr, statReplaceSpc)
	raw = bytes.TrimSpace(raw)
	if len(raw) > 0 {
		ede.recv.OnStatusEvent(watched.StatusEvent{
			Type:  t,
			Event: raw,
		})
	}
}

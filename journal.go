package watched

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"

	"git.fractalqb.de/fractalqb/qblog"
	"git.fractalqb.de/fractalqb/sllm/v3"
	"github.com/rjeczalik/notify"

	"github.com/CmdrVasquess/watched/internal"
)

type Journal struct {
	recv     EventRecv
	jdir     string
	serIndep []string
	stop     chan internal.StopEvent
	runs     int32
}

type JournalOptions struct {
	SerialIndependent []string
}

func NewJournal(dir string, r EventRecv, opt *JournalOptions) *Journal {
	res := &Journal{
		recv: r,
		jdir: dir,
		stop: make(chan internal.StopEvent),
	}
	if opt != nil {
		res.serIndep = opt.SerialIndependent
	}
	return res
}

func (ede *Journal) Start() (err error) {
	if !atomic.CompareAndSwapInt32(&ede.runs, 0, 1) {
		if atomic.LoadInt32(&ede.runs) > 0 {
			return errors.New("jdir events already running")
		}
		return errors.New("cannot restart stopped jdir events")
	}

	log.Info("Start watching files in `dir`", `dir`, ede.jdir)
	fsevents := make(chan notify.EventInfo, 32) // TODO eliminagte magic number
	if err := notify.Watch(ede.jdir, fsevents, notify.Write); err != nil {
		return err
	}

	var jfile *os.File
	defer func() {
		if err != nil {
			log.Error(err.Error())
		}
		if fsevents != nil {
			notify.Stop(fsevents)
		}
		if jfile != nil {
			if err := jfile.Close(); err != nil {
				log.Error("journal `file` close `error`",
					`file`, jfile,
					`error`, err,
				)
			}
		}
		close(ede.stop)
		log.Info("Stopped watching files in `dir`", `dir`, ede.jdir)
	}()

	var (
		jfileName string
		jeventNo  int
	)
EVENT_LOOP:
	for {
		select {
		case <-ede.stop:
			break EVENT_LOOP
		case e := <-fsevents:
			log.Trace("FS `event`", `event`, e)
			fileNm := filepath.Base(e.Path())
			if IsJournalFile(fileNm) != 0 {
				if fileNm != jfileName {
					if jfile != nil {
						if err := jfile.Close(); err != nil {
							log.Error("journal `file` close `error`",
								`file`, jfile,
								`error`, err,
							)
						}
					}
					if jfile, err = os.Open(e.Path()); err != nil {
						jfileName = ""
						log.Error("journal `file` open `error`",
							`file`, fileNm,
							`error`, err,
						)
						continue
					}
					jfileName = fileNm
					jeventNo = 0
				}
				jscan := bufio.NewScanner(jfile)
				for jscan.Scan() {
					data := bytes.TrimSpace(jscan.Bytes())
					if len(data) > 0 {
						if log.Enabled(context.Background(), qblog.LevelTrace) {
							log.Trace("journal `data`", `data`, string(data))
						}
						jeventNo++
						err = ede.recv.OnJournalEvent(JounalEvent{
							File:    jfileName,
							EventNo: jeventNo,
							Event:   data,
						})
						if err != nil {
							log.Error("`journal` `event` `error`",
								`journal`, jfileName,
								`event`, jeventNo,
								`error`, err.Error(),
							)
						}
					}
				}
			} else if sft := IsStatusFile(fileNm); sft > 0 {
				err = ede.onStatus(sft, e.Path())
				if err != nil {
					log.Error("`status` `error`",
						`status`, fileNm,
						`error`, err.Error(),
					)
				}
			} else {
				log.Trace("Ignore FS event `on`", `on`, fileNm)
			}
		}
	}
	notify.Stop(fsevents)
	fsevents = nil
	return nil
}

func (ede *Journal) Stop() {
	if !atomic.CompareAndSwapInt32(&ede.runs, 1, -1) {
		return
	}
	ede.stop <- internal.StopEvent{}
	<-ede.stop
}

func (ede *Journal) onStatus(t StatusType, file string) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return sllm.ErrorIdx("`error` reading `file`", err, file)
	}
	for i, c := range raw {
		switch c {
		case '\n', '\r':
			raw[i] = ' '
		}
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) > 0 {
		return ede.recv.OnStatusEvent(StatusEvent{
			Type:  t,
			Event: raw,
		})
	}
	return nil
}

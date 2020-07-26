package jdir

/* To be portable we need to poll the logfile. On MS Win one only gets update
 * events, if the directory is "touched", i.e. a logfile that stays open and
 * regularly receives new content will not be notified until something happens
 * to its parent directory. E.g. pressing F5 in the file explorer helps –
 * but who want's to sit at the keyboard and press F5 from time to time??? */

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	str "strings"
	"time"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal"
	"github.com/fsnotify/fsnotify"
)

var log = internal.JDirLog

type JournalDir struct {
	Dir       string
	PerJLine  func([]byte)
	OnStatChg func(event watched.StatusType, file string)
	Stop      chan internal.StopEvent
	// PollWaitMin will be set to a reasonable default
	PollWaitMin time.Duration
	// PollWaitMax will be set to a reasonable default
	PollWaitMax time.Duration
}

func MakeStopChan() chan internal.StopEvent {
	return make(chan internal.StopEvent)
}

func (jd *JournalDir) Watch(startWith string) {
	if jd.PollWaitMin <= 0 {
		jd.PollWaitMin = 700 * time.Millisecond
	}
	if jd.PollWaitMax < jd.PollWaitMin {
		jd.PollWaitMax = 4691 * time.Millisecond
	}
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatala("cannot create fs-watcher: `err`", err)
	}
	defer watch.Close()
	if err = watch.Add(jd.Dir); err != nil {
		log.Fatala("cannot watch `dir`: `err`", jd.Dir, err)
	}
	watchList := make(chan string, 12) // do we really need backlog?
	go jd.pollFile(watchList)          // careful: concurrency & shared state (const!)
	log.Infoa("watching journals in `dir`", jd.Dir)
	if startWith != "" {
		watchList <- filepath.Join(jd.Dir, startWith)
	}
	for {
		select {
		case fse := <-watch.Events:
			fseBase := filepath.Base(fse.Name)
			if evt := statsFiles[fseBase]; evt != 0 {
				log.Tracea("FS event on `stats` `tag`: `event`", fseBase, evt, fse)
				if fse.Op != fsnotify.Write {
					continue
				}
				stat, err := os.Stat(fse.Name)
				if err != nil {
					log.Errora("cannot get fstat of `event`: `err`", fse.Name, err)
				} else if stat.Size() == 0 {
					log.Tracea("empty stat `file`", fseBase)
				} else {
					log.Tracea("stat `file` `size`", fseBase, stat.Size())
					jd.OnStatChg(evt, fse.Name)
				}
			} else if !IsJournalFile(filepath.Base(fse.Name)) {
				log.Debuga("ignore `event` on non-journal `file`", fse.Op, fse.Name)
			} else if fse.Op&fsnotify.Create == fsnotify.Create {
				cleanName := filepath.Clean(fse.Name)
				log.Debuga("enqueue new `journal`", cleanName)
				watchList <- cleanName
			}
		case err = <-watch.Errors:
			log.Errora("fs-watch `err`", err)
		case <-jd.Stop:
			watchList <- ""
			<-watchList
			log.Infos("exit journal watcher")
			close(jd.Stop)
			return
		}
	}
}

var statsFiles = map[string]watched.StatusType{
	"Cargo.json":       watched.StatCargo,
	"Market.json":      watched.StatMarket,
	"ModulesInfo.json": watched.StatModules,
	"NavRoute.json":    watched.StatNavRoute,
	"Outfitting.json":  watched.StatOutfit,
	"Shipyard.json":    watched.StatShipyard,
	"Status.json":      watched.StatStatus,
}

// Unix: \n; Win: \r\n; Apple <= OS 9: \r
func splitLogLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.IndexAny(data, "\n\r"); i < 0 {
		return 0, nil, nil
	} else if len(data) == i+1 {
		return i + 1, data[0:i], nil
	} else if nc := data[i+1]; nc == '\n' || nc == '\r' {
		return i + 2, data[0:i], nil
	} else {
		return i + 1, data[0:i], nil
	}
}

func (jd *JournalDir) pollFile(watchFiles chan string) {
	log.Infos("file poller waiting for journals")
	var jrnlName string
	var jrnlFile *os.File
	var jrnlRdPos int64
	sleep := 0 * time.Millisecond
	defer func() {
		if jrnlFile != nil {
			jrnlFile.Close()
		}
	}()
	for {
		if len(jrnlName) == 0 {
			jrnlName = <-watchFiles
			if jrnlName == "" {
				log.Infos("exit logwatch file-poller")
				close(watchFiles)
				runtime.Goexit()
			}
			log.Infoa("start watching `file`", jrnlName)
			var err error
			if jrnlFile, err = os.Open(jrnlName); err != nil {
				log.Errora("cannot watch `file`: `err`", jrnlName, err)
				jrnlName = ""
			}
			jrnlRdPos = 0
			sleep = 0
		}
		jrnlStat, err := jrnlFile.Stat()
		if err != nil {
			log.Errora("cannot Stat() `file`: `err`", jrnlName, err)
			jrnlFile.Close()
			jrnlFile = nil
			jrnlName = ""
		} else {
			newRdPos := jrnlStat.Size()
			if newRdPos > jrnlRdPos {
				log.Tracea("new bytes: `count` [`start` … `end`]",
					newRdPos-jrnlRdPos,
					jrnlRdPos,
					newRdPos)
				jrnlScnr := bufio.NewScanner(jrnlFile)
				jrnlScnr.Split(splitLogLines)
				for jrnlScnr.Scan() {
					line := jrnlScnr.Bytes()
					jd.PerJLine(line)
				}
				jrnlRdPos = newRdPos
				sleep = 0
			} else if len(watchFiles) == 0 {
				switch {
				case sleep == 0:
					sleep = jd.PollWaitMin
				case sleep < jd.PollWaitMax:
					if sleep = 5 * sleep / 4; sleep > jd.PollWaitMax {
						sleep = jd.PollWaitMax
					}
				}
				log.Tracea("nothing to do, `sleep`…", sleep)
				time.Sleep(sleep)
				log.Traces("…woke up again")
			} else {
				log.Infoa("closing journal: `file`", jrnlName)
				jrnlFile.Close()
				jrnlFile = nil
				jrnlName = ""
			}
		}
	}
}

func IsJournalFile(name string) bool {
	return str.HasPrefix(name, "Journal.") &&
		str.HasSuffix(name, ".log")
}

func NewestJournal(inDir string) (res string, err error) {
	dir, err := os.Open(inDir)
	if err != nil {
		return "", err
	}
	defer dir.Close()
	var maxTime time.Time
	infos, err := dir.Readdir(1)
	for len(infos) > 0 && err == nil {
		info := infos[0]
		if IsJournalFile(info.Name()) && (info.ModTime().After(maxTime) || len(res) == 0) {
			res = info.Name()
			maxTime = info.ModTime()
		}
		infos, err = dir.Readdir(1)
	}
	return res, nil
}

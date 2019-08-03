package watched

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
	str "strings"
	"time"

	"runtime"

	l "git.fractalqb.de/fractalqb/qbsllm"
	"github.com/fsnotify/fsnotify"
)

var (
	log    = l.New(l.Lnormal, "watchED", nil, nil)
	LogCfg = l.Config(log)
)

const (
	EscrJournal  = 'J'
	EscrMarket   = 'C' // Commerce
	EscrModules  = 'M'
	EscrOutfit   = 'F'
	EscrShipyard = 'Y'
	EscrStatus   = 'S'
)

type JournalDir struct {
	Dir       string
	PerJLine  func([]byte)
	OnStatChg func(tag rune, file string)
	Quit      chan bool
	// PollWaitMin will be set to a reasonable default
	PollWaitMin time.Duration
	// PollWaitMax will be set to a reasonable default
	PollWaitMax time.Duration
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
	if len(startWith) > 0 {
		watchList <- filepath.Join(jd.Dir, startWith)
	}
	for {
		select {
		case fse := <-watch.Events:
			fseBase := filepath.Base(fse.Name)
			if ok, tag := isStatsFile(fseBase); ok {
				if fse.Op != fsnotify.Write {
					continue
				}
				log.Tracea("FSevent on `stats` (`tag`): `event`", fseBase, tag, fse)
				stat, err := os.Stat(fse.Name)
				if err != nil {
					log.Errora("cannot get fstat of `event`: `err`", fse.Name, err)
				} else if stat.Size() == 0 {
					log.Debuga("empty stat `file`", fseBase)
				} else {
					log.Tracea("stat `file` `size`", fseBase, stat.Size())
					jd.OnStatChg(tag, fse.Name)
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
		case <-jd.Quit:
			watchList <- ""
			log.Info(l.Str("exit journal watcher"))
			runtime.Goexit()
		}
	}
}

var journalStatsFiles = map[string]rune{
	"Market.json":      EscrMarket,
	"ModulesInfo.json": EscrModules,
	"Outfitting.json":  EscrOutfit,
	"Shipyard.json":    EscrShipyard,
	"Status.json":      EscrStatus,
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
	log.Info(l.Str("file poller waiting for journals"))
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
				log.Info(l.Str("exit logwatch file-poller"))
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
				log.Trace(l.Str("…woke up again"))
			} else {
				log.Infoa("closing journal: `file`", jrnlName)
				jrnlFile.Close()
				jrnlFile = nil
				jrnlName = ""
			}
		}
	}
}

func isStatsFile(name string) (flag bool, tag rune) {
	tag, ok := journalStatsFiles[name]
	return ok, tag
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

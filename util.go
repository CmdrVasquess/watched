package watched

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Format 1: Journal.201206082715.01.log
// Format 2: Journal.2023-12-27T134309.01.log
func IsJournalFile(name string) (format int) {
	if !strings.HasPrefix(name, "Journal.") || !strings.HasSuffix(name, ".log") {
		return 0
	}
	switch len(name) { // TODO just a heuristic
	case 27:
		return 1
	case 32:
		return 2
	}
	return -1 // Allows to accept synthetic names e.g. for tests
}

func JournalFileCmpr(f, g string) int {
	ff, gf := IsJournalFile(f), IsJournalFile(g)
	switch ff {
	case 2:
		if gf == 2 {
			return strings.Compare(f, g)
		}
		return 1
	case 1:
		switch gf {
		case 2:
			return -1
		case 1:
			return strings.Compare(f, g)
		}
		return 1
	case -1:
		switch gf {
		case -1:
			return strings.Compare(f, g)
		case 1, 2:
			return -1
		}
		return 1
	case 0:
		if gf == 0 {
			return 0
		}
		return -1
	}
	return 0
}

// TODO Use [JouralFileCmpr]
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
		if IsJournalFile(info.Name()) != 0 && (info.ModTime().After(maxTime) || len(res) == 0) {
			res = info.Name()
			maxTime = info.ModTime()
		}
		infos, err = dir.Readdir(1)
	}
	return res, nil
}

func IsStatusFile(name string) StatusType {
	return statsFiles[name]
}

var statsFiles = make(map[string]StatusType)

func init() {
	for i := StatusType(1); i < EndStatusType; i++ {
		statsFiles[i.String()+".json"] = i
	}
}

func PeekTime(str []byte) (t time.Time, err error) {
	idx := bytes.Index(str, timestampTag)
	if idx < 0 {
		estr := string(str)
		if len(estr) > 50 {
			estr = string(str[:50]) + "…"
		}
		return time.Time{}, fmt.Errorf("no timestamp in event: %s", estr)
	}
	val := str[idx+13 : idx+33]
	return time.Parse(time.RFC3339, string(val))
}

func PeekEvent(str []byte) (event string, err error) {
	idx := bytes.Index(str, eventTag)
	if idx < 0 {
		return "", errors.New("no event type in event")
	}
	str = str[idx+9:]
	idx = bytes.IndexByte(str, '"')
	if idx < 0 {
		return "", errors.New("cannot find end of event type")
	}
	return string(str[:idx]), nil
}

func Peek(str []byte) (t time.Time, event string, err error) {
	if t, err = PeekTime(str); err != nil {
		return t, "", err
	}
	event, err = PeekEvent(str)
	return t, event, err
}

const (
	timestamTagStr = `"timestamp":`
	eventTagStr    = `"event":`
)

var (
	timestampTag = []byte(timestamTagStr)
	eventTag     = []byte(eventTagStr)
)

type JournalProgress struct {
	file string
	eNo  int
}

func (jp *JournalProgress) File() string { return jp.file }
func (jp *JournalProgress) EventNo() int { return jp.eNo }

func (jp *JournalProgress) Reset() {
	jp.file = ""
	jp.eNo = 0
}

func (jp *JournalProgress) IsNew(file string, n int) bool {
	if IsJournalFile(file) == 0 || n < 1 {
		return false
	}
	fcmpr := JournalFileCmpr(file, jp.file)
	switch {
	case fcmpr > 0:
		return true
	case fcmpr < 0:
		return false
	}
	return n > jp.eNo
}

func (jp *JournalProgress) Set(file string, n int) error {
	switch {
	case IsJournalFile(file) == 0:
		return fmt.Errorf("illegal journal file name '%s'", file)
	case n < 1:
		return fmt.Errorf("illegal event no %d in journal file '%s'", n, file)
	}
	jp.file = file
	jp.eNo = n
	return nil
}

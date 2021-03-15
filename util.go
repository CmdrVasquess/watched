package watched

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"git.fractalqb.de/fractalqb/c4hgol"
	"github.com/CmdrVasquess/watched/internal"
)

func LogCfg() c4hgol.Configurer { return internal.LogCfg }

const Stop = internal.StopEvent(0)

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

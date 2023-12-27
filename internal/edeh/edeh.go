package edeh

import (
	"bytes"
	"strconv"

	"git.fractalqb.de/fractalqb/sllm/v3"
	"github.com/CmdrVasquess/watched"
)

func Message(r watched.EventRecv, msg []byte) error {
	name, eno, msg, err := splitHeader(msg)
	if err != nil {
		return err
	}
	if eno > 0 {
		err = r.OnJournalEvent(watched.JounalEvent{
			File:    name,
			EventNo: eno,
			Event:   bytes.Clone(msg),
		})
	} else if sty := watched.ParseStatusType(name); sty == 0 {
		err = sllm.ErrorIdx("Unknown `status type` in `line`", name, string(msg))
	} else {
		err = r.OnStatusEvent(watched.StatusEvent{
			Type:  sty,
			Event: bytes.Clone(msg),
		})
	}
	return err
}

func splitHeader(msg []byte) (file string, eno int, body []byte, err error) {
	sep := bytes.IndexByte(msg, '\t')
	if sep < 1 {
		return "", 0, msg, sllm.ErrorIdx("no event prefix in `line`", string(msg))
	}
	colon := bytes.LastIndexByte(msg[:sep], ':')
	switch {
	case colon < 0:
		return string(msg[:sep]), -1, msg[sep+1:], nil
	case colon == 0:
		return "", 0, msg, sllm.ErrorIdx("no filename in journal prefix of `line`", string(msg))
	}
	file = string(msg[:colon])
	eno, err = strconv.Atoi(string(msg[colon+1 : sep]))
	if err != nil {
		return "", 0, msg, err
	}
	return file, eno, msg[sep+1:], nil
}

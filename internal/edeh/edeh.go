package edeh

import (
	"bytes"
	"strconv"

	"git.fractalqb.de/fractalqb/sllm"
	"github.com/CmdrVasquess/watched"
)

func Messgage(r watched.EventRecv, msg []byte) (err error) {
	sep := bytes.IndexAny(msg, " \t")
	if sep < 1 {
		return sllm.Error("no event prefix in `line`", string(msg))
	}
	prefix := string(msg[:sep])
	if ser, err := strconv.ParseInt(prefix, 10, 64); err == nil {
		err = r.Journal(watched.JounalEvent{
			Serial: ser,
			Event:  bytes.Repeat(msg[sep:], 1),
		})
	} else if sty := watched.ParseStatusType(prefix); sty == 0 {
		err = sllm.Error("Unknown `status type` in `line`", prefix, string(msg))
	} else {
		err = r.Status(watched.StatusEvent{
			Type:  sty,
			Event: bytes.Repeat(msg[sep:], 1),
		})
	}
	return err
}

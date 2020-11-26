package plugin

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/watched"
)

var (
	log                      = qbsllm.New(qbsllm.Lnormal, "edehpin", nil, nil)
	LogCfg c4hgol.Configurer = qbsllm.NewConfig(log)
)

var journalPrefix = []byte("Journal ")

func RunRecv(r watched.EventRecv, rd io.Reader) error {
	if rd == nil {
		rd = os.Stdin
	}
	scn := bufio.NewScanner(rd)
	for scn.Scan() {
		sep := bytes.IndexAny(scn.Bytes(), " \t")
		if sep < 1 {
			log.Warna("no event prefix in `line`", scn.Text())
			continue
		}
		prefix := string(scn.Bytes()[:sep])
		if ser, err := strconv.ParseInt(prefix, 10, 64); err == nil {
			err = r.Journal(watched.JounalEvent{
				Serial: ser,
				Event:  bytes.TrimSpace(scn.Bytes()[sep:]),
			})
			if err != nil {
				log.Errore(err)
			}
		} else if sty := watched.ParseStatusType(prefix); sty == 0 {
			log.Warna("Unknown `status type` in `line`", prefix, scn.Text())
		} else {
			err = r.Status(watched.StatusEvent{
				Type:  sty,
				Event: bytes.TrimSpace(scn.Bytes()[sep:]),
			})
			if err != nil {
				log.Errore(err)
			}
		}
	}
	return r.Close()
}

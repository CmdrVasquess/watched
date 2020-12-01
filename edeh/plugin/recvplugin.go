package plugin

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal/edeh"
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
		err := edeh.Messgage(r, bytes.Repeat(scn.Bytes(), 1))
		if err != nil {
			log.Errore(err)
		}
	}
	return nil
}

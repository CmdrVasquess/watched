package plugin

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal/edeh"
)

var (
	log                     = qblog.New("edehpin")
	LogCfg c4hgol.LogConfig = log
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

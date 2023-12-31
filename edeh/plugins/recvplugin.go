package plugins

import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"os"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal/edeh"
)

func RunRecv(r watched.EventRecv, rd io.Reader, log *slog.Logger) error {
	if rd == nil {
		rd = os.Stdin
	}
	scn := bufio.NewScanner(rd)
	for scn.Scan() {
		err := edeh.Message(r, bytes.Clone(scn.Bytes()))
		if err != nil {
			log.Error(err.Error())
		}
	}
	return nil
}

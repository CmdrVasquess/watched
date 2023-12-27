package edehnet

import (
	"bufio"
	"bytes"
	"context"
	"log/slog"
	"net"

	"git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal/edeh"
)

var log = qblog.New(&qblog.DefaultConfig)

func SetLog(l *slog.Logger) { log = &qblog.Logger{Logger: l} }

type Receiver struct {
	Listen string
}

func (r *Receiver) Run(wrecv watched.EventRecv) (err error) {
	log.Info("waiting for event source on `addr`", `addr`, r.Listen)
	var lstn net.Listener
	lstn, err = net.Listen("tcp", r.Listen)
	if err != nil {
		return err
	}
	conn, err := lstn.Accept()
	if err != nil {
		log.Error(err.Error())
	}
	lstn.Close()
	defer conn.Close()
	log.Info("event source connected from `addr`", `addr`, conn.RemoteAddr())
	scn := bufio.NewScanner(conn)
	for scn.Scan() {
		if log.Enabled(context.Background(), qblog.LevelDebug) {
			estr := string(scn.Text())
			if len(estr) > 80 {
				estr = estr[:80] + "…"
			}
			log.Debug("received `net event`", `net event`, estr)
		}
		err := edeh.Message(wrecv, bytes.Clone(scn.Bytes()))
		if err != nil {
			log.Error(err.Error())
		}
	}
	if err = scn.Err(); err != nil {
		log.Error("event source `addr` `error`", `addr`, conn.RemoteAddr(), `error`, err)
		return err
	}
	log.Info("event source with `addr` disconnected", `addr`, conn.RemoteAddr())
	return nil
}

package edehnet

import (
	"bufio"
	"bytes"
	"net"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal/edeh"
)

var (
	log                     = qblog.New("edehnet")
	LogCfg c4hgol.LogConfig = log
)

type Receiver struct {
	Listen string
}

func (r *Receiver) Run(wrecv watched.EventRecv) (err error) {
	log.Infov("waiting for event source on `addr`", r.Listen)
	var lstn net.Listener
	lstn, err = net.Listen("tcp", r.Listen)
	if err != nil {
		return err
	}
	conn, err := lstn.Accept()
	if err != nil {
		log.Errore(err)
	}
	lstn.Close()
	defer conn.Close()
	log.Infov("event source connected from `addr`", conn.RemoteAddr())
	scn := bufio.NewScanner(conn)
	for scn.Scan() {
		if log.WouldLog(c4hgol.Trace) {
			estr := string(scn.Text())
			if len(estr) > 80 {
				estr = estr[:80] + "…"
			}
			log.Tracef("received net event: '%s'", estr)
		}
		err := edeh.Messgage(wrecv, bytes.Repeat(scn.Bytes(), 1))
		if err != nil {
			log.Errore(err)
		}
	}
	log.Infov("event source with `addr` disconnected", conn.RemoteAddr())
	return nil
}

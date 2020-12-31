package edehnet

import (
	"bufio"
	"bytes"
	"net"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal/edeh"
)

var (
	log                      = qbsllm.New(qbsllm.Lnormal, "edehnet", nil, nil)
	LogCfg c4hgol.Configurer = qbsllm.NewConfig(log)
)

type Receiver struct {
	Listen string
}

func (r *Receiver) Run(wrecv watched.EventRecv) (err error) {
	log.Infoa("waiting for event source on `addr`", r.Listen)
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
	log.Infoa("client connected from `addr`", conn.RemoteAddr())
	scn := bufio.NewScanner(conn)
	for scn.Scan() {
		err := edeh.Messgage(wrecv, bytes.Repeat(scn.Bytes(), 1))
		if err != nil {
			log.Errore(err)
		}
	}
	log.Infoa("event source with `addr` disconnected", conn.RemoteAddr())
	return nil
}

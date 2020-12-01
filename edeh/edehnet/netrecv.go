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

func (r *Receiver) Run(wrecv watched.EventRecv) error {
	lstn, err := net.Listen("tcp", r.Listen)
	if err != nil {
		return err
	}
	conn, err := lstn.Accept()
	if err != nil {
		log.Errore(err)
	}
	defer conn.Close()
	scn := bufio.NewScanner(conn)
	for scn.Scan() {
		err := edeh.Messgage(wrecv, bytes.Repeat(scn.Bytes(), 1))
		if err != nil {
			log.Errore(err)
		}
	}
	return nil
}

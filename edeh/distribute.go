package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/CmdrVasquess/watched"
)

type distributor struct {
	TCP       []tcpClient
	reconnect atomic.Value // [][]byte

	pins      []*plugin
	waitClose sync.WaitGroup
	// TODO compute and use an overall blacklist / whitelist
}

func (d *distributor) load(file string) error {
	rd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	if err = dec.Decode(d); err != nil {
		return err
	}
	return nil
}

func (d *distributor) addPlugin(pin *plugin) {
	pin.jes = make(chan *jEvent, fPinQLen)
	pin.ses = make(chan *sEvent, fPinQLen)
	d.pins = append(d.pins, pin)
}

func (d *distributor) OnJournalEvent(e watched.JounalEvent) error {
	event, err := e.Event.PeekEvent()
	if err != nil {
		return err
	}
	if event == "" {
		return errors.New("empty journal event tag")
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d ", e.Serial)
	buf.Write(e.Event)
	buf.WriteByte('\n')
	je := &jEvent{
		JounalEvent: e,
		evt:         event,
		msg:         buf.Bytes(),
	}
	for i := range d.TCP {
		tcp := &d.TCP[i]
		tcp.enqueue(je)
	}
	switch event {
	case "Fileheader":
		d.reconnect.Store([][]byte{je.msg})
	case "Commander":
		reconn := d.reconnect.Load().([][]byte)
		d.reconnect.Store(append(reconn, je.msg))
	case "Shutdown":
		d.reconnect.Store([][]byte{je.msg})
	}
	for _, pin := range d.pins {
		select {
		case pin.jes <- je:
		default:
			log.Warna("Journal event queue of `plugin` full, drop `journal event`",
				pin.cmd,
				e.Serial)
		}
	}
	return nil
}

func (d *distributor) OnStatusEvent(e watched.StatusEvent) error {
	var buf bytes.Buffer
	buf.WriteString(e.Type.String())
	buf.WriteByte(' ')
	buf.Write(e.Event)
	buf.WriteByte('\n')
	se := &sEvent{
		StatusEvent: e,
		msg:         buf.Bytes(),
	}
	for i := range d.TCP {
		tcp := &d.TCP[i]
		tcp.enqueue(se)
	}
	for _, pin := range d.pins {
		select {
		case pin.ses <- se:
		default:
			log.Warna("Status event queue of `plugin` full, frop `status event`",
				pin.cmd,
				e.Type)
		}
	}
	return nil
}

func (d *distributor) Close() error {
	for i := range d.TCP {
		c := &d.TCP[i]
		close(c.evtq)
	}
	for _, pin := range d.pins {
		close(pin.jes)
		close(pin.ses)
	}
	return nil
}

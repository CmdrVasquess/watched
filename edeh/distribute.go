package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/CmdrVasquess/watched"
)

type distributor struct {
	TCP        []tcpClient
	reconnEvts [][]byte
	reconnLock sync.RWMutex

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
	fmt.Fprintf(&buf, "%s:%d\t", e.File, e.EventNo)
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
		d.reconnSet1(je.msg)
	case "Commander", "Shutdown":
		d.reconnAdd(je.msg)
	}
	for _, pin := range d.pins {
		select {
		case pin.jes <- je:
		default:
			log.Warn("Journal event queue of `plugin` full, drop `journal event`",
				`plugin`, pin.cmd,
				`journal event`, fmt.Sprintf("%s:%d", e.File, e.EventNo))
		}
	}
	return nil
}

func (d *distributor) reconnSet1(raw []byte) {
	d.reconnLock.Lock()
	defer d.reconnLock.Unlock()
	d.reconnEvts = [][]byte{raw}
}

func (d *distributor) reconnAdd(raw []byte) {
	d.reconnLock.Lock()
	defer d.reconnLock.Unlock()
	d.reconnEvts = append(d.reconnEvts, raw)
}

func (d *distributor) reconnList() [][]byte {
	d.reconnLock.RLock()
	defer d.reconnLock.RUnlock()
	return d.reconnEvts
}

func (d *distributor) OnStatusEvent(e watched.StatusEvent) error {
	var buf bytes.Buffer
	buf.WriteString(e.Type.String())
	buf.WriteByte('\t')
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
			log.Warn("Status event queue of `plugin` full, frop `status event`",
				`plugin`, pin.cmd,
				`status event`, e.Type)
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

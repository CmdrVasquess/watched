package main

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/CmdrVasquess/watched"
)

type distributor struct {
	pins      []*plugin
	waitClose sync.WaitGroup
	// TODO compute and use an overall blacklist / whitelist
}

func (d *distributor) addPlugin(pin *plugin) {
	pin.jes = make(chan *jEvent, 16)
	pin.ses = make(chan *sEvent, 16)
	d.pins = append(d.pins, pin)
}

func (d *distributor) Journal(e watched.JounalEvent) error {
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
	for _, pin := range d.pins {
		pin.jes <- je
	}
	return nil
}

func (d *distributor) Status(e watched.StatusEvent) error {
	var buf bytes.Buffer
	buf.WriteString(e.Type.String())
	buf.WriteByte(' ')
	buf.Write(e.Event)
	buf.WriteByte('\n')
	se := &sEvent{
		StatusEvent: e,
		msg:         buf.Bytes(),
	}
	for _, pin := range d.pins {
		pin.ses <- se
	}
	return nil
}

func (d *distributor) Close() error {
	for _, pin := range d.pins {
		close(pin.jes)
		close(pin.ses)
	}
	return nil
}

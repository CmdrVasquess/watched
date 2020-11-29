package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/CmdrVasquess/watched"
)

type tcpClient struct {
	Addr    string
	Journal BlackWhiteList
	Status  BlackWhiteList

	conn net.Conn
}

func (c *tcpClient) jounrnal(event string, msg []byte) {
	if c.Journal.Filter(event) {
		var err error
		if c.conn == nil {
			log.Infoa("connect to `TCP client`", c.Addr)
			if c.conn, err = net.Dial("tcp", c.Addr); err != nil {
				log.Warne(err)
				return
			}
		}
		_, err = c.conn.Write(msg)
		if err != nil {
			log.Errora("send journal to `TCP client` `err`", c.Addr, err)
		}
	}
}

func (c *tcpClient) status(event string, msg []byte) {
	if c.Status.Filter(event) {
		var err error
		if c.conn == nil {
			log.Infoa("connect to `TCP client`", c.Addr)
			if c.conn, err = net.Dial("tcp", c.Addr); err != nil {
				log.Warne(err)
				return
			}
		}
		_, err = c.conn.Write(msg)
		if err != nil {
			log.Errora("send status to `TCP client` `err`", c.Addr, err)
		}
	}
}

type distributor struct {
	TCP []tcpClient

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
	for i := range d.TCP {
		d.TCP[i].jounrnal(event, je.msg)
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
	for i := range d.TCP {
		d.TCP[i].status(e.Type.String(), se.msg)
	}
	for _, pin := range d.pins {
		pin.ses <- se
	}
	return nil
}

func (d *distributor) Close() error {
	for i := range d.TCP {
		c := &d.TCP[i]
		if c.conn != nil {
			log.Infoa("closing TCP connection to `client`", c.Addr)
			if err := c.conn.Close(); err != nil {
				log.Errore(err)
			}
		}
	}
	for _, pin := range d.pins {
		close(pin.jes)
		close(pin.ses)
	}
	return nil
}

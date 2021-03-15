package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/CmdrVasquess/watched"
)

type tcpClient struct {
	Addr    string
	Journal BlackWhiteList
	Status  BlackWhiteList

	conn    net.Conn
	connErr time.Time
	// TODO configurable reconnect delay
}

func (c *tcpClient) jounrnal(event string, msg []byte, reconn [][]byte) {
	if c.Journal.Filter(event) {
		var err error
		if c.conn == nil {
			if time.Now().Sub(c.connErr) < time.Second {
				log.Warna("drop journal event while waiting for reconnect delay")
				return
			}
			log.Infoa("connect to `TCP consumer`", c.Addr)
			if c.conn, err = net.Dial("tcp", c.Addr); err != nil {
				log.Errore(err)
				c.connErr = time.Now()
				return
			}
			for _, rcm := range reconn {
				if _, err = c.conn.Write(rcm); err != nil {
					log.Errore(err)
				}
			}
		}
		log.Tracea("send `event` to TCP `receiver`", event, c.Addr)
		_, err = c.conn.Write(msg)
		if err != nil {
			log.Errora("disconnect: journal to `TCP consumer` `err`", c.Addr, err)
			c.conn.Close()
			c.conn = nil
			c.connErr = time.Now()
		}
	} else {
		log.Tracea("filtered `event` from TCP `receiver`", event, c.Addr)
	}
}

func (c *tcpClient) status(event string, msg []byte) {
	if c.Status.Filter(event) {
		var err error
		if c.conn == nil {
			if time.Now().Sub(c.connErr) < time.Second {
				log.Warna("drop status event while waiting for reconnect delay")
				return
			}
			log.Infoa("connect to `TCP consumer`", c.Addr)
			if c.conn, err = net.Dial("tcp", c.Addr); err != nil {
				log.Errore(err)
				c.connErr = time.Now()
				return
			}
		}
		_, err = c.conn.Write(msg)
		if err != nil {
			log.Errora("disconnect: status to `TCP consumer` `err`", c.Addr, err)
			c.conn.Close()
			c.conn = nil
			c.connErr = time.Now()
		}
	}
}

type distributor struct {
	TCP       []tcpClient
	reconnect [][]byte

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
		d.TCP[i].jounrnal(event, je.msg, d.reconnect)
	}
	switch event {
	case "Fileheader":
		d.reconnect = [][]byte{je.msg}
	case "Commander":
		d.reconnect = append(d.reconnect, je.msg)
	case "Shutdown":
		d.reconnect = [][]byte{je.msg}
	}
	for _, pin := range d.pins {
		pin.jes <- je
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

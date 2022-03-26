package main

import (
	"fmt"
	"net"
	"time"
)

type tcpClient struct {
	Addr     string
	Journal  BlackWhiteList
	Status   BlackWhiteList
	QueueLen int

	conn       net.Conn
	connErr    time.Time
	evtq       chan interface{}
	qdropCount uint
}

func (c *tcpClient) enqueue(event interface{}) {
	select {
	case c.evtq <- event:
		c.qdropCount = 0
	default:
		if c.qdropCount == 0 {
			tmpl := fmt.Sprintf(
				"Event queue of `tcp client` full, drop %T",
				event,
			)
			log.Warna(tmpl, c.Addr)
		}
		c.qdropCount++
	}
}

func (c *tcpClient) runLoop(d *distributor) {
	log.Infof("Start TCP client loop of %s", c.Addr)
	if c.QueueLen <= 0 {
		c.evtq = make(chan interface{}, fTCPQLen)
	} else {
		c.evtq = make(chan interface{}, c.QueueLen)
	}
	for e := range c.evtq {
		switch evt := e.(type) {
		case *jEvent:
			if c.Journal.Filter(evt.evt) {
				c.send(evt.evt, evt.msg, d.reconnList())
			} else {
				log.Tracea("Filtered journal `event` from TCP `receiver`",
					evt.evt,
					c.Addr)
			}
		case *sEvent:
			event := evt.Type.String()
			if c.Status.Filter(event) {
				c.send(event, evt.msg, nil)
			} else {
				log.Tracea("Filtered status `event` from TCP `receiver`",
					event,
					c.Addr)
			}
		default:
			log.Errorf("Illegal event type %T for tcp client", e)
		}
	}
	log.Infof("Exit TCP client loop of %s", c.Addr)
	if c.conn != nil {
		log.Infoa("Closing TCP connection to `client`", c.Addr)
		if err := c.conn.Close(); err != nil {
			log.Errore(err)
		}
	}
}

func (c *tcpClient) send(event string, msg []byte, reconn [][]byte) {
	var err error
	if c.conn == nil {
		if dt := time.Now().Sub(c.connErr); dt < time.Second {
			log.Warna("`Waiting` for reconnect delay", dt)
			time.Sleep(dt)
		}
		log.Infoa("Connect to `TCP consumer`", c.Addr)
		if c.conn, err = net.Dial("tcp", c.Addr); err != nil {
			log.Errore(err)
			c.conn = nil
			c.connErr = time.Now()
			return
		}
		for _, rcm := range reconn {
			if _, err = c.conn.Write(rcm); err != nil {
				log.Errore(err)
			}
		}
	}
	log.Tracea("Send `event` to TCP `receiver`", event, c.Addr)
	_, err = c.conn.Write(msg)
	if err != nil {
		log.Errora("Disconnect: `TCP consumer` `err`", c.Addr, err)
		c.conn.Close()
		c.conn = nil
		c.connErr = time.Now()
	}
}

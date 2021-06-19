package main

import (
	"net"
	"sync/atomic"
	"time"
)

type tcpClient struct {
	Addr     string
	Journal  BlackWhiteList
	Status   BlackWhiteList
	QueueLen int

	reconn  *atomic.Value
	conn    net.Conn
	connErr time.Time
	// TODO configurable reconnect delay
	evtq chan interface{}
}

func (c *tcpClient) runLoop(reconn *atomic.Value) {
	log.Infof("Start TCP client loop of %s", c.Addr)
	c.reconn = reconn
	if c.QueueLen <= 0 {
		c.evtq = make(chan interface{}, fTCPQLen)
	} else {
		c.evtq = make(chan interface{}, c.QueueLen)
	}
	for e := range c.evtq {
		switch evt := e.(type) {
		case *jEvent:
			c.journal(evt.evt, evt.msg)
		case *sEvent:
			c.status(evt.Type.String(), evt.msg)
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

func (c *tcpClient) journal(event string, msg []byte) {
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
			reconn := c.reconn.Load().([][]byte)
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

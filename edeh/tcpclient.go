package main

import (
	"bytes"
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
			log.Warn(tmpl, `tcp client`, c.Addr)
		}
		c.qdropCount++
	}
}

func (c *tcpClient) runLoop(d *distributor) {
	log.Info("Start TCP `client` loop", `client`, c.Addr)
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
				log.Trace("Filtered journal `event` from TCP `receiver`",
					`event`, evt.evt,
					`receiver`, c.Addr,
				)
			}
		case *sEvent:
			event := evt.Type.String()
			if c.Status.Filter(event) {
				c.send(event, evt.msg, nil)
			} else {
				log.Trace("Filtered status `event` from TCP `receiver`",
					`event`, event,
					`receiver`, c.Addr,
				)
			}
		default:
			log.Error("Illegal event `type` for tcp client", `type`, e)
		}
	}
	log.Info("Exit TCP `client` loop", `client`, c.Addr)
	if c.conn != nil {
		log.Info("Closing TCP connection to `client`", `client`, c.Addr)
		if err := c.conn.Close(); err != nil {
			log.Error(err.Error())
		}
	}
}

func (c *tcpClient) send(event string, msg []byte, reconn [][]byte) {
	var (
		err         error
		msgInReconn bool
	)
	if c.conn == nil {
		if dt := time.Since(c.connErr); dt < time.Second {
			log.Warn("`Waiting` for reconnect delay", `Waiting`, dt)
			time.Sleep(dt)
		}
		log.Info("Connect to `TCP consumer`", `TCP consumer`, c.Addr)
		if c.conn, err = net.Dial("tcp", c.Addr); err != nil {
			log.Error(err.Error())
			c.conn = nil
			c.connErr = time.Now()
			return
		}
		for _, rcm := range reconn {
			if _, err = c.conn.Write(rcm); err != nil {
				log.Error(err.Error())
			}
			msgInReconn = msgInReconn || bytes.Equal(rcm, msg)
		}
	}
	if msgInReconn {
		log.Trace("`event` to TCP `receiver` already in reconnect", `event`, event, `receiver`, c.Addr)
	} else {
		log.Trace("Send `event` to TCP `receiver`", `event`, event, `receiver`, c.Addr)
		_, err = c.conn.Write(msg)
	}
	if err != nil {
		log.Error("Disconnect: `TCP consumer` `err`", `TCP consumer`, c.Addr, `err`, err)
		c.conn.Close()
		c.conn = nil
		c.connErr = time.Now()
	}
}

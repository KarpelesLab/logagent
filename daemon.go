package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/KarpelesLab/logagent/logclient"
)

type logdaemon struct {
	l *net.UnixListener
}

func (d *logdaemon) start() error {
	socket := "/run/logagent"
	if id := os.Getuid(); id > 0 {
		socket = fmt.Sprintf("/tmp/.logagent-%d.sock", id)
	}

	log.Printf("initializing logagent with socket at %s", socket)

	if _, err := os.Stat(socket); err == nil {
		log.Printf("socket found, attempting to connect")
		c, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socket})
		if err != nil {
			log.Printf("connection failed: %s", err)
			os.Remove(socket)
		} else {
			// switch to upgrade mode!
			c.Close()
		}
	}

	var err error
	d.l, err = net.ListenUnix("unix", &net.UnixAddr{Name: socket})
	if err != nil {
		log.Printf("failed to listen: %s", err)
		return err
	}
	return nil
}

func (d *logdaemon) loop() {
	for {
		c, err := d.l.AcceptUnix()
		if err != nil {
			// closed?
			return
		}
		d.handleClient(c)
	}
}

func (d *logdaemon) handleClient(c *net.UnixConn) {
	defer c.Close()

	for {
		pkt := &logclient.Packet{}
		err := pkt.ReadFrom(c)
		if err != nil {
			if err != io.EOF {
				log.Printf("failed to read from client: %s", err)
			}
			return
		}
	}
}

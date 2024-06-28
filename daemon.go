package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/KarpelesLab/klbslog"
	"github.com/KarpelesLab/shutdown"
)

type logdaemon struct {
	l *net.UnixListener
}

func (d *logdaemon) start() error {
	socket := "/run/logagent.sock"
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
			err := d.upgrade(c)
			if err != nil {
				return fmt.Errorf("during upgrade: %w", err)
			}
			return nil
		}
	}

	var err error
	d.l, err = net.ListenUnix("unix", &net.UnixAddr{Name: socket})
	if err != nil {
		log.Printf("failed to listen: %s", err)
		return err
	}
	d.l.SetUnlinkOnClose(false)
	go d.loop()
	return nil
}

func (d *logdaemon) upgrade(c *net.UnixConn) error {
	defer c.Close()
	log.Printf("starting upgrade process")
	pkt := &klbslog.Packet{Type: pktTakeover}
	pkt.SendTo(c)

	for {
		pkt := &klbslog.Packet{}
		err := pkt.ReadFrom(c)
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return err
		}

		switch pkt.Type {
		case pktTakeoverListenFd:
			log.Printf("receiving main listening FD!")
			// receive the listening FD
			fd := pkt.FDs[0]
			defer fd.Close()
			l, err := net.FileListener(fd)
			if err != nil {
				return fmt.Errorf("failed to grab fd: %w", err)
			}
			l2, ok := l.(*net.UnixListener)
			if !ok {
				return fmt.Errorf("listener is of type %T, not the expected *net.UnixListener", l)
			}
			// all good
			d.l = l2
			go d.loop()
		case pktTakeoverComplete:
			log.Printf("upgrade done!")
			// all good!
			return nil
		default:
			return fmt.Errorf("unexpected takeover packet: %x", pkt.Type)
		}
	}
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
	defer func() {
		if e := recover(); e != nil {
			log.Printf("panic on client: %s", e)
		}
		c.Close()
	}()

	for {
		pkt := &klbslog.Packet{}
		err := pkt.ReadFrom(c)
		if err != nil {
			if err != io.EOF {
				log.Printf("failed to read from client: %s", err)
			}
			return
		}

		switch pkt.Type {
		case pktTakeover:
			log.Printf("received takeover command, passing along all sockets")
			// go into takeover mode (stop listening, send all sockets to peer)
			f, err := d.l.File()
			if err != nil {
				log.Printf("failed to fetch listen socket: %s", err)
				return
			}
			defer f.Close()
			d.l.Close()
			(&klbslog.Packet{Type: pktTakeoverListenFd, FDs: []*os.File{f}}).SendTo(c)
			(&klbslog.Packet{Type: pktTakeoverComplete}).SendTo(c)
			shutdown.Shutdown()
			return
		default:
			log.Printf("unhandled packet from client: %x", pkt.Type)
		}
	}
}

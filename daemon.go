package main

import (
	"fmt"
	"net"
	"os"
)

type logdaemon struct {
	l *net.UnixListener
}

func (d *logdaemon) start() error {
	socket := "/run/logagent"
	if id := os.Getuid(); id > 0 {
		socket = fmt.Sprintf("/tmp/.logagent-%d.sock", id)
	}

	var err error
	d.l, err = net.ListenUnix("unix", &net.UnixAddr{Name: socket})
	if err != nil {
		return err
	}
	return nil
}

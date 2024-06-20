package main

const (
	pktTakeover         = 0xff00 // sent from daemon to daemon
	pktTakeoverListenFd = 0xff01 // includes listening socket (name in packet)
	pktTakeoverKmsgFd   = 0xff02 // fd for /dev/kmsg
	pktTakeoverClientFd = 0xff03 // fd for a connected client
	pktTakeoverStreamFd = 0xff04 // fd for a connected stream client
)

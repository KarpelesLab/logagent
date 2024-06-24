package main

import "github.com/KarpelesLab/shutdown"

func main() {
	shutdown.SetupSignals()

	d := &logdaemon{}
	d.start()

	shutdown.Wait()
}

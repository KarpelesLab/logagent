package main

func main() {
	// determine socket name
	d := &logdaemon{}
	d.start()

	d.loop()
}

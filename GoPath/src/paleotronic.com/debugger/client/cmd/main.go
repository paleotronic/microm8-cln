package main

import (
	"log"
	"time"

	debugclient "paleotronic.com/debugger/client"
)

func main() {
	d := debugclient.NewDebugClient(0, "localhost", "9502", "/api/websocket/debug")
	err := d.Connect()
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	time.Sleep(20 * time.Second)
	d.Disconnect()
}

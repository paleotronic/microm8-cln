package main

import (
	"paleotronic.com/ducktape"
	"paleotronic.com/ducktape/server"
	"paleotronic.com/log"
	"runtime"
)

var queues server.QueueMap

func main() {

	runtime.GOMAXPROCS(6 * runtime.NumCPU())

	queues = make(server.QueueMap)

	mapper := make(server.DuckHandlerMap)
	mapper["BRF"] = HandleBRF

	server := server.NewDuckTapeServer(":9988", mapper)

	server.RunALL()
}

func HandleBRF(c *ducktape.Client, s *server.DuckTapeServer, msg *ducktape.DuckTapeBundle) error {
	// stuff
	log.Printf("Got a BRF message from %s\n", c.Name)
	return nil
}

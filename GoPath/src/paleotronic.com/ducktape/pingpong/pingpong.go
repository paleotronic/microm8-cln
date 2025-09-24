package main

import (
	//"paleotronic.com/fmt"
	"paleotronic.com/ducktape/client"
	"paleotronic.com/ducktape/server"
	"runtime"
	"time"
)

var s *server.DuckTapeServer

func RunChat1() {

	c1 := client.NewDuckTapeClient("", "", "c1", "local")

	err := c1.ConnectLocal(s)
	if err != nil {
		panic(err)
	}

	c1.SendMessage("SUB", []byte("from-c2"), false)
	c1.SendMessage("SND", []byte("from-c1"), false)

	for {
		time.Sleep(1 * time.Second)
		c1.SendMessage("MSG", []byte("C1 says hi"), true)
	}

}

func RunChat2() {

	c2 := client.NewDuckTapeClient("", "", "c2", "local")

	err := c2.ConnectLocal(s)
	if err != nil {
		panic(err)
	}

	c2.SendMessage("SUB", []byte("from-c1"), false)
	c2.SendMessage("SND", []byte("from-c2"), false)

	for {
		time.Sleep(3 * time.Second)
		c2.SendMessage("MSG", []byte("C2 says boo"), true)
	}

}

func RunServer() {

	mapper := make(server.DuckHandlerMap)

	s = server.NewDuckTapeServer(":9988", mapper)

	s.RunALL()
}

func main() {
	runtime.GOMAXPROCS(10 * runtime.NumCPU())

	go RunServer()
	time.Sleep(1 * time.Second)
	go RunChat2()
	time.Sleep(1 * time.Second)
	go RunChat1()

	for {
	}

}

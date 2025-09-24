package postoffice

import (
	"paleotronic.com/ducktape/server"
)

func Run() {

	mapper := make(server.DuckHandlerMap)

	server := server.NewDuckTapeServer(":9988", mapper)

	server.RunALL()
}

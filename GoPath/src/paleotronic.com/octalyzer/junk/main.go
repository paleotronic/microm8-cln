package main

import "os/exec"
import "paleotronic.com/fmt"

var calcs = []string{
	"4+4",
	"2+2",
}

func main() {
    cmdName := "bc"
    cmd := exec.Command("sh", "-c", cmdName)
    stdout, _ := cmd.StdoutPipe()
    stdin, _ := cmd.StdinPipe()
    cmd.Start()
    for _,v := range calcs {
    oneByte := make([]byte, 8)
    stdin.Write([]byte(v+"\n"))
    ok := false
    for !ok {
        n, _ := stdout.Read(oneByte)
	if n > 0 {
		fmt.Println(string(oneByte))
		ok = true
		continue
	}
    }
	}
    stdin.Write([]byte("quit\n"))
}

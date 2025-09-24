package errorreport

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func GetOSVersion() string {
	cmd := exec.Command("uname", "-sr")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("getInfo:", err)
	}
	return out.String()
}

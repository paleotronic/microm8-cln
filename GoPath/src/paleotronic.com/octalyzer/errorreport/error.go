package errorreport

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"paleotronic.com/update"
)

func getCaller() string {
	var buffer [4096]byte
	size := runtime.Stack(buffer[:], false)
	lines := strings.Split(string(buffer[:size]), "\n")
	return lines[6]
}

func GracefulErrorReport(
	whatHappened string,
	err error,
) {
	msg := fmt.Sprintf(`
ERROR

%v

The reported error is: 

	%v

Please report this issue to help@paleotronic.com, including this message.

Architecture: %v
OS:           %v
Version:      %v
Cores:        %-d
Version:      %v
Hash:         %v
Date:         %v
Error time:   %v

Source:       %s

We will endeavour to investigate and resolve the issue.

The microM8 Team.

This message has also been written to 'error.log'.

(Press ENTER to close this window)
			`,
		whatHappened,
		err,
		runtime.GOARCH,
		runtime.GOOS,
		GetOSVersion(),
		runtime.NumCPU(),
		update.PERCOL8_BUILD,
		update.PERCOL8_GITHASH,
		update.PERCOL8_DATE,
		time.Now(),
		getCaller(),
	)

	fmt.Println(msg)
	f, _ := os.Create("error.log")
	defer f.Close()
	f.WriteString(msg)

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(1)
}

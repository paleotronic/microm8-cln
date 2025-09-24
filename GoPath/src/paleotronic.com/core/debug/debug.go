package debug

import (
	"time"
	"paleotronic.com/fmt"
	"os"
	"paleotronic.com/files"
)

const (
	FORMAT = "[%08d.%06d.%02d] %5s %s\r\n"
	MAXLINES = 100
)

var DEBUG_ON bool = false
var file *os.File
var buffer []string

func GetLogFilename() string {

	ds := time.Now().Format("20060102_150405")

	return files.GetUserDirectory(files.BASEDIR)+"/super8_debug_"+ds+".log"

}

func SetDebug( b bool ) {
	if b {
		StartLogging()
	} else {
		StopLogging()
	}
	DEBUG_ON = b
}

func Flush() {

	if file != nil {

		for _, l := range buffer {
			_, _ = file.WriteString(l)
		}

		buffer = make([]string, 0)

		_ = file.Close()

		file = nil
	}

}

func FlushOpen() {
	if file != nil {

		for _, l := range buffer {
			_, _ = file.WriteString(l)
		}

		buffer = make([]string, 0)
	}
}

func StartLogging() {

	Flush()

	var e error

	file, e = os.Create(GetLogFilename())
	if e != nil {
		panic(e)
	}
}

func StopLogging() {
	Flush()
}

func Log( msSince int64, line int64, stmt int64, component string, message string ) {
	if !DEBUG_ON {
		return
	}

	chunk := fmt.Sprintf(FORMAT, msSince, line, stmt, component, message)

	if file != nil {
		buffer = append(buffer, chunk)
		if len(buffer) > MAXLINES {
			FlushOpen()
		}
	}
}

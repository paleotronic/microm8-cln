package log

import (
	"time"
	"os"
	"bufio"
	"runtime"
	"log"
	"fmt"
)

var logfile *os.File
var logbuff *bufio.Writer

const SHOWCALLER = true
const SILENT = true

func init() {
	if SILENT {
		return
	}
	logname := os.Args[0]+".log"

	f, err := os.Create(logname)
	if err != nil {
		log.Fatalln(err)
	}
	logbuff = bufio.NewWriterSize(f, 4096)
}

func formatStr( format string, v ...interface{} ) string {
	myfmt := "%s [%s] " +  format

	caller := "**"
		if SHOWCALLER {
		pc, _, _, ok := runtime.Caller(2)
		if ok {
			c := runtime.FuncForPC(pc)
			caller = c.Name()
		}
	}

	t := time.Now().String()

	if format == "" {
		return fmt.Sprintf("%s [%s] %v", t, caller, fmt.Sprint(v)) + "\r\n"
	}

	return fmt.Sprintf( myfmt, t, caller, v ) + "\r\n"
}

func Print(v ...interface{}) {
	//
		if SILENT {
		return
	}
	_, _ = logbuff.WriteString( formatStr("", v...) )
}

func Println(v ...interface{}) {
	//
	if SILENT {
		return
	}
	_, _ = logbuff.WriteString( formatStr("", v...) )
}

func Printf(format string, v ...interface{}) {
	//
	if SILENT {
		return
	}
	_, _ = logbuff.WriteString( formatStr(format, v...) )
}

func Fatal(v ...interface{}) {
	//
	if SILENT {
		return
	}
	_, _ = logbuff.WriteString( formatStr("", v...) )
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	//
	if SILENT {
		return
	}
	_, _ = logbuff.WriteString( formatStr("", v...) )
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	//
	if SILENT {
		return
	}
	_, _ = logbuff.WriteString( formatStr(format, v...) )
	os.Exit(1)
}

func Panic(v ...interface{}) {
	//
	if SILENT {
		panic(v)
	}
	_, _ = logbuff.WriteString( formatStr("", v...) )
	b := make([]byte, 16384)
	n := runtime.Stack(b, false)
	logbuff.Write(b[0:n])
	logbuff.Flush()
	os.Exit(1)
}

func Panicln(v ...interface{}) {
	//
	if SILENT {
		panic(v)
	}
	_, _ = logbuff.WriteString( formatStr("", v...) )
	b := make([]byte, 16384)
	n := runtime.Stack(b, false)
	logbuff.Write(b[0:n])
	logbuff.Flush()
	os.Exit(1)
}

func Panicf(format string, v ...interface{}) {
	//
	if SILENT {
		panic(v)
	}
	_, _ = logbuff.WriteString( formatStr(format, v...) )
	b := make([]byte, 16384)
	n := runtime.Stack(b, false)
	logbuff.Write(b[0:n])
	logbuff.Flush()
	os.Exit(1)
}

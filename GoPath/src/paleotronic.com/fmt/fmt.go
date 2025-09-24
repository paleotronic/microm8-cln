package fmt

import (
	"fmt"

	"paleotronic.com/core/settings"
)

func Print(v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Print(v...)
}

func Println(v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Println(v...)
}

func RPrintln(v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Println(v...)
}

func RPrintf(format string, v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Printf(format, v...)
}

func Printf(format string, v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Printf(format, v...)
}

func Sprintf(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}

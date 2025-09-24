package utils

import (
	"strings"
)

type TINISection map[string]string

type TINIFile struct {
	FileName string
	Data     map[string]TINISection
}

func (this *TINIFile) WriteString(section string, key string, value string) {
	shm, exists := this.Data[section]

	if !exists {
		shm = make(TINISection)
		this.Data[section] = shm
	}

	shm[key] = value
}

func (this *TINIFile) ReadString(section string, key string, value string) string {
	shm, exists := this.Data[section]

	if !exists {
		return value
	}

	s, exists := shm[key]

	if !exists {
		return value
	}

	return s
}

func NewTINIFileBytes(data []byte) *TINIFile {

	this := &TINIFile{}
	this.FileName = "memory"

	lines := strings.Split(string(data), "\n")

	//for _, l := range lines
	//	System.Out.Println("INI: "+l);

	this.Data = make(map[string]TINISection)

	section := ""

	for _, s := range lines {
		s = strings.Trim(s, " \r")
		if len(s) > 0 {
			if Copy(s, 1, 1) == "[" {
				section = Copy(s, 2, len(s)-2)
			} else if Copy(s, 1, 1) == ";" {
				// ignore
			} else if Pos("=", s) > 0 {
				//System.Out.Println("== line: "+s);
				parts := strings.Split(s, "=")
				if len(parts) > 1 {
					this.WriteString(section, parts[0], parts[1])
				}
			}
		}
	}

	return this
}

func NewTINIFile(filename string) *TINIFile {
	this := &TINIFile{}
	this.FileName = filename

	lines, err := ReadTextFile(filename)

	//for _, l := range lines
	//	System.Out.Println("INI: "+l);

	this.Data = make(map[string]TINISection)

	if err != nil {
		return this
	}

	section := ""

	for _, s := range lines {
		s = strings.Trim(s, " ")
		if len(s) > 0 {
			if Copy(s, 1, 1) == "[" {
				section = Copy(s, 2, len(s)-2)
			} else if Copy(s, 1, 1) == ";" {
				// ignore
			} else if Pos("=", s) > 0 {
				//System.Out.Println("== line: "+s);
				parts := strings.Split(s, "=")
				if len(parts) > 1 {
					this.WriteString(section, parts[0], parts[1])
				}
			}
		}
	}

	return this
}

func (ini *TINIFile) String() string {

	out := ""

	for sectname, sdata := range ini.Data {
		out += "[" + sectname + "]\r\n"
		for k, v := range sdata {
			out += k + "=" + v + "\r\n"
		}
		out += "\r\n"
	}

	return out

}

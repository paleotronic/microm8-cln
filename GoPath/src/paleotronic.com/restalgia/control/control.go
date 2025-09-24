package control

import (
	"os"
	"sort"
	"strconv"
	"strings"

	"paleotronic.com/fmt"

	"paleotronic.com/restalgia"
)

type ShellCommandContext int

const (
	SCCNone ShellCommandContext = 1 << iota
	SCCCommand
	SCCAny = SCCCommand
)

var context string

type shellCommand struct {
	Name             string
	Description      string
	MinArgs, MaxArgs int
	Code             func(mixer *restalgia.Mixer, args []string) int
	NeedsMount       bool
	Text             []string
	Context          ShellCommandContext
}

var CommandList map[string]*shellCommand

func init() {
	CommandList = map[string]*shellCommand{
		"quit": &shellCommand{
			Name:        "quit",
			Description: "Exit the shell",
			MinArgs:     0,
			MaxArgs:     1,
			Code:        shellQuit,
			NeedsMount:  false,
			Context:     SCCCommand,
			Text: []string{
				"quit",
				"",
				"Exit this shell",
			},
		},
		"show": &shellCommand{
			Name:        "show",
			Description: "List current structure",
			MinArgs:     0,
			MaxArgs:     0,
			Code:        shellShow,
			NeedsMount:  false,
			Context:     SCCCommand,
			Text: []string{
				"show",
				"",
				"show current param structure",
			},
		},
		"get": &shellCommand{
			Name:        "get",
			Description: "Get field value",
			MinArgs:     1,
			MaxArgs:     1,
			Code:        shellGet,
			NeedsMount:  false,
			Context:     SCCCommand,
			Text: []string{
				"get",
				"",
				"Get a field value",
			},
		},
		"use": &shellCommand{
			Name:        "use",
			Description: "Set get/set target",
			MinArgs:     0,
			MaxArgs:     1,
			Code:        shellContext,
			NeedsMount:  false,
			Context:     SCCCommand,
			Text: []string{
				"use",
				"",
				"Set get/set target",
			},
		},
		"set": &shellCommand{
			Name:        "set",
			Description: "Set field value",
			MinArgs:     2,
			MaxArgs:     2,
			Code:        shellSet,
			NeedsMount:  false,
			Context:     SCCCommand,
			Text: []string{
				"set",
				"",
				"Set a field value",
			},
		},
		"adjust": &shellCommand{
			Name:        "adjust",
			Description: "Adjust field value",
			MinArgs:     2,
			MaxArgs:     2,
			Code:        shellAdjust,
			NeedsMount:  false,
			Context:     SCCCommand,
			Text: []string{
				"adjust",
				"",
				"Adjust a field value",
			},
		},
		"help": &shellCommand{
			Name:        "help",
			Description: "Shows this help",
			MinArgs:     0,
			MaxArgs:     1,
			Code:        shellHelp,
			NeedsMount:  false,
			Context:     SCCCommand,
			Text: []string{
				"help <command>",
				"",
				"Display specific help for command or list of commands",
			},
		},
	}
}

func smartSplit(line string) (string, []string) {

	var out []string

	var inqq bool
	var lastEscape bool
	var chunk string

	add := func() {
		if chunk != "" {
			out = append(out, chunk)
			chunk = ""
		}
	}

	for _, ch := range line {
		switch {
		case ch == '"':
			inqq = !inqq
			add()
		case ch == ' ':
			if inqq || lastEscape {
				chunk += string(ch)
			} else {
				add()
			}
			lastEscape = false
		case ch == '\\' && !inqq:
			lastEscape = true
		default:
			chunk += string(ch)
		}
	}

	add()

	if len(out) == 0 {
		return "", out
	}

	return out[0], out[1:]
}

func shellQuit(mixer *restalgia.Mixer, args []string) int {
	return 999
}

func shellShow(mixer *restalgia.Mixer, args []string) int {

	q := restalgia.QueryObjectTree("mixer", mixer)

	var keys []string
	for k, _ := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		shellWriteln("%s", k)
	}

	return 0
}

func shellGet(mixer *restalgia.Mixer, args []string) int {

	path := args[0]

	if context != "" {
		path = context + "." + path
	}

	object, kind, _, exists := restalgia.ResolveField(path, mixer)
	if !exists {
		shellWriteln("Object %s not found", path)
	}

	shellWriteln("Type: %s, Content: %v", kind, object)

	return 0
}

func shellContext(mixer *restalgia.Mixer, args []string) int {

	if len(args) == 0 {
		context = ""
		return 0
	}

	path := args[0]

	_, _, attrinfo, exists := restalgia.ResolveField(path, mixer)
	if !exists || attrinfo != nil {
		shellWriteln("Object %s not found", path)
	}

	context = path

	return 0
}

func shellSet(mixer *restalgia.Mixer, args []string) int {

	path := args[0]
	value := args[1]

	if context != "" {
		path = context + "." + path
	}

	object, kind, attrinfo, exists := restalgia.ResolveField(path, mixer)
	if !exists || attrinfo == nil {
		shellWriteln("Attribute %s not found", path)
	}

	shellWriteln("Type: %s, Content: %v", kind, object)

	switch kind {
	case "bool":
		b, e := strconv.ParseBool(value)
		if e != nil {
			shellWriteln("\"%s\" not a valid %s", value, kind)
			return 0
		}
		attrinfo.Set(0, b)
	case "string":
		attrinfo.Set(0, value)
	case "float32":
		b, e := strconv.ParseFloat(value, 32)
		if e != nil {
			shellWriteln("\"%s\" not a valid %s", value, kind)
			return 0
		}
		attrinfo.Set(0, float32(b))
	case "float64":
		b, e := strconv.ParseFloat(value, 64)
		if e != nil {
			shellWriteln("\"%s\" not a valid %s", value, kind)
			return 0
		}
		attrinfo.Set(0, b)
	case "int":
		b, e := strconv.ParseInt(value, 0, 64)
		if e != nil {
			shellWriteln("\"%s\" not a valid %s", value, kind)
			return 0
		}
		attrinfo.Set(0, int(b))
	case "WAVEFORM":
		attrinfo.Set(0, restalgia.StringToWAVEFORM(value))
	default:
		shellWriteln("unrecognised type: %s", kind)
	}

	return 0
}

func shellAdjust(mixer *restalgia.Mixer, args []string) int {

	path := args[0]
	value := args[1]

	if context != "" {
		path = context + "." + path
	}

	object, kind, attrinfo, exists := restalgia.ResolveField(path, mixer)
	if !exists || attrinfo == nil {
		shellWriteln("Attribute %s not found", path)
	}

	shellWriteln("Type: %s, Content: %v", kind, object)

	switch kind {
	case "float32":
		b, e := strconv.ParseFloat(value, 32)
		if e != nil {
			shellWriteln("\"%s\" not a valid %s", value, kind)
			return 0
		}
		attrinfo.Set(0, object.(float32)+float32(b))
	case "float64":
		b, e := strconv.ParseFloat(value, 64)
		if e != nil {
			shellWriteln("\"%s\" not a valid %s", value, kind)
			return 0
		}
		attrinfo.Set(0, object.(float64)+b)
	case "int":
		b, e := strconv.ParseInt(value, 0, 64)
		if e != nil {
			shellWriteln("\"%s\" not a valid %s", value, kind)
			return 0
		}
		attrinfo.Set(0, object.(int)+int(b))
	default:
		shellWriteln("unrecognised type: %s", kind)
	}

	return 0
}

func shellHelp(mixer *restalgia.Mixer, args []string) int {

	if len(args) == 0 {
		keys := make([]string, 0)
		for k, _ := range CommandList {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			info := CommandList[k]
			fmt.Printf("%-10s %s\n", info.Name, info.Description)
		}
	} else {
		command := strings.ToLower(args[0])
		if details, ok := CommandList[command]; ok {
			if details.Text != nil {
				for _, l := range details.Text {
					fmt.Println(l)
				}
			} else {
				os.Stderr.WriteString("No help available for " + command)
			}
		} else {
			os.Stderr.WriteString("No help available for " + command)
		}
	}

	return 0
}

func ShellProcess(mixer *restalgia.Mixer, line string) int {
	line = strings.TrimSpace(line)

	verb, args := smartSplit(line)

	if verb != "" {
		verb = strings.ToLower(verb)
		command, ok := CommandList[verb]
		if ok {
			fmt.Println()
			var cok = true
			if command.MinArgs != -1 {
				if len(args) < command.MinArgs {
					os.Stderr.WriteString(fmt.Sprintf("%s expects at least %d arguments\n", verb, command.MinArgs))
					cok = false
				}
			}
			if command.MaxArgs != -1 {
				if len(args) > command.MaxArgs {
					os.Stderr.WriteString(fmt.Sprintf("%s expects at most %d arguments\n", verb, command.MaxArgs))
					cok = false
				}
			}
			if cok {
				r := command.Code(mixer, args)
				fmt.Println()
				return r
			}
			return -1
		}
		os.Stderr.WriteString(fmt.Sprintf("Unrecognized command: %s\n", verb))
		return -1
	}

	return 0
}

func shellWriteln(format string, args ...interface{}) {
	fmt.Printf(format+"\r\n", args...)
}

func shellWrite(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

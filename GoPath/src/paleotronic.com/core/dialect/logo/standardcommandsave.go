package logo

import (
	"errors"
	"strings" //	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files" //"errors"
)

type StandardCommandSAVE struct {
	dialect.Command
}

func (this *StandardCommandSAVE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	tokens, e := this.Expect(caller, tokens, 1)
	if e != nil {
		return result, e
	}

	if tokens.Size() < 1 {
		return result, errors.New("Need filename")
	}

	// Got a filename
	filename := tokens.Get(0).Content

	if !strings.HasSuffix(filename, ".lgo") {
		filename += ".lgo"
	}

	if caller.GetWorkDir() != "" && !strings.HasPrefix(filename, "/") {
		filename = strings.Trim(caller.GetWorkDir(), "/") + "/" + filename
	}

	d := this.Command.D.(*DialectLogo)

	lines := d.Driver.GetWorkspaceBody(true, true)

	s := strings.Join(lines, "\r\n")
	data := []byte(s)

	e = files.WriteBytesViaProvider(files.GetPath(filename), files.GetFilename(filename), data)

	/* enforce non void return */
	return result, e

}

func tokenToString(t *types.Token, current string, inList bool) string {

	if t.List != nil {
		current += "["
		for i, tt := range t.List.Content {
			if i > 0 {
				current += " "
			}
			current = tokenToString(tt, current, true)
		}
		current += "]"
	} else {
		if !inList {
			current += "\"" + t.Content
		} else {
			current += t.Content
		}
	}

	return current

}

func (this *StandardCommandSAVE) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}

func getDataDef(caller interfaces.Interpretable, t *types.Token, buffer string) string {

	if t.Type == types.LIST {
		buffer = buffer + "["
		for i, tt := range t.List.Content {
			if i > 0 {
				buffer = buffer + " "
			}
			getDataDef(caller, tt, buffer)
		}
		buffer = buffer + "]"
	} else {
		if t.Type == types.WORD {
			buffer = buffer + "\"" + t.Content
		} else {
			buffer = buffer + t.Content
		}
	}

	return buffer
}

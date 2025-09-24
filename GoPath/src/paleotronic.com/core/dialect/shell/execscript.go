package shell

import (
	"errors"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
)

type CommandExec struct {
	dialect.Command
}

func (this *CommandExec) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	// Exec "scriptname"
	if tokens.Size() != 1 {
		return 0, errors.New("exec script filename required")
	}
	
	tmp := tokens.Get(0).Content
	f := files.GetFilename(tmp)
	p := files.GetPath(tmp)
	
	if files.ExistsViaProvider( p, f ) {
		
//		b, e := files.ReadBytesViaProvider(p, f)
//		if e != nil {
//			return 0, e
//		}
		
		// got data, parse
		
	} else {
		return 0, errors.New("file not found: "+tmp)
	}

	return 0, nil

}

func (this *CommandExec) Syntax() string {

	/* vars */
	var result string

	result = "exec"

	/* enforce non void return */
	return result

}

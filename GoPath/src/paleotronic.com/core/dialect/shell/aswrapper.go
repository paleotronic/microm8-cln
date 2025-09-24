package shell

import (
	"errors"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
//	"paleotronic.com/utils"
)

type CommandWrapper struct {
	dialect.Command
	WrappedName string
	WrappedCommand interfaces.Commander
	WrappedParam []*types.Token
}

func (this *CommandWrapper) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* enforce non void return */
	
	if len(this.WrappedParam) != 0 {
		nt := *types.NewTokenList()
		
		for _, tt := range this.WrappedParam {
			if tt.Type == types.PLACEHOLDER {
				
				i := tt.AsInteger()
				
				if i == 0 {
					nt.Content = append(nt.Content, tokens.Content...)
				} else if i >= 1 && i <= tokens.Size() { 
					nt.Content =  append(nt.Content, tokens.Get(i-1))
				}
				
			} else {
				nt.Add( tt )
			}
		}
		
		tokens.Content = nt.Content
	}
	
	if this.WrappedCommand == nil {
		return 0, errors.New("NULL command error")
	}
	
	return this.WrappedCommand.Execute(env, caller, tokens, Scope, LPC)

}

func (this *CommandWrapper) Syntax() string {

	/* vars */
	var result string

	result = this.WrappedName

	/* enforce non void return */
	return result

}

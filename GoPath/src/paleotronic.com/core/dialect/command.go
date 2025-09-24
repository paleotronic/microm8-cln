// command.go
package dialect

import (
	"errors"

	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
)

type Command struct {
	interfaces.Commander
	Keyword       string
	NoTokens      bool
	ImmediateMode bool
	Cost          int64
	UseStates     bool // true if command uses state execution...
	D             interfaces.Dialecter
}

func NewCommand() *Command {
	return &Command{Keyword: "unknown", NoTokens: false, ImmediateMode: false, UseStates: false}
}

func (this *Command) SetD(d interfaces.Dialecter) {
	this.D = d
}

func (this *Command) HasNoTokens() bool {
	return this.NoTokens
}

func (this *Command) GetCost() int64 {
	return this.Cost
}

func (this *Command) BeforeRun(ent interfaces.Interpretable) {

}

func (this *Command) AfterRun(ent interfaces.Interpretable) {

}

func (this *Command) ImmediateModeOnly() bool {
	return this.ImmediateMode
}

func (this *Command) IsStateBased() bool {
	return this.UseStates
}

func (this *Command) IO() *Command {
	this.ImmediateMode = true
	return this
}

// Expect checks to see if we have too many parameters
func (this *Command) Expect(caller interfaces.Interpretable, tokens types.TokenList, max int) (types.TokenList, error) {
	flat, e := caller.GetDialect().ParseTokensForResult(caller, tokens)

	if e != nil {
		return tokens, e
	}

	var list = types.TokenList{Content: []*types.Token{flat}}
	if flat.Type == types.LIST {
		list = *flat.List
	}

	if list.Size() > max {
		return list, errors.New("I DON'T KNOW WHAT TO DO WITH " + list.Get(max).Content)
	}

	return list, nil

}

// Shimmy goodness

func StateInit(env interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, scope *types.Algorithm, LPC types.CodeRef) (int, types.EntitySubState, error) {
	return 0, types.ESS_DONE, nil
}

func StateExec(env interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, scope *types.Algorithm, LPC types.CodeRef) (int, types.EntitySubState, error) {
	return 0, types.ESS_DONE, nil
}

func StateDone(env interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, scope *types.Algorithm, LPC types.CodeRef) (int, types.EntitySubState, error) {
	return 0, types.ESS_DONE, nil
}

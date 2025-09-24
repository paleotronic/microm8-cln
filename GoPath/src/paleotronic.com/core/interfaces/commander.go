package interfaces

import (
	"paleotronic.com/core/types"
)

type Commander interface {
	BeforeRun(caller Interpretable)
	AfterRun(caller Interpretable)
	Syntax() string
	Execute(env *Producable, caller Interpretable, tokens types.TokenList, scope *types.Algorithm, LPC types.CodeRef) (int, error)
	HasNoTokens() bool
	ImmediateModeOnly() bool
	GetCost() int64
	IsStateBased() bool
	StateInit(env *Producable, caller Interpretable, tokens types.TokenList, scope *types.Algorithm, LPC types.CodeRef) (int, error)
	StateExec(env *Producable, caller Interpretable, tokens types.TokenList, scope *types.Algorithm, LPC types.CodeRef) (int, error)
	StateDone(env *Producable, caller Interpretable, tokens types.TokenList, scope *types.Algorithm, LPC types.CodeRef) (int, error)
	SetD(d Dialecter)
}

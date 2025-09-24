package interfaces

import (
	"paleotronic.com/core/types"
)

type Functioner interface {
	SetQuery(v bool)
	ValidateParams() (bool, error)
	FunctionExecute(params *types.TokenList) error
	FunctionParams() []types.TokenType
	GetName() string
	GetRaw() bool
	SetEntity(ent Interpretable)
	GetStack() *types.TokenList
	IsQuery() bool
	IsHidden() bool
	SetHidden(v bool)
	GetNamedParams() []string
	SetNamedParamsValues(tokens types.TokenList)
	GetNamedDefaults() []types.Token
	Syntax() string
	GetMinParams() int
	GetMaxParams() int
	Prototype() []string
	SetAllowMoreParams(b bool)
}

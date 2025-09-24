package interfaces

import (
	"paleotronic.com/core/types"
)

type DynaCoder interface {
	SeedParamsData(values types.TokenList, ent Interpretable)
	SeedParams(values types.TokenList, ent Interpretable)
	Parse(s string) error
	Init()
	HasParams() bool
	GetParamCount() int
	GetCode() *types.Algorithm
	GetRawCode() []string
	GetDialect() Dialecter
	SetDialect(dia Dialecter)
	GetFunctionSpec() (string, types.TokenList)
	SetHidden( b bool )
	IsHidden() bool
}

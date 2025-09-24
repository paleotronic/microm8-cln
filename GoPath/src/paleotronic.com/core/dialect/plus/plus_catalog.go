package plus

import (
	"paleotronic.com/core/dialect"
	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/hardware/control"
	"paleotronic.com/core/types"
	//	"paleotronic.com/utils"
)

type PlusCatalog struct {
	dialect.CoreFunction
}

func (this *PlusCatalog) FunctionExecute(params *types.TokenList) error {

	//	if e := this.CoreFunction.FunctionExecute(params); e != nil {
	//		return e
	//	}

	control.CatalogPresent(this.Interpreter)

	return nil

}

func (this *PlusCatalog) Syntax() string {

	/* vars */
	var result string

	result = "ALTCASE{n}"

	/* enforce non void return */
	return result

}

func (this *PlusCatalog) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusCatalog(a int, b int, params types.TokenList) *PlusCatalog {
	this := &PlusCatalog{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "VIDEOMODE"

	this.MinParams = 0
	this.MaxParams = 0

	return this
}

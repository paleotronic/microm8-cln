package plus

import (
//	"paleotronic.com/log"
	//"errors"
	"strings"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
//	"paleotronic.com/api"
	//"paleotronic.com/core/interfaces"
	"paleotronic.com/api"
	"paleotronic.com/fmt"
)

type PlusDisplayMOTD struct {
	dialect.CoreFunction
}

// params: 
// (1) hostname
// (2) name

func (this *PlusDisplayMOTD) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { 
		fmt.Println(e)
		return e 
	}
	
	fmt.Println("In display motd")

	motdtext := s8webclient.CONN.GetMOTD()
	
	if motdtext != "" {
	
		lines := strings.Split( motdtext, "\r\n" )
		//rows := apple2helpers.GetRows( this.Interpreter ) - 1
		cols := apple2helpers.GetColumns( this.Interpreter )
		
		//lc := 0
		
		for _, l := range lines {
			this.Interpreter.PutStr( wrap(l,cols,0) + "\r\n" )
		}
	
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))

	return nil
}

func (this *PlusDisplayMOTD) Syntax() string {

	/* vars */
	var result string

	result = "CONTROL{slot, target}"

	/* enforce non void return */
	return result

}

func (this *PlusDisplayMOTD) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusDisplayMOTD(a int, b int, params types.TokenList) *PlusDisplayMOTD {
	this := &PlusDisplayMOTD{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CONTROL"
	this.MinParams = 0
	this.MaxParams = 1

	return this
}

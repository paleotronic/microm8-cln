package plus

import (
	"strings"

	"paleotronic.com/fmt"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	//	"paleotronic.com/api"
	//	"paleotronic.com/utils"
	//"errors"
)

type PlusAppDBPrepare struct {
	dialect.CoreFunction
}

func (this *PlusAppDBPrepare) FunctionExecute(params *types.TokenList) error {

	var out string

	var e error

	if e = this.CoreFunction.FunctionExecute(params); e != nil {
		fmt.Println("Problem", e)
		return e
	}

	stmt := params.Shift().Content

	values := make([]string, 0)
	v := params.Shift()
	for v != nil {
		if v.Type == types.STRING {

			c := v.Content
			c = strings.Replace(c, "'", "\\'", -1)
			c = strings.Replace(c, ",", "\\,", -1)

			values = append(values,
				"'"+c+"'",
			)
		} else {
			values = append(values, v.Content)
		}
		v = params.Shift()
	}

	out = stmt
	for i, val := range values {
		out = strings.Replace(out, fmt.Sprintf(":%d", i+1), val, -1)
	}

	this.Stack.Push(types.NewToken(types.STRING, out))

	return nil
}

func (this *PlusAppDBPrepare) Syntax() string {

	/* vars */
	var result string

	result = "appdb.resultcount{stmtid}"

	/* enforce non void return */
	return result

}

func (this *PlusAppDBPrepare) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)

	/* enforce non void return */
	return result

}

func NewPlusAppDBPrepare(a int, b int, params types.TokenList) *PlusAppDBPrepare {
	this := &PlusAppDBPrepare{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "appdb.prepare{statement,args...}"

	this.MaxParams = 10
	this.MinParams = 2

	this.Hidden = true

	return this
}

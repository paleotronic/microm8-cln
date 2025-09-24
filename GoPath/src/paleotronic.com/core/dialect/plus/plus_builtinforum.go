package plus

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/octalyzer/ui/forumtool"
)

type PlusBuiltinForum struct {
	dialect.CoreFunction
}

func (this *PlusBuiltinForum) FunctionExecute(params *types.TokenList) error {

	//if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

	//	backend.REBOOT_NEEDED = true
	app := forumtool.NewForumApp(this.Interpreter, 1)
	app.Run()

	return nil
}

func (this *PlusBuiltinForum) Syntax() string {

	/* vars */
	var result string

	result = "REBOOT{}"

	/* enforce non void return */
	return result

}

func (this *PlusBuiltinForum) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)

	/* enforce non void return */
	return result

}

func NewPlusBuiltinForum(a int, b int, params types.TokenList) *PlusBuiltinForum {
	this := &PlusBuiltinForum{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "REBOOT"

	return this
}

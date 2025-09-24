package plus

import (
	"errors"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/settings"

	//	"paleotronic.com/core/memory"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type PlusVMSpeaker struct {
	dialect.CoreFunction
}

func (this *PlusVMSpeaker) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	st := params.Shift()
	slotid := st.AsInteger() - 1
	en := params.Shift()
	target := en.AsInteger() - 1

	if slotid < 0 || slotid >= memory.OCTALYZER_NUM_INTERPRETERS {
		return errors.New("invalid slot")
	}

	if target < 0 || target >= memory.OCTALYZER_NUM_INTERPRETERS {
		return errors.New("invalid target slot")
	}

	if target != slotid {
		settings.SpeakerRedirects[slotid] = &settings.SpeakerRedirect{
			target,
			1,
		}
	} else {
		settings.SpeakerRedirects[slotid] = nil
	}

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusVMSpeaker) Syntax() string {

	/* vars */
	var result string

	result = "SETCLASSICCPU{}"

	/* enforce non void return */
	return result

}

func (this *PlusVMSpeaker) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusVMSpeaker(a int, b int, params types.TokenList) *PlusVMSpeaker {
	this := &PlusVMSpeaker{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "SETCLASSICCPU"

	this.MaxParams = 2
	this.MinParams = 2

	return this
}

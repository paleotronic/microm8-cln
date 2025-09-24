package plus

import (
	"paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/log" //"errors"
	//	"strings"
	//"paleotronic.com/files"
	"paleotronic.com/utils" //"paleotronic.com/core/interfaces"
)

const baseport = 8580
const basehost = "paleotronic.com"
const baseslot = 0

type PlusConnect struct {
	dialect.CoreFunction
}

// params:
// (1) hostname
// (2) port
// (3) slot
// (4) 0 = current slot, 1 = allocate slot

func (this *PlusConnect) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil {
		return e
	}

	//fmt.Printf("share.connect{} called with params: %v, %v\n", params.AsString())

	host := params.Shift().Content
	iport := params.Shift().AsInteger()
	port := ":" + utils.IntToStr(iport)
	rslot := (iport - 8580) % 8
	slotid := this.Interpreter.GetMemIndex()
	if params.Size() > 0 {
		slotid = params.Shift().AsInteger() - 1
	}

	key, err := s8webclient.CONN.GetRemoteToken(host, port)
	if err != nil {
		this.Interpreter.PutStr(err.Error() + "\r\n")
		return err
	}
	log.Printf("key generated %v", key)

	//~ if !newneeded {
	e := this.Interpreter.GetProducer().GetInterpreter(slotid)
	for e.GetChild() != nil {
		e = e.GetChild()
	}

	e.ConnectRemote(host, port, rslot)
	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(e.GetMemIndex()+1)))

	log.Printf("Connected remote to vm #%d", e.GetMemIndex()+1)

	return nil
}

func (this *PlusConnect) Syntax() string {

	/* vars */
	var result string

	result = "CONNECT{address}"

	/* enforce non void return */
	return result

}

func (this *PlusConnect) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func NewPlusConnect(a int, b int, params types.TokenList) *PlusConnect {
	this := &PlusConnect{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
	this.Name = "CONNECT"
	this.MinParams = 2
	this.MaxParams = 3

	this.InitNamedParams(
		[]dialect.FunctionParamDef{
			dialect.FunctionParamDef{Name: "host", Default: *types.NewToken(types.STRING, "localhost")},
			dialect.FunctionParamDef{Name: "port", Default: *types.NewToken(types.NUMBER, "8580")},
			dialect.FunctionParamDef{Name: "vm", Default: *types.NewToken(types.NUMBER, "1")},
		},
	)

	return this
}

package dialect

import (
	"strings"

	"paleotronic.com/core/memory"

	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	"paleotronic.com/utils"
)

type CoreFunction struct {
	interfaces.Functioner
	Interpreter     interfaces.Interpretable
	Origin          interfaces.Interpretable
	Name            string
	Caller          types.CodeRef
	Return          types.CodeRef
	Raw             bool
	Hidden          bool
	Locals          types.VarManager
	Stack           *types.TokenList
	MinParams       int
	MaxParams       int
	NamedParams     []string               // named params (x, y) in default order
	NamedDefaults   []types.Token          // default values for the function
	ValueMap        map[string]types.Token // Parameters
	Query           bool
	QueryVar        string
	NameSpace       string
	EvalSingleParam bool
	NoRedirect      bool
	AllowMoreParams bool
}

func (this *CoreFunction) RawParams() bool {

	return this.Raw
}

func (this *CoreFunction) IsHidden() bool {

	return this.Hidden
}

func (this *CoreFunction) ParseParams(tl types.TokenList) map[string]types.Token {

	tla := this.Interpreter.SplitOnTokenWithBrackets(tl, *types.NewToken(types.SEPARATOR, ""))
	vmap := make(map[string]types.Token)

	if this.Query && tl.Size() == 1 {

		this.QueryVar = this.NamedParams[0]
		if tl.Size() == 1 && tl.Get(0).Type == types.VARIABLE {
			v := this.Interpreter.ParseTokensForResult(tl)
			tl.Content[0] = &v
		}

		if tl.Size() == 1 && tl.Get(0).Type == types.NUMBER {
			this.QueryVar = this.NamedParams[tl.Get(0).AsInteger()%len(this.NamedParams)]
		}

		log.Printf("ParseParams(%s)\n", this.Interpreter.TokenListAsString(tl))
		log.Printf("Query Var: %v\n", this.QueryVar)

	} //else {

	log.Printf("ParseParams(%s)\n", this.Interpreter.TokenListAsString(tl))
	log.Printf("Query Context: %v\n", this.Query)

	for p, list := range tla {
		// Process each param context
		if list.Size() > 2 && list.Get(1).Type == types.ASSIGNMENT && list.Get(0).Type == types.VARIABLE {
			// XXX = YYY
			pname := strings.ToLower(list.Get(0).Content)
			sl := list.SubList(2, list.Size())

			var v types.Token

			if sl.Size() == 1 && !this.EvalSingleParam {
				v = *sl.Get(0)
			} else {
				v = this.Interpreter.ParseTokensForResult(*sl)
			}

			vmap[pname] = v
		} else if list.Size() > 0 {
			// positional
			pname := ""
			if p < len(this.NamedParams) {
				pname = strings.ToLower(this.NamedParams[p])
			}

			var v types.Token

			//if list.Size() == 1 && !this.EvalSingleParam {
			//	v = *list.Get(0)
			//} else {
			//fmt.RPrintf("func tokens: %s\n", list.AsString())
			v = this.Interpreter.ParseTokensForResult(list)
			//}

			if pname != "" {
				vmap[pname] = v
			}
		} else {
			log.Println("Ignore empty list")
		}
	}

	// Now Iterate over values looking for missing defaults
	for p, pname := range this.NamedParams {
		_, ok := vmap[strings.ToLower(pname)]
		if !ok {
			log.Printf("%d) Filling default for %s (%v)\n", p, pname, this.NamedDefaults[p])
			vmap[strings.ToLower(pname)] = this.NamedDefaults[p]
		}
	}

	log.Println(vmap)
	//}

	return vmap
}

func NewCoreFunction(a int, b int, params types.TokenList) *CoreFunction {
	this := &CoreFunction{}

	/* vars */
	this.Name = "unknown"
	this.MinParams = -1
	this.MaxParams = -1
	this.Raw = false
	//this.Locals = *types.NewVarMap(-1, nil)
	if params.Size() > 0 {
		this.Stack = &params
	} else {
		this.Stack = types.NewTokenList()
	}

	return this
}

func (this *CoreFunction) GetStack() *types.TokenList {
	return this.Stack
}

func (this *CoreFunction) GetEntity() interfaces.Interpretable {
	return this.Interpreter
}

func (this *CoreFunction) SetEntity(ent interfaces.Interpretable) {
	this.Interpreter = ent
}

func (this *CoreFunction) SetQuery(v bool) {
	this.Query = v
}

func (this *CoreFunction) SetHidden(v bool) {
	this.Hidden = v
}

func (this *CoreFunction) IsQuery() bool {
	return this.Query
}

func (this *CoreFunction) GetName() string {
	return this.Name
}

func (this *CoreFunction) GetRaw() bool {
	return this.Raw
}

func (this *CoreFunction) SetAllowMoreParams(b bool) {
	this.AllowMoreParams = b
}

func (this *CoreFunction) ValidateParams() (bool, error) {

	/* vars */
	var result bool
	var p int

	result = false
	params := this.FunctionParams()

	if this.MinParams == -1 {
		this.MinParams = len(params)
	}
	if this.MaxParams == -1 {
		this.MaxParams = len(params)
	}

	if this.Stack.Size() < this.MinParams {
		return false, exception.NewESyntaxError("Function " + this.Name + " expects at least " + utils.IntToStr(this.MinParams) + " parameter(s).")
	}

	if this.Stack.Size() > this.MaxParams && !this.AllowMoreParams {
		return false, exception.NewESyntaxError("Function " + this.Name + " expects at most " + utils.IntToStr(this.MaxParams) + " parameter(s).")
	}

	p = 1
	for _, tok := range this.Stack.Content {
		if tok == nil {
			panic("Token was nil")
		}
		if (tok.Type != params[p-1]) && (params[p-1] != types.NOP) && !((tok.Type == types.INTEGER) && (params[p-1] == types.NUMBER)) {
			//s = params[p - 1].ToString()
			//s = s.ReplaceAll("tt", "")
			return result, exception.NewESyntaxError("Function " + this.Name + " expects parameter " + utils.IntToStr(p) + " to be " + params[p-1].String())
		}
		if p < len(params) {
			p++
		}
	}

	result = true

	/* enforce non void return */
	return result, nil

}

func (this *CoreFunction) FunctionExecute(params *types.TokenList) error {

	/* vars */
	this.Stack.Clear()
	for _, tok := range params.Content {
		this.Stack.Push(tok)
	}

	if !this.Raw {
		b, e := this.ValidateParams()
		if !b {
			return e
		}
	} else {
		//fmt.Printf("ParseParams called for (%s)\n", params.AsString())
		this.ValueMap = this.ParseParams(*params)
	}

	this.Origin = this.Interpreter

	if !this.NoRedirect {
		targetslot := this.Interpreter.GetMemoryMap().IntGetTargetSlot(this.Interpreter.GetMemIndex())
		if targetslot&128 != 0 {
			actual := targetslot & 127
			if actual != this.Interpreter.GetMemIndex() {
				//fmt.Printf("Redirecting actions to slot %d (from %d)\n", actual, this.Interpreter.GetMemIndex())
				this.Interpreter = this.Interpreter.GetProducer().GetInterpreter(actual % memory.OCTALYZER_NUM_INTERPRETERS)
			}
		}
	}

	if this.Raw {
		params.Clear()
		this.Stack.Clear()
		for _, name := range this.NamedParams {
			v := this.ValueMap[strings.ToLower(name)]
			params.Push(&v)
			this.Stack.Push(&v)
		}
	}

	return nil

}

func (this *CoreFunction) FunctionParams() []types.TokenType {

	/* vars */
	result := make([]types.TokenType, 0)

	result = append(result, types.NOP)

	/* enforce non void return */
	return result

}

func (this *CoreFunction) GetNamedParams() []string {
	return this.NamedParams
}

func (this *CoreFunction) GetNamedDefaults() []types.Token {
	return this.NamedDefaults
}

func (this *CoreFunction) Syntax() string {
	return ""
}

func (this *CoreFunction) GetMinParams() int {
	return this.MinParams
}

func (this *CoreFunction) GetMaxParams() int {
	return this.MaxParams
}

func (this *CoreFunction) SetNamedParamsValues(tokens types.TokenList) {

	this.ValueMap = this.ParseParams(tokens)

}

// Generate function prototype
func (this *CoreFunction) Prototype() []string {

	params := []string(nil)
	return params
}

type FunctionParamDef struct {
	Name    string
	Default types.Token
}

func (this *CoreFunction) InitNamedParams(m []FunctionParamDef) {

	this.NamedParams = []string(nil)
	this.NamedDefaults = []types.Token(nil)
	for _, item := range m {
		this.NamedParams = append(this.NamedParams, item.Name)
		this.NamedDefaults = append(this.NamedDefaults, item.Default)
	}
	this.Raw = true

}

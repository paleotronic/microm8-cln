package dialect

import (
	//	"paleotronic.com/fmt"

	"strings"

	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type DynaCode struct {
	Code        *types.Algorithm
	Name        string
	Dialect     interfaces.Dialecter
	ReturnOnEnd bool
	Hidden      bool
	Params      types.TokenList
	ParamCount  int
	//Dialect     Dialecter
	RawCode []string
}

func (this *DynaCode) SetDialect(dia interfaces.Dialecter) {
	this.Dialect = dia
}

func (this *DynaCode) GetDialect() interfaces.Dialecter {
	return this.Dialect
}

func (this *DynaCode) GetCode() *types.Algorithm {
	return this.Code
}

func (this *DynaCode) GetRawCode() []string {
	return this.RawCode
}

func (this *DynaCode) GetFunctionSpec() (string, types.TokenList) {
	return this.Name, this.Params
}

func (this *DynaCode) GetParamCount() int {
	return this.Params.Size()
}

func (this *DynaCode) SeedParamsData(values types.TokenList, ent interfaces.Interpretable) {

	/* vars */
	//Token vtok;
	var ptok *types.Token
	var tl types.TokenList
	//var e error

	//fmt.Println("In SeedParamsData()")

	// if values.Size() > 0 {

	// 	fmt.Println("VALUES:", values.AsString())

	// 	ptok, e = this.Dialect.ParseTokensForResult(ent, values)
	// 	if e != nil {
	// 		panic(e)
	// 	}

	// 	if ptok.Type == types.LIST && this.GetParamCount() > 1 {
	// 		tl = *ptok.List
	// 		fmt.Println("AFTER PTFR: ", tl.AsString())
	// 	} else {
	// 		tl = *types.NewTokenList()
	// 		tl.Push(ptok)
	// 	}

	// }

	tl = values

	// System.Out.Println( "DynaCode.SeedParams called with "+values.Size()+"params.");

	//fmt.Printf("~~~~~~~~~~~~~~~~~~~~~~> %s PARAMVALS: %v\n", this.Name, ent.TokenListAsString(tl))

	s := this.Params.SubList(0, this.Params.Size())

	//fmt.Printf("~~~~~~~~~~~~~~~~~~~~~~> %s PARAMS: %v\n", this.Name, ent.TokenListAsString(*s))

	for _, vtok := range s.Content {
		ptok = tl.Shift()
		//  System.Out.Println( "*** Going to seed local var "+vtok.GetContent()+" with "+ptok.GetContent() );
		//new entVar( LowerCase(vtok.Content), Newent( LowerCase(vtok.Content), vtSTRING, ptok.Content, true ) );
		//fmt.Printf("%s ->  %s\n", vtok.Content, ptok.Content)
		ent.SetData(strings.ToLower(vtok.Content), *ptok.Copy(), true)
	}

}

func NewDynaCode(n string) *DynaCode {
	this := &DynaCode{}

	this.Code = types.NewAlgorithm()
	this.Name = n
	this.Params = *types.NewTokenList()

	return this
}

func NewDynaCodeWithRootDia(n string, rootdia interfaces.Dialecter, fn string) *DynaCode {
	this := &DynaCode{}

	this.Code = types.NewAlgorithm()
	this.Name = n
	this.Dialect = rootdia
	this.Params = *types.NewTokenList()
	this.Init()

	f, err := utils.ReadTextFile(fn)
	if err == nil {
		for _, s := range f {
			s := strings.Trim(s, " ")
			if s != "" {
				this.Parse(s)
			}
		}
	}

	return this
}

func (this *DynaCode) Parse(s string) error {

	/* vars */
	var tl *types.TokenList
	//TokenList  cl;
	var cmdlist types.TokenListArray
	var tok *types.Token
	var lno int
	var ll types.Line
	var st types.Statement
	var pragma string

	// If we get a special directive starting with $/* */
	s = strings.Trim(s, " ")
	if (utils.Copy(s, 1, 2) == "${") && (utils.Copy(s, len(s), 1) == "}") {
		pragma = utils.Copy(s, 3, len(s)-3)
		if utils.Copy(pragma, 1, 8) == "include " {
			pragma = utils.Delete(pragma, 1, 8)
			f, _ := utils.ReadTextFile(pragma)
			for _, n := range f {
				n = strings.Trim(n, " ")
				if n != "" {
					this.Parse(n)
				}
			}
		}
		return nil
	}

	tl = this.Dialect.Tokenize(runestring.Cast(s))

	if tl.Size() == 0 {
		return exception.NewESyntaxError("Syntax Error")
	}

	tok = tl.Get(0)

	if tok.Type != types.NUMBER {
		tl.UnShift(types.NewToken(types.NUMBER, utils.IntToStr(this.Code.Size()+1)))
		tok = tl.Get(0)
	}

	if tok.Type == types.NUMBER {
		tok = tl.Shift()

		cmdlist = this.Dialect.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))

		lno = tok.AsInteger()
		ll = types.NewLine()

		for _, cl := range cmdlist {
			st = types.NewStatement()
			for _, ntok := range cl.Content {
				st.Push(ntok)
			}
			ll.Push(st)
		}

		this.Code.Put(lno, ll)

		return nil
	}

	return nil

}

func (this *DynaCode) Init() {
}

func (this *DynaCode) HasParams() bool {

	/* vars */
	var result bool

	result = (this.Params.Size() > 0)

	/* enforce non void return */
	return result

}

func (this *DynaCode) SetHidden(b bool) {
	this.Hidden = b
}

func (this *DynaCode) IsHidden() bool {
	return this.Hidden
}

func (this *DynaCode) HasToken(tt types.TokenType, content string) bool {
	for _, l := range this.Code.C {
		for _, s := range l {
			for _, tok := range s.Content {
				if tok.Type == tt && strings.ToLower(content) == strings.ToLower(tok.Content) {
					return true
				} else if tok.Type == types.LIST && tok.List != nil {
					if tok.List.IndexOfN(-1, tt, content) != -1 {
						return true
					}
				}
			}
		}
	}
	return false
}

func (this *DynaCode) SeedParams(values types.TokenList, ent interfaces.Interpretable) {

	/* vars */
	//Token vtok;
	var ptok *types.Token
	var tl types.TokenList

	//fmt.Println("In SeedParams()")

	ptok, _ = this.Dialect.ParseTokensForResult(ent, values)

	if ptok.Type == types.LIST {
		tl = *ptok.List
	} else {
		tl = *types.NewTokenList()
		tl.Push(ptok)
	}

	// System.Out.Println( "DynaCode.SeedParams called with "+values.Size()+"params.");

	for _, vtok := range this.Params.Content {
		ptok = tl.Shift()
		//System.Out.Println( "*** Going to seed local var "+vtok.GetContent()+" with "+ptok.GetContent() );
		//ent.CreateVar(strings.ToLower(vtok.Content), *types.NewVariableP(strings.ToLower(vtok.Content), types.VT_STRING, ptok.Content, true))
		_ = ent.GetLocal().CreateString(vtok.Content, ptok.Content)
		//fmt.Printf("Dynacode: Set %s to %s\n", vtok.Content, ptok.Content)
	}

}

func NewDynaCodeWithRootDiaS(n string, rootdia interfaces.Dialecter, code string, auto bool) *DynaCode {
	this := &DynaCode{}

	this.Code = types.NewAlgorithm()
	this.Name = n
	this.Dialect = rootdia
	this.Params = *types.NewTokenList()
	this.Init()

	count := 1

	this.RawCode = strings.Split(code, "\r\n")

	for _, s := range this.RawCode {
		s = strings.Trim(s, " ")
		if s != "" {
			if auto {
				this.Parse(utils.IntToStr(count) + " " + s)
			} else {
				this.Parse(s)
			}
		}
		count++
	}

	this.ReturnOnEnd = true

	return this
}

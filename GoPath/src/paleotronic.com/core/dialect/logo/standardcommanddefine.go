package logo

import (
	"errors"
	"paleotronic.com/core/dialect" //"paleotronic.com/core/hires"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/log"
	//	"paleotronic.com/fmt"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandDEFINE struct {
	dialect.Command
	Split bool
}

func (this *StandardCommandDEFINE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() < 2 {
		return result, nil
	}

	// DEFINE LAST FIRST :I LPUT LPUT WORD "" LAST FIRST :I [OP SETIT] [[]]

	// LPUT LPUT WORD "" LAST FIRST :I [OP SETIT] [[]]

	//fmt.Printf("%d tokens in the list\n", tokens.Size())

	var err error
	params, err := this.Command.D.(*DialectLogo).Driver.ParseExprRLCollapse(tokens.Copy(), false)
	if err != nil {
		return result, err
	}

	rest := params[len(params)-1]
	params = params[:len(params)-1]

	log.Printf("2nd param is: %s", tokenStr("", rest))

	if len(params) < 1 {
		return result, errors.New("NOT ENOUGH INPUTS TO DEFINE, EXPECT 2")
	}

	name := params[0].Content

	if rest.Type != types.LIST {
		return result, errors.New("Expect LIST as input")
	}

	if rest.List.Size() < 2 {
		return result, errors.New("Not enough inputs in LIST")
	}

	if rest.List.Get(0).Type != types.LIST {
		return result, errors.New("Expect first element of LIST to be LIST")
	}

	if rest.List.Get(1).Type != types.LIST {
		return result, errors.New("Expect second element of LIST to be LIST")
	}

	code := []string{}
	if rest.List.Size() > 1 {
		lines := rest.List.SubList(1, rest.List.Size())
		for _, l := range lines.Content {
			code = append(code, toString("", l))
		}
	}
	args := make([]string, rest.List.Get(0).List.Size())
	for i, ttt := range rest.List.Get(0).List.Content {
		args[i] = ttt.Content
	}

	d := this.Command.D.(*DialectLogo)
	_, err = d.Driver.StoreProc(
		d.Lexer,
		name,
		args,
		code,
	)

	/* enforce non void return */
	return result, err

}

func (this *StandardCommandDEFINE) Syntax() string {

	/* vars */
	var result string

	result = "GRAPHICS"

	/* enforce non void return */
	return result

}

func toString(in string, t *types.Token) string {

	var out string

	if t.Type == types.LIST {
		for i, tt := range t.List.Content {
			if i > 0 {
				out += " "
			}
			out = toString(out, tt)
		}
	} else {
		out += t.Content
	}

	return in + out

}

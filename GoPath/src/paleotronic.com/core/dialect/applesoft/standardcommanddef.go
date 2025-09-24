package applesoft

import (
	//"paleotronic.com/fmt"
	"strings"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/exception"
)

type StandardCommandDEF struct {
    dialect.Command
}

func (this *StandardCommandDEF) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

        /* vars */
        var result int


      result = 0

     /*
      DEF		FN			NAME  (		  X		  ) == <expr>
      kw		kw(nop)	var	  ob		var		cb		ass		<token list>
     var / *

     if (tokens.Size() < 7) {
      return types.NewESyntaxError("SYNTAX ERROR")
     }
	*/
	
		fn := tokens.Shift()
		if fn == nil || fn.Type != types.KEYWORD || strings.ToLower(fn.Content) != "fn" {
			return 0, exception.NewESyntaxError("SYNTAX ERROR")
		}
		
		name := tokens.Shift()
		if name == nil || name.Type != types.VARIABLE {
			return 0, exception.NewESyntaxError("SYNTAX ERROR")
		}
		
		ob := tokens.Shift() 
		if ob == nil || ob.Type != types.OBRACKET || ob.Content != "(" {
			return 0, exception.NewESyntaxError("SYNTAX ERROR")
		}
		
		varname := tokens.Shift()
		if varname == nil || varname.Type != types.VARIABLE {
			return 0, exception.NewESyntaxError("SYNTAX ERROR")
		}
		
		cb := tokens.Shift() 
		if cb == nil || cb.Type != types.CBRACKET || cb.Content != ")" {
			return 0, exception.NewESyntaxError("SYNTAX ERROR")
		}		
		
		eq := tokens.Shift() 
		if eq == nil || eq.Type != types.ASSIGNMENT || eq.Content != "=" {
			return 0, exception.NewESyntaxError("SYNTAX ERROR")
		}		
		
		tl := *types.NewTokenList()
		tl.Add( varname )
		
		maf := *interfaces.NewMultiArgumentFunction(tl, tokens)
		caller.SetMultiArgFunc(strings.ToLower(name.Content), maf)
		////fmt.Println("DEFMAP: ", caller.GetMultiArgFunc())

        /* enforce non void return */
        return result, nil

}

func (this *StandardCommandDEF) Syntax() string {

        /* vars */
        var result string

      result = "DEF"

        /* enforce non void return */
        return result

}


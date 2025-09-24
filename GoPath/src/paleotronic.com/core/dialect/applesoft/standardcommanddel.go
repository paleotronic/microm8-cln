package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/exception"
//    "paleotronic.com/fmt"
)

type StandardCommandDEL struct {
    dialect.Command
}

func (this *StandardCommandDEL) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

        /* vars */
        var result int
        //var vtok types.Token
        //var cr types.CodeRef
        var tla types.TokenListArray
        var tl types.TokenList
        //TokenList  i;
        var x int
        var y int
        var ii int


      result = 0

      tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ",") )
      tl = *types.NewTokenList()

      for _, i := range tla  {
        vtokk := caller.ParseTokensForResult( i )
        tl.Push(&vtokk)
      }

      if (tl.Size() == 0) {
        return result, exception.NewESyntaxError("DEL expects at least 1 parameter")
      }

      x = -1
      y = -1

      if (tl.Get(0).Type == types.NUMBER) || (tl.Get(0).Type == types.INTEGER) {
        x = tl.Get(0).AsInteger()
      }

      if ((tl.Size() > 1) && ( (tl.Get(tl.Size()-1).Type == types.NUMBER) || (tl.Get(tl.Size()-1).Type == types.INTEGER) ) ) {
        y = tl.Get(tl.Size()-1).AsInteger()
      }

      if ((tl.Size() == 1)) {
        y = x
      }

      if ((x == -1) || (y == -1) || (y < x)) {
        return result, exception.NewESyntaxError("SYNTAX ERROR")
      }

	  b := caller.GetCode()
      //fmt.Printf("DEL will remove between %d and %d\n", x, y)
      for ii=x;  ii <= y;  ii++ {
        if (b.ContainsKey(ii)) {
          //fmt.Printf("DEL removing line %d\n", ii)
          b.Remove(ii)
        }
      }
      caller.SetCode(b)


        /* enforce non void return */
        return result, nil

}

func (this *StandardCommandDEL) Syntax() string {

        /* vars */
        var result string

      result = "DEL x, y"

        /* enforce non void return */
        return result

}


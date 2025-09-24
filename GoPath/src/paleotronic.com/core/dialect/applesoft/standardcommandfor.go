package applesoft

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/core/exception"
	"strings"
)

type StandardCommandFOR struct {
    dialect.Command
}

func (this *StandardCommandFOR) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

        /* vars */
        var result int
        var eq_idx int
        var to_idx int
        var step_idx int
        var tvar types.Token
        //var tto types.Token
        var tlo types.Token
        var thi types.Token
        var stepv types.Token
        //var res types.Token
        var fromtok types.TokenList
        var steptok types.TokenList
        var totok types.TokenList
        var clause types.TokenList
        //var condlist types.TokenList
        var step float64
        //var lkey int
        //var current types.CodeRef
        //var nextST types.CodeRef
        //var postLOOP types.CodeRef
        var ls types.LoopState
        //var nostep bool


        //interp.VDU.PutSTR("FOR Current ref: line.Equals(\"+IntToStr(LPC.Line)+\"), statement == "+IntToStr(LPC.Statement)+PasUtil.CRLF);

      result = 0

      if (tokens.Size() < 5) {
        return result, exception.NewESyntaxError("Syntax Error")
      }

      step = 1.0

      eq_idx = tokens.IndexOf(types.ASSIGNMENT, "=")
      to_idx = tokens.IndexOf(types.KEYWORD, "TO");

      step_idx = tokens.IndexOf(types.KEYWORD, "STEP")
      //nostep = (step_idx == -1)

      tvar = *tokens.Get(0)
      //tto = *tokens.Get(to_idx)

      //writeln("fromtok");
      fromtok = *tokens.SubList(eq_idx+1, to_idx)
      tlo = caller.ParseTokensForResult(fromtok)
      // removed free call here /* free tokens */

      if (step_idx != -1) {
        //writeln("steptok");
        steptok = *tokens.SubList(step_idx+1, tokens.Size())
        stepv = caller.ParseTokensForResult(steptok)
        step = stepv.AsExtended()
        // removed free call here /*free tokens*/
      } else {
        step_idx = tokens.Size()
      }

      //writeln("totok");
      totok = *tokens.SubList( to_idx+1, step_idx )
      thi = caller.ParseTokensForResult(totok)
      // removed free call here /* free tokens */

      if (!caller.ExistsVar(strings.ToLower(tvar.Content))) {
        v, _ := types.NewVariableP(caller.GetLocal(), strings.ToLower(tvar.Content), types.VT_FLOAT, "0", true)
        v.Owner = caller.GetName();
      } else {
        if (caller.GetVar(strings.ToLower(tvar.Content)).Kind != types.VT_FLOAT) {
          return result, exception.NewESyntaxError("FOR variable must be a NUMBER")
        }
      }

      if (caller.GetLoopStates().ContainsKey(strings.ToLower(tvar.Content))) {
    	  return 0, nil; // ignore this
      }

      //if (!caller.LoopStates.ContainsKey(strings.ToLower(tvar.Content)));
      //{
        /* init loop var */

      clause = *types.NewTokenList()
      clause.Push( &tvar )
      clause.Push( types.NewToken(types.ASSIGNMENT, "=") )
      clause.Push( &tlo )
      caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)
      // removed free call here /*free tokens*/

      /* create loop state */
      if (caller.GetLoopStates().ContainsKey(strings.ToLower(tvar.Content))) {
    	  return 0, nil; // ignore this
      } else {
    	  ls = types.LoopState{}
      }
      ls.VarName = strings.ToLower(tvar.Content)
      ls.Start = tlo.AsExtended()
      ls.Finish = thi.AsExtended()
      ls.Step = step
      /* handle scope */
      ls.Code = *Scope

      ls.Entry = caller.GetNextStatement( LPC )

      caller.GetLoopStates().Put(ls.VarName, ls)

        //interp.VDU.PutStr( "Init for loop for "+ls.VarName+" with start.Equals(\"+PasUtil.FormatFloat(\"\",ls.Start)+\") && end.Equals("+PasUtil.FormatFloat(")",ls.Finish) );

      //}

      //writeln( "Entering for loop at ", ls.Entry.Line, ',', ls.Entry.Statement);

      /* if (we got here without error make sure loop is on the stack */

      // Now create a stack point;
      caller.Call( caller.GetNextStatement(LPC), Scope, caller.GetState(), false, "", *caller.GetTokenStack(), caller.GetDialect() )
      caller.SetLoopVariable(ls.VarName)
      caller.SetLoopStep(ls.Step)

        /* enforce non void return */
        return result, nil

}

func (this *StandardCommandFOR) Syntax() string {

        /* vars */
        var result string

      result = "FOR VN == N1 TO N2 [STEP N3]"

        /* enforce non void return */
        return result

}


package applesoft

import (
	"errors"
	"time"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

type StandardCommandGET struct {
	dialect.Command
}

func (this *StandardCommandGET) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	result := 0

	tla := caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	for _, tl := range tla {
		if tl.LPeek().Type != types.VARIABLE {
			return result, exception.NewESyntaxError("SYNTAX ERROR")
		}

		waiting := true

		for waiting {

			// each one handle
			apple2helpers.TextShowCursor(caller)

			for caller.GetMemory(49152) < 128 {
				time.Sleep(25 * time.Millisecond)
				if caller.VM().IsDying() {
					return result, nil
				}
			}

			apple2helpers.TextHideCursor(caller)

			// got a key
			ch := rune(caller.GetMemory(49152) & 127)

			// System.Err.Println("Got char code = "+ch);
			caller.SetMemory(49168, 0)

			if ch == 3 && caller.IsBreakable() {
				//caller.Halt()
				return result, errors.New("BREAK")
			}

			//            if ch >= 'a' && ch <= 'z' {
			//               ch -= 32
			//            } else if ch >= 'A' && ch <= 'Z' {
			//              ch += 32
			//            }

			if caller.GetDialect().IsUpperOnly() && ch >= 'a' && ch <= 'z' {
				ch -= 32
			}

			// now do an assignment

			ttl := tl.SubList(0, tl.Size())
			ttl.Push(types.NewToken(types.ASSIGNMENT, "="))
			if utils.Pos("$", ttl.LPeek().Content) <= 0 {
				ttl.Push(types.NewToken(types.NUMBER, string(ch)))
			} else {
				ttl.Push(types.NewToken(types.STRING, string(ch)))
			}
			//try {
			waiting = false
			caller.GetDialect().GetImpliedAssign().Execute(env, caller, *ttl, Scope, LPC)
			//} catch (Exception e) {
			//	  e.PrintStackTrace()
			//	  waiting = true
			//}

		}

	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandGET) Syntax() string {

	/* vars */
	var result string

	result = "GET <var>[,<var>,...]"

	/* enforce non void return */
	return result

}

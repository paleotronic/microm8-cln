package applesoft

import (
	"paleotronic.com/log"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"strings"
)

type StandardCommandREAD struct {
	dialect.Command
}

func (this *StandardCommandREAD) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var d types.Token
	var ref *types.CodeRef
	var clause types.TokenList
	var tla types.TokenListArray
	var datastring string
	var chunk string
	var ch rune
	var inqq bool
	var complete bool

	result = 0

	if caller.GetState() != types.RUNNING {
		return result, exception.NewESyntaxError("NOT DIRECT COMMAND")
	}

	if (tokens.Size() < 1) || (tokens.LPeek().Type != types.VARIABLE) {
		return result, exception.NewESyntaxError("SYNTAX ERROR")
	}

	tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	for _, tl := range tla {
		//writeln("DEBUG: NEW READ for "+caller.TokenListAsString(tl));
		//v = tl.Get(0);

		// find next data;
		if caller.GetDataRef().Line == -1 {
			return result, exception.NewESyntaxError("OUT OF DATA")
		}

		//ref = *types.NewCodeRefCopy(*caller.GetDataRef())
		ref = caller.GetDataRef()
		//writeln("DEBUG: get token");
		d = caller.GetTokenAtCodeRef(*ref)
		//writeln("DEBUG: got token");

		if (d.Type == types.KEYWORD) && (strings.ToLower(d.Content) == "data") {
			//writeln("DEBUG: get next token");
			d = caller.GetNextToken(ref)
			//writeln("DEBUG: got next token");
		}

		datastring = d.Content

		//writeln("DEBUG: Before length check");
		if (len(datastring) == 0) || ((ref.SubIndex >= len(datastring)) && (datastring[len(datastring)-1] != ',')) {
			//writeln("DEBUG: Seeking Next DATA statement...");
			/* next data statement */
			if !caller.NextTokenInstance(ref, types.KEYWORD, "data") {
				return result, exception.NewESyntaxError("OUT OF DATA")
			}
			//writeln("DEBUG: Found a Next DATA statement...");
			/* now advance */
			d = caller.GetNextToken(ref)
			datastring = d.Content
			ref.SubIndex = 0
		}
		//writeln("DEBUG: After length check");

		/* now stuff */
		chunk = ""
		complete = false
		inqq = false
		startsqq := false
		endsqq := false

		//writeln("DEBUG: just before for loop ... ", ref.SubIndex, ", ", datastring.Size());
		for !complete {

			if ref.SubIndex >= len(datastring) {
				caller.NextTokenInstance(ref, types.KEYWORD, "data")
				if ref.Line != -1 {
					d = caller.GetNextToken(ref)
					datastring = d.Content
					ref.SubIndex = 0
				}
				complete = true
				break
			}

			ch = rune(datastring[ref.SubIndex])

			switch string(ch) { /* FIXME - Switch statement needs cleanup */
			case "\"":
				{
					if len(chunk) == 0 {
						startsqq = true
                        inqq = (!inqq)
					} else if startsqq && (len(chunk) > 0) {
						endsqq = true
                        inqq = (!inqq)
					}
					
					if (len(chunk) > 0) && (!startsqq) {
						chunk = chunk + string(ch)
					}
                    
					break
				}
			case ",":
				{
					if !inqq {
						complete = true
						ref.SubIndex = ref.SubIndex + 1 /* advance past comma */
						continue
					} else {
						chunk = chunk + string(ch)
					}
					break
				}
			case " ":
				{
					if (inqq) || ((len(chunk) > 0) && (!startsqq)) {
						chunk = chunk + string(ch)
					}
					break
				}
			default:
				{
					if endsqq {
						return result, exception.NewESyntaxError("SYNTAX ERROR")
					}
					chunk = chunk + string(ch)
				}
			} /*case*/

			ref.SubIndex = ref.SubIndex + 1

		}

		// System.Out.Println("----- READ yields ["+chunk+"]");
		log.Printf("Read yields [%s]\n", chunk)

		clause = *tl.SubList(0, tl.Size())
		//riteln("DEBUG: value is now "+chunk);
		//for _, v := range tl
		clause.Push(types.NewToken(types.ASSIGNMENT, "="))

		//chunk = chunk.Trim();

		//if ((caller.Dialect.IsInteger(chunk) || caller.Dialect.IsInteger(chunk)))
		//  clause.Push( types.NewToken(types.NUMBER, chunk) );
		//else
        
        n := clause.LPeek().Content
        
        if rune(n[len(n)-1]) == '$' {
		  clause.Push(types.NewToken(types.STRING, chunk))        
        } else {
		  clause.Push(types.NewToken(types.NUMBER, chunk))
        }

		//    writeln("DEBUG: "+caller.TokenListAsString(clause));
		log.Println(caller.TokenListAsString(clause))
		caller.GetDialect().ExecuteDirectCommand(clause, caller, Scope, &LPC)
		// removed free call here;

		//caller.GetDataRef().Line = ref.Line
		//caller.GetDataRef().Statement = ref.Statement
		//caller.GetDataRef().Token = ref.Token
		//caller.GetDataRef().SubIndex = ref.SubIndex

	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandREAD) Syntax() string {

	/* vars */
	var result string

	result = "READ"

	/* enforce non void return */
	return result

}

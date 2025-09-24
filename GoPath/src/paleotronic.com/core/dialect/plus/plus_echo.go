package plus

import (
	"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/types"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/utils"
    a2 "paleotronic.com/core/hardware/apple2helpers"
    "math"
    "time"
//    "paleotronic.com/fmt"
)

type PlusEcho struct {
	dialect.CoreFunction
}

func (this *PlusEcho) FunctionExecute(params *types.TokenList) error {

	if e := this.CoreFunction.FunctionExecute(params); e != nil { return e }

    if !this.Query {

	    filename := this.ValueMap["filename"].Content
	    pc := this.ValueMap["pageclear"]
        pageclear := (pc.AsInteger() == 1)
        noprompt  := (pc.AsInteger() == 2)


	    data, err := files.ReadBytesViaProvider( files.GetPath(filename) , files.GetFilename(filename))
	    if err != nil {
		    this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(0)))
		    return err
	    }

	    s := utils.Unescape(string(data.Content)) // unescape it here
	    sl := make([]string, 0)
	    chunk := ""
	    lastCh := ' '
	    for _, ch := range s {
		    if (ch == 13 || ch == 10) && (lastCh != 6) {
			    if len(chunk) > 0 {
				    sl = append(sl, chunk)
				    chunk = ""
			    }
		    } else {
			    chunk = chunk + string(ch)
		    }
		    lastCh = ch
	    }
	    if len(chunk) > 0 {
		    sl = append(sl, chunk)
		    chunk = ""
	    }

        //fmt.Printf("ECHO read %d lines\n", len(sl))

	    if pageclear {
		    a2.Clearscreen( this.Interpreter )
	    }

        maxlines := a2.GetActualRows( this.Interpreter )

	    for i, line := range sl {
		    if i%maxlines == maxlines-1 {
                if !noprompt {
			      this.Interpreter.PutStr("(press a key)")
			      this.Interpreter.SetMemory(49168,0)
			      for this.Interpreter.GetMemory(49152) < 128 {
				      //this.Interpreter.GetVDU().ProcessKeyBuffer(this.Interpreter)
			      }
                }
			    if pageclear {
				    a2.Clearscreen( this.Interpreter )
			    } else {
				    if this.Interpreter.GetCursorX() != 0 {
               		   this.PutStr("\r\n", this.Interpreter)
            		}
			    }
		    }
		    this.PutStr(line, this.Interpreter)
            if this.Interpreter.GetCursorX() != 0 && i != len(sl)-1 {
               this.PutStr("\r\n", this.Interpreter)
            }
	    }

    }

	this.Stack.Push(types.NewToken(types.NUMBER, utils.IntToStr(1)))

	return nil
}

func (this *PlusEcho) Syntax() string {

	/* vars */
	var result string

	result = "ECHO{textfile,pageclear(1|0)}"

	/* enforce non void return */
	return result

}

func (this *PlusEcho) FunctionParams() []types.TokenType {

	/* vars */
	var result []types.TokenType

	result = make([]types.TokenType, 0)
	result = append(result, types.STRING)
	result = append(result, types.NUMBER)

	/* enforce non void return */
	return result

}

func (this *PlusEcho) PutStr(s string, caller interfaces.Interpretable) {

    speed := float64(caller.GetSpeed())

    cps := -128*math.Log((256-speed)/256)+2

    delay := 1000000 / cps

    if speed >= 255 {
       delay = 0
    }

    for _, ch := range s {
    	a2.RealPut(caller, ch)
        time.Sleep( time.Duration(delay)*time.Microsecond )
    }

	//apple2helpers.PutStr(caller, s)
}

func NewPlusEcho(a int, b int, params types.TokenList) *PlusEcho {
	this := &PlusEcho{}

	/* vars */

	this.CoreFunction = *dialect.NewCoreFunction(a, b, params)
    this.NamedParams = []string{ "filename", "pageclear" }
    this.NamedDefaults = []types.Token{
    				   *types.NewToken( types.STRING, "" ),
                       *types.NewToken( types.NUMBER, "0"),
    }
    this.Raw = true
	this.Name = "ECHO"

	return this
}

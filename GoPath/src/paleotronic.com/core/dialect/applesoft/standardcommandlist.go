package applesoft

import (
	"paleotronic.com/fmt"
	"math"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type StandardCommandLIST struct {
	dialect.Command
}

func NewStandardCommandLIST() *StandardCommandLIST {
	this := &StandardCommandLIST{}
	//this.ImmediateMode = true
	return this
}

func (this *StandardCommandLIST) PadLeft(str string, width int) string {
	for len(str) < width {
		str = str + " "
	}

	return str
}

func (this *StandardCommandLIST) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var l int
	var h int
	var w int
	var z int
	var s string
	var ft string
	var ln types.Line
	//var stmt types.Statement
	//var t types.Token
	var lt types.Token
	var ntl types.TokenList
	//var tl types.TokenList
	var tla types.TokenListArray

	result = 0

	ntl = *types.NewTokenList()

	//fmt.Println("S =", tokens.Size())

	if tokens.Size() == 3 {
		tokens.Remove(1)
	}

	if tokens.Size() == 2 {

		if tokens.RPeek().Content == "-" {

			l = int(math.Abs(float64(tokens.LPeek().AsInteger())))
			tokens = *types.NewTokenList()
			tokens.Push(types.NewToken(types.NUMBER, utils.IntToStr(l)))
			tokens.Push(types.NewToken(types.SEPARATOR, ","))
			tokens.Push(types.NewToken(types.NUMBER, utils.IntToStr(h)))

		} else {

			l = int(math.Abs(float64(tokens.LPeek().AsInteger())))
			h = int(math.Abs(float64(tokens.RPeek().AsInteger())))
			tokens = *types.NewTokenList()
			tokens.Push(types.NewToken(types.NUMBER, utils.IntToStr(l)))
			tokens.Push(types.NewToken(types.SEPARATOR, ","))
			tokens.Push(types.NewToken(types.NUMBER, utils.IntToStr(h)))

		}
	}

	//fmt.Println("S =", tokens.Size())

	tla = caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))

	for _, tl1 := range tla {
		t := caller.ParseTokensForResult(tl1)
		ntl.Push(&t)
	}

	tokens = *ntl.SubList(0, ntl.Size())

	b := caller.GetCode()

	l = b.GetLowIndex()
	h = b.GetHighIndex()

	s = utils.IntToStr(h)
	w = len(s) + 1
	if w < 4 {
		w = 4
	}

	if l < 0 {
		return result, nil
	}

	/* now take extra params */
	if (tokens.Size() == 2) && ((tokens.Get(0).Type == types.NUMBER) || (tokens.Get(0).Type == types.INTEGER)) && (tokens.Get(1).Content == "-") {
		z = tokens.Get(0).AsInteger()
		//env.VDU.PutStr(IntToStr(z)+PasUtil.CRLF);
		l = z
	}

	if (tokens.Size() > 0) && ((tokens.Get(0).Type == types.NUMBER) || (tokens.Get(0).Type == types.INTEGER)) {
		z = tokens.Get(0).AsInteger()
		//env.VDU.PutStr(IntToStr(z)+PasUtil.CRLF);
		l = z
		h = z
	}

	if (tokens.Size() > 1) && ((tokens.Get(1).Type == types.NUMBER) || (tokens.Get(1).Type == types.INTEGER)) {
		z = tokens.Get(1).AsInteger()
		//writeln( "h will be set to ", z );
		//nv.VDU.PutStr(IntToStr(z)+PasUtil.CRLF);
		h = z
	}

	//writeln( "l is set to ", l );
	//fmt.Println("L =", l)
	//fmt.Println("H =", h)

	linecount := 0

	ww := len(utils.IntToStr(h)) + 2

	columns := apple2helpers.GetColumns(caller) - ww // so we have a margin
	rows := apple2helpers.GetRows(caller)

	caller.SetIgnoreSpecial(true)

	addHistory := (l == h)

	/* got code */
	for (l != -1) && (l <= h) {
		/* display this line */

		if b.ContainsKey(l) {
			lns := this.PadLeft(utils.IntToStr(l), ww)
			//caller.PutStr(PadLeft(s, w) + " ")
			ln, _ = caller.GetCode().Get(l)
			s = ""
			for _, stmt1 := range ln {

				//wraplen := columns - len(lns)

				//ft = caller.TokenListAsString( stmt );
				ft = ""
				lt = *types.NewToken(types.INVALID, "")
				for _, t1 := range stmt1.Content {
					if (lt.Type == types.KEYWORD) || (lt.Type == types.DYNAMICKEYWORD) || (lt.Type == types.OPERATOR) || (lt.Type == types.SEPARATOR) || (lt.Type == types.OBRACKET) || (lt.Type == types.COMPARITOR) || (lt.Type == types.ASSIGNMENT) {
						ft = ft + " "
					}
					if ((t1.Type == types.FUNCTION) || (t1.Type == types.LABEL) || (t1.Type == types.UNSTRING) || (t1.Type == types.CBRACKET) || (t1.Type == types.OBRACKET) || (t1.Type == types.NUMBER) || (t1.Type == types.VARIABLE) || (t1.Type == types.STRING) || (t1.Type == types.LOGIC) || (t1.Type == types.OPERATOR) || (t1.Type == types.CBRACKET) || (t1.Type == types.KEYWORD) || (t1.Type == types.COMPARITOR) || (t1.Type == types.ASSIGNMENT)) && (len(ft) > 0) && (ft[len(ft)-1] != ' ') {
						ft = ft + " "
					}

					// now prepad
					/*ol := len( lns + s + ft )
					for ol % columns < len(lns) {
						ft += " "
						ol = len(lns + s + ft)
					}*/

					nb := t1.AsString()

					ft = ft + nb
				}

				if s != "" {
					s = s + " : "
				}

				s = s + ft

			}

			if l == h && addHistory {
				caller.AddToHistory(runestring.Cast(lns + s))
			}

			wrapped := wraptext(columns, s)
			fmt.Printf("Wrapped listing to %d cols...\n", columns)

			linecount += len(wrapped)

			if linecount >= rows-2 {
				apple2helpers.PutStr(caller, "\r\n(press a key)")
				caller.SetMemory(49168, 0)
				for caller.GetMemory(49152) < 128 {
					//caller.GetVDU().ProcessKeyBuffer(caller)
				}

				if caller.GetMemory(49152)&127 == 3 {
					caller.SetMemory(49168, 0)
					break
				}

				apple2helpers.PutStr(caller, "\r\n")
				linecount = (len(wrapped) / columns) + 1

			}

			for i, zz := range wrapped {

				if apple2helpers.GetCursorX(caller) != 0 {
					apple2helpers.PutStr(caller, "\r\n")
				}
				if i == 0 {
					apple2helpers.PutStr(caller, lns)
				} else {
					apple2helpers.PutStr(caller, this.PadLeft("", ww))
				}

				// line
				apple2helpers.PutStr(caller, zz)
			}

		}

		/* next line */
		//l = caller.Code.NextAfter(l);
		l++
	}

	caller.SetIgnoreSpecial(false)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandLIST) Syntax() string {

	/* vars */
	var result string

	result = "LIST"

	/* enforce non void return */
	return result

}

func fields(s string) []string {
	chunk := ""
	out := make([]string, 0)
	var lastCh rune
	for _, ch := range s {
		if ch == ' ' {
			chunk += string(ch)
		} else {
			if lastCh == ' ' {
				if chunk != "" {
					out = append(out, chunk)
					chunk = ""
				}
			}
			chunk += string(ch)
		}
		lastCh = ch
	}
	if chunk != "" {
		out = append(out, chunk)
		chunk = ""
	}
	return out
}

func wraptext(lineWidth int, text string) []string {

	wrapped := ""
	words := fields(text)
	fmt.Printf("Words: [%v]\n", words)
	if len(words) == 0 {
		return []string{""}
	}
	wrapped = words[0]
	spaceLeft := lineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			//~ for i := 0; i < spaceLeft; i++ {
			//~ wrapped += " "
			//~ }

			if len(word)+1 > lineWidth {

				for len(word)+1 > lineWidth && spaceLeft > 1 {

					chunk := word[:spaceLeft-1]
					word = word[spaceLeft-1:]

					wrapped += chunk
					wrapped += "\r\n"
					spaceLeft = lineWidth

				}

				wrapped += word
				spaceLeft = lineWidth - len(word)

			} else {
				wrapped += "\r\n"
				wrapped += word
				spaceLeft = lineWidth - len(word)
			}
		} else {
			wrapped += word
			spaceLeft -= 1 + len(word)
		}
	}
	return strings.Split(wrapped, "\r\n")

}

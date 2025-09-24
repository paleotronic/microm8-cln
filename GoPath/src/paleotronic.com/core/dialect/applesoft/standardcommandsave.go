package applesoft

import (
	"errors"
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/utils"
)

type StandardCommandSAVE struct {
	dialect.Command
}

func (this *StandardCommandSAVE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var l int
	var h int
	var w int
	var f []string
	var ft string
	var ln types.Line
	var res types.Token
	var filename string

	result = 0
	f = []string(nil)

	if tokens.Size() == 0 {
		return result, exception.NewESyntaxError("SAVE needs a \"filename\"")
	}

	if caller.IsRunningDirect() && (tokens.Size() > 0) {
		// collapse tokens
		out := ""

		for _, t := range tokens.Content {
			out = out + t.Content
		}

		tokens.Clear()
		tokens.Push(types.NewToken(types.STRING, out))
	}

	res = caller.ParseTokensForResult(tokens)
	if res.Content == "" {
		return result, exception.NewESyntaxError("SAVE needs a \"filename\"")
	}

	osext := string(strings.ToLower(caller.GetDialect().GetTitle())[0])
	prefext := files.GetPreferredExt(osext)

	filename = utils.Flatten7Bit(res.Content)

	if files.GetExt(filename) == "" {
		filename += "." + prefext
	}

	if caller.GetWorkDir() != "" && rune(filename[0]) != '/' {
		filename = caller.GetWorkDir() + filename
	}

	filename = strings.ToLower(filename)

	//assign(f, filename);
	//rewrite(f);

	b := caller.GetCode()

	l = b.GetLowIndex()
	h = b.GetHighIndex()

	f = make([]string, 0)

	//Str(h, s);
	s := utils.IntToStr(h)
	w = len(s) + 1
	if w < 4 {
		w = 4
	}

	if l < 0 {
		//close(f);
		return result, nil
	}

	s = ""
	/* got code */
	for l != -1 {
		/* now formatted tokens */
		ln, _ = caller.GetCode().Get(l)
		s = utils.IntToStr(l) + "  "
		z := 0
		for _, stmt := range ln {

			ft = caller.TokenListAsString(stmt.TokenList)
			if z > 0 {
				s = s + " : "
			}

			s = s + ft

			z++
		}

		f = append(f, s)
		//f.Add(s);

		/* next line */
		b := caller.GetCode()
		l = b.NextAfter(l)
	}

	//close(f);

	//FileHan//dle fh = Gdx.Files.External(filename)
	//fh.WriteString(s, false)

	str := ""
	for _, ss := range f {
		if str != "" {
			str = str + "\r\n"
		}
		str = str + utils.Escape(ss)
	}

	pp := files.GetPath(filename)
	ff := files.GetFilename(filename)

	e := files.WriteBytesViaProvider(pp, ff, []byte(str))

	//utils.WriteTextFile(filename, f)
	if e == nil {
		fr, e := files.ReadBytesViaProvider(pp, ff)
		if e != nil {
			return result, e
		}
		if string(fr.Content) == str {
			apple2helpers.PutStr(caller, "Ok: Saved \""+filename+"\""+"\r\n")
		} else {
			return result, errors.New("I/O Error")
		}
	}

	// save
	//caller.Freeze( "machine.tmp" )

	return result, e

}

func (this *StandardCommandSAVE) Syntax() string {

	/* vars */
	var result string

	result = "SAVE \"<filename>\""

	/* enforce non void return */
	return result

}

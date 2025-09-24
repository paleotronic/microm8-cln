package appleinteger

import (
	//"paleotronic.com/fmt"
	//	"paleotronic.com/fmt"

	"math"
	"regexp"
	"sort"
	"strings"

	s8webclient "paleotronic.com/api"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/dialect/applesoft"
	"paleotronic.com/core/dialect/plus"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/memory"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

const (
	LOOPSTACK_ADDRESS  = 0xff00
	CALLSTACK_ADDRESS  = 0xfe00
	LOOPSTACK_MAX      = 255
	CALLSTACK_MAX      = 255
	LOOP_PEEK_FREQ     = 86
	ADDSUB_PEEK_FREQ   = 66
	ADDSUB_PEEK_MS_PER = 8
	LOOP_PEEK_MS_PER   = 6
)

type DialectAppleInteger struct {
	dialect.Dialect
	lastLineNumber int
}

var reSimpleLoop *regexp.Regexp
var reAddSub *regexp.Regexp

func init() {
	// FOR I = 1 TO 80 : Q = PEEK( -16336 ) : NEXT I : GOTO 80
	reSimpleLoop, _ = regexp.Compile("(.* : )?FOR ([A-Za-z]+) = ([0-9]+) TO ([0-9]+) : ([A-Za-z]+) = PEEK[(] -16336 [)] : NEXT( [A-Za-z]+)?( :.*)?")
	// Z = PEEK( -16336 ) - PEEK( -16336 ) + PEEK( -16336 ) - PEEK( -16336 ) + PEEK( -16336 ) - PEEK( -16336 ) + PEEK( -16336 )
	reAddSub, _ = regexp.Compile("(.* )?([A-Za-z]+) = (PEEK[(] -16336 [)])( [+-] PEEK[(] -16336 [)])*( :.*)?")
}

func NewDialectAppleInteger() *DialectAppleInteger {
	this := &DialectAppleInteger{}
	this.Dialect = *dialect.NewDialect()
	this.Init()
	this.Dialect.DefaultCost = 1000000000 / 800
	this.Throttle = 100.0
	this.MaxVariableLength = -1
	this.UpperOnly = true
	this.GenerateNumericTokens()
	return this
}

func (this *DialectAppleInteger) BeforeRun(caller interfaces.Interpretable) {
	fixMemoryPtrs(caller)
}

func (this *DialectAppleInteger) CheckOptimize(lno int, s string, OCode types.Algorithm) {
	// stub does nothing

	if m := reSimpleLoop.FindStringSubmatch(s); len(m) > 0 {
		// (...:)? FOR ([A-Z]+) = ([0-9]+) TO ([0-9]+) : ([A-Z]+) = PEEK[(] -16336 [)] : NEXT( [A-Z]+)?( :.*)?
		//   1           2          3          4          5                                   6        7
		start := utils.StrToInt(m[3])
		end := utils.StrToInt(m[4])

		total_peeks := end - start + 1
		duration_ms := total_peeks * LOOP_PEEK_MS_PER
		freq_hz := LOOP_PEEK_FREQ

		alt_cmd := "@music.tone{" + utils.IntToStr(freq_hz) + "," + utils.IntToStr(duration_ms) + "}"
		tl := this.Tokenize(runestring.Cast(alt_cmd))
		ln := types.NewLine()
		st := types.NewStatement()

		if m[1] != "" {
			ss := m[1][0 : len(m[1])-3]

			ftl := *this.Tokenize(runestring.Cast(ss))
			tla := this.SplitOnToken(ftl, *types.NewToken(types.SEPARATOR, ":"))

			for _, tl := range tla {
				st = types.NewStatement()

				for _, t := range tl.Content {
					st.Push(t)
				}

				if st.Size() > 0 {
					ln = append(ln, st)
				}
			}

			st = types.NewStatement()

		}

		for _, t := range tl.Content {
			st.Push(t)
		}

		if st.Size() > 0 {
			ln = append(ln, st)
		}

		// final part
		if m[7] != "" {
			ss := m[7][3:]

			ftl := *this.Tokenize(runestring.Cast(ss))
			tla := this.SplitOnToken(ftl, *types.NewToken(types.SEPARATOR, ":"))

			for _, tl := range tla {
				st = types.NewStatement()

				for _, t := range tl.Content {
					st.Push(t)
				}

				if st.Size() > 0 {
					ln = append(ln, st)
				}
			}

		}

		// Add new version
		OCode.Put(lno, ln)

		//fmt.Println("Original code     :", lno, s)
		//fmt.Println("Optimizer suggests:", lno, ln.String())

	} else if m = reAddSub.FindStringSubmatch(s); len(m) > 0 {
		// (.+ : )?([A-Z]+) = (PEEK[(] -16336 [)])( [+-] (PEEK[(] -16336 [)]))*( :.*)?
		//    1       2                3              4             5             6
		//fmt.Printf("All submatches: %v\n", m)

		total_peeks := strings.Count(s, "PEEK( -16336 )")
		duration_ms := total_peeks * ADDSUB_PEEK_MS_PER
		freq_hz := ADDSUB_PEEK_FREQ

		alt_cmd := "@music.tone{" + utils.IntToStr(freq_hz) + "," + utils.IntToStr(duration_ms) + "}"
		tl := this.Tokenize(runestring.Cast(alt_cmd))
		ln := types.NewLine()
		st := types.NewStatement()

		if m[1] != "" {
			ss := m[1]

			ftl := *this.Tokenize(runestring.Cast(ss))
			tla := this.SplitOnToken(ftl, *types.NewToken(types.SEPARATOR, ":"))

			for _, tl := range tla {
				st = types.NewStatement()

				for _, t := range tl.Content {
					st.Push(t)
				}

				if st.Size() > 0 {
					ln = append(ln, st)
				}
			}

			if strings.ToLower(st.RPeek().Content) != "then" {
				st = types.NewStatement()
			}

		}

		for _, t := range tl.Content {
			st.Push(t)
		}

		if st.Size() > 0 {
			if st.Size() == tl.Size() {
				ln = append(ln, st)
			} else {
				if len(ln) == 0 {
					ln = append(ln, st)
				} else {
					ln[len(ln)-1] = st
				}
			}
		}

		//
		if m[5] != "" {
			ss := m[5][3:]

			ftl := *this.Tokenize(runestring.Cast(ss))
			tla := this.SplitOnToken(ftl, *types.NewToken(types.SEPARATOR, ":"))

			for _, tl := range tla {
				st = types.NewStatement()

				for _, t := range tl.Content {
					st.Push(t)
				}

				if st.Size() > 0 {
					ln = append(ln, st)
				}
			}

		}

		// Add new version
		OCode.Put(lno, ln)

		//fmt.Println("Original code     :", lno, s)
		//fmt.Println("Optimizer suggests:", lno, ln.String())

	}
}

func (this *DialectAppleInteger) ProcessDynamicCommand(ent interfaces.Interpretable, cmd string) error {

	//	////fmt.Printf("In ProcessDynamicCommand for [%s]\n", cmd)

	s8webclient.CONN.LogMessage("UCL", "Dos command: "+cmd)

	if utils.Copy(cmd, 1, 4) == "BRUN" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 4), " "), ",")
		addr := 16384
		length := -1
		offset := 0

		filename := parts[0]
		filename = strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		var hasaddr bool
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'A' {
				tmp = utils.Delete(tmp, 1, 1)
				addr, _ = utils.SuperStrToInt(tmp)
				hasaddr = true
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length, _ = utils.SuperStrToInt(tmp)
			} else if tmp[0] == 'B' {
				tmp = utils.Delete(tmp, 1, 1)
				offset, _ = utils.SuperStrToInt(tmp)
			}
		}

		//if !strings.HasSuffix(filename, ".s") {
		//filename += ".s"
		//}

		// try original file first, the other
		if !files.ExistsViaProvider(ent.GetWorkDir(), filename) {

			binEx := files.GetTypeBin()

			found := false
			for _, v := range binEx {
				if files.ExistsViaProvider(ent.GetWorkDir(), filename+"."+v.Ext) {
					filename += "." + v.Ext
					found = true
					break
				}
			}

			if !found {
				return exception.NewESyntaxError("FILE NOT FOUND")
			}
		}

		data, err := files.ReadBytesViaProvider(ent.GetWorkDir(), filename)
		if offset > 0 && offset <= len(data.Content) {
			data.Content = data.Content[offset:]
		}
		if err != nil {
			return err
		}

		rl := len(data.Content)
		if (length != -1) && (length <= rl) {
			rl = length
		}
		//System.Out.Println("BLOAD A="+addr+", L="+rl);
		rawdata := make([]uint64, rl)
		for i := 0; i < rl; i++ {
			//ent.SetMemory((addr+i)%65536, uint64(data.Content[i]&0xff))
			rawdata[i] = uint64(data.Content[i])
		}
		if !hasaddr && data.Address != 0 {
			addr = data.Address
		}
		ent.GetMemoryMap().BlockWritePr(ent.GetMemIndex(), addr, rawdata)

		//System.Out.Println("OK");
		// disable zero page
		apple2helpers.GetCPU(ent).BasicMode = false
		apple2helpers.DoCall(addr, ent, true)

		return nil

	}

	if utils.Copy(cmd, 1, 5) == "BLOAD" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		addr := 16384
		length := -1
		offset := 0

		filename := parts[0]
		filename = strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		var hasaddr bool
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'A' {
				tmp = utils.Delete(tmp, 1, 1)
				addr, _ = utils.SuperStrToInt(tmp)
				hasaddr = true
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length, _ = utils.SuperStrToInt(tmp)
			} else if tmp[0] == 'B' {
				tmp = utils.Delete(tmp, 1, 1)
				offset, _ = utils.SuperStrToInt(tmp)
			}
		}

		//if !strings.HasSuffix(filename, ".s") {
		//filename += ".s"
		//}

		// try original file first, the other
		if !files.ExistsViaProvider(ent.GetWorkDir(), filename) {

			binEx := files.GetTypeBin()

			found := false
			for _, v := range binEx {
				if files.ExistsViaProvider(ent.GetWorkDir(), filename+"."+v.Ext) {
					filename += "." + v.Ext
					found = true
					break
				}
			}

			if !found {
				return exception.NewESyntaxError("FILE NOT FOUND")
			}
		}

		data, err := files.ReadBytesViaProvider(ent.GetWorkDir(), filename)
		if offset > 0 && offset <= len(data.Content) {
			data.Content = data.Content[offset:]
		}
		if err != nil {
			return err
		}

		rl := len(data.Content)
		if (length != -1) && (length <= rl) {
			rl = length
		}
		//System.Out.Println("BLOAD A="+addr+", L="+rl);
		rawdata := make([]uint64, rl)
		for i := 0; i < rl; i++ {
			//ent.SetMemory((addr+i)%65536, uint64(data.Content[i]&0xff))
			rawdata[i] = uint64(data.Content[i])
		}
		if !hasaddr && data.Address != 0 {
			addr = data.Address
		}
		ent.GetMemoryMap().BlockWrite(ent.GetMemIndex(), addr, rawdata)

		//System.Out.Println("OK");

		return nil

	}

	if utils.Copy(cmd, 1, 6) == "VERIFY" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 6), " "), ",")
		//addr := 16384
		//length := -1
		offset := 0

		filename := parts[0]
		filename = strings.ToLower(filename)

		// try original file first, the other
		if !files.ExistsViaProvider(ent.GetWorkDir(), filename) {

			binEx := files.GetTypeAll()

			found := false
			for _, v := range binEx {
				if files.ExistsViaProvider(ent.GetWorkDir(), filename+"."+v.Ext) {
					filename += "." + v.Ext
					found = true
					break
				}
			}

			if !found {
				return exception.NewESyntaxError("FILE NOT FOUND")
			}
		}

		data, err := files.ReadBytesViaProvider(ent.GetWorkDir(), filename)
		if offset > 0 && offset <= len(data.Content) {
			data.Content = data.Content[offset:]
		}
		if err != nil {
			return err
		}

		return nil

	}

	if utils.Copy(cmd, 1, 5) == "BSAVE" {

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		addr := 16384
		length := -1

		filename := parts[0]
		filename = strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'A' {
				tmp = utils.Delete(tmp, 1, 1)
				addr, _ = utils.SuperStrToInt(tmp)
			} else if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				length, _ = utils.SuperStrToInt(tmp)
			}
		}

		if !strings.HasSuffix(filename, ".bin") {
			filename += ".bin"
		}

		if length == -1 {
			return nil
		}

		data := make([]byte, length)

		for i, _ := range data {
			data[i] = byte(ent.GetMemory(addr + i))
		}

		e := files.WriteBytesViaProvider(ent.GetWorkDir(), filename, data)

		//System.Out.Println("OK");

		return e

	}

	// Run shim...

	if utils.Copy(cmd, 1, 3) == "RUN" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 3), " "))
		tl := types.NewTokenList()
		tl.Push(types.NewToken(types.STRING, fn))
		a := ent.GetCode()
		_, e := ent.GetDialect().GetCommands()["load"].Execute(nil, ent, *tl, a, *ent.GetLPC())
		if e != nil {
			return e
		}
		ent.Run(false)
	}

	if utils.Copy(cmd, 1, 5) == "CHAIN" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 5), " "))
		tl := types.NewTokenList()
		tl.Push(types.NewToken(types.STRING, fn))
		a := ent.GetCode()
		_, e := ent.GetDialect().GetCommands()["load"].Execute(nil, ent, *tl, a, *ent.GetLPC())
		if e != nil {
			return e
		}
		a = ent.GetCode()
		pc := ent.GetPC()
		pc.Line = a.GetLowIndex()
		pc.Statement = 0
		pc.Token = 0
		ent.SetPC(pc)
		ent.SetState(types.RUNNING)
	}

	if utils.Copy(cmd, 1, 5) == "WRITE" {
		//fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 5), " ")) + ".d"
		//fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})

		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 5), " "), ",")
		record := 0

		filename := parts[0] + ".dat"
		filename = ent.GetWorkDir() + "/" + strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'R' {
				tmp = utils.Delete(tmp, 1, 1)
				record, _ = utils.SuperStrToInt(tmp)
				record -= 1
			}
		}

		e := files.DOSWRITE(files.GetPath(filename), files.GetFilename(filename), record)
		if e == nil {
			ent.SetOutChannel(filename)
		}
		return e
	}

	if utils.Copy(cmd, 1, 4) == "READ" {
		//		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 4), " ")) + ".d"
		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 4), " "), ",")
		record := 0

		filename := parts[0] + ".dat"
		filename = ent.GetWorkDir() + "/" + strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'R' {
				tmp = utils.Delete(tmp, 1, 1)
				record, _ = utils.SuperStrToInt(tmp)
				record -= 1
			}
		}

		e := files.DOSREAD(files.GetPath(filename), files.GetFilename(filename), record)

		if e == nil {
			ent.SetInChannel(filename)
		}

		return e
	}

	if utils.Copy(cmd, 1, 6) == "APPEND" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 6), " ")) + ".t"
		//fn := files.GetUserPath(files.BASEDIR, []string{ent.GetWorkDir(), strings.ToLower(parts[0]) + ".d"})
		e := files.DOSAPPEND(files.GetPath(fn), files.GetFilename(fn))
		if e == nil {
			ent.SetOutChannel(fn)
		}
		return e
	}

	if utils.Copy(cmd, 1, 5) == "CLOSE" {
		fn := ent.GetWorkDir() + "/" + strings.ToLower(strings.Trim(utils.Delete(cmd, 1, 5), " ")) + ".t"

		var e error
		if cmd == "CLOSE" {
			e = files.DOSCLOSEALL()
		} else {
			e = files.DOSCLOSE(files.GetPath(fn), files.GetFilename(fn))
		}

		ent.SetOutChannel("")
		ent.SetInChannel("")
		ent.SetFeedBuffer("")

		return e
	}

	if utils.Copy(cmd, 1, 4) == "OPEN" {
		parts := strings.Split(strings.Trim(utils.Delete(cmd, 1, 4), " "), ",")
		recsize := 0

		filename := parts[0] + ".dat"
		filename = ent.GetWorkDir() + "/" + strings.ToLower(filename)
		//System.Out.Println("will try to load "+filename);
		for i := 1; i < len(parts); i++ {
			tmp := strings.ToUpper(parts[i])
			if tmp[0] == 'L' {
				tmp = utils.Delete(tmp, 1, 1)
				recsize, _ = utils.SuperStrToInt(tmp)
			}
		}
		e := files.DOSOPEN(files.GetPath(filename), files.GetFilename(filename), recsize)
		ent.SetOutChannel("")
		ent.SetFeedBuffer("")
		return e
	}

	if utils.Copy(cmd, 1, 3) == "PR#" {
		cmd = strings.Trim(utils.Delete(cmd, 1, 3), " ")
		mode := utils.StrToInt(cmd)

		switch mode { /* FIXME - Switch statement needs cleanup */
		case 0:
			{
				//ent.GetVDU().SetVideoMode(ent.GetVDU().GetVideoModes()[5])
				//ent.GetVDU().ClrHome()
				//				ent.GetVDU().RegenerateWindow(ent.GetMemory())
				apple2helpers.TEXT40(ent)
				ent.SetMemory(49152, 0)
				break
			}
		case 3:
			{
				//ent.GetVDU().SetVideoMode(ent.GetVDU().GetVideoModes()[0])
				//ent.GetVDU().ClrHome()
				//				ent.GetVDU().RegenerateWindow(ent.GetMemory())
				apple2helpers.TEXT80(ent)
				ent.SetMemory(49153, 0)
				break
			}
		}

		return nil
	}

	return nil

}

func (this *DialectAppleInteger) HPOpIndex(tl types.TokenList) int {
	result := -1
	hs := 1
	var tt types.Token
	var sc int

	for i := 0; i <= tl.Size()-1; i++ {
		tt = *tl.Get(i)
		sc = 0

		if tt.Content == "^" {
			sc = 500
		} else if (tt.Content == "*") || (tt.Content == "/") {
			sc = 400
		} else if strings.ToLower(tt.Content) == "mod" {
			sc = 400
		} else if (tt.Content == "+") || (tt.Content == "-") {
			sc = 300
		} else if (tt.Type == types.COMPARITOR) || (tt.Type == types.ASSIGNMENT) {
			sc = 200
		} else if tt.Type == types.LOGIC {
			s := strings.ToLower(tt.Content)
			if s == "not" {
				sc = 150
			} else if s == "and" {
				sc = 140
			} else {
				sc = 130
			}
		} else {
			sc = 100
		}

		if sc > hs {
			hs = sc
			result = i
		}
	}

	return result
}

func (this *DialectAppleInteger) Evaluate(chunk string, tokens *types.TokenList) bool {

	/* vars */
	var result bool
	var tok *types.Token
	var ptok *types.Token
	//	var i int

	result = true // continue

	//writeln("EVALUATE: ", chunk);

	if len(chunk) == 0 {
		return result
	}

	tok = types.NewToken(types.INVALID, "")

	//  if (this.IsBoolean(chunk));
	//  begin;
	//    tok.Type = types.BOOLEAN;
	//    tok.Content = chunk;
	//  }
	//  else
	if this.IsLabel(chunk) {
		tok.Type = types.LABEL
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsFunction(chunk) {
		tok.Type = types.FUNCTION
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsDynaCommand(chunk) {
		tok.Type = types.DYNAMICKEYWORD
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsDynaFunction(chunk) {
		tok.Type = types.DYNAMICFUNCTION
		tok.Content = strings.ToUpper(chunk)
	} else if this.IsKeyword(chunk) {
		tok.Type = types.KEYWORD
		tok.Content = strings.ToUpper(chunk)

		/* determine if (we should stop parsing */
		result = !this.Commands[strings.ToLower(chunk)].HasNoTokens()
	} else if this.IsLogic(chunk) {
		tok.Type = types.LOGIC
		tok.Content = chunk
	} else if this.IsFloat(chunk) {
		if (tokens.Size() > 0) && (tokens.RPeek().Type == types.OPERATOR) && (tokens.RPeek().Content == "-") {
			tokens.RPeek().Content = "-" + chunk
			tokens.RPeek().Type = types.NUMBER
		} else {
			tok.Type = types.NUMBER
			tok.Content = chunk
		}
	} else if this.IsInteger(chunk) {
		if (tokens.Size() > 0) && (tokens.RPeek().Type == types.OPERATOR) && (tokens.RPeek().Content == "-") {
			tokens.RPeek().Content = "-" + chunk
			tokens.RPeek().Type = types.INTEGER
		} else {
			tok.Type = types.INTEGER
			tok.Content = chunk
		}
	} else if this.IsSeparator(chunk) {
		if chunk == ":" && (tokens.Size() > 0) && (strings.ToLower(tokens.RPeek().Content) == "himem" || strings.ToLower(tokens.RPeek().Content) == "lomem") {
			tokens.RPeek().Content += ":"
			tokens.RPeek().Type = types.KEYWORD
		} else {
			tok.Type = types.SEPARATOR
			tok.Content = chunk
		}
	} else if this.IsType(chunk) {
		tok.Type = types.TYPE
		tok.Content = chunk
	} else if this.IsOpenRBracket(chunk) || this.IsOpenSBracket(chunk) || this.IsOpenCBrace(chunk) {
		ptok = tokens.RPeek()
		if (tokens.Size() >= 1) && (ptok != nil) && (ptok.Type == types.VARIABLE) {
			ss := strings.ToLower(ptok.Content + chunk)
			if this.Functions[ss] != nil {
				ptok.Content = ptok.Content + chunk
				ptok.Type = types.FUNCTION
			} else if _, ex, _, _ := this.PlusFunctions.GetFunctionByNameContext(this.CurrentNameSpace, ss); ex {
				ptok.Content = ptok.Content + chunk
				ptok.Type = types.PLUSFUNCTION
			} else {
				tok.Type = types.OBRACKET
				tok.Content = chunk
			}
		} else {
			tok.Type = types.OBRACKET
			tok.Content = chunk
		}
	} else if this.IsCloseRBracket(chunk) || this.IsCloseSBracket(chunk) || this.IsCloseCBrace(chunk) {
		tok.Type = types.CBRACKET
		tok.Content = chunk
	} else if this.IsOperator(chunk) {
		tok.Type = types.OPERATOR
		tok.Content = chunk
	} else if this.IsComparator(chunk) {
		if tokens.Size() > 0 {
			ptok = tokens.RPeek()
			if (ptok.Type == types.ASSIGNMENT) && (ptok.Content == "=") && ((chunk == ">") || (chunk == "<")) {
				ptok.Content = ptok.Content + chunk
			} else if ptok.Type == types.COMPARITOR {
				ptok.Content = ptok.Content + chunk
			} else {
				tok.Type = types.COMPARITOR
				tok.Content = chunk
			}
		} else {
			tok.Type = types.COMPARITOR
			tok.Content = chunk
		}
	} else if this.IsAssignment(rune(chunk[0])) {

		/* the the preceding token was a comparitor)we merge them */
		/* eg. ">" + "=" == ">=" */

		if tokens.Size() > 0 {

			ptok = tokens.RPeek()
			posscmd := strings.ToLower(ptok.Content + chunk)
			if this.IsKeyword(posscmd) {
				ptok.Content = ptok.Content + chunk
				ptok.Type = types.KEYWORD
			} else if (ptok.Type == types.COMPARITOR) && (chunk == "=") {
				ptok.Content = ptok.Content + "="
			} else {
				tok.Type = types.ASSIGNMENT
				tok.Content = chunk
			}
		} else {
			tok.Type = types.ASSIGNMENT
			tok.Content = chunk
		}
	} else if this.IsVariableName(chunk) {
		tok.Type = types.VARIABLE
		tok.Content = chunk
	} else if this.IsPlusVariableName(chunk) {
		tok.Type = types.PLUSVAR
		tok.Content = chunk
	} else if this.IsString(chunk) {
		//System.Err.Println ( "String token ["+chunk+"]" );
		tok.Type = types.STRING
		chunk = utils.Copy(chunk, 2, utils.Len(chunk)-2)
		//char[] tmp = chunk.ToCharArray()
		//for  i=0;  i < tmp.Length;  i++  {
		//    tmp[i] = ( char ) ( ( tmp[i] & 127 ) +128 )
		//}
		//chunk = types.NewString ( tmp )
		tok.Content = chunk
	}

	if tok.Type != types.INVALID {
		//writeln("ADD: ", tok.Content);

		/* shim for -No */
		if (tok.Type == types.NUMBER) && (tokens.Size() >= 2) {
			if (tokens.RPeek().Type == types.OPERATOR) &&
				(tokens.RPeek().Content == "-") &&
				(tokens.Get(tokens.Size() - 2).IsType([]types.TokenType{types.ASSIGNMENT, types.COMPARITOR})) {
				/* merge into last token, change to number type */
				tokens.Get(tokens.Size() - 1).Type = types.NUMBER
				tokens.Get(tokens.Size() - 1).Content = "-" + tok.Content
			} else {
				tokens.Push(tok)
			}
		} else {
			tokens.Push(tok)
		}
	}

	/* enforce non void return */
	return result

}

func (this *DialectAppleInteger) HandleException(ent interfaces.Interpretable, e error) {

	/* vars */
	var msg string

	apple2helpers.NLIN(ent)

	//	e.PrintStackTrace()

	msg = e.Error()

	apple2helpers.PutStr(ent, "*** "+strings.ToUpper(msg))

	if (ent.GetState() == types.RUNNING) || (ent.GetState() == types.DIRECTRUNNING) {
		if !ent.HandleError() {
			if ent.GetState() == types.RUNNING {
				apple2helpers.PutStr(ent, " at line "+utils.IntToStr(ent.GetPC().Line))
			}

			_ = ent.Halt()

			//ent.GetVDU().PutStr(" at line "+IntToStr(ent.PC.Line));
		}
	}

	apple2helpers.PutStr(ent, "\r\n")

	r, g, b, a := ent.GetMemoryMap().GetBGColor(ent.GetMemIndex())
	ent.GetMemoryMap().SetBGColor(ent.GetMemIndex(), 255, 0, 0, 255)
	apple2helpers.Beep(ent)
	ent.GetMemoryMap().SetBGColor(ent.GetMemIndex(), r, g, b, a)

}

func (this *DialectAppleInteger) InitVDU(v interfaces.Interpretable, promptonly bool) {

	/* vars */
	apple2helpers.TEXT40(v)

	v.SetPrompt(">")
	v.SetTabWidth(8)

	if !promptonly {
		apple2helpers.Clearscreen(v)
		apple2helpers.Gotoxy(v, 0, 0)
		apple2helpers.PutStr(v, "Integer microBASIC (GAME BASIC)\r\n")
		v.SetNeedsPrompt(true)
	}

	//v.CreateVar(
	//	"speed",
	//	*types.NewVariableP("speed", types.VT_FLOAT, "255", true),
	//)
	//settings.SlotZPEmu[v.GetMemIndex()] = true

	//settings.SlotZPEmu[v.GetMemIndex()] = !settings.PureBoot(v.GetMemIndex())

	v.SaveCPOS()
	v.SetMemory(228, 255)

	settings.SpecName[v.GetMemIndex()] = "Integer microBASIC"
	settings.SetSubtitle(settings.SpecName[v.GetMemIndex()])

	v.GetMemoryMap().WriteGlobal(v.GetMemIndex(), v.GetMemoryMap().MEMBASE(v.GetMemIndex())+memory.OCTALYZER_CAMERA_GFX_BASE+0, uint64(types.CC_ResetAll))

	v.SetNeedsPrompt(true)

}

func (this *DialectAppleInteger) ExecuteDirectCommand(tl types.TokenList, ent interfaces.Interpretable, Scope *types.Algorithm, LPC *types.CodeRef) error {

	/* vars */
	var tok types.Token
	var n string
	var cmd interfaces.Commander
	var dcmd interfaces.DynaCoder
	//	var r int
	var cr types.CodeRef
	var ss types.TokenList

	if this.NetBracketCount(tl) != 0 {
		return exception.NewESyntaxError("SYNTAX ERROR")
	}

	if ent.IsDebug() && ent.IsRunning() {

		if tl.Get(0).Content == "next" {
			if ent.GetLoopStack().Size() > 0 {
				zz := ent.GetLoopStack().Get(ent.GetLoopStack().Size() - 1).Entry

				if zz.Line == ent.GetPC().Line && zz.Statement == ent.GetPC().Statement {
					ent.SetSilent(true)
					defer ent.SetSilent(false)
				}
			}
		}

		ent.Log("DO", ent.TokenListAsString(tl))

	}

	////fmt.Println( "OPT DEBUG:", ent.TokenListAsString(tl) )

	/* fix for assignment into keyword vars */
	if (tl.Size() > 0) && (tl.LPeek().Type == types.KEYWORD) && (strings.ToLower(tl.LPeek().Content) != "if") {
		//int x = tl.IndexOf ( types.ASSIGNMENT, "=" );
		x := this.FindAssignmentSymbol(tl)
		if (x > 0) && ((tl.Get(1).Type == types.ASSIGNMENT) || (tl.Get(1).Type == types.OBRACKET)) {
			tl.LPeek().Type = types.VARIABLE
		}
	}

	//System.Out.Println ( "DEBUG: -------------> ["+this.Title+"]: "+ent.PC.Line+","+ent.PC.Statement+":"+ent.TokenListAsString ( tl ) );
	if this.Trace && (ent.GetState() == types.RUNNING) {
		ent.PutStr("#" + utils.IntToStr(LPC.Line) + " ")
	}

	/* process poop monster here (@^-^@) */
	tok = *tl.Shift()

	if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) {
	} else if tok.Type == types.DYNAMICKEYWORD {
		n = strings.ToLower(tok.Content)
		if this.DynaCommands.ContainsKey(n) {
			dcmd = this.DynaCommands.Get(n)

			//ent.GetVDU().PutStr("Dynamic command parsing - Start at "+IntToStr(dcmd.Code.LowIndex)+PasUtil.CRLF);

			// try {
			/* its actually a hidden subroutine call */
			cr = *types.NewCodeRef()
			cr.Line = dcmd.GetCode().GetLowIndex()
			cr.Statement = 0
			cr.Token = 0
			if cr.Line != -1 {
				/* something to do */
				ss = *types.NewTokenList()
				for _, tok1 := range tl.Content {
					ss.Push(tok1)
				}
				ent.Call(cr, dcmd.GetCode(), ent.GetState(), false, n+ent.GetVarPrefix(), ss, dcmd.GetDialect()) // call with isolation off
			} else {
				return exception.NewESyntaxError("Dynamic Code Hook has no content")
			}
			//} catch ( Exception e ) {
			//this.HandleException ( ent, e )
			//}

		}
	} else if tok.Type == types.PLUSFUNCTION {

		// Handle plus function execution here
		fun, exists, ns, err := this.PlusFunctions.GetFunctionByNameContext(this.CurrentNameSpace, strings.ToLower(tok.Content))
		//fun, exists := this.PlusFunctions[strings.ToLower(tok.Content)]

		if !exists {
			this.HandleException(ent, err)
		}

		this.CurrentNameSpace = ns

		//fmt.Println(ent.TokenListAsString(tl))

		if tl.RPeek() != nil && tl.RPeek().Content == "}" {

			sl := tl.SubList(0, tl.Size()-1)
			////fmt.Println(ent.TokenListAsString(*sl))

			tla := ent.SplitOnTokenWithBrackets(*sl, *types.NewToken(types.SEPARATOR, ","))

			subexpr := types.NewTokenList()

			if fun.GetRaw() {
				subexpr = sl
			} else {

				for _, tt := range tla {
					v, serr := this.ParseTokensForResult(ent, tt)
					if serr != nil {
						this.HandleException(ent, serr)
					}
					subexpr.Push(v)
				}
				////fmt.Println(ent.TokenListAsString(*subexpr))
			}

			fun.SetEntity(ent)
			fun.SetQuery(false)
			fun.FunctionExecute(subexpr)

		} else {
			this.HandleException(ent, exception.NewESyntaxError("2 SYNTAX ERROR"))
		}

	} else if tok.Type == types.KEYWORD {

		if (tl.Size() > 0) && (tl.LPeek().Type == types.ASSIGNMENT) {
			tok.Type = types.VARIABLE
			tl.UnShift(&tok)
			cmd = this.ImpliedAssign

			//try {
			var e error
			_, e = cmd.Execute(nil, ent, tl, Scope, *LPC)
			if e != nil {
				this.HandleException(ent, e)
			}
			cost := cmd.GetCost()
			if cost == 0 {
				cost = this.DefaultCost
			}
			ent.Wait((int64)(float32(cost) * (100 / this.Throttle)))

		} else {
			n = strings.ToLower(tok.Content)
			if this.Commands.ContainsKey(n) {
				cmd = this.Commands[n]

				if !cmd.IsStateBased() {
					var e error
					_, e = cmd.Execute(nil, ent, tl, Scope, *LPC)
					if e != nil {
						this.HandleException(ent, e)
					}
					cost := cmd.GetCost()
					if cost == 0 {
						cost = this.DefaultCost
					}
					ent.Wait((int64)(float32(cost) * (100 / this.Throttle)))
				} else {
					// Setup state based command
					// TODO: Implement state based bootstrap
					cs := interfaces.NewCommandState(cmd)
					cs.Params = tl
					cs.Scope = Scope
					cs.PC = *LPC
					ent.SetSubState(types.ESS_INIT)
					ent.SetCommandState(cs)
					return nil
				}

			} else {
				//writeln("DOES NOT EXIST!");
			}
		}

	} else if (tok.Type == types.VARIABLE) && (this.ImpliedAssign != nil) {

		//System.Out.Println("THE TOKEN TRIGGERING ImpliedAssign IS "+tok.Content);

		/* assign variable here */
		tl.UnShift(&tok)
		cmd = this.ImpliedAssign

		var e error
		_, e = cmd.Execute(nil, ent, tl, Scope, *LPC)
		if e != nil {
			this.HandleException(ent, e)
		}
		cost := cmd.GetCost()
		if cost == 0 {
			cost = this.DefaultCost
		}
		ent.Wait((int64)(float32(cost) * (100 / this.Throttle)))

	} else if (tok.Type == types.PLUSVAR) && (this.PlusHandler != nil) {

		/* assign variable here */
		tl.UnShift(&tok)
		cmd = this.PlusHandler

		_, err := cmd.Execute(nil, ent, tl, Scope, *LPC)
		if err != nil {
			this.HandleException(ent, err)
		}

		cost := cmd.GetCost()
		if cost == 0 {
			cost = this.GetDefaultCost()
		}
		ent.Wait((int64)(float32(cost) * (100 / this.Throttle)))

	} else {
		return exception.NewESyntaxError("SYNTAX ERROR")
	}

	///// removed free call here /* clean up */

	return nil

}

func (this *DialectAppleInteger) Tokenize(s runestring.RuneString) *types.TokenList {

	/* vars */
	var result *types.TokenList
	var inq bool
	var inqq bool
	var cont bool
	var idx int
	var chunk string
	var ch rune
	var tt *types.Token

	////fmt.Println("DialectAppleInteger.Tokenize()")

	result = types.NewTokenList()

	inq = false
	inqq = false
	idx = 0
	chunk = ""
	cont = true

	for idx < len(s.Runes) {
		ch = s.Runes[idx]

		//System.Out.Println ( "inqq = "+inqq );

		//writeln("Tokenizer sees ["+ch+']');

		if this.IsWS(ch) && (inq || inqq) {
			chunk = chunk + string(ch)
		} else if this.IsVarSuffix(ch, this.VarSuffixes) && (!(inq || inqq)) {
			chunk = chunk + string(ch)
			if len(chunk) > 0 {
				cont = this.Evaluate(chunk, result)
				chunk = ""
			}
		} else if this.IsBreakingCharacter(ch, this.VarSuffixes, chunk) && (!(inq || inqq)) {

			//writeln("====== breaking char "+ch);

			/* special handling for x.Yyyye+/-xx notation */
			if ((ch == '+') || (ch == '-')) &&
				((len(chunk) >= 2) && (chunk[len(chunk)-1] == 'e') && (this.IsDigit(rune(chunk[0]))) && (this.IsDigit(rune(chunk[len(chunk)-2])))) {
				chunk = chunk + string(ch)
			} else {
				if len(chunk) > 0 {
					cont = this.Evaluate(chunk, result)

					/* if no (, treat func as variable name */
					if (result.Size() > 0) && (result.RPeek().Type == types.FUNCTION) {
						if (ch != '(') && !((idx < len(s.Runes)-1) && (s.Runes[idx+1] == '(')) {
							result.RPeek().Type = types.VARIABLE
						}
					}

					chunk = ""
				}

				if !this.IsWS(ch) {
					chunk = chunk + string(ch)
					cont = this.Evaluate(chunk, result)
					chunk = ""
				}
			}

		} else if this.IsQ(ch) && (!inqq) {
			if (len(chunk) > 0) && (!inq) {
				this.Evaluate(chunk, result)
				chunk = ""
			}
			inq = !inq
			chunk = chunk + string(ch)
		} else if this.IsQQ(ch) && (!inq) {
			if (len(chunk) > 0) && (!inqq) {
				this.Evaluate(chunk, result)
				chunk = "\""
			} else if (len(chunk) > 0) && (inqq) {
				chunk = chunk + "\""
				this.Evaluate(chunk, result)
				chunk = ""
			} else {
				chunk = chunk + string(ch)
			}
			inqq = !inqq
		} else {
			chunk = chunk + string(ch)

			/* break keywords out early */
			if this.Commands.ContainsKey(strings.ToLower(chunk)) {
				if (strings.ToLower(chunk) != "go") && (strings.ToLower(chunk) != "to") && (strings.ToLower(chunk) != "new") && (strings.ToLower(chunk) != "at") && (strings.ToLower(chunk) != "gr") {
					//System.Out.Println("Break out keyword for ["+chunk+"]");
					cont = this.Evaluate(chunk, result)
					chunk = ""
				}
			}

		}

		idx++

		if !cont {
			chunk = ""
			for idx < len(s.Runes) {
				chunk = chunk + string(s.Runes[idx])
				idx++
			}

			tt = types.NewToken(types.UNSTRING, strings.Trim(chunk, " "))
			chunk = ""
			result.Push(tt)
		}

	} /*while*/

	//writeln("chunk == ", chunk);

	if len(chunk) > 0 {
		this.Evaluate(chunk, result)
		/* if no (, treat func as variable name */
		if (result.Size() > 0) && (result.RPeek().Type == types.FUNCTION) {
			//if (ch != '(') {
			result.RPeek().Type = types.VARIABLE
			//}
		}
		chunk = ""
	}

	/* enforce non void return */
	return result

}

func (this *DialectAppleInteger) ParseTokensForResult(ent interfaces.Interpretable, tokens types.TokenList) (*types.Token, error) {

	/* vars */
	var result *types.Token
	var ops types.TokenList
	var values types.TokenList
	var tidx int
	var rbc int
	var sbc int
	//	var cbc int
	var i int
	var repeats int
	var blev int
	//var tok *types.Token
	var lasttok *types.Token
	//var ntok *types.Token
	var op types.Token
	var a types.Token
	var b types.Token
	var n string
	var rs string
	var v *types.Variable
	var exptok types.TokenList
	var subexpr types.TokenList
	var par types.TokenList
	var err bool
	var rr float64
	var aa float64
	var bb float64
	var dl []int
	var rrb bool
	var fun interfaces.Functioner
	var tla types.TokenListArray
	//	var left bool
	var defindex bool
	var lastop bool
	var hpop int
	//	var code int
	//	var hs int
	//	var sc int
	//int  i;
	//Token tt;

	var sexpr string

	result = types.NewToken(types.INVALID, "")

	/* must be 1 || more tokens in list */
	if tokens.Size() == 0 {
		return result, nil
	}

	if ent.IsDebug() && (tokens.Size() > 1 || tokens.Get(0).Type == types.VARIABLE) {
		sexpr = this.TokenListAsString(tokens)
	}

	//writeln("*** Called to parse: ", ent.TokenListAsString(tokens));

	values = *types.NewTokenList()
	ops = *types.NewTokenList()

	if tokens.Get(0).Type == types.OPERATOR {
		//writeln("BEEP");
		values.Push(types.NewToken(types.INTEGER, "0"))
	}

	/* init parser state */
	tidx = 0
	rbc = 0
	//	cbc = 0
	sbc = 0

	//fInterpreter.VDU.PutStr("Expression in: "+this.TokenListAsString(tokens)+PasUtil.CRLF);
	lasttok = nil
	lastop = false
	/*main parse loop*/
	for tidx < tokens.Size() {
		tok := tokens.Get(tidx)

		//writeln( "--------------> type of token at tidx ", tidx, " is ", tok.Type );

		if tok.Type == types.LABEL {

			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			if ent.GetLabel(tok.Content[1:]) != 0 {
				// label
				ntok := types.NewToken(types.NUMBER, utils.IntToStr(ent.GetLabel(tok.Content[1:])))
				values.Push(ntok)
			}

			lastop = false

		} else if (tok.Type == types.NUMBER) || (tok.Type == types.INTEGER) || (tok.Type == types.STRING) || (tok.Type == types.BOOLEAN) {
			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			values.Push(tok)
			lastop = false
		} else if tok.Type == types.ASSIGNMENT {
			ops.Push(tok)
			lastop = true
		} else if tok.Type == types.COMPARITOR {
			ops.Push(tok)
			lastop = true
		} else if tok.Type == types.LOGIC {
			ops.Push(tok)
			lastop = true
		} else if tok.Type == types.OPERATOR {
			if (lasttok != nil) && ((lasttok.Type == types.ASSIGNMENT) || (lasttok.Type == types.COMPARITOR)) && (tok.Content == "-") {
				values.Push(types.NewToken(types.NUMBER, "0"))
			}

			ops.Push(tok)
			lastop = true

		} else if tok.Type == types.FUNCTION {

			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			fun = this.Functions.Get(strings.ToLower(tok.Content))
			if fun == nil {
				return result, exception.NewESyntaxError("unknown function: " + tok.Content)
			}

			//fInterpreter.VDU.PutStr(fun.GetName()+"(\" + IntToStr(length(fun.FunctionParams))+\")"+PasUtil.CRLF);

			if len(fun.FunctionParams()) > 0 {

				/* okay must be bracket after */
				if tidx >= tokens.Size() {
					return result, exception.NewESyntaxError(fun.GetName() + " requires params (size)")
				}

				// must be an index;
				sbc = 1
				tidx = tidx + 1
				subexpr = *types.NewTokenList()

				for (tidx < tokens.Size()) && (sbc > 0) {
					tok = tokens.Get(tidx)
					if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
						sbc = sbc + 1
					}
					if tok.Type == types.FUNCTION {
						sbc = sbc + 1
					}
					if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
						sbc = sbc - 1
					}
					if sbc > 0 {
						subexpr.Push(tok)
					}
					if (tidx < tokens.Size()) && (sbc > 0) {
						tidx = tidx + 1
					}
				}

				/* now condense this list down */
				tla = types.NewTokenListArray()
				exptok = *types.NewTokenList()
				blev = 0

				for _, tok1 := range subexpr.Content {
					if (tok1.Type == types.SEPARATOR) && (tok1.Content == ",") && (blev == 0) {
						if exptok.Size() > 0 {
							//SetLength(tla, tla.Size()+1);
							tla = tla.Add(exptok)
							exptok = *types.NewTokenList()
						}
					} else {

						if (tok1.Type == types.OBRACKET) && (tok1.Content == "(") {
							blev = blev + 1
						}

						if tok1.Type == types.FUNCTION {
							blev = blev + 1
						}

						if (tok1.Type == types.CBRACKET) && (tok1.Content == ")") {
							blev = blev - 1
						}

						exptok.Push(tok1)
					}
				}

				if exptok.Size() > 0 {
					//SetLength(tla, tla.Size()+1);
					tla = tla.Add(exptok)
				}

				par = *types.NewTokenList()
				for _, exptok1 := range tla {
					//writeln( fun.GetName(), ": ", this.TokenListAsString(exptok) );
					if fun.GetRaw() {
						if par.Size() > 0 {
							par.Push(types.NewToken(types.SEPARATOR, ","))
						}
						for _, tok1 := range exptok1.Content {
							par.Push(tok1)
						}
					} else {
						var e error
						tok, e = this.ParseTokensForResult(ent, exptok1)
						if e != nil {
							return result, e
						}
						par.Push(tok)
					}
					//FreeAndNil(exptok);
				}
			} else {
				par = *types.NewTokenList()
			}

			fun.SetEntity(ent)
			e := fun.FunctionExecute(par.Copy())
			if e != nil {
				return result, e
			}
			values.Push(fun.GetStack().Pop())
			lastop = false

		} else if tok.Type == types.PLUSFUNCTION {

			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			fun, exists, ns, err := this.PlusFunctions.GetFunctionByNameContext(this.CurrentNameSpace, strings.ToLower(tok.Content))
			//fun, exists := this.PlusFunctions[strings.ToLower(tok.Content)]

			if fun == nil || !exists {
				return result, err
			}

			this.CurrentNameSpace = ns
			qq := fun.IsQuery()
			fun.SetQuery(true)

			//fInterpreter.VDU.PutStr(fun.GetName()+"(\" + IntToStr(length(fun.FunctionParams))+\")"+PasUtil.CRLF);

			if len(fun.FunctionParams()) > 0 {

				/* okay must be bracket after */
				if tidx >= tokens.Size() {
					return result, exception.NewESyntaxError(fun.GetName() + " requires params (size)")
				}

				// must be an index;
				sbc = 1
				tidx = tidx + 1
				subexpr = *types.NewTokenList()

				for (tidx < tokens.Size()) && (sbc > 0) {
					tok = tokens.Get(tidx)
					if (tok.Type == types.OBRACKET) && (tok.Content == "{") {
						sbc = sbc + 1
					}
					if tok.Type == types.PLUSFUNCTION {
						sbc = sbc + 1
					}
					if (tok.Type == types.CBRACKET) && (tok.Content == "}") {
						sbc = sbc - 1
					}
					if sbc > 0 {
						subexpr.Push(tok)
					}
					if (tidx < tokens.Size()) && (sbc > 0) {
						tidx = tidx + 1
					}
				}

				/* now condense this list down */
				tla = types.NewTokenListArray()
				exptok = *types.NewTokenList()
				blev = 0

				for _, tok1 := range subexpr.Content {
					if (tok1.Type == types.SEPARATOR) && (tok1.Content == ",") && (blev == 0) {
						if exptok.Size() > 0 {
							//SetLength(tla, tla.Size()+1);
							tla = tla.Add(exptok)
							exptok = *types.NewTokenList()
						}
					} else {

						if (tok1.Type == types.OBRACKET) && (tok1.Content == "(") {
							blev = blev + 1
						}

						if tok1.Type == types.FUNCTION {
							blev = blev + 1
						}

						if (tok1.Type == types.CBRACKET) && (tok1.Content == ")") {
							blev = blev - 1
						}

						exptok.Push(tok1)
					}
				}

				if exptok.Size() > 0 {
					//SetLength(tla, tla.Size()+1);
					tla = tla.Add(exptok)
				}

				par = *types.NewTokenList()
				for _, exptok1 := range tla {
					//writeln( fun.GetName(), ": ", this.TokenListAsString(exptok) );
					if fun.GetRaw() {
						if par.Size() > 0 {
							par.Push(types.NewToken(types.SEPARATOR, ","))
						}
						for _, tok1 := range exptok1.Content {
							par.Push(tok1)
						}
					} else {
						var e error
						tok, e = this.ParseTokensForResult(ent, exptok1)
						if e != nil {
							return result, e
						}
						par.Push(tok)
					}
					//FreeAndNil(exptok);
				}
			} else {
				par = *types.NewTokenList()
			}

			fun.SetEntity(ent)
			fun.FunctionExecute(par.Copy())
			fun.SetQuery(qq)
			values.Push(fun.GetStack().Pop())
			lastop = false
		} else if (tok.Type == types.VARIABLE) || (tok.Type == types.KEYWORD) {

			// fix for missing + || separators;
			if (lastop == false) && (lasttok != nil) {
				ops.Push(types.NewToken(types.OPERATOR, "+"))
			}

			//~ if ent.GetLabel( tok.Content ) != 0 {
			//~ // label
			//~ ntok := types.NewToken( types.NUMBER, utils.IntToStr(ent.GetLabel( tok.Content )) )
			//~ values.Push(ntok)
			//~ tidx = tidx + 1
			//~ continue
			//~ }

			/* first try entity local */
			n = strings.ToLower(tok.Content)
			v = nil
			if ent.GetLocal().ContainsKey(n) {
				v = ent.GetLocal().Get(n)
			} else {
				//if (this.Producer.Global.ContainsKey(n));
				if ent.ExistsVar(n) {
					//v = this.Producer.Global.Get(n);
					v = ent.GetVar(n)
				}
			}
			/* fall out if (var does not exist */
			if v == nil {
				//result.Type = types.INVALID;
				//return result;
				if n[len(n)-1] == '$' {
					values.Push(types.NewToken(types.STRING, ""))
				} else {
					values.Push(types.NewToken(types.NUMBER, "0"))
				}
			} else if v.IsArray() {
				defindex = false
				tidx = tidx + 1
				if tidx >= tokens.Size() {
					if v.AssumeLowIndex {
						defindex = true
					} else {
						return result, exception.NewESyntaxError(v.Name + " requires index")
					}
				} else {
					tok = tokens.Get(tidx)
				}

				if (tok.Type != types.OBRACKET) || (tok.Content != "(") {
					if v.AssumeLowIndex {
						defindex = true
					} else {
						return result, exception.NewESyntaxError(v.Name + " requires index")
					}
				}

				if defindex {
					tidx = tidx - 1
				}

				// must be an index;
				if !defindex {
					sbc = 1
					tidx = tidx + 1
					subexpr = *types.NewTokenList()
					subexpr.Push(types.NewToken(types.OBRACKET, "("))
					for (tidx < tokens.Size()) && (sbc > 0) {
						tok = tokens.Get(tidx)
						if tok.Type == types.FUNCTION {
							sbc = sbc + 1
						}
						if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
							sbc = sbc + 1
						}
						if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
							sbc = sbc - 1
						}
						if sbc > 0 {
							subexpr.Push(tok)
						}
						if (tidx < tokens.Size()) && (sbc > 0) {
							tidx = tidx + 1
						}
					}
				}

				//this.VDU.PutStr('!');

				if defindex {
					dl = make([]int, len(v.Dimensions()))
					for i = 0; i < len(dl); i++ {
						dl[i] = 0
					}
					//this.VDU.PutStr("[\"+v.Name+\" assume def index 0]");
				} else {
					subexpr.Push(types.NewToken(types.CBRACKET, ")"))
					var e error
					dl, e = ent.IndicesFromTokens(subexpr, "(", ")")
					if e != nil {
						return result, exception.NewESyntaxError("invalid indices")
					}
					//FreeAndNil(subexpr);
				}

				/* var exists */
				vv, ee := v.GetContentByIndex(this.ArrayDimDefault, this.ArrayDimMax, dl)
				if ee != nil {
					return result, ee
				}
				ntok := types.NewToken(types.INVALID, vv)
				//  VariableType == (vtString, vtBoolean, vtFloat, vtInteger, vtExpression);
				switch v.Kind { /* FIXME - Switch statement needs cleanup */
				case types.VT_STRING:
					ntok.Type = types.STRING
					break
				case types.VT_FLOAT:
					ntok.Type = types.NUMBER
					break
				case types.VT_INTEGER:
					ntok.Type = types.INTEGER
					break
				case types.VT_BOOLEAN:
					ntok.Type = types.BOOLEAN
					break
				case types.VT_EXPRESSION:
					{
						vv, ee = v.GetContentByIndex(this.ArrayDimDefault, this.ArrayDimMax, dl)
						if ee != nil {
							return result, ee
						}
						exptok = *this.Tokenize(runestring.Cast(vv))
						ntok, ee = this.ParseTokensForResult(ent, exptok)
						if ee != nil {
							return result, ee
						}
					}
				}
				//this.VDU.PutStr("// Adding value from array "+ntok.Content);
				values.Push(ntok)

			} else {

				/* var exists */
				vv, ee := v.GetContentScalar()
				if ee != nil {
					return result, ee
				}
				ntok := types.NewToken(types.INVALID, vv)
				//  VariableType == (vtString, vtBoolean, vtFloat, vtInteger, vtExpression);
				switch v.Kind { /* FIXME - Switch statement needs cleanup */
				case types.VT_STRING:
					{
						ntok.Type = types.STRING
						tidx = tidx + 1

						if tidx < tokens.Size() {

							tok = tokens.Get(tidx)

							if (tok.Type == types.OBRACKET) && (tok.Content == "(") {

								sbc = 1
								tidx = tidx + 1
								subexpr = *types.NewTokenList()
								subexpr.Push(types.NewToken(types.OBRACKET, "("))
								for (tidx < tokens.Size()) && (sbc > 0) {
									tok = tokens.Get(tidx)
									if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
										sbc = sbc + 1
									}
									if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
										sbc = sbc - 1
									}
									if sbc > 0 {
										subexpr.Push(tok)
									}
									if (tidx < tokens.Size()) && (sbc > 0) {
										tidx = tidx + 1
									}
								}

								subexpr.Push(types.NewToken(types.CBRACKET, ")"))

								//ent.GetVDU().PutStr("*** "+ent.TokenListAsString(subexpr)+#13+#10);

								/* have indices */
								var ee error
								dl, ee = ent.IndicesFromTokens(subexpr, "(", ")")
								if ee != nil {
									return result, ee
								}

								//ent.GetVDU().PutStr(IntToStr(dl.Size())+PasUtil.CRLF);

								if len(dl) > 2 {
									return result, exception.NewESyntaxError("Bad string indices")
								}

								if len(dl) == 1 {
									ntok.Content = utils.Copy(ntok.Content, dl[0], len(ntok.Content)-dl[0]+1)
								} else {
									ntok.Content = utils.Copy(ntok.Content, dl[0], dl[1]-dl[0]+1)
								}

							} else {
								tidx = tidx - 1
							}

						} else {
							tidx = tidx - 1
						}
						break
					}
				case types.VT_FLOAT:
					ntok.Type = types.NUMBER
					break
				case types.VT_INTEGER:
					ntok.Type = types.INTEGER
					break
				case types.VT_BOOLEAN:
					ntok.Type = types.BOOLEAN
					break
				case types.VT_EXPRESSION:
					{
						vv, ee := v.GetContentScalar()
						if ee != nil {
							return result, ee
						}
						exptok = *this.Tokenize(runestring.Cast(vv))
						ntok, ee = this.ParseTokensForResult(ent, exptok)
						if ee != nil {
							return result, ee
						}
					}
				}
				values.Push(ntok)

			}

			lastop = false
		} else if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
			rbc = 1
			tidx = tidx + 1
			subexpr = *types.NewTokenList()
			for (tidx < tokens.Size()) && (rbc > 0) {
				tok = tokens.Get(tidx)
				if (tok.Type == types.OBRACKET) && (tok.Content == "(") {
					rbc = rbc + 1
				}
				if tok.Type == types.FUNCTION {
					rbc = rbc + 1
				}
				if (tok.Type == types.CBRACKET) && (tok.Content == ")") {
					rbc = rbc - 1
				}
				if rbc > 0 {
					subexpr.Push(tok)
				}
				if (tidx < tokens.Size()) && (rbc > 0) {
					tidx = tidx + 1
				}
			}

			//writeln( "*** Must parse bracketed subexpression: ", this.TokenListAsString(subexpr) );
			//writeln( "=== rbc == ",rbc );

			if rbc > 0 {
				return result, exception.NewESyntaxError("SYNTAX ERROR")
			}

			/* solve the expression */
			nntok, ee := this.ParseTokensForResult(ent, subexpr)
			if ee != nil {
				return result, ee
			}
			//FreeAndNil(subexpr);
			values.Push(nntok)
			lastop = false
		}

		lasttok = tok
		tidx = tidx + 1
	}

	/* now we have ops, values etc - lets actually parse the expression */
	//writeln("Op stack");
	//for _, ntok := range ops
	//    Writeln("OP: ", ntok.Type, ", ", ntok.Content);
	//for _, ntok := range values
	//    Writeln("VALUE: ", ntok.Type, ", ", ntok.Content);
	//writeln("--------");

	/* This bit is some magic || something yea!!! Whoo ---------------- */
	/* End magic bits ------------------------------------------------- */

	/* process */
	err = false

	for (ops.Size() > 0) && (!err) {

		hpop = this.HPOpIndex(ops)
		op = *ops.Remove(hpop)

		if op.Type == types.LOGIC {

			if strings.ToLower(op.Content) == "not" {

				//a = values.Pop();
				a = *values.Remove(hpop)

				/*if (a.AsInteger() != 0)
				      values.Push( types.NewToken(types.NUMBER, "0") )
				  else {
				      values.Push( types.NewToken(types.NUMBER, '1') );*/
				//}
				var ntok *types.Token
				if a.AsInteger() != 0 {
					ntok = types.NewToken(types.NUMBER, "0")
				} else {
					ntok = types.NewToken(types.NUMBER, "1")
				}

				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "and" {
				//b = values.Pop();
				//b = values.Pop();
				//a = values.Pop();
				if hpop == values.Size()-1 {
					hpop--
				}
				a = *values.Remove(hpop)
				b = *values.Remove(hpop)

				//this.VDU.PutStr(a.Content+" \"+op.Content+\" "+b.Content+PasUtil.CRLF);
				vv := 0
				if (a.AsInteger() != 0) && (b.AsInteger() != 0) {
					vv = 1
				}
				//values.Push( types.NewToken(types.NUMBER, IntToStr(a.AsInteger() && b.AsInteger())) );
				ntok := types.NewToken(types.NUMBER, utils.IntToStr(vv))
				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "or" {
				//b = values.Pop();
				//a = values.Pop();
				if hpop == values.Size()-1 {
					hpop--
				}
				a = *values.Remove(hpop)
				b = *values.Remove(hpop)
				vv := 0
				if (a.AsInteger() != 0) || (b.AsInteger() != 0) {
					vv = 1
				}
				//values.Push( types.NewToken(types.NUMBER, IntToStr(a.AsInteger() || b.AsInteger())) );
				ntok := types.NewToken(types.NUMBER, utils.IntToStr(vv))
				values.Insert(hpop, ntok)
			} else if strings.ToLower(op.Content) == "xor" {
				//b = values.Pop();
				//a = values.Pop();
				if hpop == values.Size()-1 {
					hpop--
				}
				a = *values.Remove(hpop)
				b = *values.Remove(hpop)

				//values.Push( types.NewToken(types.NUMBER, IntToStr(a.AsInteger() xor b.AsInteger())) );
				ntok := types.NewToken(types.NUMBER, utils.IntToStr(a.AsInteger()^b.AsInteger()))
				values.Insert(hpop, ntok)
			}

		} else if (op.Type == types.COMPARITOR) || (op.Type == types.ASSIGNMENT) {

			if values.Size() < 2 {
				return result, exception.NewESyntaxError("invalid expression")
			}

			//writeln("@@@@@@@@@@@@@@@@@@@@@ COMPARE");

			/*if (left) {
			      a = values.Left
			      b = values.Left
			  } else {
			      b = values.Right
			      a = values.Right
			  }*/
			if hpop == values.Size()-1 {
				hpop--
			}
			a = *values.Remove(hpop)
			b = *values.Remove(hpop)

			//this.VDU.PutStr(a.Content+" \"+op.Content+\" "+b.Content+PasUtil.CRLF);

			if a.IsNumeric() && b.IsNumeric() {

				aa = float64(a.AsNumeric())
				bb = float64(b.AsNumeric())
				rrb = false

				//writeln("====================> NUMBER COMPARE a == [",a.Content,"], b == [",b.Content,']');

				if op.Content == ">" {
					rrb = (aa > bb)
				} else if op.Content == "<" {
					rrb = (aa < bb)
				} else if op.Content == ">=" {
					rrb = (aa >= bb)
				} else if op.Content == "<=" {
					rrb = (aa <= bb)
				} else if (op.Content == "<>") || (op.Content == "#") {
					rrb = (aa != bb)
				} else if op.Content == "=" {
					rrb = (aa == bb)
				} else {
					err = true
					break
				}

				n = "0"
				if rrb {
					n = "1"
				}
				ntok := types.NewToken(types.NUMBER, n)

				/*if (left)
				     values.UnShift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)
			} else if (!a.IsNumeric()) || (!b.IsNumeric()) {

				//				tt := types.NUMBER /* most results are string */
				rrb = false

				//writeln("====================> STRING COMPARE a == [",a.Content,"], b == [",b.Content,']');

				if op.Content == ">" {
					rrb = (a.Content > b.Content)
				} else if op.Content == "<" {
					rrb = (a.Content < b.Content)
				} else if op.Content == "<>" || op.Content == "#" {
					rrb = (a.Content != b.Content)
				} else if op.Content == ">=" {
					rrb = (a.Content >= b.Content)
				} else if op.Content == "<=" {
					rrb = (a.Content <= b.Content)
				} else if op.Content == "=" {
					rrb = (a.Content == b.Content)
				} else {
					err = true
					break
				}

				n = "0"
				if rrb {
					n = "1"
				}
				ntok := types.NewToken(types.NUMBER, n)
				/*if (left)
				     values.UnShift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)

			}

		} else if op.Type == types.OPERATOR {

			//writeln("Currently ",values.Size()," values in stack... About to pop 2");

			if values.Size() < 2 {
				return result, exception.NewESyntaxError("invalid expression")
			}

			/*if (left) {
			      a = values.Left
			      b = values.Left
			  } else {
			      b = values.Right
			      a = values.Right
			  }*/
			if hpop == values.Size()-1 {
				hpop--
			}
			a = *values.Remove(hpop)
			b = *values.Remove(hpop)

			//fInterpreter.VDU.PutStr("Op is: "+a.Content+op.Content+b.Content+PasUtil.CRLF);

			if a.IsNumeric() && b.IsNumeric() {

				aa = float64(a.AsNumeric())
				bb = float64(b.AsNumeric())
				rr = 0

				if op.Content == "+" {
					rr = aa + bb
				} else if op.Content == "-" {
					rr = aa - bb
				} else if op.Content == "*" {
					rr = aa * bb
				} else if op.Content == "/" {
					if bb == 0 {
						return result, exception.NewESyntaxError(">32767 ERROR")
					} else {
						rr = aa / bb
					}
				} else if op.Content == "^" {
					rr = math.Pow(aa, bb)
				} else if strings.ToLower(op.Content) == "mod" {
					rr = float64(int(aa) % int(bb))
				} else {
					err = true
					break
				}

				rr = math.Trunc(rr)

				n = utils.FloatToStr(rr)
				ntok := types.NewToken(types.NUMBER, n)
				//}
				values.Insert(hpop, ntok)
			} else if (a.IsNumeric() && (!b.IsNumeric())) || (b.IsNumeric() && (!a.IsNumeric())) {

				tt := types.STRING /* most results are string */

				if op.Content == "+" {
					rs = a.Content + b.Content
				} else if op.Content == "-" {
					// remove b from a, ignoring case;
					// rs = StringReplace( a.Content, b.Content, "", [rfReplaceAll, rfIgnoreCase] );
					rs = strings.Replace(a.Content, b.Content, "", -1)
				} else if op.Content == "*" {
					repeats = 0
					rs = ""
					n = ""
					if a.IsNumeric() {
						repeats = int(a.AsNumeric())
						n = b.Content
					} else {
						repeats = int(b.AsNumeric())
						n = a.Content
					}
					for i = 1; i <= repeats; i++ {
						rs = rs + n
					}
				} else if op.Content == "/" {
					// divide string by substring - count occurrences;
					tt = types.INTEGER

					repeats = 0

					rs = utils.IntToStr(repeats)
				} else {
					err = true
					break
				}

				ntok := types.NewToken(tt, rs)
				//writeln("CREATE RESULT: ", rs);
				/*if (left)
				     values.UnShift(ntok)
				  else {
				      values.Push(ntok);*/
				// }
				values.Insert(hpop, ntok)

			} else if (!a.IsNumeric()) && (!b.IsNumeric()) {

				tt := types.STRING /* most results are string */

				if op.Content == "+" {
					rs = a.Content + b.Content
				} else if op.Content == "-" {
					// remove b from a, ignoring case;
					// rs = StringReplace( a.Content, b.Content, "", [rfReplaceAll, rfIgnoreCase] );
					rs = strings.Replace(a.Content, b.Content, "", -1)
				} else if op.Content == "/" {
					// divide string by substring - count occurrences;
					tt = types.INTEGER

					repeats = 0

					rs = utils.IntToStr(repeats)
				} else {
					err = true
					break
				}

				ntok := types.NewToken(tt, rs)
				/*if (left)
				     values.UnShift(ntok)
				  else {
				      values.Push(ntok);*/
				//}
				values.Insert(hpop, ntok)
			}

		}

	}

	if err || (values.Size() > 1) {
		if ent.IsDebug() && (tokens.Size() > 1 || tokens.Get(0).Type == types.VARIABLE) {
			ent.Log("EVAL", sexpr+" == Error!")
		}
		return result, exception.NewESyntaxError("syntax error")
	}

	result = values.Pop()

	result = types.NewToken(result.Type, ""+result.Content)

	if result.IsNumeric() {

		//fmt.Println("RES =", result.AsExtended())

		v := int(math.Trunc(result.AsExtended()))

		result.Content = utils.IntToStr(v)
	}

	// removed free call here;
	// removed free call here;
	if ent.IsDebug() && (tokens.Size() > 1 || tokens.Get(0).Type == types.VARIABLE) {
		ent.Log("EVAL", sexpr+" == "+result.AsString())
	}

	/* enforce non void return */
	return result, nil

}

func (this *DialectAppleInteger) IsBreakingCharacter(ch rune, vs string, chunk string) bool {
	result := false
	bc := " \r\n\t+-*/^()[]{}=;:,<>?"
	if strings.IndexRune(bc, ch) > -1 {
		result = true
	}
	if ch == '#' {
		if strings.ToLower(chunk) == "pr" || strings.ToLower(chunk) == "in" {
			result = false
		} else {
			result = true
		}
	}
	log.Println("IsBreakingCharacter("+string(ch)+") =", result)
	return result
}

func (this *DialectAppleInteger) IsOperator(ch string) bool {

	/* vars */
	var result bool

	result = (ch == "+") || (ch == "-") || (ch == "/") || (ch == "*") ||
		(ch == "^") || (ch == "&") || (ch == "|") || (strings.ToLower(ch) == "mod") ||
		(ch == ";") || (ch == "@")

	/* enforce non void return */
	return result

}

func (this *DialectAppleInteger) Init() {

	this.ShortName = "int"
	this.LongName = "Integer BASIC"

	/* vars */

	/*title*/
	//this.AddCommand("tracker", applesoft.NewStandardCommandTRACKER())
	this.AddCommand("exit", applesoft.NewStandardCommandEXIT())
	//this.AddCommand("spawn", &applesoft.StandardCommandSPAWN{})
	this.AddCommand("flush", &applesoft.StandardCommandFLUSH{})
	this.AddCommand("cat", applesoft.NewStandardCommandCAT())
	this.AddCommand("package", applesoft.NewStandardCommandPACKAGE())
	this.AddCommand("catalog", applesoft.NewStandardCommandCAT())
	this.AddCommand("mon", applesoft.NewStandardCommandMON())
	this.AddCommand("xlist", applesoft.NewStandardCommandXLIST())
	this.AddCommand("edit", applesoft.NewStandardCommandEDIT())
	this.AddCommand("feedback", applesoft.NewStandardCommandFEEDBACK())
	this.AddCommand("print", &applesoft.StandardCommandPRINT{})
	this.AddCommand("input", applesoft.NewStandardCommandINPUTH())
	this.AddCommand("?", &applesoft.StandardCommandPRINT{})
	this.AddCommand("gr5", &applesoft.StandardCommandGR5{})
	this.AddCommand("gr6", &applesoft.StandardCommandGR6{})
	this.AddCommand("gr7", &applesoft.StandardCommandGR7{})
	this.AddCommand("autosave", &applesoft.StandardCommandAUTOSAVE{})
	//this.AddCommand("gr8", &applesoft.StandardCommandGR8{}) //this.AddCommand( "declare", applesoft.NewStandardCommandDECLARE());
	//this.AddCommand( "home", applesoft.NewStandardCommandCLS());
	this.AddCommand("list", applesoft.NewStandardCommandLIST())
	this.AddCommand("run", &applesoft.StandardCommandRUN{})
	this.AddCommand("new", &applesoft.StandardCommandNEW{})
	this.AddCommand("goto", &applesoft.StandardCommandGOTO{})
	this.AddCommand("gosub", &applesoft.StandardCommandGOSUB{})
	this.AddCommand("return", &StandardCommandWozRETURN{})
	this.AddCommand("rem", applesoft.NewStandardCommandREM())
	this.AddCommand("end", &applesoft.StandardCommandEND{})
	this.AddCommand("stop", &applesoft.StandardCommandSTOP{})
	this.AddCommand("con", &applesoft.StandardCommandCONT{})
	//this.AddCommand( "else", &applesoft.StandardCommandNOP{} );
	this.AddCommand("then", &applesoft.StandardCommandNOP{})
	this.AddCommand("trace", &applesoft.StandardCommandNOP{})
	this.AddCommand("notrace", &applesoft.StandardCommandNOP{})
	//this.AddCommand( "store", &applesoft.StandardCommandNOP{} );
	//this.AddCommand( "recall", &applesoft.StandardCommandNOP{} );
	//this.AddCommand( "usr", &applesoft.StandardCommandNOP{} );
	this.AddCommand("call", &applesoft.StandardCommandCALL{})
	this.AddCommand("at", &applesoft.StandardCommandNOP{})
	this.AddCommand("to", &applesoft.StandardCommandNOP{})
	this.AddCommand("step", &applesoft.StandardCommandNOP{})
	//this.AddCommand( "wait", &applesoft.StandardCommandNOP{} );
	this.AddCommand("if", &StandardCommandWozIF{})
	this.AddCommand("save", &applesoft.StandardCommandSAVE{})
	this.AddCommand("load", &StandardCommandLOAD{})
	this.AddCommand("for", &StandardCommandWozFOR{})
	this.AddCommand("next", &StandardCommandWozNEXT{})
	//this.AddCommand( "data", applesoft.NewStandardCommandDATA() );
	//this.AddCommand( "read", applesoft.NewStandardCommandREAD() );
	//this.AddCommand( "restore", applesoft.NewStandardCommandRESTORE() );
	this.AddCommand("dim", &StandardCommandWozDIM{})
	this.AddCommand("pop", &StandardCommandWozPOP{})
	this.AddCommand("text", &applesoft.StandardCommandTEXT{})
	this.AddCommand("gr", &applesoft.StandardCommandGR{})
	this.AddCommand("gr2", &applesoft.StandardCommandGR2{})
	this.AddCommand("gr3", &applesoft.StandardCommandGR2{})
	//this.AddCommand( "hgr", applesoft.NewStandardCommandHGR() );
	this.AddCommand("plot", &applesoft.StandardCommandPLOT{})
	//this.AddCommand( "hplot", applesoft.NewStandardCommandHPLOT() );
	this.AddCommand("hlin", &applesoft.StandardCommandHLIN{})
	this.AddCommand("vlin", &applesoft.StandardCommandVLIN{})
	this.AddCommand("poke", &applesoft.StandardCommandPOKE{})
	this.AddCommand("tab", &applesoft.StandardCommandHTAB{})
	this.AddCommand("vtab", &applesoft.StandardCommandVTAB{})
	this.AddCommand("clr", &applesoft.StandardCommandCLEAR{})
	//this.AddCommand("del", &applesoft.StandardCommandDEL{})
	this.AddCommand("pr#", &applesoft.StandardCommandPR{})
	this.AddCommand("in#", &applesoft.StandardCommandQNOP{})
	//this.AddCommand( "onerr", applesoft.NewStandardCommandONERR() );
	//this.AddCommand( "resume", applesoft.NewStandardCommandRESUME() );
	//this.AddCommand( "on", applesoft.NewStandardCommandON() );
	this.AddCommand("lang", applesoft.NewStandardCommandDIALECT())
	this.AddCommand("let", &StandardCommandWozIMPLIEDASSIGN{})
	this.AddCommand("trace", &applesoft.StandardCommandTRACE{})
	this.AddCommand("notrace", &applesoft.StandardCommandNOTRACE{})
	this.AddCommand("dsp", &applesoft.StandardCommandDSP{})
	this.AddCommand("nodsp", &applesoft.StandardCommandNODSP{})
	this.AddCommand("help", &applesoft.StandardCommandHELP{})

	this.AddCommand("renumber", &applesoft.StandardCommandRENUMBER{})
	this.AddCommand("reorganize", &applesoft.StandardCommandREORGANIZE{})

	this.AddCommand("color=", &applesoft.StandardCommandCOLOR{})
	this.AddCommand("speed=", &applesoft.StandardCommandSPEED{})

	/* dummies for now */
	//this.AddCommand( "inverse", applesoft.NewStandardCommandVIDEOINVERSE() );
	//this.AddCommand( "normal", applesoft.NewStandardCommandVIDEONORMAL() );
	//this.AddCommand( "flash", applesoft.NewStandardCommandVIDEOFLASH() );
	this.AddCommand("himem:", &applesoft.StandardCommandHIMEM{})
	this.AddCommand("lomem:", &applesoft.StandardCommandLOMEM{})

	this.ImpliedAssign = &StandardCommandWozIMPLIEDASSIGN{}
	//this.PlusHandler = &applesoft.PlusHandler{}

	/* math functions - TRS-80 LEVEL II */
	this.AddFunction("abs(", applesoft.NewStandardFunctionABS(0, 0, types.TokenList{}))
	this.AddFunction("rnd(", NewStandardFunctionRND(0, 0, types.TokenList{}))
	this.AddFunction("sgn(", applesoft.NewStandardFunctionSGN(0, 0, types.TokenList{}))

	/* string functions - TRS-80 LEVEL II */
	this.AddFunction("asc(", NewStandardFunctionASC(0, 0, types.TokenList{}))           // valid
	this.AddFunction("len(", applesoft.NewStandardFunctionLEN(0, 0, types.TokenList{})) // valid
	this.AddFunction("peek(", applesoft.NewStandardFunctionPEEK())
	this.AddFunction("scrn(", applesoft.NewStandardFunctionSCRN(0, 0, types.TokenList{}))
	this.AddFunction("pdl(", applesoft.NewStandardFunctionPDL(0, 0, types.TokenList{}))

	/*int basic*/
	plus.RegisterFunctions(this)

	/* added functions for fun */
	//this.AddFunction( "time$", types.NewStandardFunctionTIMEDollar(0,0,nil) );

	this.VarSuffixes = "%$"
	this.Logicals["or"] = 1
	this.Logicals["and"] = 1
	this.Logicals["not"] = 1

	/* dynacode test shim */
	//this.AddDynaCommand( "get", applesoft.NewStandardCommandGET() );
	//this.AddDynaCommand( "input", applesoft.NewStandardCommandINPUT() );

	this.ReverseCase = true

	this.ArrayDimDefault = 10
	this.ArrayDimMax = 65535
	this.DefaultCost = 1000000000 / 800

	this.IPS = -1
	this.Title = "INTEGER"

}

func (this *DialectAppleInteger) Parse(ent interfaces.Interpretable, s string) error {

	/* vars */
	var tl *types.TokenList
	//TokenList  cl;
	var cmdlist types.TokenListArray
	//Token tok;
	var lno int
	var ll types.Line
	var st types.Statement

	tl = this.Tokenize(runestring.Cast(s))

	log.Printf("Dialect [%s] received [%s] to parse from [%s]\n", this.GetTitle(), s, ent.GetName())

	//for _, tok := range tl
	//{
	//  Str( tok.Type, n );
	//  Producer.GetInstance().VDU.PutStr("T: "+n+", V: "+tok.Content+"\r\n");
	//}

	if tl.Size() == 0 {
		return nil
	}

	tok := tl.Get(0)

	if tok.Type == types.SEPARATOR && tok.Content == ":" {

		if tl.Size() > 0 {

			// fix for immediate only commands
			for i := 0; i < tl.Size(); i++ {
				if tl.Get(i).Type == types.KEYWORD {
					cmd := this.Commands[strings.ToLower(tl.Get(i).Content)]
					if cmd.ImmediateModeOnly() {
						tl.Get(i).Type = types.VARIABLE
					}
				}
			}

			cmdlist = ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))

			//			////fmt.Println(len(cmdlist))

			lno = this.lastLineNumber

			ll, exists := ent.GetCode().Get(lno)

			if !exists {
				ll = types.NewLine()
			}

			for _, cl := range cmdlist {
				if cl.Size() > 0 {
					st = types.NewStatement()
					for _, tok1 := range cl.Content {
						st.Push(tok1)
					}
					ll = append(ll, st)
				}
			}

			z := ent.GetCode()
			z.Put(lno, ll)
			this.lastLineNumber = lno

			//ent.PutStr("+" + utils.IntToStr(lno) + "\r\n")
			ent.SetState(types.STOPPED)
		}

		//data := this.GetMemoryRepresentation(ent.GetCode())

		if !this.SkipMemParse() {
			data := this.GetMemoryRepresentation(ent.GetCode())

			// write to memory
			addr := 38400 - len(data) - 1
			for i, v := range data {
				ent.SetMemory(addr+1+i, v)
			}
			ent.SetMemory(addr, uint64(addr+len(data)))
		}

		//        types.RebuildAlgo()

		return nil

	} else if tok.Type == types.NUMBER {

		tok = tl.Shift()

		if tl.Size() > 0 {

			// fix for immediate only commands
			for i := 0; i < tl.Size(); i++ {
				if tl.Get(i).Type == types.KEYWORD {
					cmd := this.Commands[strings.ToLower(tl.Get(i).Content)]
					if cmd.ImmediateModeOnly() {
						tl.Get(i).Type = types.VARIABLE
					}
				}
			}

			cmdlist = ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))

			////fmt.Println(len(cmdlist))

			lno = tok.AsInteger()
			this.lastLineNumber = lno
			ll = types.NewLine()

			for _, cl := range cmdlist {
				if cl.Size() > 0 {
					st = types.NewStatement()
					for _, tok1 := range cl.Content {
						st.Push(tok1)
					}
					ll = append(ll, st)
				} /*else {
					st = types.NewStatement()
					st.Push(types.NewToken(types.KEYWORD, "REM"))
					ll = append(ll, st)
				}*/
			}

			z := ent.GetCode()
			z.Put(lno, ll)

			//ent.GetVDU().PutStr("+" + utils.IntToStr(lno) + "\r\n")
			ent.SetState(types.STOPPED)
			//fInterpreter.VDU.Dump;

			if !this.SkipMemParse() {

				if settings.AutosaveFilename[ent.GetMemIndex()] != "" {
					data := this.GetWorkspace(ent)
					files.AutoSave(ent.GetMemIndex(), data)
				}

				data := this.GetMemoryRepresentation(ent.GetCode())

				// write to memory
				addr := 38400 - len(data) - 1
				for i, v := range data {
					ent.SetMemory(addr+1+i, v)
				}
				ent.SetMemory(addr, uint64(addr+len(data)))
			}

		} else {
			lno = tok.AsInteger()
			if _, ok := ent.GetCode().Get(lno); ok {
				z := ent.GetCode()
				z.Remove(lno)
				//ent.GetVDU().PutStr("-"+PasUtil.IntToStr(lno)+"\r\n");
			}
		}

		//        types.RebuildAlgo()
		return nil
	}

	/* at this point we break it into statements */
	cmdlist = ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))
	lno = 999
	ent.SetDirectAlgorithm(types.NewAlgorithm())
	ent.SetLPC(types.NewCodeRef())

	ll = types.NewLine()

	for zz, cl := range cmdlist {
		log.Println(zz)
		if cl.Size() > 0 {
			st = types.NewStatement()
			for _, zz := range cl.Content {
				st.Add(zz)
			}
			ll.Push(st)
			log.Println("Added:", st)
			log.Println("Statement count ", len(ll))
		} else {
			log.Println("cmdlist has no content")
		}
	}

	//ent.GetDirectAlgorithm()[lno] = ll
	a := ent.GetDirectAlgorithm()
	a.Put(lno, ll)
	ent.SetDirectAlgorithm(a)
	ent.GetLPC().Line = lno
	ent.GetLPC().Statement = 0
	ent.GetLPC().Token = 0
	ent.PreOptimizer()

	// start running
	ent.SetState(types.DIRECTRUNNING)

	//ent.GetVDU().PutStr(fmt.Sprintf("Interpreter state is now %v\r\n", ent.GetState()))

	return nil

}

func (this *DialectAppleInteger) PutStr(ent interfaces.Interpretable, s string) {
	apple2helpers.PutStr(ent, s)
}

func (this *DialectAppleInteger) RealPut(ent interfaces.Interpretable, ch rune) {
	apple2helpers.Put(ent, ch)
}

func (this *DialectAppleInteger) Backspace(ent interfaces.Interpretable) {
	apple2helpers.Backspace(ent)
}

func (this *DialectAppleInteger) ClearToBottom(ent interfaces.Interpretable) {
	apple2helpers.ClearToBottom(ent)
}

func (this *DialectAppleInteger) SetCursorX(ent interfaces.Interpretable, xx int) {
	x := (80 / apple2helpers.GetColumns(ent)) * xx

	apple2helpers.SetCursorX(ent, x)
}

func (this *DialectAppleInteger) SetCursorY(ent interfaces.Interpretable, yy int) {
	y := (48 / apple2helpers.GetRows(ent)) * yy

	apple2helpers.SetCursorY(ent, y)
}

func (this *DialectAppleInteger) GetColumns(ent interfaces.Interpretable) int {
	return apple2helpers.GetColumns(ent)
}

func (this *DialectAppleInteger) GetRows(ent interfaces.Interpretable) int {
	return apple2helpers.GetRows(ent)
}

func (this *DialectAppleInteger) Repos(ent interfaces.Interpretable) {
	apple2helpers.Gotoxy(ent, int(ent.GetMemory(36)), int(ent.GetMemory(37)))
}

func (this *DialectAppleInteger) GetCursorX(ent interfaces.Interpretable) int {
	return apple2helpers.GetCursorX(ent) / (80 / apple2helpers.GetColumns(ent))
}

func (this *DialectAppleInteger) GetCursorY(ent interfaces.Interpretable) int {
	return apple2helpers.GetCursorY(ent) / (48 / apple2helpers.GetRows(ent))
}

func fixMemoryPtrs(caller interfaces.Interpretable) {
	data := caller.GetDialect().GetMemoryRepresentation(caller.GetCode())
	addr := 38400 - len(data) - 1
	for i, v := range data {
		caller.SetMemory(addr+1+i, v)
	}
	caller.SetMemory(addr, uint64(addr+len(data)))

	//// Set Lomem after program load
	//caller.SetMemory(106, lm/256)
	//caller.SetMemory(105, lm%256)

	//// Set himem to default after program load
	//hm := uint64(0x9600)
	//caller.SetMemory(116, hm/256)
	//caller.SetMemory(115, hm%256)

	caller.GetDialect().InitVarmap(caller, caller.GetLocal())
}

func (this *DialectAppleInteger) InitVarmap(ent interfaces.Interpretable, vm types.VarManager) {

	fretop := 38400
	varmem := 2048

	// Create an Applesoft compatible memory map
	vmgr := types.NewVarManagerWOZ(
		ent.GetMemoryMap(),
		ent.GetMemIndex(),
		74,
		204,
		202,
		76,
		types.VUR_QUIET,
	)

	vmgr.SetVector(vmgr.VARBOT, varmem)
	vmgr.SetVector(vmgr.VARTOP, varmem)
	vmgr.SetVector(vmgr.BASBOT, fretop)
	vmgr.SetVector(vmgr.BASTOP, fretop)

	ent.SetLocal(vmgr)

}

// NTokenize tokenize a group of tokens to uints
func (this *DialectAppleInteger) NTokenize(tl types.TokenList) []uint64 {

	var values []uint64

	//var lasttok *types.Token

	for _, t := range tl.Content {
		switch t.Type {
		case types.LOGIC:
			values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
		case types.KEYWORD:
			values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
		case types.FUNCTION:
			values = append(values, this.TokenMapping[strings.ToLower(t.Content)])
		case types.STRING:
			values = append(values, 34)
			for _, ch := range t.Content {
				values = append(values, uint64(ch))
			}
			values = append(values, 34)
		default:
			//if lasttok != nil && lasttok.Type != types.KEYWORD && lasttok.Type != types.FUNCTION {
			//	values = append(values, 32)
			//}
			for _, ch := range t.Content {
				values = append(values, uint64(ch))
			}
		}
		//lasttok = t
	}

	return values

}

func (this *DialectAppleInteger) GetMemoryRepresentation(a *types.Algorithm) []uint64 {
	data := make([]uint64, 0)

	// iterate through code
	l := a.GetLowIndex()
	h := a.GetHighIndex()

	for l <= h && l != -1 {
		ln, _ := a.Get(l)

		buffer := make([]uint64, 2) // allocate 2 slots
		buffer[0] = uint64(l)       // line number
		for i, s := range ln {
			tl := *s.SubList(0, s.Size())
			encoded := this.NTokenize(tl)
			buffer = append(buffer, encoded...)
			if i < len(ln)-1 {
				buffer = append(buffer, uint64(':'))
			}
		}
		buffer = append(buffer, 0)
		buffer[1] = uint64(len(data) + len(buffer))
		data = append(data, buffer...)

		// increment
		l = a.NextAfter(l)
	}

	return data
}

func (this *DialectAppleInteger) ParseMemoryRepresentation(data []uint64) types.Algorithm {

	var lno int
	var pos, nextpos int
	var ll types.Line
	var st types.Statement

	if len(data) < 3 {
		return *types.NewAlgorithm()
	}

	a := *types.NewAlgorithm()

	if len(data) < 3 {
		return a
	}

	for pos < len(data) {
		lno = int(data[pos])
		nextpos = int(data[pos+1])
		buffer := data[pos+2 : nextpos-1]

		tl := this.UnNTokenize(buffer)

		cmdlist := this.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))

		////fmt.Println(len(cmdlist))

		ll = types.NewLine()

		for _, cl := range cmdlist {
			if cl.Size() > 0 {
				st = types.NewStatement()
				for _, tok1 := range cl.Content {
					st.Push(tok1)
				}
				ll = append(ll, st)
			} else {
				st = types.NewStatement()
				st.Push(types.NewToken(types.KEYWORD, "REM"))
				ll = append(ll, st)
			}
		}

		a.Put(lno, ll)

		pos = nextpos
	}

	return a

}

func (this *DialectAppleInteger) UnNTokenize(values []uint64) *types.TokenList {

	s := ""
	var lastcode uint64 = 0

	var skipspace bool

	for _, v := range values {
		if v < dialect.TID_COMMAND_BASE {
			if lastcode >= dialect.TID_COMMAND_BASE && !skipspace {
				s = s + " "
			}
			skipspace = false
			if v > 0 {
				s = s + string(rune(v))
			}
		} else {
			s = s + " " + this.MappingToken[v]
			if this.MappingToken[v] == "rem" || this.MappingToken[v] == "data" {
				skipspace = true
			}
		}
		lastcode = v
	}

	tl := this.Tokenize(runestring.Cast(s))

	return tl
}

func (this *DialectAppleInteger) UpdateRuntimeState(ent interfaces.Interpretable) {

	// Run state (227)
	ent.SetMemory(227, uint64(ent.GetState()))

	// Line/Stmt
	ent.SetMemory(117, uint64(ent.GetPC().Line))
	ent.SetMemory(118, uint64(ent.GetPC().Statement))

	// Data/Pointer
	ent.SetMemory(123, uint64(ent.GetDataRef().Line))
	ent.SetMemory(124, uint64(ent.GetDataRef().Statement))
	ent.SetMemory(125, uint64(ent.GetDataRef().Token))
	ent.SetMemory(126, uint64(ent.GetDataRef().SubIndex))

	// Onerr address
	ent.SetMemory(220, uint64(ent.GetErrorTrap().Line))

	// Loop stack
	data, _ := ent.GetLoopStack().MarshalBinary()
	for i, v := range data {
		ent.SetMemory(LOOPSTACK_ADDRESS+i, v)
	}

	// Call stack
	data, _ = ent.GetStack().MarshalBinary()
	for i, v := range data {
		ent.SetMemory(CALLSTACK_ADDRESS+i, v)
	}

	// Loop vars
	data = types.PackName(ent.GetLoopVariable(), 16)
	data = append(data, types.Float2uint(float32(ent.GetLoopStep())))
	for i, v := range data {
		ent.SetMemory(0xfd00+i, v)
	}

}

func (this *DialectAppleInteger) DumpState(ent interfaces.Interpretable) {

	//	//fmt.Printf("- Entity State: %v\n", ent.GetState())
	//	//fmt.Printf("- Entity PC   : line %d, stmt %d\n", ent.GetPC().Line, ent.GetPC().Statement)
	//	//fmt.Printf("- Error trap  : line %d, stmt %d\n", ent.GetErrorTrap().Line, ent.GetErrorTrap().Statement)
	//	//fmt.Printf("- Data pointer: line %d, stmt %d, token %d, subindex %d\n", ent.GetDataRef().Line, ent.GetDataRef().Statement, ent.GetDataRef().Token, ent.GetDataRef().SubIndex)
	//	//fmt.Printf("- Loopstack   : %d entries, %f, %s\n", ent.GetLoopStack().Size(), ent.GetLoopStep(), ent.GetLoopVariable())
	//	//fmt.Printf("- Callstack   : %d entries\n", ent.GetStack().Size())
	//	vkeys := ent.GetLocal().GetVarNames()
	//	//fmt.Printf("- Variables   : %d, %v\n", len(vkeys), vkeys)
	//	//fmt.Printf("- Program size: %d bytes\n", ent.GetMemory(2048)-2049)

}

func (this *DialectAppleInteger) PreFreeze(ent interfaces.Interpretable) {

	this.UpdateRuntimeState(ent)
	this.DumpState(ent)

}

func (this *DialectAppleInteger) PostThaw(ent interfaces.Interpretable) {

	// reload softswitches
	_ = ent.GetMemory(65536 - 16304)

	data := make([]uint64, 0)
	e := int(ent.GetMemory(2048))
	for i := 2049; i < e; i++ {
		data = append(data, ent.GetMemory(i))
	}

	//	//fmt.Printf("PostThaw, says program ends at %d\n", e)

	// now unpack program again
	if len(data) > 0 {
		a := this.ParseMemoryRepresentation(data)
		ent.SetCode(&a)
	} else {
		ent.SetCode(types.NewAlgorithm())
	}

	fixMemoryPtrs(ent)

	//ent.GetLocal().Mgr = ent
	//ent.GetLocal().Defrost(true)

	// Cursor
	apple2helpers.ResyncCursor(ent)
	ent.SaveCPOS()

	// Now state registers
	ent.GetErrorTrap().Line = int(ent.GetMemory(220))

	ent.GetDataRef().Line = int(ent.GetMemory(123))
	ent.GetDataRef().Statement = int(ent.GetMemory(124))
	ent.GetDataRef().Token = int(ent.GetMemory(125))
	ent.GetDataRef().SubIndex = int(ent.GetMemory(126))

	// PC
	ent.GetPC().Line = int(ent.GetMemory(117))
	ent.GetPC().Statement = int(ent.GetMemory(118))

	// stacks here
	data = make([]uint64, 0)
	for i := 0; i < LOOPSTACK_MAX; i++ {
		data = append(data, ent.GetMemory(LOOPSTACK_ADDRESS+i))
	}
	_ = ent.GetLoopStack().UnmarshalBinary(data)

	data = make([]uint64, 0)
	for i := 0; i < CALLSTACK_MAX; i++ {
		data = append(data, ent.GetMemory(CALLSTACK_ADDRESS+i))
	}
	_ = ent.GetStack().UnmarshalBinary(data)

	// Fix stuff on stack
	c := ent.GetCode()
	for i := 0; i < ent.GetStack().Size(); i++ {
		ent.GetStack().Get(i).State = ent.GetState()
		ent.GetStack().Get(i).Locals = ent.GetLocal()
		ent.GetStack().Get(i).Code = c
	}

	// Get loop var
	data = make([]uint64, 0)
	for i := 0; i < 5; i++ {
		data = append(data, ent.GetMemory(0xfd00+i))
	}
	//fmt.Println(data)
	//fmt.Printf("LOOPVAR = %s\n", types.UnpackName(data[0:4]))
	ent.SetLoopVariable(types.UnpackName(data[0:4]))
	ent.SetLoopStep(float64(types.Uint2Float(data[4])))

	// PC
	ent.SetState(types.EntityState(ent.GetMemory(227)))

	this.DumpState(ent)
}

func (this *DialectAppleInteger) ThawVideoConfig(ent interfaces.Interpretable) {
	apple2helpers.RestoreSoftSwitches(ent)
}

func (this *DialectAppleInteger) HomeLeft(ent interfaces.Interpretable) {
	apple2helpers.HomeLeft(ent)
}

func Fields(str string) []string {

	sepkeepers := ":=+-*"

	chunk := ""
	parts := make([]string, 0)
	for _, ch := range str {
		switch {
		case ch == ' ':
			if chunk != "" {
				parts = append(parts, chunk)
				chunk = ""
			}
		case strings.IndexRune(sepkeepers, ch) > -1:
			if chunk != "" {
				parts = append(parts, chunk)
				chunk = ""
			}
			chunk = string(ch)
		default:
			if len(chunk) == 1 && strings.Index(sepkeepers, chunk) > -1 {
				parts = append(parts, chunk)
				chunk = ""
			}
			chunk += string(ch)
		}
	}

	if chunk != "" {
		parts = append(parts, chunk)
	}

	return parts
}

func in(str string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(v) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func typeIn(t types.TokenType, list []types.TokenType) bool {
	for _, v := range list {
		if v == t {
			return true
		}
	}
	return false
}

const CTX_START_LINE = 0
const CTX_AFTER_KW = 1
const CTX_AFTER_ASSIGN = 2
const CTX_AFTER_FUNC = 3
const CTX_WANT_FILE = 4

// Build completion list...
// Priority given to language keywords
func (this *DialectAppleInteger) GetCompletions(ent interfaces.Interpretable, line runestring.RuneString, index int) (int, *types.TokenList) {
	r := types.NewTokenList()

	// 10 pr/int/#/
	if index > len(line.Runes) {
		return 0, r
	}
	tmp := string(line.Runes[0:index])

	if len(tmp) > 0 && tmp[len(tmp)-1] == 32 {
		return 0, r
	}

	var inqq bool
	var bc int
	for _, ch := range line.Runes[0:index] {
		switch {
		case ch == '{' || ch == '[' || ch == '(':
			bc++
		case ch == '}' || ch == ']' || ch == ')':
			bc--
		case ch == '"':
			inqq = !inqq
		}
	}
	if bc > 0 || inqq {
		return 0, r
	}

	p := Fields(tmp)
	if len(p) == 0 {
		return 0, r
	}
	// get last word part before insert pos
	last := strings.ToLower(p[len(p)-1])

	// now tokenize the rest - we need to know what if any the last token was
	rest := tmp[0 : len(tmp)-len(last)]
	tl := this.Tokenize(runestring.Cast(rest))

	tla := ent.SplitOnToken(*tl, *types.NewToken(types.SEPARATOR, ":"))
	tl = &tla[len(tla)-1]

	nca := []string{"load", "save", "run", "next", "for", "help"}
	efn := []string{"load", "save", "run"}

	searchbase := last

	ctx := CTX_START_LINE

	if tl.Size() > 0 {
		t := tl.RPeek()
		f := tl.LPeek()

		if typeIn(t.Type, []types.TokenType{types.KEYWORD}) {
			ctx = CTX_AFTER_KW

		} else if typeIn(t.Type, []types.TokenType{types.FUNCTION, types.PLUSFUNCTION}) {
			ctx = CTX_AFTER_FUNC
		} else if typeIn(t.Type, []types.TokenType{types.ASSIGNMENT, types.COMPARITOR}) {
			ctx = CTX_AFTER_ASSIGN
		} else if typeIn(t.Type, []types.TokenType{types.SEPARATOR}) {
			ctx = CTX_START_LINE
		}

		if f.Type == types.KEYWORD && in(f.Content, efn) {
			ctx = CTX_WANT_FILE

			searchbase = strings.Replace(searchbase, "\"", "", -1)

			base := searchbase

			var addworkdir bool

			if !strings.HasPrefix(base, "/") {
				base = ent.GetWorkDir() + "/" + base
				addworkdir = true
			}

			// Do the search
			//fmt.Printf("I will try autocomplete for files based on a path of [%s]\n", base)
			ditems, fitems := files.GetCompletions(base)
			for _, i := range ditems {
				isDir := true

				// strip auto prefix
				if addworkdir {
					i = i[len(ent.GetWorkDir()+"/"):]
				}

				if i != "/" && i != "" && !strings.HasSuffix(i, "..") && i != searchbase {
					if isDir {
						r.Push(types.NewToken(types.STRING, i+"/"))
					} else {
						r.Push(types.NewToken(types.STRING, i))
					}
				}
			}
			for _, i := range fitems {
				isDir := false

				// strip auto prefix
				if addworkdir {
					i = i[len(ent.GetWorkDir()+"/"):]
				}

				if i != "/" && i != "" && !strings.HasSuffix(i, "..") && i != searchbase {
					if isDir {
						r.Push(types.NewToken(types.STRING, i+"/"))
					} else {
						r.Push(types.NewToken(types.STRING, i))
					}
				}
			}

			return len(searchbase), r

		} else if t.Type == types.KEYWORD {
			// last was keyword
			if strings.ToLower(t.Content) == "then" || strings.ToLower(t.Content) == "def" {
				ctx = CTX_START_LINE
			}

			if in(t.Content, nca) {
				return 0, r
			}
		} else {
			t = tl.LPeek()
			if in(t.Content, nca) {
				return 0, r
			}
		}
	} else {

	}

	// now we look for it as the start of something
	//
	// #1 keywords

	scom := make([]string, 0)

	if ctx == CTX_START_LINE {
		for k, _ := range this.Commands {
			scom = append(scom, k)
		}
		sort.Strings(scom)

		for _, k := range scom {
			//fmt.Printf("[%s] vs [%s]\n", k, last)
			if strings.HasPrefix(k, last) {
				// Add to list
				r.Push(types.NewToken(types.KEYWORD, k))
			}
		}
	}
	//
	// #2 functions
	scom = make([]string, 0)
	if ctx == CTX_AFTER_ASSIGN || ctx == CTX_AFTER_FUNC || ctx == CTX_AFTER_KW {
		for k, _ := range this.Functions {
			scom = append(scom, k)
		}
		sort.Strings(scom)
		for _, k := range scom {
			if strings.HasPrefix(k, last) {
				// Add to list
				r.Push(types.NewToken(types.FUNCTION, k))
			}
		}
	}
	// #3 Plus functions
	scom = make([]string, 0)
	for k, v := range this.PlusFunctions {
		//scom = append(scom, k)
		for kk, vv := range v {
			if !vv.IsHidden() {
				scom = append(scom, k+"."+kk)
			}
		}
	}
	sort.Strings(scom)
	for _, k := range scom {
		if strings.HasPrefix(k, last) {
			// Add to list
			r.Push(types.NewToken(types.PLUSFUNCTION, k))
		}
	}

	return len(last), r
}

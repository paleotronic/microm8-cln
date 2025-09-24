package appleinteger

import (
	"paleotronic.com/fmt"

	"paleotronic.com/log"
	//"paleotronic.com/fmt"
	//"os"
	"github.com/atotto/clipboard"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/filerecord"
	"paleotronic.com/files"
	"paleotronic.com/utils"
	//"strconv"
	"strings"
	//	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandLOAD struct {
	dialect.Command
}

func (this *StandardCommandLOAD) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var res types.Token
	var filename string
	var bb []byte
	var ferr error
	var useClip bool
	//var src string
	//Text f;
	var fr filerecord.FileRecord

	//    types.RebuildAlgo()

	// //System.Out.Println("WORKDIR PRE LOAD = "+caller.WorkDir);

	result = 0

	if tokens.Size() == 0 {
		return result, exception.NewESyntaxError("LOAD needs a \"filename\"")
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
		return result, exception.NewESyntaxError("LOAD needs a \"filename\"")
	}

	//	if caller.IsRunningDirect() {
	//		caller.SetWorkDir("")
	//	}

	filename = utils.Flatten7Bit(res.Content)

	thefile := filename

	if filename == "@" {
		stmp, e := clipboard.ReadAll()
		if e != nil {
			return 0, e
		}
		bb = []byte(stmp)
		ferr = nil
		useClip = true
		caller.SetFileRecord(fr)
	}

	if filename == "!" {
		bb = []byte(caller.GetFeedBuffer())
		ferr = nil
		useClip = true
		caller.SetFileRecord(fr)
	}

	if utils.PosRune('/', thefile) == 0 {
		thefile = caller.GetWorkDir() + thefile
	}

	// Clear program memory ala - NEW
	filename = strings.ToLower(thefile)
	path := files.GetPath(filename)
	filename = files.GetFilename(filename)

	codeTypes := files.GetTypeCode()
	fmt.Printf("code types: %v\n", codeTypes)

	found := false

	if matches, _ := files.ResolveFileViaProvider(path, filename); len(matches) > 0 && files.GetExt(filename) != "" {

		ext := files.GetExt(filename)
		info, ok := files.GetInfo(ext)
		if ok {
			fmt.Printf("*** Found as %s, dialect %s\n", filename, info.Dialect)

			log.Println("LOAD needs to switch to " + info.Dialect)

			// bootstrap dialect
			cd := caller.GetDialect().GetTitle()

			if (cd != "Shell" && info.Dialect == "shell") || (cd != "Applesoft" && info.Dialect == "fp") || (cd != "INTEGER" && info.Dialect == "int") {
				tl := types.NewTokenList()
				tl.Push(types.NewToken(types.VARIABLE, info.Dialect))
				caller.GetDialect().GetCommands()["lang"].Execute(env, caller, *tl, Scope, LPC)
			}
			found = true
		}

	} else {

		for _, info := range codeTypes {
			fmt.Printf("Trying type %s\n", info.Ext)

			if matches, _ := files.ResolveFileViaProvider(path, filename+"."+info.Ext); len(matches) > 0 {
				//filename = filename + "." + info.Ext
				filename = matches[0]

				fmt.Printf("*** Found as %s, dialect %s\n", filename, info.Dialect)

				log.Println("LOAD needs to switch to " + info.Dialect)

				// bootstrap dialect
				cd := caller.GetDialect().GetTitle()

				if (cd != "Shell" && info.Dialect == "shell") || (cd != "Applesoft" && info.Dialect == "fp") || (cd != "INTEGER" && info.Dialect == "int") {
					tl := types.NewTokenList()
					tl.Push(types.NewToken(types.VARIABLE, info.Dialect))
					caller.GetDialect().GetCommands()["lang"].Execute(env, caller, *tl, Scope, LPC)
				}
				found = true
				break
			}
		}

	}

	if !found && !useClip {
		fmt.Println("Not found...")
		return result, exception.NewESyntaxError("FILE NOT FOUND")
	}

	if useClip {
		goto loadbypass
	}

	//var fr filerecord.FileRecord
	fr, ferr = files.ReadBytesViaProvider(path, filename)
	if ferr != nil {
		return result, ferr
	}
	bb = fr.Content
	caller.SetFileRecord(fr) // set file record

loadbypass:

	log.Println(string(bb))

	sl := utils.SplitLines(bb)

	caller.Clear()

	caller.GetDialect().SetSkipMemParse(true)
	for i, src1 := range sl {
		if strings.Trim(src1, "") != "" {
			//src1, _ = strconv.Unquote(src1)

			ss := utils.Unescape(src1)

			if caller.GetDialect().GetTitle() == "Shell" {
				ss = utils.IntToStr(i+1) + "  " + ss
				log.Printf("Autonumber shell: %s\n", ss)
			}

			caller.Parse(ss)
		}
	}
	caller.GetDialect().SetSkipMemParse(false)

	fixMemoryPtrs(caller)

	//	apple2helpers.PutStr(caller, "Ok: Loaded \""+path+"/"+filename+"\"\r\n")

	//System.Out.Println("WORKDIR POST 1 LOAD = "+caller.WorkDir);

	caller.SetWorkDir(path + "/")
	caller.SetProgramDir(path + "/")
	//System.Out.Println("WORKDIR POST 2 LOAD = "+caller.WorkDir);

	// now wipe the memory
	if !caller.IsRunning() {
		for i := 768; i < 1024; i++ {
			caller.SetMemory(i, 0)
		}
	}

	//caller.GetLocal().Put("speed", types.NewVariableP("speed", types.VT_FLOAT, "255", true))

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandLOAD) Syntax() string {

	/* vars */
	var result string

	result = "LOAD \"<filename>\""

	/* enforce non void return */
	return result

}

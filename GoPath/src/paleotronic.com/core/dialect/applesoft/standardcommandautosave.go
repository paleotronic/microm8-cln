package applesoft

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/utils"
)

type StandardCommandAUTOSAVE struct {
	dialect.Command
}

func (this *StandardCommandAUTOSAVE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int
	var res types.Token
	var filename string

	result = 0

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
		return result, exception.NewESyntaxError("AUTOSAVE needs a \"filename\"")
	}

	osext := string(strings.ToLower(caller.GetDialect().GetTitle())[0])
	prefext := files.GetPreferredExt(osext)

	filename = utils.Flatten7Bit(res.Content)

	if files.GetExt(filename) == "" {
		filename += "." + prefext
	}

	if caller.GetWorkDir() != "" && rune(filename[0]) != '/' {
		if caller.GetWorkDir() == "/" {
			filename = "/local/" + filename
		} else {
			filename = caller.GetWorkDir() + filename
		}
	}

	filename = strings.ToLower(filename)

	settings.AutosaveFilename[caller.GetMemIndex()] = filename
	caller.PutStr("AUTOSAVE TO " + filename + "\r\n")

	return result, nil

}

func (this *StandardCommandAUTOSAVE) Syntax() string {

	/* vars */
	var result string

	result = "SAVE \"<filename>\""

	/* enforce non void return */
	return result

}

package applesoft

import (
	"strings"

	"paleotronic.com/core/dialect"
	"paleotronic.com/core/editor"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/runestring"
)

type StandardCommandCAT struct {
	dialect.Command
	selected  int
	execute   int
	Scope     *types.Algorithm
	lastindex int
	filepanel *editor.FileCatalog
}

func NewStandardCommandCAT() *StandardCommandCAT {
	this := &StandardCommandCAT{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandCAT) Syntax() string {
	// TODO Auto-generated method stub
	return "XCAT"
}

func (this *StandardCommandCAT) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {
	// Purge keys

	var result int

	caller.SetBuffer(runestring.NewRuneString())

	path := ""
	ext := files.GetPatternAll()

	if caller.GetWorkDir() != "" {
		path = caller.GetWorkDir()
	}

	for strings.HasSuffix(path, "//") {
		path = strings.Replace(path, "//", "/", -1)
	}

	if tokens.Size() > 0 {
		tok := caller.ParseTokensForResult(tokens)
		ext = tok.Content
	}

	if this.filepanel == nil {
		s := editor.FileCatalogSettings{
			DiskExtensions: files.GetExtensionsDisk(),
			Title:          "microFile File Manager",
			Pattern:        ext,
			Path:           path,
			BootstrapDisk:  false,
			HidePaths:      []string{"system", "FILECACHE"},
		}
		this.filepanel = editor.NewFileCatalog(caller, s)
	}
	settings.DisableMetaMode[caller.GetMemIndex()] = true
	this.lastindex, _ = this.filepanel.Do(this.lastindex)
	settings.DisableMetaMode[caller.GetMemIndex()] = false

	return result, nil
}

func resetState(caller interfaces.Interpretable) {
	// for i := 0; i < 131072; i++ {
	// 	caller.GetMemoryMap().WriteInterpreterMemorySilent(caller.GetMemIndex(), i, 0)
	// }

	caller.LoadSpec(caller.GetSpec())
	caller.SetMemory(103, 1)
	caller.SetMemory(104, 8)
	caller.GetDialect().InitVarmap(caller, nil)
	caller.Zero(false)
}

package applesoft

import (
	"strings"

	"paleotronic.com/files"
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/exception"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//"paleotronic.com/fmt"
	"paleotronic.com/core/hardware/apple2helpers"
)

type StandardCommandPACKAGE struct {
	dialect.Command
}

func NewStandardCommandPACKAGE() *StandardCommandPACKAGE {
	this := &StandardCommandPACKAGE{}
	this.ImmediateMode = true
	return this
}

func (this *StandardCommandPACKAGE) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	/*
	 * PACKAGE CREATE "packname"
	 * PACKAGE ADD "packname", "file"
	 * PACKAGE DEL "packname", "file"
	 * PACKAGE TAG "packname", "name", "value"
	 * PACKAGE TAGFILE "packname", "file", "name", "value"
	 */

	if tokens.Size() < 2 {
		return 0, exception.NewESyntaxError("SYNTAX ERROR")
	}

	t := *tokens.Shift()
	verb := strings.ToLower(t.Content)

	p := *tokens.Shift()
	packname := p.Content

	tla := caller.SplitOnTokenWithBrackets(tokens, *types.NewToken(types.SEPARATOR, ","))
	tl := *types.NewTokenList()
	for _, l := range tla {
		v := caller.ParseTokensForResult(l)
		tl.Push(&v)
	}

	switch {
	case verb == "create":
		// create "empty" package
		p := &files.Package{}
		p.Name = packname
		p.Meta = ""
		b, e := p.MarshalBinary()
		if e != nil {
			return 0, e
		}
		e = files.WriteBytesViaProvider("", packname+".s8p", b)
		apple2helpers.PutStr(caller, "Created "+packname+".s8p")
	case verb == "add":
		// add file to package, optional meta data
		if tl.Size() == 0 {
			return 0, exception.NewESyntaxError("PACKAGE ADD needs a file")
		}

		r, e := files.ReadBytesViaProvider("", packname+".s8p")
		if e != nil {
			return 0, e
		}

		pk := &files.Package{}
		e = pk.UnmarshalBinary(r.Content)
		if e != nil {
			return 0, e
		}

		for tl.Size() > 0 {
			t := *tl.Shift()
			p := files.GetPath(t.Content)
			f := files.GetFilename(t.Content)
			b, e := files.ReadBytesViaProvider(p, f)
			if e != nil {
				return 0, e
			}
			pk.Add(f, "", b.Content)
		}

		b, e := pk.MarshalBinary()
		if e != nil {
			return 0, e
		}
		e = files.WriteBytesViaProvider("", packname+".s8p", b)
		apple2helpers.PutStr(caller, "Updated "+packname+".s8p")

	case verb == "del":
		// delete file from package
		if tl.Size() == 0 {
			return 0, exception.NewESyntaxError("PACKAGE DEL needs a file")
		}

		r, e := files.ReadBytesViaProvider("", packname+".s8p")
		if e != nil {
			return 0, e
		}

		pk := &files.Package{}
		e = pk.UnmarshalBinary(r.Content)
		if e != nil {
			return 0, e
		}

		for tl.Size() > 0 {
			t := *tl.Shift()
			//p := files.GetPath(t.Content)
			f := files.GetFilename(t.Content)
			pk.Remove(f)
		}

		b, e := pk.MarshalBinary()
		if e != nil {
			return 0, e
		}
		e = files.WriteBytesViaProvider("", packname+".s8p", b)
		apple2helpers.PutStr(caller, "Updated "+packname+".s8p")
	case verb == "tag":
		//
		if tl.Size() < 2 {
			return 0, exception.NewESyntaxError("PACKAGE TAG needs a name and value")
		}

		r, e := files.ReadBytesViaProvider("", packname+".s8p")
		if e != nil {
			return 0, e
		}

		pk := &files.Package{}
		e = pk.UnmarshalBinary(r.Content)
		if e != nil {
			return 0, e
		}

		pk.SetMetadata(tl.Get(0).Content, tl.Get(1).Content)

		b, e := pk.MarshalBinary()
		if e != nil {
			return 0, e
		}
		e = files.WriteBytesViaProvider("", packname+".s8p", b)
		apple2helpers.PutStr(caller, "Updated "+packname+".s8p")
	case verb == "tagfile":
		//
		if tl.Size() < 3 {
			return 0, exception.NewESyntaxError("PACKAGE TAGFILE needs a file, name and value")
		}

		r, e := files.ReadBytesViaProvider("", packname+".s8p")
		if e != nil {
			return 0, e
		}

		pk := &files.Package{}
		e = pk.UnmarshalBinary(r.Content)
		if e != nil {
			return 0, e
		}

		file := tl.Shift().Content
		n := tl.Shift().Content
		v := tl.Shift().Content

		////fmt.Printf("file=%s,n=%s,v=%s\n", file, n, v)

		e = pk.SetFileMetadata(file, n, v)
		if e != nil {
			return 0, e
		}

		b, e := pk.MarshalBinary()
		if e != nil {
			return 0, e
		}
		e = files.WriteBytesViaProvider("", packname+".s8p", b)
		apple2helpers.PutStr(caller, "Updated "+packname+".s8p")
	default:
		return 0, exception.NewESyntaxError("No such PACKAGE command: " + verb)
	}

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPACKAGE) Syntax() string {

	/* vars */
	var result string

	result = "PACKAGE <ADD|DEL|META> "

	/* enforce non void return */
	return result

}

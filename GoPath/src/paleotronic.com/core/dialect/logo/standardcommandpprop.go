package logo

import (
	"paleotronic.com/core/dialect"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	//	"paleotronic.com/core/hardware/apple2helpers"
	"errors"
	"strings"
)

type StandardCommandPPROP struct {
	dialect.Command
}

func (this *StandardCommandPPROP) Execute(env *interfaces.Producable, caller interfaces.Interpretable, tokens types.TokenList, Scope *types.Algorithm, LPC types.CodeRef) (int, error) {

	/* vars */
	var result int

	result = 0

	if tokens.Size() < 1 {
		return result, errors.New("I NEED A VALUE")
	}

	// get result
	tt, e := this.Command.D.(*DialectLogo).ParseTokensForResult(caller, tokens)
	if e != nil {
		return result, e
	}

	if tt.Type != types.LIST || tt.List.Size() != 3 {
		return result, errors.New("I NEED 3 VALUES")
	}

	plist := tt.List.Get(0)
	pname := plist.Content
	if !strings.HasPrefix(pname, ":") {
		pname = ":" + pname
	}
	pprop := tt.List.Get(1)
	pvalue := tt.List.Get(2)

	lobj := caller.GetData(pname)
	if lobj == nil {
		lobj = types.NewToken(types.LIST, "")
		lobj.List = types.NewTokenList()
	}

	// now either the prop name exists or not
	// since the structure is name, value, name, value
	// we only need to check even indices
	foundIdx := -1
	for i := 0; i < lobj.List.Size(); i += 2 {
		t := lobj.List.Get(i)
		if strings.ToLower(t.Content) == strings.ToLower(pprop.Content) {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		// new property
		lobj.List.Push(pprop)
		lobj.List.Push(pvalue)
	} else {
		// existing property
		lobj.List.Content[foundIdx+1] = pvalue
	}

	lobj.IsPropList = true // mark as property list
	caller.SetData(pname, *lobj, false)

	/* enforce non void return */
	return result, nil

}

func (this *StandardCommandPPROP) Syntax() string {

	/* vars */
	var result string

	result = "TEXT"

	/* enforce non void return */
	return result

}

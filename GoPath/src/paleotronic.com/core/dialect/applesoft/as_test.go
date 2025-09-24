package applesoft

import (
	"paleotronic.com/fmt"
	"testing"
)
import "paleotronic.com/runestring"

func TestPackUnpack(t *testing.T) {

	as := NewDialectApplesoft()

	//	tl := *as.Tokenize(runestring.Cast("PRINT LEFT$(\"FROG\", 2)"))

	//	for i, t := range tl.Content {
	//		//fmt.Printf("%d) %s: %s\n", i, t.Content, t.Type.String())
	//	}

	//	data := as.NTokenize(tl)

	//	//fmt.Println(data)

	//	ntl := *as.UnNTokenize(data)

	//	for i, t := range ntl.Content {
	//		//fmt.Printf("%d) %s: %s\n", i, t.Content, t.Type.String())
	//	}

	//	data2 := as.NTokenize(ntl)

	//	//fmt.Println(data2)

	line := runestring.Cast("10 PRINT A")
	clist := as.GetCompletions(nil, line, len(line.Runes))

	for i, t := range clist.Content {
		fmt.Printf("(%d) %s %s\n", i, t.Content, t.Type.String())
	}

	t.Error("Force")

}

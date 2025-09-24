// algorithm.go
package types

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"sync"

	"paleotronic.com/utils"
	//    "paleotronic.com/fmt"
)

type AlgorithmHolder map[int]Line
type AlgorithmKeys []int

type Algorithm struct {
	C       AlgorithmHolder
	keys    AlgorithmKeys
	Changed bool
	m       sync.Mutex
}

func (this *Algorithm) Remove(i int) {
	delete(this.C, i)
	this.Changed = true
}

func (this *Algorithm) GetLowIndex() int {
	this.Changed = true
	keys := this.GetSortedKeys()
	if len(keys) == 0 {
		return -1
	}
	return keys[0]
}

func (this *Algorithm) GetHighIndex() int {
	this.Changed = true
	keys := this.GetSortedKeys()
	if len(keys) == 0 {
		return -1
	}
	return keys[len(keys)-1]
}

func (this *Algorithm) GetSortedKeys() AlgorithmKeys {

	var keys AlgorithmKeys

	if !this.Changed {
		return this.keys
	}

	this.m.Lock()
	defer this.m.Unlock()

	for k, _ := range this.C {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	this.keys = keys

	//fmt.Println("Built sorted keys")
	this.Changed = false

	return keys
}

func (this Algorithm) Size() int {
	return len(this.C)
}

func (this *Algorithm) Put(i int, ll Line) {
	this.m.Lock()
	defer this.m.Unlock()
	this.C[i] = ll
	this.Changed = true
}

func (this Algorithm) Get(i int) (Line, bool) {
	this.m.Lock()
	defer this.m.Unlock()
	l, ok := this.C[i]
	return l, ok
}

func NewAlgorithm() *Algorithm {

	this := &Algorithm{
		C:       make(AlgorithmHolder),
		keys:    make([]int, 0),
		Changed: true,
	}

	return this
}

func (this Algorithm) ContainsKey(key int) bool {
	this.m.Lock()
	defer this.m.Unlock()
	_, ok := this.C[key]
	return ok
}

func (slice AlgorithmKeys) IndexOf(value int) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func (this Algorithm) PrevAfter(line int) int {
	var i, l int
	keys := this.GetSortedKeys()
	i = keys.IndexOf(line)

	l = -1
	if len(this.keys) > 0 {
		l = this.keys[0]
	}

	if i == -1 {
		for (i == -1) && (line > l) {
			line--
			i = keys.IndexOf(line)
		}
		if i == -1 {
			return -1
		}
		return line
	}
	// last line?
	if i == 0 {
		return -1
	}
	// in the middle
	return keys[i-1]
}

func (this Algorithm) NextAfter(line int) int {
	var i, h int
	keys := this.GetSortedKeys()
	i = keys.IndexOf(line)
	//h = this.GetHighIndex()
	h = -1
	if len(this.keys) > 0 {
		h = this.keys[len(keys)-1]
	}
	if i == -1 {
		for (i == -1) && (line < h) {
			line++
			i = keys.IndexOf(line)
		}
		if i == -1 {
			return -1
		}
		return line
	}
	// last line?
	if i == len(keys)-1 {
		return -1
	}
	// in the middle
	return keys[i+1]
}

func (this Algorithm) String() string {
	out := ""

	l := this.GetLowIndex()
	h := this.GetHighIndex()

	s := utils.IntToStr(h)
	w := len(s) + 1
	if w < 4 {
		w = 4
	}

	if l < 0 {
		return ""
	}

	//writeln( "l is set to ", l );

	/* got code */
	for (l != -1) && (l <= h) {
		/* display this line */

		if this.ContainsKey(l) {
			lns := utils.IntToStr(l) + "  "
			//caller.GetVDU().PutStr(PadLeft(s, w) + " ")
			ln := this.C[l]
			s = ""
			for _, stmt1 := range ln {

				//ft = caller.TokenListAsString( stmt );
				ft := ""
				lt := *NewToken(INVALID, "")
				for _, t1 := range stmt1.Content {
					if (lt.Type == KEYWORD) || (lt.Type == DYNAMICKEYWORD) || (lt.Type == OPERATOR) || (lt.Type == SEPARATOR) || (lt.Type == OBRACKET) || (lt.Type == COMPARITOR) || (lt.Type == ASSIGNMENT) {
						ft = ft + " "
					}
					if ((t1.Type == FUNCTION) || (t1.Type == CBRACKET) || (t1.Type == OBRACKET) || (t1.Type == NUMBER) || (t1.Type == VARIABLE) || (t1.Type == STRING) || (t1.Type == LOGIC) || (t1.Type == OPERATOR) || (t1.Type == KEYWORD) || (t1.Type == COMPARITOR) || (t1.Type == ASSIGNMENT)) && (len(ft) > 0) && (ft[len(ft)-1] != ' ') {
						ft = ft + " "
					}
					ft = ft + t1.AsString()
					//lt := t1
				}

				if s != "" {
					s = s + " : "
				}

				s = s + ft

			}

			out = out + lns + s
			out = out + "\r\n"
		}

		/* next line */
		//l = caller.Code.NextAfter(l);
		l++
	}

	return out
}

func (this Algorithm) Checksum() string {

	b := []byte(this.String())

	ck := md5.Sum(b)

	return hex.EncodeToString(ck[0:16])
}

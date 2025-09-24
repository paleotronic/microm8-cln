// tokenlist.go
package types

import (
	"strings"

	"paleotronic.com/fmt"
)

type TokenList struct {
	Content []*Token
	Open    bool
}

// Creates an empty TokenList ready for use in the Producer
func NewTokenList() *TokenList {
	return &TokenList{Open: true, Content: make([]*Token, 0)}
}

// Create a new TokenList which is a copy of a sublist
func (src TokenList) SubList(start, end int) *TokenList {
	dest := NewTokenList()
	if start < 0 {
		start = 0
	}
	if end > len(src.Content) {
		end = len(src.Content)
	}
	dest.Content = append(dest.Content, src.Content[start:end]...)
	return dest
}

func (src TokenList) Copy() *TokenList {

	tl := NewTokenList()
	for _, v := range src.Content {
		n := v.Copy()
		if v.List != nil {
			n.List = v.List.Copy()
		}
		tl.Push(n)
	}

	return tl
}

func (src TokenList) Strings() []string {
	out := []string(nil)
	for _, t := range src.Content {
		out = append(out, t.Content)
	}
	return out
}

func (tl TokenList) AsString() string {
	var out string

	for _, tok := range tl.Content {
		if out != "" {
			out = out + " "
		}
		out = out + tok.AsString()
	}

	return out
}

// Index of first occurrence of Token in TokenList starting at offset N
func (tl TokenList) IndexOfN(start int, ttype TokenType, tcontent string) int {
	var res int = -1

	for i, tok := range tl.Content {
		if i > start {
			fmt.Printf(">> tok.Content=%s, tcontent=%s\n", tok.Content, tcontent)
			if ((tok.Type == ttype) || (ttype == INVALID)) && ((strings.ToLower(tcontent) == strings.ToLower(tok.Content)) || (tcontent == "")) {
				res = i
				return res
			}
		}
	}

	return res
}

func (tl TokenList) IndexOfTokenN(start int, t *Token) int {
	for i, tok := range tl.Content {
		if i >= start {
			fmt.Printf("Comparing tok=%+v, t=%+v\n", *tok, *t)
			if t.Type == LIST && tok.Type == LIST && t.List.Equals(tok.List) {
				return i
			}
			if t.Type != LIST && strings.ToLower(t.Content) == strings.ToLower(tok.Content) {
				return i
			}
		}
	}
	return -1
}

func (tl TokenList) IndexOfToken(t *Token) int {
	return tl.IndexOfTokenN(0, t)
}

// Index of first occurrence of Token in TokenList
func (tl TokenList) IndexOf(ttype TokenType, tcontent string) int {
	return tl.IndexOfN(0, ttype, tcontent)
}

func (tl *TokenList) Equals(l *TokenList) bool {
	if tl.Size() != l.Size() {
		return false
	}
	for i, _ := range tl.Content {
		fmt.Printf("List.Equals.compare tl[%d]=%s, l[%d]=%s\n", i, tl.Content[i].Content, i, l.Content[i].Content)
		if strings.ToLower(tl.Content[i].Content) != strings.ToLower(l.Content[i].Content) {
			return false
		}
	}
	return true
}

// Insert a Token into a list at a given index
func (tl *TokenList) Insert(i int, tok *Token) {
	tl.Content = append(tl.Content[:i], append([]*Token{tok}, tl.Content[i:]...)...)
}

// Unshift token
func (tl *TokenList) UnShift(tok *Token) {
	tl.Insert(0, tok)
}

// Shift token off start of list
func (tl *TokenList) Shift() *Token {

	var res *Token = nil
	if len(tl.Content) < 1 {
		return res
	}
	res = tl.Content[0]

	tl.Content = append([]*Token(nil), tl.Content[1:]...)

	return res
}

// Push token to end of list
func (tl *TokenList) Push(tok *Token) {
	tl.Content = append(tl.Content, tok)
}

func (tl *TokenList) Add(tok *Token) {
	tl.Content = append(tl.Content, tok)
}

// Pop token off end of list
func (tl *TokenList) Pop() *Token {
	var res *Token = nil
	if len(tl.Content) < 1 {
		return res
	}

	res, tl.Content = tl.Content[len(tl.Content)-1], tl.Content[:len(tl.Content)-1]
	return res
}

// Same as Shift()
func (tl *TokenList) Left() *Token {
	return tl.Shift()
}

// Same as Pop()
func (tl *TokenList) Right() *Token {
	return tl.Pop()
}

// Mark list as closed
func (tl *TokenList) Close() {
	tl.Open = false
}

// Return true if list is open
func (tl TokenList) IsOpen() bool {
	return tl.Open
}

// Return true if list is closed
func (tl TokenList) IsClosed() bool {
	return !tl.Open
}

// Empty the list
func (tl *TokenList) Clear() {
	tl.Content = []*Token(nil)
}

func (tl *TokenList) RPeek() *Token {
	var t *Token = NewToken(INVALID, "")
	if tl.Size() > 0 {
		t = tl.Content[len(tl.Content)-1]
	}
	return t
}

func (tl *TokenList) LPeek() *Token {
	var t *Token = NewToken(INVALID, "")
	if tl.Size() > 0 {
		t = tl.Content[0]
	}
	return t
}

func (tl *TokenList) Size() int {
	return len(tl.Content)
}

func (tl *TokenList) Get(index int) *Token {
	if index >= 0 && index < tl.Size() {
		return tl.Content[index]
	}
	return nil
}

func (tl *TokenList) Remove(index int) *Token {
	var tok *Token
	if index >= 0 && index < tl.Size() {
		tok = tl.Content[index]
		tl.Content[index] = nil
		tl.Content = append(tl.Content[:index], tl.Content[index+1:]...)
		return tok
	}
	return nil
}

func (tl *TokenList) FindCommandLists() []int {
	out := []int{}
	for i, t := range tl.Content {
		if t.Type == LIST && t.List.Size() > 0 {
			ft := t.List.Get(0)
			if ft.Type == KEYWORD || ft.Type == DYNAMICKEYWORD {
				out = append(out, i)
			} else if (ft.Type == LIST || ft.Type == COMMANDLIST) && ft.List.Size() > 0 && (ft.List.Get(0).Type == KEYWORD || ft.List.Get(0).Type == DYNAMICKEYWORD) {
				out = append(out, i)
			}
		}
	}
	return out
}

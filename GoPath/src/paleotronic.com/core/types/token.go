// token.go
package types

import (
	"math"
	"strconv"

	"paleotronic.com/fmt"

	. "paleotronic.com/utils"
)

type TokenType int

const (
	INVALID TokenType = iota // 1
	STRING
	INTEGER
	NUMBER
	BOOLEAN
	KEYWORD
	FUNCTION
	LOGIC
	OPERATOR
	ASSIGNMENT
	COMPARITOR
	SEPARATOR
	OBRACKET
	CBRACKET
	UNSTRING
	VARIABLE
	DYNAMICKEYWORD
	DYNAMICFUNCTION
	USERKEYWORD
	EXPRESSION
	NOP
	TYPE
	WORD
	LIST
	PLUSVAR
	PLUSFUNCTION
	PLACEHOLDER
	LOGOPARAMPREFIX
	LOGOVARQUOTE
	LABEL
	COMMANDLIST
	EXPRESSIONLIST
	IDENTIFIER
	TABLE
	TABLEROW
	TABLECOL
)

func (this TokenType) String() string {
	switch this {
	case STRING:
		return "STRING"
	case INTEGER:
		return "INTEGER"
	case NUMBER:
		return "NUMBER"
	case BOOLEAN:
		return "BOOLEAN"
	case LIST:
		return "LIST"
	case KEYWORD:
		return "KEYWORD"
	case FUNCTION:
		return "FUNCTION"
	case DYNAMICKEYWORD:
		return "DYNAMICKEYWORD"
	case DYNAMICFUNCTION:
		return "DYNAMICFUNCTION"
	case VARIABLE:
		return "VARIABLE"
	case OPERATOR:
		return "OPERATOR"
	case ASSIGNMENT:
		return "ASSIGNMENT"
	case LOGIC:
		return "LOGIC"
	case COMPARITOR:
		return "COMPARITOR"
	case SEPARATOR:
		return "SEPARATOR"
	case LOGOPARAMPREFIX:
		return "LOGOPARAM"
	case LOGOVARQUOTE:
		return "LOGOVAR"
	case WORD:
		return "WORD"
	case NOP:
		return "NOP"
	default:
		return fmt.Sprintf("(other id=%d)", int(this))
	}
}

type Token struct {
	Content            string
	Type               TokenType
	List               *TokenList
	IsPropList         bool
	Hidden             bool
	WSPrefix, WSSuffix string
}

func NewToken(t TokenType, c string) *Token {
	tok := new(Token)
	tok.Type = t
	tok.Content = c
	if (tok.Content == "") && (tok.IsIn([]TokenType{NUMBER, INTEGER})) {
		tok.Content = "0"
	}
	return tok
}

func (tok *Token) Copy() *Token {
	t := NewToken(tok.Type, tok.Content)
	if tok.List != nil {
		t.List = tok.List.Copy()
	}
	t.IsPropList = tok.IsPropList
	t.Hidden = tok.Hidden
	t.WSPrefix = tok.WSPrefix
	t.WSSuffix = tok.WSSuffix
	return t
}

func (tok *Token) IsNumeric() bool {
	if tok == nil {
		return false
	}
	return tok.IsIn([]TokenType{BOOLEAN, INTEGER, NUMBER})
}

func (tok *Token) AsString() string {

	if tok == nil {
		return ""
	}

	switch tok.Type {
	case BOOLEAN:
		if tok.AsInteger() > 0 {
			return "true"
		} else {
			return "false"
		}
	case STRING:
		return string('"') + tok.Content + string('"')
	default:
		return tok.Content
	}

}

func (tok *Token) AsNumeric() float64 {

	if tok == nil {
		return 0
	}

	var f float64
	var err error

	switch tok.Type {
	case NUMBER:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
	case INTEGER:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
		f = math.Floor(f)
	case STRING:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
		f = math.Floor(f)
	case WORD:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
		f = math.Floor(f)
	case BOOLEAN:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
		if (tok.Content == "true") || (tok.Content == "yes") || (f != 0) {
			f = 1
		} else {
			f = 0
		}
	default:
		f = 0
	}

	return f

}

func (tok *Token) AsNumeric64() float64 {

	if tok == nil {
		return 0
	}

	var f float64
	var err error

	switch tok.Type {
	case NUMBER:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
	case INTEGER:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
		f = math.Floor(f)
	case STRING:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
		f = math.Floor(f)
	case BOOLEAN:
		f, err = strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
		if err != nil {
			f = 0
		}
		if (tok.Content == "true") || (tok.Content == "yes") || (f != 0) {
			f = 1
		} else {
			f = 0
		}
	default:
		f = 0
	}

	return f

}

func (tok *Token) AsInteger() int {

	if tok == nil {
		return 0
	}

	if tok.Type == STRING {
		if len(tok.Content) > 0 {
			return 1
		} else {
			return 0
		}
	}

	if tok.Content == "" {
		tok.Content = "0"
	}

	f, err := strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
	if err != nil {
		f = 0
	}
	return int(math.Floor(f))
}

func (tok *Token) AsExtended() float64 {
	//return float64(tok.AsFloat());

	if tok == nil {
		return 0
	}

	f, err := strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
	if err != nil {
		f = 0
	}
	return f
}

func (tok *Token) AsFloat() float64 {

	if tok == nil {
		return 0
	}

	if tok.Type == STRING {
		if len(tok.Content) > 0 {
			return 1
		} else {
			return 0
		}
	}

	f, err := strconv.ParseFloat(NumberPart(FlattenASCII(tok.Content)), 64)
	if err != nil {
		f = 0
	}
	return f
}

func (tok *Token) IsType(list []TokenType) bool {
	return tok.IsIn(list)
}

func (tok *Token) IsIn(list []TokenType) bool {
	for _, a := range list {
		if a == tok.Type {
			return true
		}
	}
	return false
}

func (tok *Token) IsNotIn(list []TokenType) bool {
	return !tok.IsIn(list)
}

func (tok *Token) SubListAsToken(start, end int) *Token {
	if tok.List != nil {
		l := tok.List.SubList(start, end)
		t := NewToken(LIST, tok.Content)
		t.List = l
		return t
	}
	t := NewToken(LIST, tok.Content)
	t.List = NewTokenList()
	return t
}

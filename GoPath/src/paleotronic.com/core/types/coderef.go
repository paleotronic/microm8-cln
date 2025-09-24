// coderef.go
package types

type CodeRef struct {
	Line, Statement, Token, SubIndex int
}

func NewCodeRef() *CodeRef {
	return &CodeRef{}
}

func NewCodeRefCopy(c CodeRef) *CodeRef {
	return &CodeRef{Line: c.Line, Statement: c.Statement, Token: c.Token, SubIndex: 0}
}

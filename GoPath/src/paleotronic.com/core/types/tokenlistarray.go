package types

type TokenListArray []TokenList

func NewTokenListArray() TokenListArray {
	return make(TokenListArray, 0)
}

func (this TokenListArray) Add(tl TokenList) TokenListArray {
	this = append(this, tl)
	return this
}

func (this TokenListArray) Get(index int) *TokenList {
	return &this[index]
}

func (this TokenListArray) Size() int {
	return len(this)
}

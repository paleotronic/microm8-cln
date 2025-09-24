package types

type CompareFlag int

const (
	EQUAL       CompareFlag = 1 << iota
	LESSTHAN    CompareFlag = 1 << iota
	GREATERTHAN CompareFlag = 1 << iota
)

type CompareFlags struct {
	Value CompareFlag
}

func (this *CompareFlags) IsSet(v CompareFlag) bool {
	return ((this.Value & v) == v)
}

func (this *CompareFlags) Set(v CompareFlag) {
	this.Value = this.Value | v
}

func (this *CompareFlags) Clear(v CompareFlag) {
	this.Value = this.Value & (^v)
}

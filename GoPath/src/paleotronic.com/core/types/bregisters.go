package types

type BRegisters struct {
	Buffer *TokenList
	TREG   []Token
	SREG   []string
	IREG   []int
	EREG   []float64
	COMP   CompareFlags
}

func NewBRegisters(o BRegisters) *BRegisters {
	this := &BRegisters{}
	this.TREG = make([]Token, 16)
	this.SREG = make([]string, 16)
	this.IREG = make([]int, 16)
	this.EREG = make([]float64, 16)

	/* vars */
	var i int

	/* create based on existing set */
	this.Buffer = o.Buffer.SubList(0, o.Buffer.Size())

	/* fill in */
	for i = 0; i <= 15; i++ {
		this.IREG[i] = o.IREG[i]
		this.EREG[i] = o.EREG[i]
		this.SREG[i] = o.SREG[i]
		this.TREG[i] = o.TREG[i]
	}

	return this
}

func NewBRegistersBlank() *BRegisters {
	this := &BRegisters{}
	this.TREG = make([]Token, 16)
	this.SREG = make([]string, 16)
	this.IREG = make([]int, 16)
	this.EREG = make([]float64, 16)

	/* create based on existing set */
	this.Buffer = NewTokenList()

	return this
}

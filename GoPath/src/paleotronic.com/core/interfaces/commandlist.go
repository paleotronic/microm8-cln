package interfaces

type CommandList map[string]Commander

func NewCommandList() CommandList {
	return make(CommandList)
}

func (this CommandList) ContainsKey(s string) bool {
	_, ok := this[s]
	return ok
}

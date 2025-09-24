package types

type NameList map[string]int

func NewNameList() NameList {
	return make(NameList, 0)
}

func (this NameList) ContainsKey(s string) bool {
	_, ok := this[s]
	return ok
}

func (this NameList) Push(s string) {
		this[s] = 1
}

func (this NameList) Clear() {
	this = make(NameList);
}

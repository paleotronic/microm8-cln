package types

type LoopStateMap map[string]LoopState

func NewLoopStateMap() LoopStateMap {
	return make(LoopStateMap)
}

func (this LoopStateMap) ContainsKey( s string ) bool {
	_, ok := this[s]
	return ok
}

func (this LoopStateMap) Put( s string, st LoopState) {
	this[s] = st
}

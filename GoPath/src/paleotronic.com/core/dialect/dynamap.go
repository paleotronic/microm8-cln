package dialect

import (
	"paleotronic.com/core/interfaces"
)

type DynaMap map[string]interfaces.DynaCoder

func NewDynaMap() DynaMap {
	return make(DynaMap)
}

func (this DynaMap) Put(n string, v interfaces.DynaCoder) {
	this[n] = v
}

func (this DynaMap) Get(n string) interfaces.DynaCoder {
	return this[n]
}

func (this DynaMap) ContainsKey(n string) bool {
	_, ok := this[n]
	return ok
}

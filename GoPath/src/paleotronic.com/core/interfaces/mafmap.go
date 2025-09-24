package interfaces

type MafMap map[string]MultiArgumentFunction

func NewMafMap() MafMap {
	return make(MafMap)
}

func (this MafMap) ContainsKey(s string) bool {
	_, ok := this[s]
	return ok
}

func (this MafMap) Get(s string) MultiArgumentFunction {
	m, _ := this[s]
	return m
}

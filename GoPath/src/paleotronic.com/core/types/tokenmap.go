package types

type TokenMap struct {
	Content map[string]*Token
}

func NewTokenMap() *TokenMap {
	this := &TokenMap{Content: make(map[string]*Token)}
	return this
}

// True if map contains variable with that name
func (this TokenMap) Contains(n string) bool {
	_, ok := this.Content[n]
	return ok
}

func (this TokenMap) ContainsKey(n string) bool {
	_, ok := this.Content[n]
	return ok
}

// Returns a Token for a given name, or nil
func (this *TokenMap) Get(n string) *Token {
	cn := n
	if this.Contains(cn) {
		return this.Content[cn]
	}
	return nil
}

// Maps a Token to a particular key
func (this *TokenMap) Put(n string, v *Token) {
	cn := n
	this.Content[cn] = v
}

// Returns a list of Token names (keys)
func (this *TokenMap) Keys() []string {
	v := make([]string, len(this.Content))
	i := 0
	for k := range this.Content {
		v[i] = k
		i++
	}
	return v
}

// Returns a list of Variables (values)
func (this *TokenMap) Values() []*Token {
	v := make([]*Token, len(this.Content))
	i := 0
	for _, c := range this.Content {
		v[i] = c
		i++
	}
	return v
}

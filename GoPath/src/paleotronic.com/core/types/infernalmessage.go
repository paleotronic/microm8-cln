package types

type InfernalMessage struct {
	Kind    string
	Content string
	Sender  string
}

func NewInfernalMessage(name string, kind string, content string) *InfernalMessage {
	this := &InfernalMessage{}
	this.Sender = name
	this.Kind = kind
	this.Content = content
	return this
}

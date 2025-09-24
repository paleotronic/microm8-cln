package vduproto

type Streamable interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary() error
	Identity() byte
}

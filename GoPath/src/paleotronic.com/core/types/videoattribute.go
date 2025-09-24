package types

type VideoAttribute int

const (
	VA_NORMAL  VideoAttribute = 1 << iota
	VA_INVERSE VideoAttribute = 1 << iota
	VA_BLINK   VideoAttribute = 1 << iota
)
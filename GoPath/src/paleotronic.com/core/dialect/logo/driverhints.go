package logo

type LogoCmdHint struct {
	Min int
	Max int
}

var minCommandLengths = map[string]*LogoCmdHint{
	"to":     &LogoCmdHint{2, 0},
	"repeat": &LogoCmdHint{3, 0},
	"pr":     &LogoCmdHint{2, 0},
	"erase":  &LogoCmdHint{2, 0},
	"bury":   &LogoCmdHint{2, 0},
}

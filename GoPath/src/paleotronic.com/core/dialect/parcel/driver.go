package parcel

// GrammarDriver handles output from the tokenizer
type GrammarDriver interface {
	OnTokenRecognized(t *TokenMatchResult) error
	OnTokenUnrecognized(line string, pos int, err error) error
	OnBeginStream(l *Lexer) error
	OnEndStream(l *Lexer) error
}

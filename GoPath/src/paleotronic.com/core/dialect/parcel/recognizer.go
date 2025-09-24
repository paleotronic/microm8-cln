package parcel

// Recognizer implements recognizer for a particular type
type Recognizer func(r []rune) (int, []rune, []rune)


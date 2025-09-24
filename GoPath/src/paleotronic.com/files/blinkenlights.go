package files

var BlinkenCallback0 [10]func(b bool)
var BlinkenCallback1 [10]func(b bool)

func SetLED0(b bool) {
	for index := 0; index < 10; index++ {
		cf := BlinkenCallback0[index]
		if cf != nil {
			cf(b)
		}
	}
}

func SetLED1(b bool) {
	for index := 0; index < 10; index++ {
		cf := BlinkenCallback1[index]
		if cf != nil {
			cf(b)
		}
	}
}

func SetBlink0Callback(index int, f func(b bool)) {
	BlinkenCallback0[index] = f
}

func SetBlink1Callback(index int, f func(b bool)) {
	BlinkenCallback1[index] = f
}

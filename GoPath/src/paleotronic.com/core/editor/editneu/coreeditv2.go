package editneu

type EditBuffer struct {
	TEXT  []rune
	ATTR  []rune
	LINES []int
}

// Grow the buffer by inserting count characters
func (b *EditBuffer) grow(index int, count int) {
	cs := 0
	ce := len(b.TEXT)

	attchunk := make([]rune, count)
	txtchunk := make([]rune, count)

	if index < ce {
		// break insert
		txta := b.TEXT[cs:index]
		atta := b.ATTR[cs:index]
		txtb := b.TEXT[index:ce]
		attb := b.ATTR[index:ce]

		b.TEXT = append(txta, txtchunk...)
		b.TEXT = append(b.TEXT, txtb...)

		b.ATTR = append(atta, attchunk...)
		b.ATTR = append(b.ATTR, attb...)

	} else {
		// append data
		b.TEXT = append(b.TEXT, txtchunk...)
		b.ATTR = append(b.ATTR, attchunk...)
	}
}

func (b *EditBuffer) shrink(index int, count int) {
	cs := 0
	ce := len(b.TEXT)

	if index <= len(b.TEXT) && index+count > ce {
		// trim tail
		b.TEXT = b.TEXT[0:index]
		b.ATTR = b.ATTR[0:index]
	} else {
		// chunk out the middle
		txta := b.TEXT[cs:index]         // start
		atta := b.ATTR[cs:index]         // start
		txtb := b.TEXT[index+count : ce] // end
		attb := b.ATTR[index+count : ce] // end
		b.TEXT = append(txta, txtb...)
		b.ATTR = append(atta, attb...)
	}
}

func (b *EditBuffer) scanLines() {
	// update lines from
	l := make([]int, 1)
	l[0] = 0
	for i, v := range b.TEXT {
		if v == 10 {
			l = append(l, i+1) // next char is start of line
		}
	}
	b.LINES = l
}

func (b *EditBuffer) Count() int {
	if b.LINES == nil {
		b.scanLines()
	}
	return len(b.LINES)
}

func (b *EditBuffer) GetLine(i int) ([]rune, []rune) {

	var txt []rune
	var att []rune

	max := b.Count()

	if i < 0 || i >= max {
		return txt, att
	}

	var s int
	var e int
	if l == max-1 {
		// last line
		s = b.LINES[l]
		e = len(b.TEXT)
	} else {
		s = b.LINES[l]
		e = b.LINES[l+1]
	}

}

func (b *EditBuffer) InsertString(index int)

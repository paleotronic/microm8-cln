package runestring

type RuneString struct {
	Runes []rune
}

func (this *RuneString) Assign(s string) {
	this.Runes = make([]rune, 0)
	this.Append(s)
}

func (this *RuneString) Append(s string) {
	for _, r := range s {
		this.Runes = append(this.Runes, r)
	}
}

func (this *RuneString) AppendRunes(s RuneString) {
	for _, r := range s.Runes {
		this.Runes = append(this.Runes, r)
	}

}

func (this *RuneString) AppendSlice(s []rune) {
	for _, r := range s {
		this.Runes = append(this.Runes, r)
	}
}

func (this RuneString) Length() int {
	return len(this.Runes)
}

func (this RuneString) String() string {
	return string(this.Runes)
}

func (this RuneString) SubString(s, e int) RuneString {

	if e > len(this.Runes) {
		e = len(this.Runes)
	}

	ss := make([]rune, 0)
	for i := s; i < e; i++ {
		ss = append(ss, this.Runes[i])
	}

	return RuneString{Runes: ss}
}

func NewRuneString() RuneString {
	return RuneString{Runes: []rune(nil)}
}

func Cast(s string) RuneString {
	ss := NewRuneString()
	ss.Append(s)
	return ss
}

func Copy(str RuneString, start int, count int) RuneString {
	sindex := start - 1
	eindex := sindex + count

	if sindex > str.Length() {
		return NewRuneString()
	}

	if eindex < sindex {
		return NewRuneString()
	}

	if eindex > str.Length() {
		eindex = str.Length()
	}

	return str.SubString(sindex, eindex)
}

func Concat(s1, s2 RuneString) RuneString {
	tmp := s1
	tmp.AppendRunes(s2)
	return tmp
}

func Delete(str RuneString, start int, count int) RuneString {
	sindex := start - 1
	eindex := sindex + count

	if sindex > str.Length() {
		return str
	}

	if eindex < sindex {
		return str
	}

	if eindex > str.Length() {
		eindex = str.Length()
	}

	return Concat(str.SubString(0, sindex), str.SubString(eindex, str.Length()))
}

func Pos(ch rune, str RuneString) int {
	for i, r := range str.Runes {
		if r == ch {
			return i + 1
		}
	}
	return 0
}

func l(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

func (r RuneString) HasPrefix(a RuneString) bool {
	if a.Length() == 0 || a.Length() > r.Length() {
		return false
	}
	for i, ch := range a.Runes {
		if l(r.Runes[i]) != l(ch) {
			return false
		}
	}
	return true
}

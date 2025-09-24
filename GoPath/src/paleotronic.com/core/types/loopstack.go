package types

import "errors"

type LoopStack []LoopState

func (this *LoopStack) Add(ls LoopState) {
	s := *this
	s = append(s, ls)
	*this = s
}

func (this *LoopStack) Remove(i int) *LoopState {
	s := *this

	if (i < 0) || (i >= len(s)) {
		return nil
	}

	ls := s[i]

	s = append(s[:i], s[i+1:]...)

	*this = s

	return &ls
}

func (this *LoopStack) Get(i int) *LoopState {
	s := *this

	if (i < 0) || (i >= len(s)) {
		return nil
	}

	ls := s[i]

	return &ls
}

func (this *LoopStack) Size() int {
	return len(*this)
}

func (this *LoopStack) Clear() {
	*this = make(LoopStack, 0)
}

func (this *LoopStack) MarshalBinary() ([]uint64, error) {

	var data []uint64
	data = append(data, uint64(this.Size()))

	for _, v := range *this {
		b, _ := v.MarshalBinary()
		data = append(data, b...)
	}

	return data, nil

}

func (this *LoopStack) UnmarshalBinary(data []uint64) error {

	if len(data) < 1 || len(data) < 1+int(data[0])*9 {
		return errors.New("not enough data")
	}

	this.Clear()
	count := int(data[0])
	for i := 0; i < count; i++ {
		offs := 1 + i*9
		chunk := data[offs : offs+9]
		ls := &LoopState{}
		_ = ls.UnmarshalBinary(chunk)
		this.Add(*ls)
	}

	return nil
}

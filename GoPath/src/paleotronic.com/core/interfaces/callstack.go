package interfaces

import (
	"errors"
)

type CallStack struct {
	Content []*StackEntry
}

func NewCallStack() *CallStack {
	return &CallStack{}
}

// pop entry off stack
func (this *CallStack) Pop() (*StackEntry, error) {
	if this.Size() == 0 {
		return nil, errors.New("Nothing on stack")
	}
	var e *StackEntry
	e, this.Content = this.Content[len(this.Content)-1], this.Content[:len(this.Content)-1]
	return e, nil
}

func (this *CallStack) Push(entry *StackEntry) {
	this.Content = append(this.Content, entry)
}

func (this *CallStack) Clear() {
	this.Content = []*StackEntry(nil)
}

func (this *CallStack) Size() int {
	return len(this.Content)
}

func (this *CallStack) Get(index int) *StackEntry {
	return this.Content[index]
}

// --
func (this *CallStack) MarshalBinary() ([]uint64, error) {

	var data []uint64
	data = append(data, uint64(this.Size()))

	for _, v := range this.Content {
		b, _ := v.MarshalBinary()
		data = append(data, b...)
	}

	return data, nil

}

func (this *CallStack) UnmarshalBinary(data []uint64) error {

	if len(data) < 1 || len(data) < 1+int(data[0])*8 {
		return errors.New("not enough data")
	}

	this.Clear()
	count := int(data[0])
	for i := 0; i < count; i++ {
		offs := 1 + i*8
		chunk := data[offs : offs+9]
		ls := &StackEntry{}
		_ = ls.UnmarshalBinary(chunk)
		this.Push(ls)
	}

	return nil
}

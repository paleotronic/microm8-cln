package logo

import (
	"errors"
)

type LogoStack struct {
	s []*LogoScope
}

func NewLogoStack() *LogoStack {
	return &LogoStack{
		s: []*LogoScope{},
	}
}

func (s *LogoStack) Empty() {
	s.s = []*LogoScope{}
}

func (s *LogoStack) Pop() *LogoScope {
	if len(s.s) == 0 {
		return nil
	}
	p := s.s[len(s.s)-1]
	s.s = s.s[:len(s.s)-1]
	return p
}

const StackLimit = 512

func (s *LogoStack) Push(p *LogoScope) error {
	if s.Size() >= StackLimit {
		return errors.New("stack exceeded")
	}
	s.s = append(s.s, p)
	return nil
}

func (s *LogoStack) Bottom() *LogoScope {
	if len(s.s) == 0 {
		return nil
	}
	p := s.s[0]
	return p
}

func (s *LogoStack) Top() *LogoScope {
	if len(s.s) == 0 {
		return nil
	}
	p := s.s[len(s.s)-1]
	return p
}

func (s *LogoStack) Size() int {
	return len(s.s)
}

func (s *LogoStack) Get(i int) *LogoScope {
	if i < 0 || i >= len(s.s) {
		return nil
	}
	p := s.s[i]
	return p
}

func (d *LogoDriver) Reset() {
	d.State = lsStopped
	d.Stack.Empty()
	d.LastReturn = nil
}

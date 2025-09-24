package main

import "paleotronic.com/core/memory"

type MockInt struct{}

var m = memory.NewMemoryMap()

func (i *MockInt) GetMemoryMap() *memory.MemoryMap {
	return m
}

func (i *MockInt) GetMemIndex() int {
	return 0
}

package types

import (
	"paleotronic.com/fmt"

	"paleotronic.com/core/memory"
)

type InlineImageManager struct {
	Index int
	M     *memory.MemoryMap
}

func NewInlineImageManager(index int, mm *memory.MemoryMap) *InlineImageManager {
	this := &InlineImageManager{
		Index: index,
		M:     mm,
	}
	return this
}

func (iim *InlineImageManager) Count() int {

	return int(iim.M.ReadInterpreterMemory(iim.Index, memory.OCTALYZER_MAP_FILENAME_COUNT))

}

func (iim *InlineImageManager) Empty() {
	iim.M.WriteInterpreterMemory(iim.Index, memory.OCTALYZER_MAP_FILENAME_COUNT, 0)
}

func (iim *InlineImageManager) Add(ch rune, imagesource string) {
	c := iim.Count()

	if c == 10 {
		return
	}

	imagesource += "\r"

	offsetS := memory.OCTALYZER_MAP_FILENAME_0 + (memory.OCTALYZER_MAP_FILENAME_SIZE * c)
	offsetR := memory.OCTALYZER_MAP_CHAR_0 + c

	iim.M.WriteInterpreterMemory(iim.Index, offsetR, uint64(ch))

	for i, r := range imagesource {
		iim.M.WriteInterpreterMemory(iim.Index, offsetS+i, uint64(r))
	}

	c += 1

	iim.M.WriteInterpreterMemory(iim.Index, memory.OCTALYZER_MAP_FILENAME_COUNT, uint64(c))
}

func (iim *InlineImageManager) GetImage(num int) (rune, string) {
	c := iim.Count()

	if num >= c {
		return 0, ""
	}

	offsetS := memory.OCTALYZER_MAP_FILENAME_0 + (memory.OCTALYZER_MAP_FILENAME_SIZE * num)
	offsetR := memory.OCTALYZER_MAP_CHAR_0 + num

	ch := rune(iim.M.ReadInterpreterMemory(iim.Index, offsetR))
	s := ""
	i := 0
	r := rune(iim.M.ReadInterpreterMemory(iim.Index, offsetS+i))
	for i < memory.OCTALYZER_MAP_FILENAME_SIZE && r != '\r' {
		s += string(r)
		i++
		r = rune(iim.M.ReadInterpreterMemory(iim.Index, offsetS+i))
	}

	fmt.Printf("Map entry %s -> %s\n", string(ch), s)

	return ch, s
}

func (iim *InlineImageManager) GetMap() map[rune]string {
	c := iim.Count()

	m := make(map[rune]string)

	for i := 0; i < c; i++ {
		ch, s := iim.GetImage(i)
		m[ch] = s
	}

	return m
}

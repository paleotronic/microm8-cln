package editor

import (
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
)

const (
	pcfillchar = 1129
	pcbgchar   = 1060
)

func DrawProgressBar(
	txt *types.TextBuffer,
	x, y int,
	label string,
	percent float64,
	width int,
) {
	w := width - 4
	txt.GotoXY(x, y)
	txt.PutStr(label)
	txt.GotoXY(x, y+1)
	npc := fmt.Sprintf("%3d%%", int(percent*100))
	txt.PutStr(npc)
	filled := int(float64(w) * percent)
	bare := w - filled
	for i := 0; i < filled; i++ {
		txt.Put(pcfillchar)
	}
	for i := 0; i < bare; i++ {
		txt.Put(pcbgchar)
	}
}

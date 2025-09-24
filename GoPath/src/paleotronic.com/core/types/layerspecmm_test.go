package types

import (
	"paleotronic.com/fmt"
	"paleotronic.com/core/memory"
	"paleotronic.com/utils"
	"testing"
)

func TestLayerSpec(t *testing.T) {

	ram := memory.NewMemoryMap()
	
	p := NewVideoPalette()
	p.Add( NewVideoColor(128,0,0,255) )
	
	spec := &LayerSpec{
		ID: "FROG",
		Width: 320,
		Type: 1,
		Height: 200,
		Bounds: LayerRect{ 0, 0, 319, 199 },
		Active: 1,
		Index: 0,
		Format: LF_HGR_X,
		Blocks: []memory.MemoryRange{ memory.MemoryRange{Base: 0x400, Size: 0x800} },
		Palette: *p,
	}
	
	ls := NewLayerSpecMapped( ram, spec, 0, memory.OCTALYZER_HUD_BASE )
	
	if ls.GetID() != "FROG" {
		t.Error( "Expecting FROG, got ["+ls.GetID()+"]" )
	}
	
	if ls.GetWidth() != 320 {
		t.Error( "Expecting 320, got ["+utils.IntToStr(int(ls.GetWidth()))+"]" )
	}
	
	if ls.GetHeight() != 200 {
		t.Error( "Expecting 200, got ["+utils.IntToStr(int(ls.GetHeight()))+"]" )
	}
	
	if !ls.GetActive() {
		t.Error( "Block should be active" )
	}
	
	if ls.GetFormat() != LF_HGR_X {
		t.Error( "Layer should have LF_HGR_X format" )
	}
	
	x0, y0, x1, y1 := ls.GetBounds()
	if x0 != 0 || x1 != 319 || y0 != 0 || y1 != 199 {
		fmt.Println( x0, y0, x1, y1 )
		t.Error( "Bounds failure" )
	}
	
	if ls.GetNumBlocks() != 1 {
		t.Error("Block count invalid")
	}
	
	base, size := ls.GetBlock(0)
	if base != 0x400 {
		t.Error( "Block base incorrect" )
	}
	if size != 0x800 {
		t.Error( "Block size incorrect" )
	}
	
	c := ls.GetPaletteColor( 0 )
	fmt.Println(c)

	fmt.Println(ls.String())

}

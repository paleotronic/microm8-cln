package types

/*
import (
	"paleotronic.com/fmt"
	"testing"
)

func TestWozDeWoz(t *testing.T) {
	x, y := 23, 11
	offset := baseOffsetWoz(x, y)
	nx, ny := offsetToXYWoz(offset)

	if nx != x || ny != y {
		t.Error("Did not get original x, y back from Woz/UnWoz")
	}
}

func TestScreenPrintNN(t *testing.T) {
	tb := NewTextBuffer(false, 0, W_NORMAL_H_NORMAL)

	for c := 0; c < 80; c++ {
		tb.Put(rune(65 + c))
	}

	//raw := tb.GetStrings()
	var r rune
	var e error

	r, e = tb.RuneAt(0, 0)
	if e != nil || r != 'A' {
		t.Error("A not where it should be")
	}

	r, e = tb.RuneAt(0, 2)
	if e != nil || r != 'i' {
		t.Error("i not where it should be")
	}

	r, e = tb.RuneAt(2, 2)
	if e != nil || r != 'j' {
		t.Error("j not where it should be")
	}

	//fmt.Println(tb.GetStrings())

}

func TestScreenPrintHH(t *testing.T) {
	tb := NewTextBuffer(false, 0, W_HALF_H_HALF)

	for c := 0; c < 80; c++ {
		tb.Put(rune(65 + c))
	}

	//raw := tb.GetStrings()
	var r rune
	var e error

	r, e = tb.RuneAt(0, 0)
	if e != nil || r != 'A' {
		t.Error("A not where it should be")
	}

	r, e = tb.RuneAt(40, 0)
	if e != nil || r != 'i' {
		t.Error("i not where it should be")
	}

	r, e = tb.RuneAt(41,0)
	if e != nil || r != 'j' {
		t.Error("j not where it should be")
	}

	//fmt.Println(tb.GetStrings())

}

func TestScreenPrintNNScroll(t *testing.T) {
	tb := NewTextBuffer(false, 0, W_NORMAL_H_NORMAL)

	tb.GotoXY(0, 46)

	for c := 0; c < 80; c++ {
		tb.Put(rune(65 + c))
	}

	//raw := tb.GetStrings()
	var r rune
	var e error

	r, e = tb.RuneAt(0, 42)
	if e != nil || r != 'A' {
		t.Error("A not where it should be")
	}

	r, e = tb.RuneAt(0, 44)
	if e != nil || r != 'i' {
		t.Error("i not where it should be")
	}

	r, e = tb.RuneAt(2, 44)
	if e != nil || r != 'j' {
		t.Error("j not where it should be")
	}

	//fmt.Println(tb.GetStrings())

}

func TestScreenPrintNHScroll(t *testing.T) {
	tb := NewTextBuffer(false, 0, W_NORMAL_H_HALF)

	tb.GotoXY(0, 46)

	for c := 0; c < 80; c++ {
		tb.Put(rune(65 + c))
	}

	//raw := tb.GetStrings()
	var r rune
	var e error

	r, e = tb.RuneAt(0, 45)
	if e != nil || r != 'A' {
		t.Error("A not where it should be")
	}

	r, e = tb.RuneAt(0, 46)
	if e != nil || r != 'i' {
		t.Error("i not where it should be")
	}

	r, e = tb.RuneAt(2, 46)
	if e != nil || r != 'j' {
		t.Error("j not where it should be")
	}

	//fmt.Println(tb.GetStrings())

}

func TestScreenPrintDHScroll(t *testing.T) {
	tb := NewTextBuffer(false, 0, W_DOUBLE_H_HALF)

	tb.GotoXY(0, 46)

	for c := 0; c < 80; c++ {
		tb.Put(rune(65 + c))
	}

	//raw := tb.GetStrings()
	var r rune
	var e error

	r, e = tb.RuneAt(0, 43)
	if e != nil || r != 'A' {
		t.Error("A not where it should be")
	}

	r, e = tb.RuneAt(0, 45)
	if e != nil || r != 'i' {
		t.Error("i not where it should be")
	}

	r, e = tb.RuneAt(4, 45)
	if e != nil || r != 'j' {
		t.Error("j not where it should be")
	}

	//fmt.Println(tb.GetStrings())

}

func TestScreenPrintNNClrEOL(t *testing.T) {
	tb := NewTextBuffer(false, 0, W_NORMAL_H_NORMAL)

	tb.GotoXY(40, 0)   // Half way across screen

	tb.Fill(true, 0, 0, bufferWidth-1, bufferHeight-1, '@', VA_NORMAL, 0, 15)
	tb.ClearToEOLWindow()

	var r rune
	var e error

	r, e = tb.RuneAt(40, 0)
	if e != nil || r != ' ' {
		t.Error("clreol not rooit")
	}

	r, e = tb.RuneAt(40, 2)
	if e != nil || r != '@' {
		t.Error("clreol not rooit")
	}

	//fmt.Println(tb.GetStrings())
}

func TestScreenPrintNNClrBottom(t *testing.T) {
	tb := NewTextBuffer(false, 0, W_NORMAL_H_NORMAL)

	tb.GotoXY(40, 0)   // Half way across screen

	tb.Fill(true, 0, 0, bufferWidth-1, bufferHeight-1, '@', VA_NORMAL, 0, 15)
	tb.ClearToBottomWindow()

	var r rune
	var e error

	r, e = tb.RuneAt(40, 0)
	if e != nil || r != ' ' {
		t.Error("clrtobottom not right")
	}

	r, e = tb.RuneAt(0, 2)
	if e != nil || r != ' ' {
		t.Error("clrtobottom not right")
	}

	//fmt.Println(tb.GetStrings())
}

func TestScreenTabNN(t *testing.T) {
	tb := NewTextBuffer(false, 0, W_NORMAL_H_NORMAL)
	tb.VerticalTab(5)
	tb.HorizontalTab(6)

	if tb.CX != 10 {
		t.Error("CX wrong after htab")
	}

	if tb.CY != 8 {
		t.Error("CY wrong after vtab")
		//fmt.Println(tb.CY)
	}

	tb.Put('X')

	//fmt.Println( tb.GetStrings() )

}


func TestError(t *testing.T) {
	t.Error("moo")
}*/

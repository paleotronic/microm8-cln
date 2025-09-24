package chat

import (
	"strings"

	"paleotronic.com/core/editor"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/core/types"
	"paleotronic.com/fmt"
	"paleotronic.com/log"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"
)

type ChatBuffer struct {
	edit *editor.CoreEdit
	//fg, bg       uint64
	SuppressDraw bool
	WrapMargin   int
}

func NewChatBuffer(e interfaces.Interpretable, width, rt, rb int) *ChatBuffer {
	cb := &ChatBuffer{
		edit: editor.NewCoreEdit(e, "", "", true, false),
	}
	cb.edit.Width = width
	cb.edit.ReservedTop = rt
	cb.edit.ReservedBot = rb
	return cb
}

func (cb *ChatBuffer) Printf(pattern string, args ...interface{}) {

	s := utils.Unescape(fmt.Sprintf(pattern, args...))

	for _, ch := range s {
		cb.edit.OnEditorKeypress(ch)
		if cb.edit.Column > cb.edit.Width-1 {
			cb.edit.OnEditorKeypress(13)
			cb.edit.OnEditorKeypress(32)
		}
	}

}

func (cb *ChatBuffer) Printfr(pattern string, args ...interface{}) {

	s := fmt.Sprintf(pattern, args...)

	for _, ch := range s {
		cb.edit.OnEditorKeypress(ch)
	}

}

func (cb *ChatBuffer) Printfw(pattern string, args ...interface{}) {

	s := fmt.Sprintf(pattern, args...)

	words := strings.Fields(s)

	for i, w := range words {

		if i > 0 {
			cb.edit.OnEditorKeypress(32)
		}
		if len(w) < cb.edit.Width && cb.edit.Column+len(w) >= cb.edit.Width {
			cb.edit.OnEditorKeypress(13)
			cb.edit.OnEditorKeypress(32)
		}
		log.Printf("putting word (length %d) at %d: %s (width=%d)", len(w), cb.edit.Column, w, cb.edit.Width)
		cb.Printf("%s", w)
	}

}

func (cb *ChatBuffer) Empty() {
	cb.edit.SetText("\r\n")
	cb.edit.Voffset = 0
	cb.edit.Hoffset = 0
	cb.edit.Line = 0
	cb.edit.Column = 0
}

func (c *ChatBuffer) Draw(cc *ChatClient) {
	if c.SuppressDraw {
		return
	}
	cc.ChatlogDo(
		func(txt *types.TextBuffer) {
			maxline := txt.EY - txt.SY
			txt.ClearScreenWindow()
			txt.GotoXYWindow(0, 0)
			log.Printf("MaxLine = %d", maxline)

			if c.edit.Voffset >= len(c.edit.Content)-1 {
				c.edit.Voffset = len(c.edit.Content) - maxline - 1
			}
			if c.edit.Voffset < 0 {
				c.edit.Voffset = 0
			}

			for i := 0; i <= maxline; i++ {

				txt.FGColor = cc.chatlogFG
				txt.BGColor = cc.chatlogBG
				txt.Shade = 0

				txt.GotoXYWindow(0, i)

				realLine := c.edit.Voffset + i

				txt.Attribute = types.VA_NORMAL

				//log.Printf("Voffset = %d", c.edit.Voffset)

				if (realLine >= 0) && (realLine < len(c.edit.Content)) {

					//log.Printf("Display line %d -> real line %d", i, realLine)

					s := c.edit.Content[realLine]

					r := c.edit.ProcessHighlight(realLine, s)

					if c.edit.Hoffset > 0 {
						if len(r.Data.Runes) > c.edit.Hoffset {
							r.Data = runestring.Delete(r.Data, 1, c.edit.Hoffset)
							r.Colour = runestring.Delete(r.Colour, 1, c.edit.Hoffset)
						} else {
							r.Data.Assign(" ")
							r.Colour.Assign(string(rune(c.edit.FGColor | (c.edit.BGColor << 4))))
						}
					}
					r.Data = runestring.Copy(r.Data, 1, c.edit.Width)
					r.Colour = runestring.Copy(r.Colour, 1, c.edit.Width)

					// now display it
					if len(r.Data.Runes) > 0 {
						for zz := 0; zz < len(r.Data.Runes); zz++ {
							// if r.Colour.Runes[zz]&256 != 0 {
							// 	apple2helpers.ColorFlip = true
							// } else {
							// 	apple2helpers.ColorFlip = false
							// }

							fgcol := r.Colour.Runes[zz] & 15
							bgcol := (r.Colour.Runes[zz] >> 4) & 15

							txt.FGColor = uint64(fgcol)
							txt.BGColor = uint64(bgcol)
							// txt.Shade = uint64((r.Colour.Runes[zz] >> 16) & 7)
							//txt.Shade = 0
							if r.Data.Runes[zz] != 7 {
								txt.PutStr(string(r.Data.Runes[zz]))
							}
						}

					}

					// for zz := len(r.Data.Runes); zz < cc.ChatWidth; zz++ {
					// 	fgcol := cc.chatlogFG
					// 	bgcol := cc.chatlogBG

					// 	txt.FGColor = uint64(fgcol)
					// 	txt.BGColor = uint64(bgcol)
					// 	txt.Shade = 0
					// 	txt.PutStr(" ")
					// }
					txt.ClearToEOLWindow()

					txt.FGColor = cc.chatlogFG
					txt.BGColor = cc.chatlogBG
					txt.Shade = 0

					// for apple2helpers.GetCursorX(c.edit.Int) != 0 {
					// 	txt.PutStr(" ")
					// }
				} else {
					s := " "
					txt.PutStr(s)
					// for apple2helpers.GetCursorX(c.edit.Int) != 0 {
					// 	txt.PutStr(" ")
					// }
				}

			}
		},
	)
}

func (c *ChatBuffer) Dump() {
	//log.Printf("Text: %s", c.edit.GetText())
}

func (c *ChatBuffer) GotoBottom() {
	c.edit.Voffset = 0
	c.edit.Line = 0
	for i := 0; i < len(c.edit.Content); i++ {
		c.edit.CursorDown()
	}
}

func (c *ChatBuffer) Up() {
	if c.edit.Voffset > 0 {
		c.edit.Voffset--
	}
}

func (c *ChatBuffer) Down() {
	if c.edit.Voffset < len(c.edit.Content) {
		c.edit.Voffset++
	}
}

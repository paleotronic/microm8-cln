// +build !nox

package ui

import "paleotronic.com/core/settings"

func NewMenu(parent *Menu) *Menu {
	if settings.HighContrastUI {
		return &Menu{
			Items:    make([]*MenuItem, 0, 16),
			Fg:       0,
			Bg:       15,
			SelBg:    0,
			SelFg:    15,
			TitBg:    15,
			TitFg:    15,
			selected: 0,
			parent:   parent,
		}
	}
	return &Menu{
		Items:    make([]*MenuItem, 0, 16),
		Fg:       15,
		Bg:       2,
		SelBg:    15,
		SelFg:    2,
		TitBg:    2,
		TitFg:    12,
		selected: 0,
		parent:   parent,
	}
}

package logo

import (
	"errors"
	"fmt"
	"strings"

	"paleotronic.com/core/types"
	"paleotronic.com/utils"
)

/*
Table driver for logo:
A table is a 2D data-structure.
*/

// CreateTable creates a table in a token...
func TableCreate(name string, rows int, columns int) *types.Token {
	table := types.NewToken(types.TABLE, fmt.Sprintf("%d,%d", rows, columns))
	table.List = types.NewTokenList()
	rowsdata := types.NewTokenList()
	for r := 0; r < rows; r++ {
		row := types.NewToken(types.LIST, "")
		row.List = types.NewTokenList()
		for c := 0; c < columns; c++ {
			row.List.Push(types.NewToken(types.WORD, ""))
		}
		rowsdata.Push(row)
	}
	table.List = rowsdata
	return table
}

var ErrNotATable = errors.New("var is not a table")
var ErrNotEnoughRows = errors.New("invalid row index in table")
var ErrNotEnoughCols = errors.New("invalid column index in table")

func TableGetSize(t *types.Token) (rows int, cols int, err error) {
	if t.Type != types.TABLE {
		err = ErrNotATable
		return
	}
	parts := strings.Split(t.Content, ",")
	rows, cols = utils.StrToInt(parts[0]), utils.StrToInt(parts[1])
	return
}

func TableGetCell(t *types.Token, row, column int) (cell *types.Token, err error) {
	var rows, cols int
	rows, cols, err = TableGetSize(t)
	if err != nil {
		return
	}
	if row >= rows {
		err = ErrNotEnoughRows
		return
	}
	if column >= cols {
		err = ErrNotEnoughCols
		return
	}
	cell = t.List.Content[row].List.Content[column]
	return
}

func TableSetCell(t *types.Token, row, column int, cell *types.Token) error {
	var rows, cols int
	var err error
	rows, cols, err = TableGetSize(t)
	if err != nil {
		return err
	}
	if row >= rows {
		err = ErrNotEnoughRows
		return err
	}
	if column >= cols {
		err = ErrNotEnoughCols
		return err
	}
	t.List.Content[row].List.Content[column] = cell
	return nil
}

// TableSearchRow searches the specified row for the value ... if a number type, then variance is used if supplied
// (variance can be nil, and will be straight == if nil)
func TableSearchRow(t *types.Token, row int, value *types.Token, variance *types.Token) (list *types.TokenList, err error) {
	var rows, cols int
	rows, cols, err = TableGetSize(t)
	if err != nil {
		return
	}
	if row >= rows {
		err = ErrNotEnoughRows
		return
	}

	list = types.NewTokenList()

	var v float64
	var fv float64
	var cv float64

	if value != nil && (value.Type == types.NUMBER || value.Type == types.INTEGER) {
		v = utils.StrToFloat64(value.Content)
	}
	if variance != nil && (variance.Type == types.NUMBER || variance.Type == types.INTEGER) {
		fv = utils.StrToFloat64(variance.Content)
	}

	for c := 0; c < cols; c++ {
		tt := t.List.Content[row].List.Content[c]
		var match bool
		if value.Type == types.NUMBER || value.Type == types.INTEGER {
			cv = utils.StrToFloat64(tt.Content)
			match = cv >= (v-fv) && cv <= (v+fv)
		} else {
			// simple string match
			match = strings.ToLower(tt.Content) == strings.ToLower(value.Content)
		}
		if match {
			list.Push(types.NewToken(types.NUMBER, utils.IntToStr(c)))
		}
	}
	return
}

// TableSearchColumn searches the specified row for the value ... if a number type, then variance is used if supplied
// (variance can be nil, and will be straight == if nil)
func TableSearchColumn(t *types.Token, col int, value *types.Token, variance *types.Token) (list *types.TokenList, err error) {
	var rows, cols int
	rows, cols, err = TableGetSize(t)
	if err != nil {
		return
	}
	if col >= cols {
		err = ErrNotEnoughCols
		return
	}

	list = types.NewTokenList()

	var v float64
	var fv float64
	var cv float64

	if value != nil && (value.Type == types.NUMBER || value.Type == types.INTEGER) {
		v = utils.StrToFloat64(value.Content)
	}
	if variance != nil && (variance.Type == types.NUMBER || variance.Type == types.INTEGER) {
		fv = utils.StrToFloat64(variance.Content)
	}

	for r := 0; r < rows; r++ {
		tt := t.List.Content[r].List.Content[col]
		var match bool
		if value.Type == types.NUMBER || value.Type == types.INTEGER {
			cv = utils.StrToFloat64(tt.Content)
			match = cv >= (v-fv) && cv <= (v+fv)
		} else {
			// simple string match
			match = strings.ToLower(tt.Content) == strings.ToLower(value.Content)
		}
		if match {
			list.Push(types.NewToken(types.NUMBER, utils.IntToStr(r)))
		}
	}
	return
}

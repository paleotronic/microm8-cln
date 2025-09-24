package logo

import (
	"testing"

	"paleotronic.com/core/types"
)

func TestTableGetSet(t *testing.T) {
	tb := TableCreate("frog", 10, 5)
	err := TableSetCell(tb, 6, 2, types.NewToken(types.NUMBER, "4.9"))
	if err != nil {
		t.Fatalf("error setting cell: %v", err)
	}
	cell, err := TableGetCell(tb, 6, 2)
	if err != nil {
		t.Fatalf("error getting cell: %v", err)
	}
	if cell == nil || cell.Type != types.NUMBER || cell.Content != "4.9" {
		t.Fatalf("did not get expected cell")
	}
	cols, err := TableSearchRow(tb, 6, types.NewToken(types.NUMBER, "5.0"), types.NewToken(types.NUMBER, "0.5"))
	if err != nil {
		t.Fatalf("failed to search row: %v", err)
	}
	if cols.Size() != 1 {
		t.Fatalf("expected one item in result")
	}
	if cols.Get(0).Content != "2" {
		t.Fatalf("Didn't get expected column")
	}
	rows, err := TableSearchColumn(tb, 2, types.NewToken(types.NUMBER, "5.0"), types.NewToken(types.NUMBER, "0.5"))
	if err != nil {
		t.Fatalf("failed to search row: %v", err)
	}
	if rows.Size() != 1 {
		t.Fatalf("expected one item in result")
	}
	if rows.Get(0).Content != "6" {
		t.Fatalf("Didn't get expected row")
	}
}

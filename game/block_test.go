package game

import "testing"

func TestNewBlock(t *testing.T) {
	block := NewBlock(1, 2, 3, 4, 5)
	if !(block.X == 1 && block.Y == 2 && block.Dy == 3 && block.w == 4 && block.h == 5) {
		t.Fatalf("Error test for block creating")
	}
}

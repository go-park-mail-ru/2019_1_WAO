package game

import (
	"testing"
)

func TestSize(t *testing.T) {

	player := &Player{
		W: 50,
		H: 40,
	}

	if player.W != 50 {
		t.Errorf("Expected: 50, but got: %f\n", player.W)
	}

}

func TestJump(t *testing.T) {
	players := []*Player{
		&Player{
			Dy: -0.99,
		},
		&Player{
			Dy: 555.46,
		},

		&Player{
			Dy: -0.35,
		},
		&Player{
			Dy: 0,
		},
	}
	expected := -0.35
	for _, player := range players {
		player.Jump()
		if player.Dy != -0.35 {
			t.Errorf("Expected: %f, but got: %f", expected, player.Dy)
		}
	}
}

func TestNewPlayer(t *testing.T) {
	player := NewPlayer(nil)
	if player == nil {
		t.Fatalf("Unexpected nil player\n")
	}
}

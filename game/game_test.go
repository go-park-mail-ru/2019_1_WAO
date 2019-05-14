package game

import (
	"testing"
)

func TestProcessSpeed(t *testing.T) {
	test := struct {
		player   *Player
		delay    float64
		expected float64
	}{
		player: &Player{
			Dy: 10,
		},
		delay:    1.0,
		expected: 10.0004,
	}
	ProcessSpeed(test.delay, test.player)
	if test.player.Dy != test.expected {
		t.Errorf("Expected dy: %f, but got: %f\n", test.expected, test.player.Dy)
	}
	// player.Dy += (gravity * delay)
}

func TestCircleDraw(t *testing.T) {
	tests := []struct {
		player    *Player
		expectedX float64
	}{
		{
			player: &Player{
				X: 397,
			},
			expectedX: 397,
		},
		{
			player: &Player{
				X: 405,
			},
			expectedX: 0,
		},
		{
			player: &Player{
				X: 400,
			},
			expectedX: 400,
		},
		{
			player: &Player{
				X: 0,
			},
			expectedX: 0,
		},
		{
			player: &Player{
				X: -75,
			},
			expectedX: WidthField,
		},
	}

	for _, test := range tests {
		CircleDraw(test.player)
		if test.player.X != test.expectedX {
			t.Errorf("Expected x: %f, but got: %f", test.expectedX, test.player.X)
		}
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

func TestSetPlayerOnPlate(t *testing.T) {
	tests := []struct {
		player    *Player
		block     *Block
		expectedX float64
		expectedY float64
	}{
		{
			player: &Player{
				X: 1,
				Y: 2,
			},
			block: &Block{
				X: 5,
				Y: 6,
				w: 90,
				h: 15,
			},
			expectedX: 50,
			expectedY: -9,
		},
		{
			player: &Player{
				X: 0,
				Y: 17.5,
			},
			block: &Block{
				X: 10,
				Y: 15,
				w: 90,
				h: 15,
			},
			expectedX: 55,
			expectedY: 0,
		},
		{
			player: &Player{
				X: 1,
				Y: 2,
			},
			block: &Block{
				X: -45,
				Y: 105,
				w: 90,
				h: 15,
			},
			expectedX: 0,
			expectedY: 90,
		},
	}
	for _, test := range tests {
		test.player.SetPlayerOnPlate(test.block)
		if test.player.X != test.expectedX {
			t.Errorf("Expected x: %f, but got: %f", test.expectedX, test.player.X)
		}
		if test.player.Y != test.expectedY {
			t.Errorf("Expected y: %f, but got: %f", test.expectedY, test.player.Y)
		}
	}
}

func (player *Player) TestSelectNearestBlock(t *testing.T) {
	// nearestBlock = nil
	// minY := math.MaxFloat64
	// // canvasY := player.canvas.y
	// for _, block := range *blocks {

	// 	// if (block.Y-block.h > canvasY+700) || (block.Y < canvasY) {
	// 	// 	continue
	// 	// }
	// 	if player.X <= block.X+block.w && player.X+player.W >= block.X {
	// 		if math.Abs(block.Y-player.Y) < minY && player.Y <= block.Y {
	// 			minY = math.Abs(block.Y - player.Y)
	// 			nearestBlock = block
	// 		}
	// 	}
	// }
	// return
}

// func TestFieldGenerator(t *testing.T) {
// 	var player *Player
// 	player = FieldGenerator(100, 100, 20)
// 	players = append(players, player)
// 	for _, value := range blocks {
// 		fmt.Println(*value)
// 	}
// 	fmt.Println("Players:")
// 	for _, plr := range players {
// 		fmt.Println(*plr)
// 	}

// }

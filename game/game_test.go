package game

import (
	"testing"

	"github.com/spf13/viper"
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
	ProcessSpeed(test.delay, test.player, 0.0004)
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
			expectedX: 400,
		},
	}

	for id, test := range tests {
		CircleDraw(test.player, 400)
		if test.player.X != test.expectedX {
			t.Errorf("test_id: %d -> Expected x: %f, but got: %f", id, test.expectedX, test.player.X)
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

func TestSelectNearestBlock(t *testing.T) {
	player := &Player{
		X:  120,
		Y:  350,
		Dy: 1,
	}
	blocks := &[]*Block{
		&Block{
			X: 120,
			Y: 200,
			w: 90,
			h: 15,
		},
		&Block{
			X: 60,
			Y: 380,
			w: 90,
			h: 15,
		},
		&Block{
			X: 400,
			Y: 200,
			w: 90,
			h: 15,
		},
	}
	block := player.SelectNearestBlock(blocks)
	if block != (*blocks)[1] {
		t.Fatalf("Expected block x: %f, y: %f; but got block x: %f, y: %f\n", (*blocks)[1].X, (*blocks)[1].Y, block.X, block.Y)
	}
}

func TestCollision(t *testing.T) {

	tests := []struct {
		player   *Player
		block    *Block
		delay    float64
		expected bool
	}{
		{
			player: &Player{
				X:  10,
				Y:  200,
				Dy: -10,
				W:  50,
				H:  40,
			},
			block: &Block{
				X: 20,
				Y: 10,
				w: 90,
				h: 15,
			},
			delay:    1.0,
			expected: false,
		},
		{
			player: &Player{
				X:  120,
				Y:  350,
				Dy: 1,
				W:  50,
				H:  40,
			},
			block: &Block{
				X: 60,
				Y: 351,
				w: 90,
				h: 15,
			},
			delay:    1.0,
			expected: true,
		},
	}
	for id, test := range tests {
		Collision(test.delay, test.player, test.block)
		if (test.player.Dy == -0.35) != test.expected {
			t.Fatalf("test_id: %d\n", id)
		}
	}
}

func TestFieldGenerator(t *testing.T) {
	viper.SetConfigFile("../config/test.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	// beginY was sended as the parameter
	beginY := 400.0
	b := viper.GetFloat64("settings.koefHeightOfMaxGenerateSlice")
	k := uint16(10)
	blocks := FieldGenerator(beginY, b, k)
	// for _, b := range blocks {
	// 	fmt.Printf("x: %f, y: %f, dy: %f, w: %f, h: %f\n", b.X, b.Y, b.Dy, b.w, b.h)
	// }
	if blocks[len(blocks)-1].Y != (beginY - b + (b / float64(k))) {
		t.Fatalf("Last block has incorrect position")
	}
	if len(blocks) != int(k) {
		t.Fatalf("Wrong count of blocks!")
	}
}

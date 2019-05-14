package game

import (
	"testing"
)

func TestProcessSpeed(t *testing.T) {
	test := []struct {
		player   *Player
		delay    float64
		expected float64
	}{
		player: &Player{
			Dy: 10,
		},
		delay:    1.0,
		expected: 19.8,
	}
	if ProcessSpeed(test.delay, test.player) != test.expected {
		t.Errorf("Expected: %f, but got: %f\n", test.expected, ProcessSpeed(test.delay, test.player))
	}
	// player.Dy += (gravity * delay)
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

package game

import "testing"

func TestRemoveRoom(t *testing.T) {
	game := &Game{
		rooms: []*Room{
			&Room{},
			&Room{},
			&Room{},
			&Room{},
		},
	}
	for _, room := range game.rooms {
		room.game = game
	}

	for i := 0; i < 4; i++ {
		game.RemoveRoom(game.rooms[0])
		if len(game.rooms) != 3-i {
			t.Errorf("Expected count of rooms: %d, but got: %d\n", 3-i, len(game.rooms))
		}
	}
}
func TestAddRoom(t *testing.T) {
	game := NewGame(4)
	for i := 0; i < 4; i++ {
		game.AddRoom(NewRoom(2, game))
		if len(game.rooms) != i+1 {
			t.Errorf("Expected count of rooms: %d, but got: %d\n", i+1, len(game.rooms))
		}

	}
	for _, room := range game.rooms {
		game.RemoveRoom(room)
	}
}

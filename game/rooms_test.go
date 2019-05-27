package game

import (
	"sync"
	"testing"

	"github.com/spf13/viper"
)

// func TestAddPlayer(t *testing.T) {
// 	game := NewGame(5)
// 	tests := struct {
// 		room     *Room
// 		players  []*Player
// 		expected int
// 	}{
// 		room: NewRoom(5, game),
// 		players: []*Player{
// 			&Player{},
// 			&Player{},
// 			&Player{},
// 		},
// 		expected: 4,
// 	}
// 	tests.room.AddPlayer(&Player{
// 		room: tests.room,
// 	})
// 	if len(tests.room.Players) != tests.expected {
// 		t.Errorf("Expected count of players: %d, but got: %d\n", tests.expected, len(tests.room.Players))
// 	}
// }

func TestLength(t *testing.T) {

	m := &sync.Map{}
	m.Store("1", 2)
	m.Store(3, "4")
	m.Store(5, struct {
		value int
	}{
		value: 6,
	})

	if length(m) != 3 {
		t.Fatalf("Expected length: 3, but got: %d", length(m))
	}
}

func TestAddPlayer(t *testing.T) {
	viper.SetConfigFile("../config/test.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	player := &Player{}
	game := NewGame(2)
	go game.Run()
	room := NewRoom(2, game)
	go room.Run()
	game.AddRoom(room)
	room.AddPlayer(player)
	if !(player.room == room) {
		t.Fatalf("Player is not in the room")
	}
	if length(&room.Players) != 1 {
		t.Fatalf("Count of players is wrong")
	}
}

func TestRemovePlayer(t *testing.T) {

	viper.SetConfigFile("../config/test.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	player := NewPlayer(nil)
	game := NewGame(2)
	go game.Run()
	room := NewRoom(2, game)
	go room.Run()
	game.AddRoom(room)
	room.AddPlayer(player)
	player.Listen()
	if length(&room.Players) != 0 {
		t.Fatalf("Player still in a room!")
	}
}

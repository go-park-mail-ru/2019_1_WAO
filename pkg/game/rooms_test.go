package game

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/spf13/viper"
)
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
	player := NewPlayer(nil)
	player2 := NewPlayer(nil)
	game := NewGame(2)
	go game.Run()
	room := NewRoom(2, game)
	go room.Run()
	game.AddRoom(room)
	room.AddPlayer(player)
	room.AddPlayer(player2)
	if !(player.room == room) {
		t.Fatalf("Player is not in the room")
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
}

func TestGetCommandStruct(t *testing.T) {
	player := &Player{
		IdP: 9,
	}
	commandData, _ := json.Marshal(Command{
		Direction: "LEFT",
		Delay:     1.15,
	})
	command := (*player.GetCommandStruct(commandData))
	if command.Direction != "LEFT" || command.Delay != 1.15 {
		t.Fatalf("Old content was damaged! New data - direction: %s, delay: %f", command.Direction, command.Delay)
	}
	if command.IdP != player.IdP {
		t.Fatalf("A command has not received player_id: %d, command_id: %d", player.IdP, command.IdP)
	}
}

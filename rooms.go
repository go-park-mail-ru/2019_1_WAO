package main

import (
	"fmt"

	game "./game"
)

// type RoomController struct{}

var Rooms map[string]game.Connections

func InitGame(roomName string) {
	room := Rooms[roomName]
	if room == nil {
		fmt.Println("Error with game init was occured")
		return
	}
	game.GameLoop(&room) // Init a cycle for the room
}

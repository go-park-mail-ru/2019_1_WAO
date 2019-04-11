package main

import (
	"log"
	"sync"
)

type Game struct {
	MaxRooms uint
	rooms    []*Room
	mutex    *sync.Mutex
	register chan *Player
}

func NewGame(maxRooms uint) *Game {
	return &Game{
		register: make(chan *Player),
	}
}

func (g *Game) Run() {
	log.Println("Main loop started")

LOOP:
	for {
		player := <-g.register

		for _, room := range g.rooms {
			if len(room.Players) < int(room.MaxPlayers) {
				room.AddPlayer(player)
				continue LOOP
			}
		}

		room := NewRoom(2)
		g.AddRoom(room)

		go room.Run()

		room.AddPlayer(player)
	}
}

func (g *Game) AddPlayer(player *Player) {
	log.Printf("Player %s queued to add", player.ID)
	g.register <- player
}

func (g *Game) AddRoom(room *Room) {
	g.rooms = append(g.rooms, room)
}

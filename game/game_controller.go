package game

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

type Game struct {
	MaxRooms uint
	rooms    []*Room
	mutex    sync.Mutex
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
			if len(room.Players) < room.MaxPlayers {
				room.AddPlayer(player)
				continue LOOP
			}
		}

		room := NewRoom(2, g)
		g.AddRoom(room)

		go room.Run()

		room.AddPlayer(player)
	}
}

func (g *Game) AddPlayer(player *Player) {
	log.Println("Player %d queued to add")
	go player.Listen()
	g.register <- player
}

func (g *Game) AddRoom(room *Room) {
	g.mutex.Lock()
	g.rooms = append(g.rooms, room)
	g.mutex.Unlock()
}

func (g *Game) RemoveRoom(room *Room) error {
	rooms := &g.rooms
	lastIndex := len(*rooms) - 1
	for index, r := range g.rooms {
		if r == room {
			g.mutex.Lock()
			(*rooms)[index] = (*rooms)[lastIndex]
			g.rooms = (*rooms)[:lastIndex]
			g.mutex.Unlock()
			fmt.Println("The room was deleted")
			return nil
		}
	}
	return errors.New("The room is not found")
}

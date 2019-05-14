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
				g.mutex.Lock()
				room.AddPlayer(player)
				g.mutex.Unlock()
				continue LOOP
			}
		}

		room := NewRoom(2, g)
		g.mutex.Lock()
		g.AddRoom(room)
		go room.Run()
		room.AddPlayer(player)
		g.mutex.Unlock()
	}
}

func (g *Game) AddPlayer(player *Player) {
	log.Println("Player %d queued to add")
	go player.Listen()
	g.register <- player
}

func (g *Game) AddRoom(room *Room) {
	g.rooms = append(g.rooms, room)
}

func (g *Game) RemoveRoom(room *Room) error {

	rooms := &g.rooms
	lastIndex := len(*rooms) - 1
	g.mutex.Lock()
	for index, r := range g.rooms {
		if r == room {

			(*rooms)[index] = (*rooms)[lastIndex]
			g.rooms = (*rooms)[:lastIndex]

			fmt.Println("The room was deleted")
			g.mutex.Unlock()
			return nil
		}
	}
	g.mutex.Unlock()
	return errors.New("The room is not found")
}

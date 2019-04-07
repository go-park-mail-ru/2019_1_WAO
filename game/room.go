package main

import (
	"log"
	"sync"
	"time"
)

type PlayerState struct {
	ID string
	X  int
	Y  int
}

type ObjectState struct {
	ID   string
	Type string
	X    int
	Y    int
}

type RoomState struct {
	Players     []PlayerState
	Objects     []ObjectState
	CurrentTime time.Time
}

type Room struct {
	ID         string
	MaxPlayers uint
	Players    map[string]*Player
	mutex      *sync.Mutex
	register   chan *Player
	unregister chan *Player
	ticker     *time.Ticker
	state      *RoomState
}

func NewRoom(maxPlayers uint) *Room {
	return &Room{
		MaxPlayers: maxPlayers,
		Players:    make(map[string]*Player),
		register:   make(chan *Player),
		unregister: make(chan *Player),
		ticker:     time.NewTicker(1 * time.Second),
	}
}

func (r *Room) Run() {
	log.Println("Room loop started")
	for {
		select {
		case player := <-r.unregister:
			delete(r.Players, player.ID)
			log.Printf("Player %s was remoted from room", player.ID)
		case player := <-r.register:
			r.Players[player.ID] = player
			log.Printf("Player %s added to game", player.ID)
			player.SendMessage(&Message{"Connected", nil})
		case <-r.ticker.C:
			// log.Println("tick")

			// игровая механика
			// взять у player'a команды и обработать их

			// for _, player := range r.Players {
			// 	player.SendState(r.state)
			// }
		}
	}
}

func (r *Room) AddPlayer(player *Player) {
	player.room = r
	r.register <- player
}

func (r *Room) RemovePlayer(player *Player) {
	r.unregister <- player
}

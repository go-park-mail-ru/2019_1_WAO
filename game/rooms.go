package game

import (
	"log"
	"sync"
	"time"
)

// type RoomController struct{}

type Room struct {
	ID         string
	MaxPlayers int
	Players    map[int]*Player
	Blocks     []*Block
	mutex      *sync.Mutex
	register   chan *Player
	unregister chan *Player
	init       chan bool
	ticker     *time.Ticker
	// isRun      bool
}

var Rooms []*Room

func NewRoom(maxPlayers int) *Room {
	return &Room{
		MaxPlayers: maxPlayers,
		Players:    make(map[int]*Player),
		register:   make(chan *Player),
		unregister: make(chan *Player),
		ticker:     time.NewTicker(1 * time.Second),
	}
}

func (room *Room) Run() {
	log.Println("Room loop started")
	// room.isRun = false
	for {
		select {
		case player := <-room.unregister:
			delete(room.Players, player.Id)
			log.Printf("Player %d was remoted from room", player.Id)
		case player := <-room.register:
			room.Players[player.Id] = player
			log.Printf("Player %d added to game\n", player.Id)
			log.Printf("Count of players: %d", len(room.Players))
			// player.connection.SendMessage(&Message{"Connected", nil})
		case <-room.init:
			var wg sync.WaitGroup
			for _, player := range room.Players {
				wg.Add(1)
				go Engine(player, &wg)
			}
			wg.Wait()
		default:
			// log.Println("tick")

			// игровая механика
			// взять у player'a команды и обработать их
			if len(room.Players) == room.MaxPlayers {
				room.init <- true
			}
		}
	}
}

func (room *Room) AddPlayer(player *Player) {
	player.room = room
	room.register <- player
}

func (room *Room) RemovePlayer(player *Player) {
	room.unregister <- player
}

// func InitGame(roomName string) {
// 	room := Rooms[roomName]
// 	if room == nil {
// 		fmt.Println("Error with game init was occured")
// 		return
// 	}
// 	game.GameLoop(&room) // Init a cycle for the room
// }

package game

import (
	"encoding/json"
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
	init       chan struct{}
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
		init:       make(chan struct{}, 1),
		ticker:     time.NewTicker(1 * time.Second),
	}
}

func (room *Room) Run() {
	log.Println("Room loop started")
	// room.isRun = false
	for {
		select {
		case player := <-room.unregister:
			// player.connection.Close()
			log.Println("Unregistering...")
			player.connection.Close()
			delete(room.Players, player.IdP)
			log.Printf("Player %d was remoted from room", player.IdP)
			log.Printf("Count of players: %d", len(room.Players))
		case player := <-room.register:
			player.IdP = len(room.Players)
			room.Players[player.IdP] = player
			log.Printf("Player %d added to game\n", player.IdP)
			log.Printf("Count of players: %d", len(room.Players))
			log.Printf("len(room.Players): %d, room.MaxPlayers: %d", len(room.Players), room.MaxPlayers)
			if len(room.Players) == room.MaxPlayers {
				room.init <- struct{}{}
			}
			// player.connection.SendMessage(&Message{"Connected", nil})
		case <-room.init:
			log.Println("room init")
			type BlocksAndPlayers struct {
				Blocks  []*Block  `json:"blocks"`
				Players []*Player `json:"players"`
			}
			room.Blocks = FieldGenerator(100, 100, 10)
			var players []*Player
			for _, p := range room.Players {
				players = append(players, p) // The
				p.SetPlayerOnPlate(room.Blocks[0])
			}
			blocksAndPlayers := BlocksAndPlayers{
				Blocks:  room.Blocks,
				Players: players,
			}
			payload, err := json.Marshal(blocksAndPlayers)
			if err != nil {
				log.Println("Error blocks and players is occured", err)
				return
			}
			msg := &Message{
				Type:    "init",
				Payload: payload,
			}
			temp := blocksAndPlayers.Players[0]
			blocksAndPlayers.Players[0] = blocksAndPlayers.Players[1]
			blocksAndPlayers.Players[1] = temp
			payload2, err := json.Marshal(blocksAndPlayers)
			if err != nil {
				log.Println("Error blocks and players is occured", err)
				return
			}
			msg2 := &Message{
				Type:    "init",
				Payload: payload2,
			}

			// room.Blocks = blocks
			// for _, player := range room.Players {
			// 	player.SetPlayerOnPlate(room.Blocks[0])
			// 	if player != p {
			// 		player.SendMessage(&msg2)
			// 		continue
			// 	}
			// 	player.SendMessage(&msg)
			// }
			room.Players[0].SendMessage(msg)
			room.Players[1].SendMessage(msg2)
			for {
				var wg sync.WaitGroup
				for _, player := range room.Players {
					wg.Add(1)
					go Engine(player, &wg)
				}
				wg.Wait()
			}
			log.Println("wait finished")
			// default:
			// log.Println("tick")

			// игровая механика
			// взять у player'a команды и обработать их

		}
	}
}

func (room *Room) AddPlayer(player *Player) {
	player.room = room
	room.register <- player
}

func (room *Room) RemovePlayer(player *Player) {

	log.Println("Player was removed!")
	player.room = nil
	log.Printf("Data: id: %d\n", player.IdP)
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

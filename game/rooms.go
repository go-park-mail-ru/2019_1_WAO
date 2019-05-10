package game

import (
	"encoding/json"
	"log"
	"sync"
)

// type RoomController struct{}

type Room struct {
	ID         string
	game       *Game
	MaxPlayers int
	Players    map[int]*Player
	Blocks     []*Block
	mutex      *sync.Mutex
	register   chan *Player
	unregister chan *Player
	init       chan struct{}
	finish     chan struct{}
	mut        sync.Mutex
	// isRun      bool
}

func NewRoom(maxPlayers int, game *Game) *Room {
	return &Room{
		MaxPlayers: maxPlayers,
		game:       game,
		Players:    make(map[int]*Player),
		register:   make(chan *Player),
		unregister: make(chan *Player),
		init:       make(chan struct{}, 1),
		finish:     make(chan struct{}, 1),
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
			room.mut.Lock()
			delete(room.Players, player.IdP)
			room.mut.Unlock()
			log.Printf("Player %d was remoted from room\n", player.IdP)
			log.Printf("Count of players: %d\n", len(room.Players))
			if len(room.Players) == 0 {
				room.finish <- struct{}{}
			}
		case player := <-room.register:
			player.IdP = len(room.Players)
			room.Players[player.IdP] = player
			log.Printf("Player %d added to game\n", player.IdP)
			log.Printf("len(room.Players): %d, room.MaxPlayers: %d\n", len(room.Players), room.MaxPlayers)
			if len(room.Players) == room.MaxPlayers {
				room.init <- struct{}{}
			}
			// player.connection.SendMessage(&Message{"Connected", nil})
		case <-room.init:
			go func() {
				log.Println("room init")
				type BlocksAndPlayers struct {
					Blocks  []*Block  `json:"blocks"`
					Players []*Player `json:"players"`
				}
				room.Blocks = FieldGenerator(HeightField-20, 2000, 2000*0.01)
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
				room.Players[0].SendMessage(msg)
				room.Players[1].SendMessage(msg2)
				for _, player := range room.Players {
					// wg.Add(1)
					go Engine(player)
				}
				for {
					select {
					case <-room.finish:
						room.game.RemoveRoom(room)
					}
				}
			}()

		}
	}
}

func (room *Room) AddPlayer(player *Player) {
	player.room = room
	room.register <- player
}

func (room *Room) RemovePlayer(player *Player) {

	log.Println("Player was removed!")

	log.Printf("Data: id: %d\n", player.IdP)

	room.unregister <- player
	player.engineDone <- struct{}{}
	player.room = nil
}

// func InitGame(roomName string) {
// 	room := Rooms[roomName]
// 	if room == nil {
// 		fmt.Println("Error with game init was occured")
// 		return
// 	}
// 	game.GameLoop(&room) // Init a cycle for the room
// }

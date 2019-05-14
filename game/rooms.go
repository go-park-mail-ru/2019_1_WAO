package game

import (
	"encoding/json"
	"log"
	"sync"
)

// type RoomController struct{}

type Room struct {
	ID                   string
	game                 *Game
	MaxPlayers           int
	Players              map[int]*Player
	Blocks               []*Block
	register             chan *Player
	unregister           chan *Player
	canvasControllerDone chan struct{}
	init                 chan struct{}
	finish               chan struct{}
	mutexRoom            *sync.Mutex
	mutexEngine          *sync.Mutex
	scrollCount          int
	// scroller             *Player
}

func NewRoom(maxPlayers int, game *Game) *Room {
	return &Room{
		MaxPlayers:           maxPlayers,
		game:                 game,
		Players:              make(map[int]*Player),
		register:             make(chan *Player),
		unregister:           make(chan *Player),
		init:                 make(chan struct{}, 1),
		finish:               make(chan struct{}, 1),
		canvasControllerDone: make(chan struct{}, 1),
		mutexRoom:            &sync.Mutex{},
		mutexEngine:          &sync.Mutex{},
		scrollCount:          0,
		// scroller:             nil,
	}
}

func (room *Room) Run() {
	log.Println("Room loop started")
	for {
		select {
		case player := <-room.unregister:
			log.Println("Unregistering...")
			room.mutexEngine.Lock()
			delete(room.Players, player.IdP)
			room.mutexEngine.Unlock()
			log.Printf("Player %d was remoted from room\n", player.IdP)
			log.Printf("Count of players: %d\n", len(room.Players))
			room.mutexEngine.Lock()
			countOfPlayers := len(room.Players)
			room.mutexEngine.Unlock()
			player.messagesClose <- struct{}{}
			if countOfPlayers == 0 {
				room.finish <- struct{}{}
			}
		case player := <-room.register:
			room.mutexRoom.Lock()
			player.IdP = len(room.Players)
			room.Players[player.IdP] = player
			log.Printf("Player %d added to game\n", player.IdP)
			log.Printf("len(room.Players): %d, room.MaxPlayers: %d\n", len(room.Players), room.MaxPlayers)
			room.mutexRoom.Unlock()
			if len(room.Players) == room.MaxPlayers {
				room.init <- struct{}{}
			}
		case <-room.init:
			go func() {

				type BlocksAndPlayers struct {
					Blocks  []*Block  `json:"blocks"`
					Players []*Player `json:"players"`
				}
				room.mutexRoom.Lock()
				log.Println("room init")
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
				room.mutexRoom.Unlock()
				room.Players[0].SendMessage(msg)
				room.Players[1].SendMessage(msg2)
				room.mutexRoom.Lock()
				for _, player := range room.Players {
					go Engine(player)
				}
				room.mutexRoom.Unlock()
				<-room.finish
				room.game.RemoveRoom(room)
			}()

		}
	}
}

func (room *Room) AddPlayer(player *Player) {
	player.room = room
	room.register <- player
}

func RemovePlayer(player *Player) {

	log.Printf("id deleting player: %d\n", player.IdP)
	player.room.unregister <- player

	log.Println("Player was removed!")
}

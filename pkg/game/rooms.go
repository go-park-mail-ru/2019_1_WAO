package game

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/spf13/viper"
)

// type RoomController struct{}

type Room struct {
	ID                   string
	game                 *Game
	MaxPlayers           int
	Players              sync.Map
	Blocks               []*Block
	register             chan *Player
	unregister           chan *Player
	canvasControllerDone chan struct{}
	init                 chan struct{}
	finish               chan struct{}
	mutexRoom            *sync.Mutex
	mutexEngine          *sync.Mutex
	scrollCount          int
	isStarted            bool
	// scroller             *Player
}

func NewRoom(maxPlayers int, game *Game) *Room {
	return &Room{
		MaxPlayers:           maxPlayers,
		game:                 game,
		register:             make(chan *Player),
		unregister:           make(chan *Player),
		init:                 make(chan struct{}, 1),
		finish:               make(chan struct{}, 1),
		canvasControllerDone: make(chan struct{}, 1),
		mutexRoom:            &sync.Mutex{},
		mutexEngine:          &sync.Mutex{},
		scrollCount:          0,
		isStarted:            false,
		// scroller:             nil,
	}
}
func length(m *sync.Map) int {
	counter := 0
	m.Range(func(_, _ interface{}) bool {
		counter++
		return true
	})
	return counter
}

func (room *Room) Run() {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Error at room.Run was occured (at room logic)", e)
		}
	}()
	log.Println("Room loop started")
	initFinish := make(chan struct{}, 1)
	for {
		select {
		case <-room.finish:
			initFinish <- struct{}{}
			room.game.RemoveRoom(room)
		case player := <-room.unregister:
			log.Println("Unregistering...")
			room.Players.Delete(player.IdP)
			log.Printf("Player %d was remoted from room\n", player.IdP)
			log.Printf("Count of players: %d\n", length(&room.Players))

			if length(&room.Players) == 0 {
				room.finish <- struct{}{}
			}
		case player := <-room.register:
			player.IdP = length(&room.Players)
			room.Players.Store(player.IdP, player)
			log.Printf("Player %d added to game\n", player.IdP)
			log.Printf("Count of players: %d, MaxPlayersCount: %d\n", length(&room.Players), room.MaxPlayers)
			if length(&room.Players) == room.MaxPlayers {
				room.init <- struct{}{}
			}
		case <-room.init:
			go func() {

				type BlocksAndPlayers struct {
					Blocks  []*Block  `json:"blocks"`
					Players []*Player `json:"players"`
				}
				room.mutexRoom.Lock()
				room.isStarted = true
				log.Println("room init")
				var HeightField float64 = viper.GetFloat64("canvas.heightField")
				var spacing float64 = viper.GetFloat64("settings.spacing")
				var koefGeneratePlates float64 = viper.GetFloat64("settings.koefGeneratePlates")
				var koefHeightOfMaxGenerateSlice int = viper.GetInt("settings.koefHeightOfMaxGenerateSlice")
				room.Blocks = FieldGenerator(HeightField-spacing, float64(koefHeightOfMaxGenerateSlice), uint16(float64(koefHeightOfMaxGenerateSlice)*koefGeneratePlates))
				var players []*Player
				log.Printf("len_blocks: %d", len(room.Blocks))
				room.Players.Range(func(_, p interface{}) bool {
					players = append(players, p.(*Player))
					p.(*Player).SetPlayerOnPlate(room.Blocks[0])
					return true
				})
				blocksAndPlayers := BlocksAndPlayers{
					Blocks:  room.Blocks,
					Players: players,
				}
				if viper.ConfigFileUsed() == "./config/test.yml" {
					room.mutexRoom.Unlock()
					return
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
					room.mutexRoom.Unlock()
					return
				}
				msg2 := &Message{
					Type:    "init",
					Payload: payload2,
				}
				room.mutexRoom.Unlock()
				var p0, p1 interface{}
				var ok bool
				if p0, ok = room.Players.Load(0); !ok {
					log.Println("Player id 0 was not found!")
				}
				p0.(*Player).SendMessage(msg)
				if p1, ok = room.Players.Load(1); !ok {
					log.Println("Player id 1 was not found!")
				}
				p1.(*Player).SendMessage(msg2)
				room.mutexRoom.Lock()
				room.Players.Range(func(_, player interface{}) bool {
					go Engine(player.(*Player))
					return true
				})
				room.mutexRoom.Unlock()
				<-initFinish
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
	player.messagesClose <- struct{}{}
	player.room.unregister <- player
	log.Println("Player was removed!")
}

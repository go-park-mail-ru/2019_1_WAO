package game

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

var Blocks []*Block
var Commands []*Command
var Players []*Player

var widthField float64 = 400
var heightField float64 = 700
var gravity float64 = 0.0004

func GetParams() (float64, uint16) {
	return (heightField - 20) - 20, 5
}

func FieldGenerator(beginY float64, b float64, k uint16) (newBlocks []*Block) {
	// beginY was sended as the parameter
	p := b / float64(k) // Плотность
	r := rand.New(rand.NewSource(99))
	var currentX float64
	currentY := beginY
	var i uint16
	for i = 0; i < k; i++ {
		currentX = r.Float64()*(widthField-91) + 1.0
		newBlocks = append(newBlocks, &Block{
			X: currentX,
			Y: currentY,
			w: 90,
			h: 15,
		})
		currentY -= p
	}
	// blocks = append(blocks, newBlocks)
	// for _, block := range newBlocks {
	// 	Blocks = append(Blocks, block)
	// }
	return
}

// Scroll down the whole map

func ScrollMap(delay float64) {
	for _, block := range Blocks {
		block.Y += block.Vy * delay
	}
}

// Функция изменения скорости

func ProcessSpeed(command *Command) {
	player := FoundPlayer(command.IdP)
	player.Vy += (gravity * command.Delay)
}

// Отрисовка по кругу

func CircleDraw() {
	for _, player := range Players {
		if player.X > widthField {
			player.X = 0
		} else if player.X < 0 {
			player.X = widthField
		}
	}
}

func Collision(command *Command) {
	var player *Player = FoundPlayer(command.IdP)
	var plate *Block = player.SelectNearestBlock()
	if plate == nil {
		return
	}
	if player.Vy >= 0 {
		if player.Y+player.Vy*command.Delay < plate.Y-15 {
			return
		}
		player.Y = plate.Y - 15
		player.Jump()
	}
}

func Engine(player *Player, wg *sync.WaitGroup) {
	defer wg.Done()
	CircleDraw()
	queue := player.queue
	log.Printf("Len of queue for player with *ID%d*: %d", player.Id, queue.Count)
	for i := 0; i < queue.Count; i++ {
		command := queue.Pop()
		if command == nil {
			fmt.Println("Command's error was occured")
			return
		}
		// player = FoundPlayer(command.IdP)
		if command.Direction == "LEFT" {
			player.X -= player.Vx * command.Delay
		} else if command.Direction == "RIGHT" {
			player.X += player.Vx * command.Delay
		}
		ProcessSpeed(command)
		Collision(command)
	}
	if Blocks[0].Vy != 0 {
		ScrollMap(Commands[0].Delay)
	}
	for _, command := range Commands {
		player.Y += (player.Vy * command.Delay)
	}
	// return this.state;
}

// Main gameloop for each player

type Connections map[*Player]*websocket.Conn

// func GameLoop(connections *Connections) {

// 	for {
// 		var wg sync.WaitGroup
// 		for player, _ := range *connections {
// 			wg.Add(1)
// 			go Engine(player, &wg)
// 		}
// 		wg.Wait()
// 	}
// }

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
			if len(room.Players) < room.MaxPlayers {
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
	log.Printf("Player %d queued to add", player.Id)
	g.register <- player
}

func (g *Game) AddRoom(room *Room) {
	g.rooms = append(g.rooms, room)
}

package game

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var WidthField float64 = 400
var HeightField float64 = 700

var randomGame *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano())) // Randomizer initialize
// var koefHeightOfMaxGenerateSlice float64 = 2000
var gravity float64 = 0.0004

var koefScrollSpeed float64 = 0.5 // Скорость с которой все объекты будут падать вниз
// this.state = true;
// this.stateScrollMap = false;  // Нужен для отслеживания другими классами состояния скроллинга
// this.stateGenerateNewMap = false; // Нужен для отслеживания другими классами момента когда надо добавить к своей карте вновь сгенерированный кусок this.state.newPlates
// Настройки генерации карты
var koefGeneratePlates float64 = 0.01
var koefHeightOfMaxGenerateSlice int = 2000

var leftIndent float64 = 91
var rightIndent float64 = 91

// this.idPhysicBlockCounter = 0;  // Уникальный идентификатор нужен для отрисовки новых объектов

func GetParams() (float64, uint16) {
	return (HeightField - 20) - 20, 5
}

func FieldGenerator(beginY float64, b float64, k uint16) (newBlocks []*Block) {
	// beginY was sended as the parameter
	p := b / float64(k) // Плотность
	var currentX float64
	currentY := beginY
	var i uint16
	for i = 0; i < k; i++ {
		currentX = randomGame.Float64() * (((WidthField - rightIndent) - leftIndent + 1) + leftIndent)
		newBlocks = append(newBlocks, &Block{
			X:  currentX,
			Y:  currentY,
			Dy: 0,
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

func ScrollMap(delay float64, room *Room) {
	for _, block := range room.Blocks {
		block.Y += block.Dy * delay
	}
}

// Функция изменения скорости

func ProcessSpeed(delay float64, player *Player) {
	player.Dy += (gravity * delay)
}

// Отрисовка по кругу

func CircleDraw(player *Player) {
	if player.X > WidthField {
		player.X = 0
	} else if player.X < 0 {
		player.X = WidthField
	}
}

func Collision(delay float64, player *Player) {
	var plate *Block = player.SelectNearestBlock()
	if plate == nil {
		return
	}
	if player.Dy >= 0 {
		if player.Y+player.Dy*delay < plate.Y-15 {
			return
		}
		player.Y = plate.Y - 15
		player.Jump()
	}
}

func Engine(player *Player) {
	// defer wg.Done()
	for {
		select {
		case <-player.engineDone:
			return
		default:
			CircleDraw(player)
			select {
			case command := <-player.commands:
				// command := commands[i]
				if command == nil {
					fmt.Println("Command's error was occured")
					return
				}
				// player = FoundPlayer(command.IdP)
				fmt.Println("Command was catched")
				if command.Direction == "LEFT" {
					player.X -= player.Dx * command.Delay

				} else if command.Direction == "RIGHT" {
					player.X += player.Dx * command.Delay
				}
				ProcessSpeed(command.Delay, player)
				Collision(command.Delay, player)
				player.Y += (player.Dy * command.Delay)
			}
			log.Printf("*Player* id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", player.IdP, player.X, player.Y, player.Dx, player.Dy)
		}
	}
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

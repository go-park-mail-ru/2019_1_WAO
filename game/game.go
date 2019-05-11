package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

var WidthField float64 = 400
var HeightField float64 = 700

var maxScrollHeight float64 = 0.25 * HeightField
var minScrollHeight float64 = 0.25 * HeightField

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

func FieldGenerator(beginY float64, b float64, k uint16) (newBlocks []*Block) {
	// beginY was sended as the parameter
	p := b / float64(k) // Плотность
	var currentX float64
	currentY := beginY
	var i uint16
	for i = 0; i < k; i++ {
		currentX = randomGame.Float64()*((WidthField-rightIndent)-leftIndent+1) + leftIndent
		newBlocks = append(newBlocks, &Block{
			X:  currentX,
			Y:  currentY,
			Dy: 0,
			w:  90,
			h:  15,
		})
		currentY -= p
	}
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
	player.room.mutex.Lock()
	player.Dy += (gravity * delay)
	player.room.mutex.Unlock()
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
		if player.Y+player.Dy*delay < plate.Y-plate.h {
			return
		}
		player.room.mutex.Lock()
		player.Y = plate.Y - plate.h
		player.room.mutex.Unlock()
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
			room := player.room
			if player.Y-player.canvas.y <= maxScrollHeight && !player.room.stateScrollMap {
				room.stateScrollMap = true // Сигнал запрещающий выполнять этот код еще раз пока не выполнится else
				player.canvas.dy = -koefScrollSpeed
				// Send new map to players

				lastBlock := room.Blocks[len(room.Blocks)-1]
				beginY := lastBlock.Y - 20
				b := float64(koefHeightOfMaxGenerateSlice) + lastBlock.Y
				k := uint16(koefGeneratePlates * (float64(koefHeightOfMaxGenerateSlice) + lastBlock.Y))
				newBlocks := FieldGenerator(beginY, b, k)
				room.Blocks = append(room.Blocks, newBlocks...)

				for _, playerWithCanvas := range room.Players {
					var players []Player
					for _, player := range room.Players {
						player.room.mutex.Lock()
						playerCopy := *player
						player.room.mutex.Unlock()
						playerCopy.Y -= playerWithCanvas.canvas.y
						players = append(players, playerCopy)

						buffer, err := json.Marshal(struct {
							Blocks  []*Block `json:"blocks"`
							Players []Player `json:"players"`
						}{
							Blocks:  newBlocks,
							Players: players,
						})
						if err != nil {
							fmt.Println("Error encoding new blocks", err)
							return
						}
						playerWithCanvas.SendMessage(&Message{
							Type:    "map",
							Payload: buffer,
						})
					}
				}
			} else if player.Y-player.canvas.y >= minScrollHeight && room.stateScrollMap {
				room.stateScrollMap = false // Scrolling was finished
				player.canvas.dy = 0
			}
			CircleDraw(player)
			select {
			case command := <-player.commands:
				if command == nil {
					fmt.Println("Command's error was occured")
					return
				}
				player.room.mutex.Lock()
				if command.Direction == "LEFT" {
					player.X -= player.Dx * command.Delay
				} else if command.Direction == "RIGHT" {
					player.X += player.Dx * command.Delay
				}
				player.room.mutex.Unlock()
				ProcessSpeed(command.Delay, player)
				Collision(command.Delay, player)
				player.room.mutex.Lock()
				player.Y += (player.Dy * command.Delay)
				player.canvas.y += player.canvas.dy * command.Delay
				player.room.mutex.Unlock()
			}
			log.Printf("*Player* id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", player.IdP, player.X, player.Y, player.Dx, player.Dy)
		}
	}
}

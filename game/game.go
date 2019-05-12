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
var minScrollHeight float64 = 0.75 * HeightField

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

// Функция изменения скорости

func ProcessSpeed(delay float64, player *Player) {
	player.room.mutex.Lock()
	player.Dy += (gravity * delay)
	player.room.mutex.Unlock()
}

// Отрисовка по кругу

func CircleDraw(player *Player) {
	if player.X > WidthField {
		player.room.mutex.Lock()
		player.X = 0
		player.room.mutex.Unlock()
	} else if player.X < 0 {
		player.room.mutex.Lock()
		player.X = WidthField
		player.room.mutex.Unlock()
	}
}

func Collision(delay float64, player *Player) {
	var plate *Block = player.SelectNearestBlock()
	if plate == nil {
		log.Println("************ Plate is nil ************")
		// panic("AAA")
		return
	}
	if player.Dy >= 0 {
		if player.Y+player.Dy*delay < plate.Y-plate.h {
			// fmt.Println("Player is not on a plate")
			return
		}
		player.room.mutex.Lock()
		player.Y = plate.Y - plate.h
		// fmt.Println("******** COLLISION WAS OCCURED ********")
		player.room.mutex.Unlock()
		player.Jump()
	}
}

func (canvas *Canvas) BlocksToAnotherCanvas(blocks []*Block) []*Block {
	var newBlocks []*Block
	for _, block := range blocks {
		blockCopy := *block
		blockCopy.Y -= canvas.y
		newBlocks = append(newBlocks, &blockCopy)
	}
	return newBlocks
}
func (room *Room) HighestPlayer() (Player *Player) {
	players := room.Players
	maxYPlayer := players[0]
	maxY := maxYPlayer.Y
	for i := 1; i < len(players); i++ {
		if players[i].Y < maxY {
			room.mutex.Lock()
			maxY = players[i].Y
			maxYPlayer = players[i]
			room.mutex.Unlock()
		}
	}
	return maxYPlayer
}
func Engine(player *Player) {
	// defer wg.Done()
	for {
		select {
		case <-player.engineDone:
			return
		default:

			if player.Y-player.canvas.y <= maxScrollHeight && player.stateScrollMap == false {

				player.room.mutex.Lock()

				player.stateScrollMap = true // Сигнал запрещающий выполнять этот код еще раз пока не выполнится else
				player.canvas.dy = -koefScrollSpeed
				log.Printf("Canvas with player id%d is moving...\n", player.IdP)
				player.room.mutex.Unlock()
				if player == player.room.HighestPlayer() {
					player.room.mutex.Lock()
					player.room.scroller = player
					player.room.scrollCount++
					player.room.mutex.Unlock()
					player.room.mutex.Lock()
					log.Println("Map scrolling is starting...")
					fmt.Printf("Count of scrolling: %d\n", player.room.scrollCount)
					fmt.Println("Players:")
					player.room.mutex.Unlock()
					for _, plr := range player.room.Players {
						player.room.mutex.Lock()
						fmt.Printf("id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", plr.IdP, plr.X, plr.Y, plr.Dx, plr.Dy)
						fmt.Printf("Canvas for id%d y: %f, dy: %f\n", plr.IdP, plr.canvas.y, plr.canvas.dy)
						player.room.mutex.Unlock()
					}
					// Send new map to players
					lastBlock := player.room.Blocks[len(player.room.Blocks)-1]
					beginY := lastBlock.Y - 20
					b := float64(koefHeightOfMaxGenerateSlice) + (lastBlock.Y - player.canvas.y)
					k := uint16(koefGeneratePlates * (float64(koefHeightOfMaxGenerateSlice) + (lastBlock.Y - player.canvas.y)))
					newBlocks := FieldGenerator(beginY, b, k)
					player.room.mutex.Lock()
					player.room.Blocks = append(player.room.Blocks, newBlocks...)
					player.room.mutex.Unlock()
					var buffer []byte
					var err error

					for _, playerWithCanvas := range player.room.Players {
						var players []*Player
						for _, player := range player.room.Players {
							player.room.mutex.Lock()
							playerCopy := *player

							playerCopy.Y -= playerWithCanvas.canvas.y
							player.room.mutex.Unlock()
							player.room.mutex.Lock()
							players = append(players, &playerCopy)
							player.room.mutex.Unlock()
						}
						newBlocksForPlayer := playerWithCanvas.canvas.BlocksToAnotherCanvas(newBlocks)
						if buffer, err = json.Marshal(struct {
							Blocks  []*Block  `json:"blocks"`
							Players []*Player `json:"players"`
						}{
							Blocks:  newBlocksForPlayer,
							Players: players,
						}); err != nil {
							fmt.Println("Error encoding new blocks", err)
							return
						}
						playerWithCanvas.SendMessage(&Message{
							Type:    "map",
							Payload: buffer,
						})
						log.Printf("New blocks for id %d:\n", playerWithCanvas.IdP)
						for _, block := range newBlocksForPlayer {
							player.room.mutex.Lock()
							fmt.Printf("x: %f, y: %f, w: %f, h: %f\n", block.X, block.Y, block.w, block.h)
							player.room.mutex.Unlock()
						}
					}
					player.room.mutex.Lock()
					log.Println("******* MAP WAS SENDED *******")
					log.Println("New blocks:")
					player.room.mutex.Unlock()
					for _, block := range newBlocks {
						player.room.mutex.Lock()
						fmt.Printf("x: %f, y: %f, w: %f, h: %f\n", block.X, block.Y, block.w, block.h)
						player.room.mutex.Unlock()
					}

				}
			} else if player.Y-player.canvas.y >= minScrollHeight && player.stateScrollMap == true {

				player.room.mutex.Lock()
				player.canvas.dy = 0
				log.Printf("Canvas with player id%d was stopped...\n", player.IdP)
				player.stateScrollMap = false // Scrolling was finished
				player.room.mutex.Unlock()
				if player.room.scroller == player {
					player.room.mutex.Lock()
					player.room.scroller = nil
					player.room.mutex.Unlock()
				}
				// player.room.mutex.Lock()
				// log.Println("Map scrolling is finishing...")
				// fmt.Printf("Count of scrolling: %d\n", player.room.scrollCount)
				// fmt.Println("Players:")
				// player.room.mutex.Unlock()
				// for _, plr := range player.room.Players {
				// 	player.room.mutex.Lock()
				// 	// fmt.Printf("id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", plr.IdP, plr.X, plr.Y, plr.Dx, plr.Dy)
				// 	// fmt.Printf("Canva for id%d y: %f, dy: %f\n", plr.IdP, plr.canvas.y, plr.canvas.dy)
				// 	player.room.mutex.Unlock()
				// }
			}
			CircleDraw(player)
			select {
			case command := <-player.commands:
				if command == nil {
					fmt.Println("Command's error was occured")
					return
				}
				if command.Direction == "LEFT" {
					player.room.mutex.Lock()
					player.X -= player.Dx * command.Delay
					player.room.mutex.Unlock()
				} else if command.Direction == "RIGHT" {
					player.room.mutex.Lock()
					player.X += player.Dx * command.Delay
					player.room.mutex.Unlock()
				}
				ProcessSpeed(command.Delay, player)
				Collision(command.Delay, player)
				player.room.mutex.Lock()
				player.Y += (player.Dy * command.Delay)
				player.canvas.y += player.canvas.dy * command.Delay
				player.room.mutex.Unlock()
			}
			if player.Dy > 1 {
				player.room.mutex.Lock()
				fmt.Println("Blocks:")
				player.room.mutex.Unlock()
				for _, block := range player.room.Blocks {
					player.room.mutex.Lock()
					fmt.Printf("Block x: %f, y: %f\n", block.X, block.Y)
					player.room.mutex.Unlock()
				}
				fmt.Println("Players:")
				for _, plr := range player.room.Players {
					player.room.mutex.Lock()
					fmt.Printf("id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", plr.IdP, plr.X, plr.Y, plr.Dx, plr.Dy)
					fmt.Printf("Canva for id%d y: %f, dy: %f\n", plr.IdP, plr.canvas.y, plr.canvas.dy)
					player.room.mutex.Unlock()
				}
				// player.connection.Close()
				panic("Dy >>>>>")
			}
			//

			// log.Printf("*Player* id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", player.IdP, player.X, player.Y, player.Dx, player.Dy)
		}
	}
}

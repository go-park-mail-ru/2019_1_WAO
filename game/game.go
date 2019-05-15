package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
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

// func Collision(delay float64, player *Player) {
//     // if (command.idP !== 0) {
//     //   alert("1");
//     // }
//     var plate *Block = player.SelectNearestBlock(&player.room.Blocks)
// 	if plate == nil {
// 		log.Printf("* Plate is nil * for player id%d", player.IdP)
// 		return
// 	}
//     if (player.Dy >= 0 && !player.StateScrollMap || player.Dy - koefScrollSpeed >= 0 && player.stateScrollMap) {

//       if ((player.Dy >= 0 && !player.StateScrollMap) && (player.Y + player.Dy * command.Delay < plate.Y - plate.h) || (player.Dy - koefScrollSpeed >= 0 && player.StateScrollMap) && (player.Y + (player.Dy - koefScrollSpeed) * command.Delay < plate.Y - plate.)) {
//         return
//       }
//       // Если был включен счетчик очков, то передать в него пластину от которой будем прыгать
//     //   if (this.score) {
//     //     this.score.giveCurrentPlate(plate);
//     //   }
//       player.Y = plate.Y - plate.h
// 		// fmt.Println("******** COLLISION WAS OCCURED ********")
// 		player.Jump()
//     }
//   }

func Collision(delay float64, player *Player) {
	var plate *Block = player.SelectNearestBlock(&player.room.Blocks)
	if plate == nil {
		log.Printf("* Plate is nil * for player id%d", player.IdP)
		return
	}
	if player.Dy >= 0 {
		if player.Y+player.Dy*delay < plate.Y-plate.h {
			// fmt.Println("Player is not on a plate")
			return
		}
		player.Y = plate.Y - plate.h
		// fmt.Println("******** COLLISION WAS OCCURED ********")
		player.Jump()
	}
}

func (player *Player) BlocksToAnotherCanvas(blocks []*Block, b float64) []*Block {
	var newBlocks []*Block
	for _, block := range blocks {
		blockCopy := *block

		// blockCopy.Y = blockCopy.Y - (blocks[0].Y - b) + canvas.y
		highestPlayer := player.room.HighestPlayer()
		blockCopy.Y = blocks[0].Y - (2000 + blocks[0].Y - highestPlayer.canvas.y) - highestPlayer.Y
		newBlocks = append(newBlocks, &blockCopy)
	}
	return newBlocks
}

// Virtual transfer player to anotherPlayer's canvas
func (player *Player) playerToAnotherCanvas(anotherPlayer *Player) *Player {
	player.room.mutexEngine.Lock()
	playerCopy := *player
	playerCopy.Y += (anotherPlayer.canvas.y - player.canvas.y)
	player.room.mutexEngine.Unlock()
	return &playerCopy
}

func (room *Room) AllPlayersToAnotherCanvas(player *Player) []*Player {
	var players []*Player
	room.Players.Range(func(_, plr interface{}) bool { // plr - a current player
		players = append(players, plr.(*Player).playerToAnotherCanvas(player))
		return true
	})
	return players
}

func (room *Room) HighestPlayer() *Player {
	var maxYPlayer *Player = nil
	maxY := math.MaxFloat64
	room.Players.Range(func(_, player interface{}) bool {
		if player.(*Player).Y < maxY {
			maxY = player.(*Player).Y
			maxYPlayer = player.(*Player)
		}
		return true
	})
	return maxYPlayer
}
func Engine(player *Player) {
	for {
		select {
		case <-player.engineDone:
			return
		default:
			if player.Y-player.canvas.y <= maxScrollHeight && player.stateScrollMap == false {
				player.stateScrollMap = true // Сигнал запрещающий выполнять этот код еще раз пока не выполнится else
				player.canvas.dy = -koefScrollSpeed
				log.Printf("Canvas with player id%d is moving...\n", player.IdP)
				if player == player.room.HighestPlayer() {
					player.room.mutexEngine.Lock()
					player.room.scrollCount++
					player.room.mutexEngine.Unlock()
					log.Println("Map scrolling is starting...")
					fmt.Printf("Count of scrolling: %d\n", player.room.scrollCount)
					fmt.Println("Players:")
					player.room.Players.Range(func(_, plr interface{}) bool {
						fmt.Printf("id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", plr.(*Player).IdP, plr.(*Player).X, plr.(*Player).Y, plr.(*Player).Dx, plr.(*Player).Dy)
						fmt.Printf("Canvas for id%d y: %f, dy: %f\n", plr.(*Player).IdP, plr.(*Player).canvas.y, plr.(*Player).canvas.dy)
						return true
					})
					// Send new map to players
					lastBlock := player.room.Blocks[len(player.room.Blocks)-1]
					beginY := lastBlock.Y - 20
					b := float64(koefHeightOfMaxGenerateSlice) + (lastBlock.Y - player.canvas.y)
					k := uint16(koefGeneratePlates * (float64(koefHeightOfMaxGenerateSlice) + (lastBlock.Y - player.canvas.y)))
					newBlocks := FieldGenerator(beginY, b, k)
					player.room.mutexEngine.Lock()
					player.room.Blocks = append(player.room.Blocks, newBlocks...)
					player.room.mutexEngine.Unlock()
					var buffer []byte
					var err error

					player.room.Players.Range(func(_, playerWithCanvas interface{}) bool {
						var players []*Player
						player.room.Players.Range(func(_, player interface{}) bool {
							playerCopy := player.(*Player).playerToAnotherCanvas(playerWithCanvas.(*Player))
							players = append(players, playerCopy)
							return true
						})
						newBlocksForPlayer := playerWithCanvas.(*Player).BlocksToAnotherCanvas(newBlocks, b)
						if buffer, err = json.Marshal(struct {
							Blocks  []*Block  `json:"blocks"`
							Players []*Player `json:"players"`
						}{
							Blocks:  newBlocksForPlayer,
							Players: players,
						}); err != nil {
							fmt.Println("Error encoding new blocks", err)
							return false // ?
						}
						playerWithCanvas.(*Player).SendMessage(&Message{
							Type:    "map",
							Payload: buffer,
						})
						log.Printf("New blocks for id %d:\n", playerWithCanvas.(*Player).IdP)
						for _, block := range newBlocksForPlayer {
							fmt.Printf("x: %f, y: %f, w: %f, h: %f\n", block.X, block.Y, block.w, block.h)
						}
						return true
					})
					log.Println("******* MAP WAS SENDED *******")
					log.Println("New blocks:")
					for _, block := range newBlocks {
						fmt.Printf("x: %f, y: %f, w: %f, h: %f\n", block.X, block.Y, block.w, block.h)
					}

				}
			} else if player.Y-player.canvas.y >= minScrollHeight && player.stateScrollMap == true {
				player.stateScrollMap = false // Scrolling was finished
				player.canvas.dy = 0
				log.Printf("Canvas with player id%d was stopped...\n", player.IdP)
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
					continue
				}
				if player.commandCounter == 10 {
					fmt.Println("For Player id", player.IdP)

					players := player.room.AllPlayersToAnotherCanvas(player)
					player.room.mutexEngine.Lock()
					for _, plr := range players {
						fmt.Printf("id: %d, x: %f, y: %f, dy: %f\n", plr.IdP, plr.X, plr.Y, plr.Dy)
					}
					buf, err := json.Marshal(players)
					if err != nil {
						log.Println("Error players to encoding")
						player.room.mutexEngine.Unlock()
						continue
					}

					player.SendMessage(&Message{
						Type:    "updatePositions",
						Payload: buf,
					})
					player.commandCounter = 0
					player.room.mutexEngine.Unlock()
				} else {
					player.room.mutexEngine.Lock()
					player.commandCounter++
					player.room.mutexEngine.Unlock()
				}
				if command.Direction == "LEFT" {
					player.room.mutexEngine.Lock()
					player.X -= player.Dx * command.Delay
					player.room.mutexEngine.Unlock()
				} else if command.Direction == "RIGHT" {
					player.room.mutexEngine.Lock()
					player.X += player.Dx * command.Delay
					player.room.mutexEngine.Unlock()
				}
				ProcessSpeed(command.Delay, player)
				Collision(command.Delay, player)
				player.room.mutexEngine.Lock()
				player.Y += (player.Dy * command.Delay)
				player.canvas.y += player.canvas.dy * command.Delay
				player.room.mutexEngine.Unlock()
			}
			if player.Dy > 1 {
				fmt.Println("Blocks:")
				for _, block := range player.room.Blocks {
					fmt.Printf("Block x: %f, y: %f\n", block.X, block.Y)
				}
				fmt.Println("Players:")
				player.room.Players.Range(func(_, plr interface{}) bool {
					fmt.Printf("id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", plr.(*Player).IdP, plr.(*Player).X, plr.(*Player).Y, plr.(*Player).Dx, plr.(*Player).Dy)
					fmt.Printf("Canvas for id%d y: %f, dy: %f\n", plr.(*Player).IdP, plr.(*Player).canvas.y, plr.(*Player).canvas.dy)
					return true
				})
				// panic("Dy >>>>>")
			}
			// for logss
			// log.Printf("*Player* id%d	-	x: %f, y: %f, yC: %f, Dy: %f\n", player.IdP, player.X, player.Y, player.Y-player.canvas.y, player.Dy)
		}
	}
}

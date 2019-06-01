package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/spf13/viper"
)

var WidthField float64 = viper.GetFloat64("canvas.widthField")
var HeightField float64 = viper.GetFloat64("canvas.heightField")

var maxScrollHeight float64 = 0.25 * HeightField
var minScrollHeight float64 = 0.75 * HeightField

var randomGame *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano())) // Randomizer initialize
var gravity float64 = viper.GetFloat64("settings.gravity")
var spacing float64 = viper.GetFloat64("settings.spacing")

var koefScrollSpeed float64 = viper.GetFloat64("settings.koefScrollSpeed") // Скорость с которой все объекты будут падать вниз
// this.state = true;
// this.stateScrollMap = false;  // Нужен для отслеживания другими классами состояния скроллинга
// this.stateGenerateNewMap = false; // Нужен для отслеживания другими классами момента когда надо добавить к своей карте вновь сгенерированный кусок this.state.newPlates
// Настройки генерации карты
var koefGeneratePlates float64 = viper.GetFloat64("settings.koefGeneratePlates")
var koefHeightOfMaxGenerateSlice int = viper.GetInt("settings.koefHeightOfMaxGenerateSlice")

var leftIndent float64 = viper.GetFloat64("settings.leftIndent")
var rightIndent float64 = viper.GetFloat64("settings.rightIndent")

// this.idPhysicBlockCounter = 0;  // Уникальный идентификатор нужен для отрисовки новых объектов

func CheckDevelopmentEnvironment() bool {
	if viper.ConfigFileUsed() != "../config/test.yml" {
		return true
	}
	return false
}

func FieldGenerator(beginY float64, b float64, k uint16) (newBlocks []*Block) {
	var WidthField float64 = viper.GetFloat64("canvas.widthField")
	var leftIndent float64 = viper.GetFloat64("settings.leftIndent")
	var rightIndent float64 = viper.GetFloat64("settings.rightIndent")
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
			w:  viper.GetFloat64("block.width"),
			h:  viper.GetFloat64("block.height"),
		})
		currentY -= p
	}
	return
}

func (player *Player) updateScore(plate *Block) {
	if player.lastBlock != nil {
		player.AddScore(int(player.lastBlock.Y - plate.Y))
	} else {
		player.AddScore(int((viper.GetFloat64("canvas.heightField") - viper.GetFloat64("settings.spacing")) - plate.Y))
	}
	player.lastBlock = plate
	log.Printf("player id %d - score: %d\n", player.IdP, player.scoreCounter)
}

// Функция изменения скорости

func ProcessSpeed(delay float64, player *Player, gravity float64) {
	player.Dy += (gravity * delay)
}

// Отрисовка по кругу

func CircleDraw(player *Player) {
	WidthField := viper.GetFloat64("canvas.widthField")
	if player.X > WidthField {
		player.X = 0
	} else if player.X < 0 {
		player.X = WidthField
	}
}

func (player *Player) SendEndGame(typeName string) {
	payload, err := json.Marshal(struct {
		IdP   int `json:"idP"`
		Score int `json:"score"`
	}{
		IdP:   player.IdP,
		Score: player.scoreCounter,
	})
	if err != nil {
		log.Println("Error with encoding player's id and score was occured", err)
		return
	}
	player.SendMessage(&Message{
		Type:    typeName,
		Payload: payload,
	})
}
func KillPlayer(player *Player) {
	player.room.Players.Range(func(_, plr interface{}) bool {
		if plr == player {
			plr.(*Player).SendEndGame("endgame")
		} else {
			plr.(*Player).SendEndGame("lose")
		}
		return true
	})
	log.Printf("! player id %d lose !\n", player.IdP)
}

func Collision(delay float64, player *Player, plate *Block) {
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

		// Score
		if player.lastBlock == nil || player.lastBlock.Y >= plate.Y {
			player.updateScore(plate)
		}
		// fmt.Println("******** COLLISION WAS OCCURED ********")
		player.Jump()
	}
}

func (player *Player) BlocksToAnotherCanvas(blocks []*Block, b float64) []*Block {
	var newBlocks []*Block
	for _, block := range blocks {
		blockCopy := *block

		// blockCopy.Y = blockCopy.Y - (blocks[0].Y - b) + canvas.y
		// highestPlayer := player.room.HighestPlayer()
		// blockCopy.Y = blocks[0].Y - (2000 + blocks[0].Y - highestPlayer.canvas.y) - highestPlayer.Y
		blockCopy.Y = blockCopy.Y - (blocks[0].Y - b) + player.canvas.y
		newBlocks = append(newBlocks, &blockCopy)
	}
	return newBlocks
}

// Virtual transfer player to anotherPlayer's canvas
func (player *Player) playerToAnotherCanvas(anotherPlayer *Player) *Player {
	// player.room.mutexEngine.Lock()
	playerCopy := *player
	playerCopy.Y += (anotherPlayer.canvas.y - player.canvas.y)
	// player.room.mutexEngine.Unlock()
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
	room.mutexEngine.Lock()
	room.Players.Range(func(_, player interface{}) bool {
		if player.(*Player).Y < maxY {
			maxY = player.(*Player).Y
			maxYPlayer = player.(*Player)
		}
		return true
	})
	room.mutexEngine.Unlock()
	return maxYPlayer
}

func (player *Player) Move(command *Command) {
	if command.Direction == "LEFT" {
		player.X -= player.Dx * command.Delay
	} else if command.Direction == "RIGHT" {
		player.X += player.Dx * command.Delay
	}
}

func (player *Player) PlayerMoveWithGravity(delay float64) {
	player.Y += (player.Dy * delay)
}

func (canvas *Canvas) CanvasMove(delay float64) {
	canvas.y += canvas.dy * delay
}

func (player *Player) MapUpdate(lastBlock *Block) (newBlocks []*Block, b float64) {
	k := uint16(koefGeneratePlates * (float64(koefHeightOfMaxGenerateSlice) + (lastBlock.Y - player.canvas.y)))
	if k == 0 {
		newBlocks = nil
		b = 0
		return
	}
	beginY := lastBlock.Y - viper.GetFloat64("settings.spacing")
	koefHeightOfMaxGenerateSlice := viper.GetInt("settings.koefHeightOfMaxGenerateSlice")
	koefGeneratePlates := viper.GetFloat64("settings.koefGeneratePlates")
	b = float64(koefHeightOfMaxGenerateSlice) + (lastBlock.Y - player.canvas.y)
	newBlocks = FieldGenerator(beginY, b, k)
	for id, block := range newBlocks {
		fmt.Printf("block %d - x: %f, y: %f\n", id, block.X, block.Y)
	}
	// ! TO CATCH NIL !
	if len(newBlocks) == 0 {
		log.Panicf("beginY: %f, b: %f, k: %d\n lastBlock.Y: %f, player.y: %f, player.canvas.y: %f, player.canvas.dy: %f\nkoefHeightOfMaxGenerateSlice: %d, koefGeneratePlates: %f\n", beginY, b, k, lastBlock.Y, player.Y,
			player.canvas.y, player.canvas.dy, koefHeightOfMaxGenerateSlice, koefGeneratePlates)
	}
	return
}

func (player *Player) StartScrolling() {
	player.stateScrollMap = true // Сигнал запрещающий выполнять этот код еще раз пока не выполнится else
	player.canvas.dy = -viper.GetFloat64("settings.koefScrollSpeed")
}

func (player *Player) StopScrolling() {
	player.stateScrollMap = false // Scrolling was finished
	player.canvas.dy = 0
}

func Engine(player *Player) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Error at physic treatment was occured (function Engine)", e)
			// if CheckDevelopmentEnvironment() {
			// 	KillPlayer(player)
			// }
			RemovePlayer(player)
		}
	}()
	var maxCountOfCommands uint64 = viper.GetUint64("player.maxCountOfCommands")
	var HeightField float64 = viper.GetFloat64("canvas.heightField")
	var gravity float64 = viper.GetFloat64("settings.gravity")
	var maxScrollHeight float64 = 0.25 * HeightField
	var minScrollHeight float64 = 0.75 * HeightField
	for {
		select {
		case <-player.engineDone:
			return
		default:
			if player.Y-player.H > player.canvas.y+HeightField {
				// log.Printf("Player with id %d lose!\n", player.IdP)
				// if CheckDevelopmentEnvironment() {
				// 	KillPlayer(player)
				// }
				RemovePlayer(player)
				return
			}
			if player.Y-player.canvas.y <= maxScrollHeight && player.stateScrollMap == false {
				player.room.mutexEngine.Lock()
				player.StartScrolling()
				player.room.mutexEngine.Unlock()
				log.Printf("Canvas with player id%d is moving...\n", player.IdP)
				if player == player.room.HighestPlayer() {
					player.room.mutexEngine.Lock()
					player.room.scrollCount++
					player.room.mutexEngine.Unlock()
					log.Println("Map scrolling is starting...")
					// fmt.Printf("Count of scrolling: %d\n", player.room.scrollCount)
					fmt.Println("Players:")
					player.room.Players.Range(func(_, plr interface{}) bool {
						fmt.Printf("id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", plr.(*Player).IdP, plr.(*Player).X, plr.(*Player).Y, plr.(*Player).Dx, plr.(*Player).Dy)
						fmt.Printf("Canvas for id%d y: %f, dy: %f\n", plr.(*Player).IdP, plr.(*Player).canvas.y, plr.(*Player).canvas.dy)
						return true
					})
					// Send new map to players
					player.room.mutexEngine.Lock()
					lastBlock := player.room.Blocks[len(player.room.Blocks)-1]
					player.room.mutexEngine.Unlock()
					newBlocks, b := player.MapUpdate(lastBlock)
					if newBlocks != nil {
						player.room.mutexEngine.Lock()
						player.room.Blocks = append(player.room.Blocks, newBlocks...)
						player.room.mutexEngine.Unlock()
						var buffer []byte
						var err error
						player.room.mutexEngine.Lock()
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
								player.room.mutexEngine.Unlock()
								return false // ?
							}
							if CheckDevelopmentEnvironment() {
								playerWithCanvas.(*Player).SendMessage(&Message{
									Type:    "map",
									Payload: buffer,
								})
							}
							// log.Printf("New blocks for id %d:\n", playerWithCanvas.(*Player).IdP)
							// for _, block := range newBlocksForPlayer {
							// 	fmt.Printf("x: %f, y: %f, w: %f, h: %f\n", block.X, block.Y, block.w, block.h)
							// }
							return true
						})
						player.room.mutexEngine.Unlock()
						log.Println("******* MAP WAS SENDED *******")
						log.Println("New blocks:")
						for _, block := range newBlocks {
							fmt.Printf("x: %f, y: %f, w: %f, h: %f\n", block.X, block.Y, block.w, block.h)
						}
					} else {
						log.Println("New blocks are absent. ! k = 0 !")
					}
				}
			} else if player.Y-player.canvas.y >= minScrollHeight && player.stateScrollMap == true {
				player.room.mutexEngine.Lock()
				player.StopScrolling()
				player.room.mutexEngine.Unlock()
				log.Printf("Canvas with player id%d was stopped...\n", player.IdP)
				// player.room.mutex.Lock()
				log.Println("Map scrolling is finishing...")
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
			player.room.mutexEngine.Lock()
			CircleDraw(player)
			player.room.mutexEngine.Unlock()
			select {
			case command := <-player.commands:
				if command == nil {
					fmt.Println("Command's error was occured")
					continue
				}
				if player.commandCounter == maxCountOfCommands {
					// fmt.Println("For Player id", player.IdP)

					player.room.mutexEngine.Lock()
					players := player.room.AllPlayersToAnotherCanvas(player)
					// for _, plr := range players {
					// 	fmt.Printf("id: %d, x: %f, y: %f, dy: %f\n", plr.IdP, plr.X, plr.Y, plr.Dy)
					// }
					buf, err := json.Marshal(players)
					if err != nil {
						log.Println("Error players to encoding")
						player.room.mutexEngine.Unlock()
						continue
					}
					if CheckDevelopmentEnvironment() {
						player.SendMessage(&Message{
							Type:    "updatePositions",
							Payload: buf,
						})
					}
					player.commandCounter = 0
					player.room.mutexEngine.Unlock()
				} else {
					player.room.mutexEngine.Lock()
					player.commandCounter++
					player.room.mutexEngine.Unlock()
				}
				if command.Direction == "LEFT" || command.Direction == "RIGHT" {
					player.room.mutexEngine.Lock()
					player.Move(command)
					player.room.mutexEngine.Unlock()
				}
				player.room.mutexEngine.Lock()
				ProcessSpeed(command.Delay, player, gravity)
				Collision(command.Delay, player, player.SelectNearestBlock(&player.room.Blocks))
				player.PlayerMoveWithGravity(command.Delay)
				player.canvas.CanvasMove(command.Delay)
				player.room.mutexEngine.Unlock()
			}
			// if player.Dy > viper.GetFloat64("player.speedLimit") {
			// 	fmt.Println("Blocks:")
			// 	for _, block := range player.room.Blocks {
			// 		fmt.Printf("Block x: %f, y: %f\n", block.X, block.Y)
			// 	}
			// 	fmt.Println("Players:")
			// 	player.room.Players.Range(func(_, plr interface{}) bool {
			// 		fmt.Printf("id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", plr.(*Player).IdP, plr.(*Player).X, plr.(*Player).Y, plr.(*Player).Dx, plr.(*Player).Dy)
			// 		fmt.Printf("Canvas for id%d y: %f, dy: %f\n", plr.(*Player).IdP, plr.(*Player).canvas.y, plr.(*Player).canvas.dy)
			// 		return true
			// 	})
			// 	// panic("Dy >>>>>")
			// }
			// for logss
			// log.Printf("*Player* id%d	-	x: %f, y: %f, yC: %f, Dy: %f\n", player.IdP, player.X, player.Y, player.Y-player.canvas.y, player.Dy)
			// fmt.Println("Players:")
			// player.room.Players.Range(func(_, plr interface{}) bool {
			// 	if plr.(*Player).IdP == 0 {
			// 		fmt.Printf("id%d	-	x: %f, y: %f, Dx: %f, Dy: %f\n", plr.(*Player).IdP, plr.(*Player).X, plr.(*Player).Y, plr.(*Player).Dx, plr.(*Player).Dy)
			// 		fmt.Printf("Canvas for id%d y: %f, dy: %f\n", plr.(*Player).IdP, plr.(*Player).canvas.y, plr.(*Player).canvas.dy)
			// 	}
			// 	return true
			// })
		}
	}
}

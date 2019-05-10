package game

import (
	"encoding/json"
	"fmt"
)

type Canvas struct {
	y  float64
	dy float64
}

// CanvasController deletes old blocks (below the lowest canvas)
func (room *Room) CanvasController() {
	var canvases []*Canvas

	for _, player := range room.Players {
		canvases = append(canvases, player.canvas)
	}
	for {
		select {
		case <-room.canvasControllerDone:
			room.game.RemoveRoom(room)
			return
		default:
			minCanvas := canvases[0]
			for i := 1; i < len(canvases); i++ { // At first minCanvas is canvases[0]
				if canvases[i].y > minCanvas.y {
					minCanvas = canvases[i]
				}
			}
			// Clear old blocks
			blocks := &room.Blocks
			var survivorBlocks []*Block
			for _, block := range *blocks {
				if !(block.Y > minCanvas.y+HeightField) {
					survivorBlocks = append(survivorBlocks, block)
				}
			}
			*blocks = survivorBlocks
		}
	}
}

func (player *Player) MapPlayerListen() {
	room := player.room
	for {
		select {
		case <-player.mapPlayerListenEnd:
			return
		default:
			if player.Y <= maxScrollHeight && !room.stateScrollMap {
				room.stateScrollMap = true // Сигнал запрещающий выполнять этот код еще раз пока не выполнится else
				for _, block := range room.Blocks {
					block.Dy = koefScrollSpeed
				}
				for _, player := range room.Players {
					player.Dy += koefScrollSpeed
				}
				player.canvas.dy = koefScrollSpeed
				// Send new map to players
				room := player.room
				lastBlock := room.Blocks[len(room.Blocks)-1]
				beginY := lastBlock.Y - 20
				b := float64(koefHeightOfMaxGenerateSlice) + lastBlock.Y
				k := uint16(koefGeneratePlates * (float64(koefHeightOfMaxGenerateSlice) + lastBlock.Y))
				newBlocks := FieldGenerator(beginY, b, k)
				room.Blocks = append(room.Blocks, newBlocks...)
				var players []*Player
				for _, player := range room.Players {
					players = append(players, player)
				}
				buffer, err := json.Marshal(struct {
					Blocks  []*Block  `json:"blocks"`
					Players []*Player `json:"players"`
				}{
					Blocks:  newBlocks,
					Players: players,
				})
				if err != nil {
					fmt.Println("Error encoding new blocks", err)
					return
				}
				for _, player := range room.Players {
					player.SendMessage(&Message{
						Type:    "map",
						Payload: buffer,
					})
				}
			} else if player.Y >= minScrollHeight && room.stateScrollMap {
				room.stateScrollMap = false // Scrolling was finished
				for _, block := range room.Blocks {
					block.Dy = 0
				}
				for _, player := range room.Players {
					player.Dy -= koefScrollSpeed
				}
				player.canvas.dy = 0
			}
		}

	}
}

package game

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
			var survivorBlocks []*Block
			for _, block := range room.Blocks {
				if !(block.Y > minCanvas.y+HeightField) {
					room.mutex.Lock()
					survivorBlocks = append(survivorBlocks, block)
					room.mutex.Unlock()
				}
			}
			room.Blocks = survivorBlocks
		}
	}
}

// func (player *Player) MapPlayerListen() {
// 	room := player.room
// 	for {
// 		select {
// 		case <-player.mapPlayerListenEnd:
// 			return
// 		default:

// 	}
// }

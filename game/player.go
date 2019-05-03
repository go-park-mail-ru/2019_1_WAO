package game

import (
	"fmt"
)

type Vector struct {
	x float64
	y float64
}

type Point struct {
	x float64
	y float64
}

type Player struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	vx float64
	vy float64
	w  float64
	h  float64
}

func (player *Player) Move(vector Vector, dt float64) {
	player.X += vector.x * dt
	player.Y += vector.y * dt
	fmt.Printf("x: %f, y: %f\n", player.X, player.Y)
}

func CheckPointCollision(playerPoint, blockUpPoint, blockDownPoint Point) bool {
	if blockUpPoint.x <= playerPoint.x && playerPoint.x <= blockDownPoint.x && blockUpPoint.y <= playerPoint.y && playerPoint.y <= blockDownPoint.y {
		return true
	}
	return false
}

// The function checks all of the player points, but we need only bottom two points
// func (player *Player) CheckCollision(block Block) bool {
// 	var playerPoints []Point

// 	playerPoints = append(playerPoints, Point{player.X, player.Y}, Point{player.X + player.w, player.Y},
// 		Point{player.X, player.Y + player.h}, Point{player.X + player.w, player.Y + player.h})
// 	// We will check collisions between the block and each player's point
// 	isCollision := false
// 	blockUpPoint := Point{block.X, block.Y}
// 	blockDownPoint := Point{block.X + block.w, block.Y + block.h}
// 	for _, point := range playerPoints {
// 		if CheckPointCollision(point, blockUpPoint, blockDownPoint) {
// 			isCollision = true
// 			break
// 		}
// 	}
// 	return isCollision
// }

func (player *Player) SelectNearestBlock() (nearestBlock *Block) {
	nearestBlock = nil
	var minY float64
	for _, block := range blocks {
		if player.X+player.w >= block.X && player.X <= block.X+block.w {
			if block.Y-player.Y < minY && player.Y <= block.Y {
				minY = block.Y - player.Y
				nearestBlock = block
			}
		}
	}
	return
}

func (player *Player) Jump() {
	player.vy = -0.35 // Change a vertical speed (for jump)
}

func (player *Player) SetPlayerOnPlate(block *Block) {
	player.Y = block.Y - block.h
}

// func (player *Player) CheckCollision(block Block) bool {
// 	var playerPoints []Point

// 	playerPoints = append(playerPoints, Point{player.X, player.Y + player.h}, Point{player.X + player.w, player.Y + player.h})
// 	// We will check collisions between the block and each player's point
// 	isCollision := false
// 	blockUpPoint := Point{block.X, block.Y}
// 	blockDownPoint := Point{block.X + block.w, block.Y + block.h}
// 	for _, point := range playerPoints {
// 		if CheckPointCollision(point, blockUpPoint, blockDownPoint) {
// 			isCollision = true
// 			break
// 		}
// 	}
// 	return isCollision
// }

func (player *Player) CheckCollision(block *Block, dt float64) bool {
	// block := player.SelectNearestBlock()
	// if block {
	// 	return false
	// }
	if player.vy >= 0 { // If the collision will occur
		if player.Y+player.vy*dt < block.Y-15 {
			return true
		}
		player.SetPlayerOnPlate(block)
		player.Jump()
	}
	return false
}

func (player *Player) Gravity(g float64, dt float64) {
	player.vy += g * dt
	// player.Move(Vector{0, player.vy})
	// fmt.Printf("x: %f, y: %f\n", player.X, player.Y)
	// nearestBlock := player.SelectNearestBlock()
	// player.CheckCollision(nearestBlock, dt)
}

func (player *Player) CircleDraw() {
	if player.X > widthField {
		player.X = 0
	} else if player.X < 0 {
		player.X = widthField
	}
}

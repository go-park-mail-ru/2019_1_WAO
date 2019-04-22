package game

import (
	"fmt"
)

type Vector struct {
	x float32
	y float32
}

type Point struct {
	x float32
	y float32
}

type Player struct {
	X  float32 `json:"x"`
	Y  float32 `json:"y"`
	vx float32
	vy float32
	w  float32
	h  float32
}

func (player *Player) Move(vector Vector) {
	player.X += vector.x
	player.Y += vector.y
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

func (player *Player) CheckCollision(block Block) bool {
	var playerPoints []Point

	playerPoints = append(playerPoints, Point{player.X, player.Y + player.h}, Point{player.X + player.w, player.Y + player.h})
	// We will check collisions between the block and each player's point
	isCollision := false
	blockUpPoint := Point{block.X, block.Y}
	blockDownPoint := Point{block.X + block.w, block.Y + block.h}
	for _, point := range playerPoints {
		if CheckPointCollision(point, blockUpPoint, blockDownPoint) {
			isCollision = true
			break
		}
	}
	return isCollision
}

func (player *Player) Gravity(g float32) {
	player.vy += g
	// player.Move(Vector{0, player.vy})
	// fmt.Printf("x: %f, y: %f\n", player.X, player.Y)
}

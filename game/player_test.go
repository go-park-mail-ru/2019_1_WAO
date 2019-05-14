package game

// import (
// 	"testing"
// )

// type Moves struct {
// 	vector    Vector
// 	expectedX float32
// 	expectedY float32
// }

// var moves = []Moves{
// 	{Vector{-5, 5}, 5, 15},
// 	{Vector{-10, -10}, 0, 0},
// 	{Vector{10, 10}, 20, 20},
// 	{Vector{3.55, -6.1}, 13.55, 3.9},
// 	{Vector{0, 0}, 10, 10},
// 	{Vector{0.00, 0.00}, 10, 10},
// }

// func TestMove(t *testing.T) {

// 	for _, pair := range moves {
// 		player := Player{10, 10, 0, 0, 10, 10}
// 		player.Move(pair.vector)
// 		if player.X != pair.expectedX || player.Y != pair.expectedY {
// 			t.Error("Expected x:", pair.expectedX, "y:", pair.expectedY,
// 				"but got x:", player.X, "y:", player.Y)
// 		}
// 	}

// }

// type BlockForCollision struct {
// 	block    Block
// 	expected bool
// }

// var blocksForCollisions = []BlockForCollision{

// 	{Block{3, 10, 1, 1}, true},
// 	{Block{1, 10, 1, 2}, true},
// 	{Block{0, 10, 6.5, 0.5}, true},
// 	{Block{1, 9.5, 4, 1}, true},
// 	{Block{0, 7, 2, 1}, false},
// 	{Block{2.5, 7.5, 3.5, 1}, false},
// 	{Block{0, 0, 10, 2}, false},
// }

// func TestCheckCollision(t *testing.T) {
// 	player := Player{2, 8, 0, 0, 1, 2}
// 	for _, pair := range blocksForCollisions {
// 		result := player.CheckCollision(pair.block)
// 		if result != pair.expected {
// 			t.Error("Expected", pair.expected,
// 				"for x:", pair.block.X,
// 				"y:", pair.block.Y,
// 				"w:", pair.block.w,
// 				"h:", pair.block.h,
// 				"but got:", result)
// 		}
// 	}
// }

// type GravityPlayer struct {
// 	player   Player
// 	g        float32
// 	expected float32
// }

// var GravityPlayers = []GravityPlayer{
// 	{Player{10, 10, 0, 0, 0, 0}, 9.81, 9.81},
// 	{Player{10, 10, 0, 5, 0, 0}, 9.81, 14.81},
// 	{Player{10, 10, 0, 0, 0, 0}, 15, 15},
// }

// func TestGravity(t *testing.T) {
// 	for _, test := range GravityPlayers {
// 		test.player.Gravity(test.g)
// 		if test.player.vy != test.expected {
// 			t.Error("Expected", test.expected,
// 				"but got:", test.player.vy)
// 		}
// 	}
// }

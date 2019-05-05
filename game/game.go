package game

import (
	"math/rand"
	"sync"
)

var Players []*Player

// func FieldGenerator(w int, h int, jumpLength float64) (player *Player) {
// 	r := rand.New(rand.NewSource(99))
// 	currentY := float64(h - 20)
// 	fmt.Printf("Iterations: %d\n", ((h-20)-20)/10)
// 	for i := 0; i < ((h-20)-20)/10; i++ {
// 		currentX := float64(r.Intn(w-21) + 21)
// 		if i == 0 {
// 			player = &Player{currentX + 10, currentY - 5, 0, 0, 10, 10}
// 		}
// 		blocks = append(blocks, &Block{currentX, currentY, 20, 5})
// 		currentY -= (jumpLength / 2)
// 	}
// 	return
// }

var widthField float64 = 400
var heightField float64 = 700
var g float64 = 0.0004

func GetParams() (float64, uint16) {
	return (heightField - 20) - 20, 5
}

func FieldGenerator(beginY float64, b float64, k uint16) (newBlocks []*Block) {
	// beginY was sended as the parameter
	p := b / float64(k) // Плотность
	r := rand.New(rand.NewSource(99))
	var currentX float64
	currentY := beginY
	var i uint16
	for i = 0; i < k; i++ {
		currentX = r.Float64()*(widthField-91) + 1.0
		newBlocks = append(newBlocks, &Block{currentX, currentY, 90, 15})
		currentY -= p
	}
	// blocks = append(blocks, newBlocks)
	for _, block := range newBlocks {
		blocks = append(blocks, block)
	}
	return
}

func GameLoop() {
	var wg sync.WaitGroup
	for _, player := range Players {
		wg.Add(1)
		go func(pl *Player) {
			defer wg.Done()
			player.CircleDraw()
			player.Gravity(g, 1000/16) // dt = 1000 / 16
			player.Move(Vector{pl.Vx, pl.Vy}, 1000/16)
			nearestBlock := player.SelectNearestBlock()
			player.CheckCollision(nearestBlock, 1000/16)
		}(player)
	}
	wg.Wait()

}

package game

import (
	"fmt"
	"math/rand"
	"sync"
)

var players []*Player

func FieldGenerator(w int, h int, jumpLength float32) (player *Player) {
	r := rand.New(rand.NewSource(99))
	currentY := float32(h - 20)
	fmt.Printf("Iterations: %d\n", ((h-20)-20)/10)
	for i := 0; i < ((h-20)-20)/10; i++ {
		currentX := float32(r.Intn(w-21) + 21)
		if i == 0 {
			player = &Player{currentX + 10, currentY - 5, 0, 0, 10, 10}
		}
		blocks = append(blocks, &Block{currentX, currentY, 20, 5})
		currentY -= (jumpLength / 2)
	}
	return
}
func GameLoop() {
	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(pl *Player) {
			defer wg.Done()
			pl.Gravity(9.81)
			pl.Move(Vector{pl.vx, pl.vy})
		}(player)
	}
	wg.Wait()

}

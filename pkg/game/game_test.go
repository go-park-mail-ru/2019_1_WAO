package game

import (
	"fmt"
	"sync"
	"testing"
)

func TestGameLoop(t *testing.T) {
	player1 := Player{2, 6, 0, 0, 1, 1}
	block := Block{2, 36.43, 1, 2}
	players = append(players, &player1)
	wgr := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		fmt.Print(i+1, " sec\n")
		GameLoop()
		for _, plr := range players {
			wgr.Add(1)
			go func(pl *Player) {
				defer wgr.Done()
				if pl.CheckCollision(block) {
					pl.Y = block.Y - 1
					pl.vy = -15
					fmt.Println("Collision was occured")
				}
			}(plr)
		}
		wgr.Wait()
	}
	for index, player := range players {
		fmt.Printf("player %x, x: %f, y: %f\n", index+1, player.X, player.Y)
	}
}

func TestFieldGenerator(t *testing.T) {
	var player *Player
	player = FieldGenerator(100, 100, 20)
	players = append(players, player)
	for _, value := range blocks {
		fmt.Println(*value)
	}
	fmt.Println("Players:")
	for _, plr := range players {
		fmt.Println(*plr)
	}

}

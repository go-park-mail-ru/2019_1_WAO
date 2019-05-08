package game

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	_ "net/http"

	"github.com/gorilla/websocket"
)

type Vector struct {
	x float64
	y float64
}

type Point struct {
	x float64
	y float64
}

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Player struct {
	connection *websocket.Conn `json:"-"`
	room       *Room           `json:"-"`
	queue      *Queue          `json:"-"` // Commands queue for players
	in         chan []byte     `json:"-"`
	out        chan []byte     `json:"-"`
	Id         int             `json:"id`
	X          float64         `json:"x"`
	Y          float64         `json:"y"`
	Vx         float64         `json:"vx"`
	Vy         float64         `json:"vy"`
	W          float64         `json:"w"`
	H          float64         `json:"h"`
	// conn *websocket.Conn
}

// func (player *Player) Move(vector Vector, dt float64) {
// 	player.X += vector.x * dt
// 	player.Y += vector.y * dt
// 	fmt.Printf("x: %f, y: %f\n", player.X, player.Y)
// }

func Move(command *Command) {
	var player *Player = FoundPlayer(command.IdP)
	if player == nil {
		return
	}
	player.Y += (player.Vy * command.Delay)
}

func CheckPointCollision(playerPoint, blockUpPoint, blockDownPoint Point) bool {
	if blockUpPoint.x <= playerPoint.x && playerPoint.x <= blockDownPoint.x && blockUpPoint.y <= playerPoint.y && playerPoint.y <= blockDownPoint.y {
		return true
	}
	return false
}

func (player *Player) SelectNearestBlock() (nearestBlock *Block) {
	nearestBlock = nil
	var minY float64
	for _, block := range Blocks {
		if player.X+player.W >= block.X && player.X <= block.X+block.w {
			if block.Y-player.Y < minY && player.Y <= block.Y {
				minY = block.Y - player.Y
				nearestBlock = block
			}
		}
	}
	return
}

func (player *Player) Jump() {
	if Blocks[0].Vy != 0 {
		player.Vy = -0.35 + Blocks[0].Vy
		return
	}
	player.Vy = -0.35 // Change a vertical speed (for jump)
}

func (player *Player) SetPlayerOnPlate(block *Block) {
	player.Y = block.Y - block.h
}

func (player *Player) Gravity(g float64, dt float64) {
	player.Vy += g * dt
	// player.Move(Vector{0, player.Vy})
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

func FoundPlayer(id int) *Player {
	for _, player := range Players {
		if player.Id == id {
			return player
		}
	}
	return nil
}

// Сдвиг персонажа вниз

// move(command) {
//     let player = this.foundPlayer(command.idP);
//     player.y += (player.dy * command.delay);
//   }

func NewPlayer(conn *websocket.Conn, id int) *Player {
	newPlayer := &Player{
		connection: conn,
		Id:         id,
		in:         make(chan []byte),
		out:        make(chan []byte),
	}
	Players = append(Players, newPlayer)
	return newPlayer
}

func (p *Player) Listen() {
	go func() {
		defer p.room.RemovePlayer(p)
		for {
			_, buffer, err := p.connection.ReadMessage()
			if err != nil {
				fmt.Println("Error connection", err)
				return
			}
			fmt.Println(string(buffer))
			var msg Message
			if err = json.Unmarshal(buffer, &msg); err != nil {
				fmt.Println("Error message parsing", err)
				return
			}
			if _, ok := err.(*net.OpError); ok {
				log.Println("My Life is a pain")
				// p.room.RemovePlayer(p)
				log.Printf("Player %s disconnected", p.Id)
				return
			}

			if websocket.IsUnexpectedCloseError(err) {
				// p.room.RemovePlayer(p)
				log.Printf("Player %s disconnected", p.Id)
				return
			}
			if err != nil {
				log.Printf("cannot read json: %v", err)
				continue
			}

			switch msg.Type {
			case "move":
				var command Command
				if err := json.Unmarshal([]byte(msg.Payload), &command); err != nil {
					fmt.Println("Moving error was occured", err)
					return
				}
				fmt.Println(" was received")
				fmt.Printf("Direction: %s, dt: %f\n", command.Direction, command.Delay)
				command.IdP = p.Id
				payload, err := json.Marshal(command)
				if err != nil {
					fmt.Println("Error with encoding command", err)
					return
				}
				msg.Payload = payload
				for _, player := range p.room.Players {
					if player != p {
						player.SendMessage(&msg)
					}
				}
			case "init":
				type BlocksAndPlayers struct {
					Blocks  []*Block  `json:"blocks`
					Players []*Player `json:"players`
				}
				blocks := FieldGenerator(100, 100, 10)
				for _, block := range blocks {
					p.room.Blocks = append(p.room.Blocks, block)
				}
				var players []*Player
				players = append(players, p) // The
				for _, player := range p.room.Players {
					if player == p {
						continue
					}
					players = append(players, player)
				}
				blocksAndPlayers := BlocksAndPlayers{
					Blocks:  blocks,
					Players: players,
				}
				payload, err := json.Marshal(blocksAndPlayers)
				if err != nil {
					log.Println("Error blocks and players is occured", err)
					return
				}
				msg.Payload = payload
				temp := blocksAndPlayers.Players[0]
				blocksAndPlayers.Players[0] = blocksAndPlayers.Players[1]
				blocksAndPlayers.Players[1] = temp
				payload2, err := json.Marshal(blocksAndPlayers)
				if err != nil {
					log.Println("Error blocks and players is occured", err)
					return
				}
				msg2 := Message{
					Type:    "init",
					Payload: payload2,
				}
				for _, player := range p.room.Players {
					if player != p {
						player.SendMessage(&msg2)
						continue
					}
					player.SendMessage(&msg)
				}
			case "map":
				newBlocks := FieldGenerator(100, 100, 10)
				for _, newBlock := range newBlocks {
					p.room.Blocks = append(p.room.Blocks, newBlock)
				}
				buffer, err := json.Marshal(newBlocks)
				if err != nil {
					fmt.Println("Error encoding new blocks", err)
					return
				}
				JsonNewBlocks := Message{
					Type:    "map",
					Payload: buffer,
				}

				// BlocksToSend, err := json.Marshal(JsonNewBlocks)
				// if err != nil {
				// 	fmt.Println("Error encoding new blocks", err)
				// 	return
				// }
				for _, player := range p.room.Players {
					player.SendMessage(&JsonNewBlocks)
				}
			case "lose":
				fmt.Println("!Player lose!")
				p.room.RemovePlayer(p)
				// case "blocks":
				// 	var blocks []game.Block
				// 	if err := json.Unmarshal([]byte(msg.Payload), &blocks); err != nil {
				// 		fmt.Println("Moving error was occured", err)
				// 		return
				// 	}
				// 	fmt.Println("Blocks:")
				// 	for index, block := range blocks {
				// 		fmt.Printf("*Block %d*\nx: %f, y: %f\n", index, block.X, block.Y)
				// 	}

			}
			// p.in <- message
		}
	}()

	for {
		select {
		case message := <-p.out:
			p.connection.WriteMessage(websocket.TextMessage, message)
		case message := <-p.in:
			log.Printf("Income: %#v", message)
		}
	}
}

func (p *Player) SendMessage(message *Message) {
	data, err := json.Marshal(*message)
	if err != nil {
		fmt.Println("Error with encoding struct was occured", err)
		return
	}
	p.out <- data
}

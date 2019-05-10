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
	connection         *websocket.Conn `json:"-"`
	room               *Room           `json:"-"`
	commands           chan *Command   `json:"-"`
	engineDone         chan struct{}   `json:"-"`
	mapPlayerListenEnd chan struct{}   `json:"-"`
	canvas             *Canvas         `json:"-"`
	out                chan []byte     `json:"-"`
	IdP                int             `json:"idP"`
	X                  float64         `json:"x"`
	Y                  float64         `json:"y"`
	Dx                 float64         `json:"dx"`
	Dy                 float64         `json:"dy"`
	W                  float64         `json:"-"`
	H                  float64         `json:"-"`
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
	for _, block := range player.room.Blocks {
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
	anyBlockDy := player.room.Blocks[0].Dy
	if anyBlockDy != 0 {
		player.Dy = -0.35 + anyBlockDy
		return
	}
	player.Dy = -0.35 // Change a vertical speed (for jump)
}

func (player *Player) SetPlayerOnPlate(block *Block) {
	player.Y = block.Y - block.h
	player.X = block.X + block.w/2 // Отцентровка игрока по середине
}

func (player *Player) CircleDraw() {
	if player.X > WidthField {
		player.X = 0
	} else if player.X < 0 {
		player.X = WidthField
	}
}

func NewPlayer(conn *websocket.Conn) *Player {
	newPlayer := &Player{
		connection:         conn,
		out:                make(chan []byte),
		commands:           make(chan *Command, 10),
		engineDone:         make(chan struct{}, 1),
		mapPlayerListenEnd: make(chan struct{}),
		Dx:                 0.2,
		Dy:                 0.002,
		canvas: &Canvas{
			y:  0,
			dy: 0,
		},
	}
	return newPlayer
}

func (p *Player) Listen() {
	go func() {
		defer RemovePlayer(p)
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
				log.Printf("Player %s disconnected\n", p.IdP)
				return
			}

			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Player %s disconnected\n", p.IdP)
				return
			}
			if err != nil {
				log.Printf("cannot read json: %v\n", err)
				continue
			}

			switch msg.Type {
			case "move":
				var command Command
				if err := json.Unmarshal([]byte(msg.Payload), &command); err != nil {
					fmt.Println("Moving error was occured", err)
					return
				}
				fmt.Printf("Direction: %s, dt: %f\n", command.Direction, command.Delay)
				if p.IdP == 1 {
					p.IdP = 1
				}
				command.IdP = p.IdP
				p.commands <- &command
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
			case "lose":
				fmt.Println("!Player lose!")
				return
			}
		}
	}()

	go func() {

	}()
	for {
		select {
		case message := <-p.out:
			p.connection.WriteMessage(websocket.TextMessage, message)
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

package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	_ "net/http"

	"github.com/spf13/viper"

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
	connection     *websocket.Conn `json:"-"`
	room           *Room           `json:"-"`
	commands       chan *Command   `json:"-"`
	engineDone     chan struct{}   `json:"-"`
	messagesClose  chan struct{}   `json"-"`
	canvas         *Canvas         `json:"-"`
	out            chan []byte     `json:"-"`
	IdP            int             `json:"idP"`
	X              float64         `json:"x"`
	Y              float64         `json:"y"`
	Dx             float64         `json:"dx"`
	Dy             float64         `json:"dy"`
	W              float64         `json:"-"`
	H              float64         `json:"-"`
	stateScrollMap bool            `json:"-"`
	commandCounter uint64          `json:"-"`
}

func (player *Player) SelectNearestBlock(blocks *[]*Block) (nearestBlock *Block) {
	nearestBlock = nil
	minY := math.MaxFloat64
	// canvasY := player.canvas.y
	var playerInAreaOfPlatefunc = func(plate *Block) bool {
		return (player.X+player.W >= plate.X && player.X <= plate.X+plate.w)
	}
	var playerAbouteAPlate = func(plate *Block) bool {
		return (plate.Y-player.Y < minY && player.Y <= plate.Y)
	}
	for _, block := range *blocks {
		if playerInAreaOfPlatefunc(block) {
			if playerAbouteAPlate(block) {
				minY = block.Y - player.Y
				nearestBlock = block
			}
		}
	}
	return
}

func (player *Player) Jump() {
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
		connection:    conn,
		out:           make(chan []byte),
		commands:      make(chan *Command, 10),
		engineDone:    make(chan struct{}, 1),
		messagesClose: make(chan struct{}),
		Dx:            viper.GetFloat64("player.dx"),
		Dy:            viper.GetFloat64("player.dy"),
		W:             viper.GetFloat64("player.width"),
		H:             viper.GetFloat64("player.height"),
		canvas: &Canvas{
			y:  0,
			dy: 0,
		},
		stateScrollMap: false,
		commandCounter: 0,
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
			var msg Message
			if err = json.Unmarshal(buffer, &msg); err != nil {
				fmt.Println("Error message parsing", err)
				return
			}
			if _, ok := err.(*net.OpError); ok {
				log.Println("My Life is a pain")
				log.Printf("Player %d disconnected\n", p.IdP)
				return
			}

			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Player %d disconnected\n", p.IdP)
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
				// fmt.Printf("Direction: %s, dt: %f\n", command.Direction, command.Delay)
				command.IdP = p.IdP
				p.commands <- &command
				payload, err := json.Marshal(command)
				if err != nil {
					fmt.Println("Error with encoding command", err)
					return
				}
				msg.Payload = payload
				p.room.Players.Range(func(_, player interface{}) bool {
					if player.(*Player) != p {
						player.(*Player).SendMessage(&msg)
					}
					return true
				})
			case "lose":
				fmt.Println("!Player lose!")
				return
			}
		}
	}()

	for {
		select {
		case <-p.messagesClose:
			p.engineDone <- struct{}{}
			return
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

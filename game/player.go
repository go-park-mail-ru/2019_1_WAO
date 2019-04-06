package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Player struct {
	connection *websocket.Conn
	ID         string
	in         chan *Message
	out        chan *Message
}

func NewPlayer(conn *websocket.Conn, id string) *Player {
	return &Player{
		connection: conn,
		ID:         id,
		in:         make(chan *IncomeMessage),
		out:        make(chan *Message),
	}
}

func (p *Player) Listen() {
	go func() {
		for {
			message := &IncomeMessage{}
			err := p.connection.ReadJSON(message)
			_, ok := err.(*websocket.CloseError)
			log.Println(ok)

			if websocket.IsUnexpectedCloseError(err) {
				p.room.RemovePlayer(p)
				log.Printf("Player %s disconnected", p.ID)
				return
			}
			if err != nil {
				log.Printf("cannot read json: %v", err)
				continue
			}
			p.in <- message
		}
	}()

	for {
		select {
		case message := <-p.out:
			p.connection.WriteJSON(message)
		case message := <-p.in:
			log.Printf("Income: %#v", message)
		}
	}
}

func (p Player) SendState(state *RoomState) {
	p.out <- &Message{"STATE", state}
}

func (p *Player) SendMessage(message *Message) {
	p.out <- message
}

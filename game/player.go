package main

import (
	"log"
	"net"

	"github.com/gorilla/websocket"
)

type Player struct {
	connection *websocket.Conn
	ID         string
	in         chan *IncomeMessage
	out        chan *Message
	room       *Room
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
			if _, ok := err.(*net.OpError); ok {
				log.Println("My Life is a pain")
				p.room.RemovePlayer(p)
				log.Printf("Player %s disconnected", p.ID)
				return
			}

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

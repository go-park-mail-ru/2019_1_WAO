package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	game "./game"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type connections map[*game.Player]*websocket.Conn

// var rooms map[string]connections

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Move struct {
	Direction string  `json:"direction"`
	Dt        float64 `json:"dt"`
}

// func UpdateField() {
// 	newBlocks := game.FieldGenerator(20, 100, 5)
// 	for _, player := range game.Players {
// 		if err := rooms["first"][player].WriteJSON(newBlocks); err != nil {
// 			fmt.Println("Error JSON array encoding", err)
// 			break
// 		}
//
// }

// func InitGame() {
// 	UpdateField()

// }

// func ConnectPlayerToRoom(socketPlayer *websocket.Conn, roomName string) {
// 	var room connections
// 	// var ok bool
// 	// if room, ok = rooms[roomName]; !ok {
// 	// 	fmt.Println("This room already exists!")
// 	// 	return
// 	// }
// 	newPlayer := &game.Player{
// 		X: 0,
// 		Y: 0,
// 	}
// 	room[newPlayer] = socketPlayer
// }
func SocketFunc(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error", err)
	}
	// defer ws.Close()
	// ConnectPlayerToRoom(ws, "first") // Connect this player
	go ReadSocket(ws)
}

func ReadSocket(ws *websocket.Conn) {
	fmt.Println("TEST")
	defer ws.Close()
	for {
		_, buf, err := ws.ReadMessage()
		fmt.Println(buf)
		if err != nil {
			fmt.Println("Error reading message", err)
			return
		}
		var msg Message
		if err := json.Unmarshal(buf, &msg); err != nil {
			fmt.Println("Error message parsing", err)
			return
		}
		// for the love of Gopher DO NOT DO THIS
		switch msg.Type {
		case "move":
			var moving Move
			if err := json.Unmarshal([]byte(msg.Payload), &moving); err != nil {
				fmt.Println("Moving error was occured", err)
				return
			}
			fmt.Println("Move was received")
			fmt.Printf("Direction: %s, dt: %f\n", moving.Direction, moving.Dt)
		case "player":
			var plr game.Player
			if err := json.Unmarshal([]byte(msg.Payload), &plr); err != nil {
				fmt.Println("Moving error was occured", err)
				return
			}
			fmt.Printf("***Player***\nx: %f, y: %f, vx: %f, vy: %f\n", plr.X, plr.Y, plr.Vx, plr.Vy)
		case "blocks":
			var blocks []game.Block
			if err := json.Unmarshal([]byte(msg.Payload), &blocks); err != nil {
				fmt.Println("Moving error was occured", err)
				return
			}
			fmt.Println("Blocks:")
			for index, block := range blocks {
				fmt.Printf("*Block %d*\nx: %f, y: %f\n", index, block.X, block.Y)
			}

		}
	}
}

func WriteSocket(ws *websocket.Conn) {
	// ticker := time.NewTicker(3 * time.Second)
	// for {
	// 	w, err := ws.NextWriter(websocket.TextMessage)
	// 	if err != nil {
	// 		ticker.Stop()
	// 		break
	// 	}
	// 	w.Write([]byte("Hello!"))
	// 	w.Close()
	// 	<-ticker.C
	// }
	player := game.FieldGenerator(100, 100, 10.0)

	err := ws.WriteJSON(player)
	if err != nil {
		fmt.Println("Error was occured", err)
	}
}
func MainFunc(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id := vars["id"]
	// tmpl, _ := template.ParseFiles("./templates/index.html")
	// tmpl.Execute(w, "")

	// fmt.Fprintf(w, "ok")
	http.ServeFile(w, r, "index.html")

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", MainFunc)
	router.HandleFunc("/websocket", SocketFunc)
	http.Handle("/", router)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	fmt.Println("Server is listening")

	http.ListenAndServe(":8080", nil)
}
